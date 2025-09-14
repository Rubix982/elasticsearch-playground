package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"go.uber.org/zap"

	"github.com/saif-islam/es-playground/projects/search-api/internal/cache"
	"github.com/saif-islam/es-playground/projects/search-api/internal/metrics"
	"github.com/saif-islam/es-playground/projects/search-api/internal/models"
	"github.com/saif-islam/es-playground/projects/search-api/internal/realtime"
	"github.com/saif-islam/es-playground/projects/search-api/internal/tracing"
	"github.com/saif-islam/es-playground/shared"
)

// SearchService handles advanced search operations with optimization focus
type SearchService struct {
	esClient      shared.ESClientInterface
	logger        *zap.Logger
	analyticsHub  *realtime.AnalyticsHub
	tracer        *tracing.SearchOperationTracer
	cacheManager  *cache.CacheManager
}

// NewSearchService creates a new search service
func NewSearchService(esClient shared.ESClientInterface, logger *zap.Logger, analyticsHub *realtime.AnalyticsHub, tracer *tracing.SearchOperationTracer, cacheManager *cache.CacheManager) *SearchService {
	return &SearchService{
		esClient:     esClient,
		logger:       logger,
		analyticsHub: analyticsHub,
		tracer:       tracer,
		cacheManager: cacheManager,
	}
}

