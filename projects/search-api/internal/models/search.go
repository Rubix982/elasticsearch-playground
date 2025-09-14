package models

import "time"

// SearchRequest represents a comprehensive search query request
type SearchRequest struct {
	// Basic search parameters
	Query       string            `json:"query" form:"q"`
	Index       string            `json:"index" form:"index"`
	Size        int               `json:"size" form:"size"`
	From        int               `json:"from" form:"from"`
	
	// Advanced query options
	QueryType   string            `json:"query_type,omitempty" form:"query_type"` // match, multi_match, bool, etc.
	Fields      []string          `json:"fields,omitempty" form:"fields"`         // fields to search in
	Operator    string            `json:"operator,omitempty" form:"operator"`     // AND, OR
	Fuzziness   string            `json:"fuzziness,omitempty" form:"fuzziness"`   // AUTO, 0, 1, 2
	MinScore    float64           `json:"min_score,omitempty" form:"min_score"`
	
	// Filtering and sorting
	Sort        []SortField       `json:"sort,omitempty" form:"sort"`
	Filters     []Filter          `json:"filters,omitempty"`
	PostFilter  []Filter          `json:"post_filter,omitempty"` // Applied after aggregations
	
	// Aggregations
	Aggregations map[string]AggregationConfig `json:"aggregations,omitempty"`
	
	// Result customization
	Highlight   HighlightConfig   `json:"highlight,omitempty"`
	Source      []string          `json:"_source,omitempty"`        // Fields to include/exclude
	ExcludeSource []string        `json:"_source_excludes,omitempty"`
	
	// Performance options
	Preference  string            `json:"preference,omitempty"`     // _local, _primary, custom
	Timeout     string            `json:"timeout,omitempty"`        // 1s, 100ms, etc.
	
	// Analytics
	TrackScores bool              `json:"track_scores,omitempty"`
	TrackTotalHits bool           `json:"track_total_hits,omitempty"`
	
	// Advanced features
	Suggest     map[string]SuggesterConfig `json:"suggest,omitempty"`
	Rescore     []RescoreConfig   `json:"rescore,omitempty"`
	
	RequestID   string            `json:"request_id,omitempty"`
	
	// A/B testing and experimentation
	ABTestVariant string                 `json:"ab_test_variant,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// Filter represents a search filter
type Filter struct {
	Field    string      `json:"field"`
	Type     string      `json:"type"`     // term, terms, range, exists, wildcard, etc.
	Value    interface{} `json:"value"`
	Operator string      `json:"operator,omitempty"` // gte, lte, gt, lt for range
}

// HighlightConfig represents highlighting configuration
type HighlightConfig struct {
	Enabled        bool                       `json:"enabled"`
	Fields         []string                   `json:"fields,omitempty"`
	PreTags        []string                   `json:"pre_tags,omitempty"`
	PostTags       []string                   `json:"post_tags,omitempty"`
	FragmentSize   int                        `json:"fragment_size,omitempty"`
	NumFragments   int                        `json:"number_of_fragments,omitempty"`
	HighlightType  string                     `json:"type,omitempty"` // unified, plain, fvh
	Settings       map[string]interface{}     `json:"settings,omitempty"`
}

// AggregationConfig represents aggregation configuration
type AggregationConfig struct {
	Type     string                 `json:"type"`     // terms, date_histogram, stats, etc.
	Field    string                 `json:"field"`
	Size     int                    `json:"size,omitempty"`
	Settings map[string]interface{} `json:"settings,omitempty"`
	SubAggs  map[string]AggregationConfig `json:"aggs,omitempty"`
}

// SuggesterConfig represents suggester configuration
type SuggesterConfig struct {
	Text       string `json:"text"`
	Field      string `json:"field"`
	Size       int    `json:"size,omitempty"`
	Type       string `json:"type"`       // term, phrase, completion
	Fuzziness  string `json:"fuzziness,omitempty"`
}

// RescoreConfig represents rescoring configuration
type RescoreConfig struct {
	WindowSize int     `json:"window_size"`
	Query      string  `json:"query"`
	Weight     float64 `json:"query_weight,omitempty"`
}

// SortField represents a sort configuration
type SortField struct {
	Field string `json:"field"`
	Order string `json:"order"` // "asc" or "desc"
}

// SearchResponse represents comprehensive search results
type SearchResponse struct {
	// Basic response info
	Query        string                 `json:"query"`
	Total        HitsTotal              `json:"total"`
	MaxScore     *float64               `json:"max_score"`
	Took         int                    `json:"took"`
	TimedOut     bool                   `json:"timed_out"`
	
	// Results
	Hits         []SearchHit            `json:"hits"`
	Aggregations map[string]interface{} `json:"aggregations,omitempty"`
	Suggest      map[string][]SuggestOption `json:"suggest,omitempty"`
	
	// Performance and debug info
	Shards       ShardInfo              `json:"_shards"`
	Profile      ProfileInfo            `json:"profile,omitempty"`
	
	// Analytics
	SearchType   string                 `json:"search_type,omitempty"`
	Warnings     []string               `json:"warnings,omitempty"`
	
	// Caching
	CacheHit     bool                   `json:"cache_hit,omitempty"`
	
	// Request tracking
	RequestID    string                 `json:"request_id"`
	Timestamp    time.Time              `json:"timestamp"`
	ResponseTime time.Duration          `json:"response_time"`
}

// HitsTotal represents the total hits information
type HitsTotal struct {
	Value    int64  `json:"value"`
	Relation string `json:"relation"` // eq, gte
}

// ShardInfo represents shard execution information
type ShardInfo struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Skipped    int `json:"skipped"`
	Failed     int `json:"failed"`
}

// ProfileInfo represents query profiling information
type ProfileInfo struct {
	Enabled     bool                   `json:"enabled"`
	QueryTime   int64                  `json:"query_time_in_nanos,omitempty"`
	FetchTime   int64                  `json:"fetch_time_in_nanos,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// SuggestOption represents a single suggestion option
type SuggestOption struct {
	Text    string             `json:"text"`
	Score   float64            `json:"score"`
	Freq    int                `json:"freq,omitempty"`
	Options []SuggestionOption `json:"options,omitempty"`
}

// SuggestionOption represents individual suggestion
type SuggestionOption struct {
	Text  string  `json:"text"`
	Score float64 `json:"_score"`
}

// SearchHit represents a single search result
type SearchHit struct {
	Index     string          `json:"_index"`
	ID        string          `json:"_id"`
	Score     *float64        `json:"_score"`
	Source    interface{}     `json:"_source"`
	Highlight map[string][]string `json:"highlight,omitempty"`
}

// SuggestRequest represents an autocomplete/suggestion request
type SuggestRequest struct {
	Text  string `json:"text" form:"text"`
	Index string `json:"index" form:"index"`
	Field string `json:"field" form:"field"`
	Size  int    `json:"size" form:"size"`
}

// SuggestResponse represents suggestion results
type SuggestResponse struct {
	Suggestions []Suggestion `json:"suggestions"`
	RequestID   string       `json:"request_id"`
	Timestamp   time.Time    `json:"timestamp"`
}

// Suggestion represents a single suggestion
type Suggestion struct {
	Text  string  `json:"text"`
	Score float64 `json:"score"`
}

// IndexRequest represents a document indexing request
type IndexRequest struct {
	Index    string      `json:"index"`
	ID       string      `json:"id,omitempty"`
	Document interface{} `json:"document"`
}

// IndexResponse represents the indexing result
type IndexResponse struct {
	Index     string    `json:"_index"`
	ID        string    `json:"_id"`
	Version   int       `json:"_version"`
	Result    string    `json:"result"`
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error     string    `json:"error"`
	Message   string    `json:"message"`
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status        string                 `json:"status"`
	Elasticsearch ElasticsearchHealth    `json:"elasticsearch"`
	Redis         RedisHealth            `json:"redis,omitempty"`
	RequestID     string                 `json:"request_id"`
	Timestamp     time.Time              `json:"timestamp"`
}

// ElasticsearchHealth represents Elasticsearch health status
type ElasticsearchHealth struct {
	Status    string `json:"status"`
	Nodes     int    `json:"nodes"`
	DataNodes int    `json:"data_nodes"`
}

// RedisHealth represents Redis health status
type RedisHealth struct {
	Status      string `json:"status"`
	Connections int    `json:"connections"`
}

// SearchAnalytics represents search analytics and optimization data
type SearchAnalytics struct {
	QueryID          string                 `json:"query_id"`
	Query            string                 `json:"query"`
	Index            string                 `json:"index"`
	ExecutionTime    time.Duration          `json:"execution_time"`
	ResultCount      int64                  `json:"result_count"`
	ClickThroughRate float64                `json:"ctr,omitempty"`
	ConversionRate   float64                `json:"conversion_rate,omitempty"`
	UserID           string                 `json:"user_id,omitempty"`
	SessionID        string                 `json:"session_id,omitempty"`
	Timestamp        time.Time              `json:"timestamp"`
	Performance      SearchPerformanceMetrics `json:"performance"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// SearchPerformanceMetrics represents detailed performance metrics
type SearchPerformanceMetrics struct {
	QueryTime        int64   `json:"query_time_ms"`
	FetchTime        int64   `json:"fetch_time_ms"`
	TotalShards      int     `json:"total_shards"`
	SuccessfulShards int     `json:"successful_shards"`
	FailedShards     int     `json:"failed_shards"`
	CacheHits        int     `json:"cache_hits,omitempty"`
	CacheMisses      int     `json:"cache_misses,omitempty"`
	IndexSize        int64   `json:"index_size_bytes,omitempty"`
	MemoryUsage      int64   `json:"memory_usage_bytes,omitempty"`
	CPUTime          int64   `json:"cpu_time_ms,omitempty"`
	OptimizationTips []string `json:"optimization_tips,omitempty"`
}

// QueryBuilder represents a visual query builder request
type QueryBuilder struct {
	Conditions []QueryCondition           `json:"conditions"`
	Logic      string                     `json:"logic"` // AND, OR
	Boost      float64                    `json:"boost,omitempty"`
	Settings   map[string]interface{}     `json:"settings,omitempty"`
}

// QueryCondition represents a single query condition
type QueryCondition struct {
	Field     string      `json:"field"`
	Operator  string      `json:"operator"` // equals, contains, starts_with, range, etc.
	Value     interface{} `json:"value"`
	Boost     float64     `json:"boost,omitempty"`
	Fuzziness string      `json:"fuzziness,omitempty"`
}

// QueryOptimizationSuggestion represents optimization suggestions
type QueryOptimizationSuggestion struct {
	Type         string  `json:"type"`         // performance, accuracy, relevance
	Priority     string  `json:"priority"`     // high, medium, low
	Description  string  `json:"description"`
	Impact       string  `json:"impact"`       // performance_gain, accuracy_improvement
	Suggestion   string  `json:"suggestion"`
	BeforeQuery  string  `json:"before_query,omitempty"`
	AfterQuery   string  `json:"after_query,omitempty"`
	EstimatedGain float64 `json:"estimated_gain,omitempty"`
}

// SearchTemplate represents a saved search template
type SearchTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Template    string                 `json:"template"`    // Mustache template
	Parameters  map[string]interface{} `json:"parameters"`  // Default parameters
	Tags        []string               `json:"tags"`
	CreatedBy   string                 `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	UsageCount  int64                  `json:"usage_count"`
	IsPublic    bool                   `json:"is_public"`
}

// SearchExplain represents query explanation
type SearchExplain struct {
	QueryID       string                 `json:"query_id"`
	Query         string                 `json:"query"`
	Explanation   QueryExplanation       `json:"explanation"`
	Suggestions   []QueryOptimizationSuggestion `json:"suggestions"`
	Performance   SearchPerformanceMetrics `json:"performance"`
	AlternativeQueries []string          `json:"alternative_queries,omitempty"`
}

// QueryExplanation represents detailed query explanation
type QueryExplanation struct {
	QueryType    string                 `json:"query_type"`
	ParsedQuery  map[string]interface{} `json:"parsed_query"`
	IndexesUsed  []string               `json:"indexes_used"`
	ShardsQueried []string              `json:"shards_queried"`
	FieldsSearched []string             `json:"fields_searched"`
	Complexity   string                 `json:"complexity"` // simple, moderate, complex
	EstimatedCost float64               `json:"estimated_cost"`
}