// Search performs advanced search with comprehensive features
func (s *SearchService) Search(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error) {
	// Start search operation span
	ctx, span := s.tracer.TraceSearchOperation(ctx, "search", req)
	defer span.End()
	
	startTime := time.Now()
	
	// Try cache first
	if cachedResponse, found := s.cacheManager.GetCache().GetSearchResult(ctx, req); found {
		s.tracer.RecordCacheOperation(ctx, "get", true, "search_result")
		s.tracer.RecordSearchResult(ctx, cachedResponse.Total.Value, time.Since(startTime), true)
		
		// Update analytics for cache hit
		if s.analyticsHub != nil {
			analyticsEvent := realtime.SearchEvent{
				Timestamp:    startTime,
				QueryID:      req.RequestID,
				Index:        req.Index,
				Query:        req.Query,
				QueryType:    req.QueryType,
				ResponseTime: time.Since(startTime),
				ResultCount:  cachedResponse.Total.Value,
				Success:      true,
				CacheHit:     true,
				TraceID:      span.SpanContext().TraceID().String(),
			}
			s.analyticsHub.RecordSearchEvent(analyticsEvent)
		}
		
		return cachedResponse, nil
	}
	
	// Cache miss - record it
	s.tracer.RecordCacheOperation(ctx, "get", false, "search_result")
	
	// Build Elasticsearch query
	query, err := s.buildElasticsearchQuery(req)
	if err != nil {
		s.logger.Error("Failed to build query", zap.Error(err))
		s.tracer.RecordError(ctx, err, map[string]interface{}{
			"operation": "build_query",
			"index": req.Index,
		})
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	// Execute search with tracing
	ctx, esSpan := s.tracer.TraceElasticsearchOperation(ctx, "POST", fmt.Sprintf("/%s/_search", req.Index), query)
	defer esSpan.End()
	
	searchReq := elasticsearch.Search{
		Index: []string{req.Index},
		Body:  strings.NewReader(query),
	}
	
	if req.Timeout != "" {
		searchReq.Timeout = req.Timeout
	}
	
	res := searchReq.Do(ctx, s.esClient.(*elasticsearch.Client))
	if res.IsError() {
		err := fmt.Errorf("search failed: %s", res.String())
		s.tracer.RecordElasticsearchResult(ctx, res.StatusCode, 0, time.Since(startTime))
		s.tracer.RecordError(ctx, err, map[string]interface{}{
			"elasticsearch.status_code": res.StatusCode,
			"elasticsearch.error": res.String(),
		})
		return nil, err
	}
	defer res.Body.Close()

	// Parse response
	var esResponse map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&esResponse); err != nil {
		s.tracer.RecordError(ctx, err, map[string]interface{}{
			"operation": "parse_response",
		})
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Transform to our response format
	response := s.transformSearchResponse(esResponse, req)
	response.ResponseTime = time.Since(startTime)
	response.RequestID = req.RequestID
	response.Timestamp = time.Now()
	
	// Record tracing results
	s.tracer.RecordElasticsearchResult(ctx, res.StatusCode, len(res.String()), time.Since(startTime))
	s.tracer.RecordSearchResult(ctx, response.Total.Value, time.Duration(response.Took)*time.Millisecond, true)

	// Record metrics
	queryType := req.QueryType
	if queryType == "" {
		queryType = "simple_query_string"
	}
	metrics.RecordElasticsearchSearch(req.Index, queryType, response.ResponseTime, response.Total.Value)

	// Cache the successful result
	if err := s.cacheManager.GetCache().SetSearchResult(ctx, req, response); err != nil {
		s.logger.Warn("Failed to cache search result", zap.Error(err))
	} else {
		s.tracer.RecordCacheOperation(ctx, "set", true, "search_result")
	}
	
	// Record real-time analytics event
	if s.analyticsHub != nil {
		analyticsEvent := realtime.SearchEvent{
			Timestamp:    startTime,
			QueryID:      req.RequestID,
			Index:        req.Index,
			Query:        req.Query,
			QueryType:    queryType,
			ResponseTime: response.ResponseTime,
			ResultCount:  response.Total.Value,
			Success:      true,
			CacheHit:     false,
			TraceID:      span.SpanContext().TraceID().String(),
		}
		s.analyticsHub.RecordSearchEvent(analyticsEvent)
	}

	// Log search analytics
	s.logSearchAnalytics(req, response, startTime)

	return response, nil
}

// buildElasticsearchQuery builds comprehensive Elasticsearch query JSON
func (s *SearchService) buildElasticsearchQuery(req *models.SearchRequest) (string, error) {
	query := map[string]interface{}{
		"size": req.Size,
		"from": req.From,
	}

	// Build main query
	mainQuery := s.buildMainQuery(req)
	if mainQuery != nil {
		query["query"] = mainQuery
	}

	// Add sorting
	if len(req.Sort) > 0 {
		sorts := make([]map[string]interface{}, len(req.Sort))
		for i, sort := range req.Sort {
			sorts[i] = map[string]interface{}{
				sort.Field: map[string]interface{}{
					"order": sort.Order,
				},
			}
		}
		query["sort"] = sorts
	}

	// Add highlighting
	if req.Highlight.Enabled {
		highlight := s.buildHighlightConfig(req.Highlight)
		query["highlight"] = highlight
	}

	// Add aggregations
	if len(req.Aggregations) > 0 {
		aggs := make(map[string]interface{})
		for name, aggConfig := range req.Aggregations {
			aggs[name] = s.buildAggregation(aggConfig)
		}
		query["aggs"] = aggs
	}

	// Add post filters
	if len(req.PostFilter) > 0 {
		postFilter := s.buildFilters(req.PostFilter)
		query["post_filter"] = postFilter
	}

	// Add suggestions
	if len(req.Suggest) > 0 {
		suggest := make(map[string]interface{})
		for name, suggestConfig := range req.Suggest {
			suggest[name] = s.buildSuggester(suggestConfig)
		}
		query["suggest"] = suggest
	}

	// Add rescoring
	if len(req.Rescore) > 0 {
		rescores := make([]map[string]interface{}, len(req.Rescore))
		for i, rescore := range req.Rescore {
			rescores[i] = s.buildRescore(rescore)
		}
		query["rescore"] = rescores
	}

	// Add source filtering
	if len(req.Source) > 0 || len(req.ExcludeSource) > 0 {
		source := make(map[string]interface{})
		if len(req.Source) > 0 {
			source["includes"] = req.Source
		}
		if len(req.ExcludeSource) > 0 {
			source["excludes"] = req.ExcludeSource
		}
		query["_source"] = source
	}

	// Add performance options
	if req.MinScore > 0 {
		query["min_score"] = req.MinScore
	}

	if req.TrackTotalHits {
		query["track_total_hits"] = true
	}

	if req.TrackScores {
		query["track_scores"] = true
	}

	// Convert to JSON
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return "", fmt.Errorf("failed to marshal query: %w", err)
	}

	return string(queryJSON), nil
}

// buildMainQuery builds the main query part based on request
func (s *SearchService) buildMainQuery(req *models.SearchRequest) map[string]interface{} {
	if req.Query == "" && len(req.Filters) == 0 {
		return map[string]interface{}{
			"match_all": map[string]interface{}{},
		}
	}

	boolQuery := map[string]interface{}{
		"must": []interface{}{},
		"filter": []interface{}{},
	}

	// Add main query
	if req.Query != "" {
		var mainQuery map[string]interface{}
		
		switch req.QueryType {
		case "match":
			mainQuery = map[string]interface{}{
				"match": map[string]interface{}{
					"_all": map[string]interface{}{
						"query": req.Query,
						"operator": req.Operator,
						"fuzziness": req.Fuzziness,
					},
				},
			}
		case "multi_match":
			queryConfig := map[string]interface{}{
				"query": req.Query,
			}
			if len(req.Fields) > 0 {
				queryConfig["fields"] = req.Fields
			}
			if req.Operator != "" {
				queryConfig["operator"] = req.Operator
			}
			if req.Fuzziness != "" {
				queryConfig["fuzziness"] = req.Fuzziness
			}
			mainQuery = map[string]interface{}{
				"multi_match": queryConfig,
			}
		case "query_string":
			mainQuery = map[string]interface{}{
				"query_string": map[string]interface{}{
					"query": req.Query,
					"default_operator": req.Operator,
				},
			}
		default: // Simple query string
			mainQuery = map[string]interface{}{
				"simple_query_string": map[string]interface{}{
					"query": req.Query,
					"default_operator": req.Operator,
				},
			}
		}
		
		boolQuery["must"] = []interface{}{mainQuery}
	}

	// Add filters
	if len(req.Filters) > 0 {
		filters := s.buildFilters(req.Filters)
		boolQuery["filter"] = []interface{}{filters}
	}

	return map[string]interface{}{
		"bool": boolQuery,
	}
}

// buildFilters builds filter queries from filter array
func (s *SearchService) buildFilters(filters []models.Filter) map[string]interface{} {
	if len(filters) == 1 {
		return s.buildSingleFilter(filters[0])
	}

	boolFilter := map[string]interface{}{
		"must": []interface{}{},
	}

	for _, filter := range filters {
		boolFilter["must"] = append(boolFilter["must"].([]interface{}), s.buildSingleFilter(filter))
	}

	return map[string]interface{}{
		"bool": boolFilter,
	}
}

// buildSingleFilter builds a single filter based on type
func (s *SearchService) buildSingleFilter(filter models.Filter) map[string]interface{} {
	switch filter.Type {
	case "term":
		return map[string]interface{}{
			"term": map[string]interface{}{
				filter.Field: filter.Value,
			},
		}
	case "terms":
		return map[string]interface{}{
			"terms": map[string]interface{}{
				filter.Field: filter.Value,
			},
		}
	case "range":
		rangeFilter := make(map[string]interface{})
		if filter.Operator != "" {
			rangeFilter[filter.Operator] = filter.Value
		} else {
			rangeFilter["gte"] = filter.Value
		}
		return map[string]interface{}{
			"range": map[string]interface{}{
				filter.Field: rangeFilter,
			},
		}
	case "exists":
		return map[string]interface{}{
			"exists": map[string]interface{}{
				"field": filter.Field,
			},
		}
	case "wildcard":
		return map[string]interface{}{
			"wildcard": map[string]interface{}{
				filter.Field: filter.Value,
			},
		}
	case "prefix":
		return map[string]interface{}{
			"prefix": map[string]interface{}{
				filter.Field: filter.Value,
			},
		}
	case "match":
		return map[string]interface{}{
			"match": map[string]interface{}{
				filter.Field: filter.Value,
			},
		}
	default:
		return map[string]interface{}{
			"term": map[string]interface{}{
				filter.Field: filter.Value,
			},
		}
	}
}

// buildHighlightConfig builds highlighting configuration
func (s *SearchService) buildHighlightConfig(config models.HighlightConfig) map[string]interface{} {
	highlight := make(map[string]interface{})

	if config.HighlightType != "" {
		highlight["type"] = config.HighlightType
	}

	if len(config.PreTags) > 0 {
		highlight["pre_tags"] = config.PreTags
	}

	if len(config.PostTags) > 0 {
		highlight["post_tags"] = config.PostTags
	}

	if config.FragmentSize > 0 {
		highlight["fragment_size"] = config.FragmentSize
	}

	if config.NumFragments > 0 {
		highlight["number_of_fragments"] = config.NumFragments
	}

	// Fields to highlight
	fields := make(map[string]interface{})
	if len(config.Fields) > 0 {
		for _, field := range config.Fields {
			fields[field] = map[string]interface{}{}
		}
	} else {
		// Default to highlighting all text fields
		fields["*"] = map[string]interface{}{}
	}
	highlight["fields"] = fields

	return highlight
}

// buildAggregation builds aggregation configuration
func (s *SearchService) buildAggregation(config models.AggregationConfig) map[string]interface{} {
	agg := make(map[string]interface{})

	switch config.Type {
	case "terms":
		termsAgg := map[string]interface{}{
			"field": config.Field,
		}
		if config.Size > 0 {
			termsAgg["size"] = config.Size
		}
		agg["terms"] = termsAgg

	case "date_histogram":
		dateHistAgg := map[string]interface{}{
			"field": config.Field,
		}
		// Add settings from config
		for key, value := range config.Settings {
			dateHistAgg[key] = value
		}
		agg["date_histogram"] = dateHistAgg

	case "stats":
		agg["stats"] = map[string]interface{}{
			"field": config.Field,
		}

	case "histogram":
		histAgg := map[string]interface{}{
			"field": config.Field,
		}
		for key, value := range config.Settings {
			histAgg[key] = value
		}
		agg["histogram"] = histAgg

	default:
		// Generic aggregation
		agg[config.Type] = map[string]interface{}{
			"field": config.Field,
		}
	}

	// Add sub-aggregations
	if len(config.SubAggs) > 0 {
		subAggs := make(map[string]interface{})
		for name, subAgg := range config.SubAggs {
			subAggs[name] = s.buildAggregation(subAgg)
		}
		agg["aggs"] = subAggs
	}

	return agg
}

// buildSuggester builds suggester configuration
func (s *SearchService) buildSuggester(config models.SuggesterConfig) map[string]interface{} {
	suggest := map[string]interface{}{
		"text": config.Text,
	}

	switch config.Type {
	case "term":
		termSuggest := map[string]interface{}{
			"field": config.Field,
		}
		if config.Size > 0 {
			termSuggest["size"] = config.Size
		}
		suggest["term"] = termSuggest

	case "phrase":
		phraseSuggest := map[string]interface{}{
			"field": config.Field,
		}
		if config.Size > 0 {
			phraseSuggest["size"] = config.Size
		}
		suggest["phrase"] = phraseSuggest

	case "completion":
		completionSuggest := map[string]interface{}{
			"field": config.Field,
		}
		if config.Size > 0 {
			completionSuggest["size"] = config.Size
		}
		if config.Fuzziness != "" {
			completionSuggest["fuzzy"] = map[string]interface{}{
				"fuzziness": config.Fuzziness,
			}
		}
		suggest["completion"] = completionSuggest
	}

	return suggest
}

// buildRescore builds rescoring configuration
func (s *SearchService) buildRescore(config models.RescoreConfig) map[string]interface{} {
	rescore := map[string]interface{}{
		"window_size": config.WindowSize,
		"query": map[string]interface{}{
			"rescore_query": map[string]interface{}{
				"simple_query_string": map[string]interface{}{
					"query": config.Query,
				},
			},
		},
	}

	if config.Weight > 0 {
		rescore["query"].(map[string]interface{})["query_weight"] = config.Weight
	}

	return rescore
}

// transformSearchResponse transforms Elasticsearch response to our format
func (s *SearchService) transformSearchResponse(esResponse map[string]interface{}, req *models.SearchRequest) *models.SearchResponse {
	response := &models.SearchResponse{
		Query: req.Query,
	}

	// Parse hits
	if hits, ok := esResponse["hits"].(map[string]interface{}); ok {
		// Total hits
		if total, ok := hits["total"].(map[string]interface{}); ok {
			if value, ok := total["value"].(float64); ok {
				response.Total.Value = int64(value)
			}
			if relation, ok := total["relation"].(string); ok {
				response.Total.Relation = relation
			}
		} else if total, ok := hits["total"].(float64); ok {
			// Older ES versions return just a number
			response.Total.Value = int64(total)
			response.Total.Relation = "eq"
		}

		// Max score
		if maxScore, ok := hits["max_score"].(float64); ok {
			response.MaxScore = &maxScore
		}

		// Individual hits
		if hitsList, ok := hits["hits"].([]interface{}); ok {
			response.Hits = make([]models.SearchHit, len(hitsList))
			for i, hit := range hitsList {
				if hitMap, ok := hit.(map[string]interface{}); ok {
					searchHit := models.SearchHit{
						Index:  getString(hitMap, "_index"),
						ID:     getString(hitMap, "_id"),
						Source: hitMap["_source"],
					}
					
					if score, ok := hitMap["_score"].(float64); ok {
						searchHit.Score = &score
					}
					
					if highlight, ok := hitMap["highlight"].(map[string]interface{}); ok {
						searchHit.Highlight = make(map[string][]string)
						for field, fragments := range highlight {
							if fragList, ok := fragments.([]interface{}); ok {
								searchHit.Highlight[field] = make([]string, len(fragList))
								for j, frag := range fragList {
									if fragStr, ok := frag.(string); ok {
										searchHit.Highlight[field][j] = fragStr
									}
								}
							}
						}
					}
					
					response.Hits[i] = searchHit
				}
			}
		}
	}

	// Parse aggregations
	if aggs, ok := esResponse["aggregations"].(map[string]interface{}); ok {
		response.Aggregations = aggs
	}

	// Parse suggestions
	if suggest, ok := esResponse["suggest"].(map[string]interface{}); ok {
		response.Suggest = make(map[string][]models.SuggestOption)
		for name, suggestions := range suggest {
			if suggList, ok := suggestions.([]interface{}); ok {
				response.Suggest[name] = make([]models.SuggestOption, len(suggList))
				for i, sugg := range suggList {
					if suggMap, ok := sugg.(map[string]interface{}); ok {
						option := models.SuggestOption{
							Text: getString(suggMap, "text"),
						}
						if score, ok := suggMap["score"].(float64); ok {
							option.Score = score
						}
						response.Suggest[name][i] = option
					}
				}
			}
		}
	}

	// Parse timing
	if took, ok := esResponse["took"].(float64); ok {
		response.Took = int(took)
	}

	if timedOut, ok := esResponse["timed_out"].(bool); ok {
		response.TimedOut = timedOut
	}

	// Parse shard info
	if shards, ok := esResponse["_shards"].(map[string]interface{}); ok {
		response.Shards = models.ShardInfo{
			Total:      getInt(shards, "total"),
			Successful: getInt(shards, "successful"),
			Skipped:    getInt(shards, "skipped"),
			Failed:     getInt(shards, "failed"),
		}
	}

	return response
}

// logSearchAnalytics logs search analytics for performance monitoring
func (s *SearchService) logSearchAnalytics(req *models.SearchRequest, resp *models.SearchResponse, startTime time.Time) {
	analytics := models.SearchAnalytics{
		QueryID:       req.RequestID,
		Query:         req.Query,
		Index:         req.Index,
		ExecutionTime: resp.ResponseTime,
		ResultCount:   resp.Total.Value,
		Timestamp:     time.Now(),
		Performance: models.SearchPerformanceMetrics{
			QueryTime:        int64(resp.Took),
			TotalShards:      resp.Shards.Total,
			SuccessfulShards: resp.Shards.Successful,
			FailedShards:     resp.Shards.Failed,
		},
	}

	s.logger.Info("Search analytics",
		zap.String("query_id", analytics.QueryID),
		zap.String("query", analytics.Query),
		zap.Duration("execution_time", analytics.ExecutionTime),
		zap.Int64("result_count", analytics.ResultCount),
		zap.Int64("query_time_ms", analytics.Performance.QueryTime))
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if val, ok := m[key].(float64); ok {
		return int(val)
	}
	return 0
}