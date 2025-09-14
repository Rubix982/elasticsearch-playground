package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/saif-islam/es-playground/projects/search-api/internal/middleware"
	"github.com/saif-islam/es-playground/projects/search-api/internal/models"
	"github.com/saif-islam/es-playground/projects/search-api/internal/services"
)

// SearchHandler handles all search-related HTTP requests
type SearchHandler struct {
	searchService *services.SearchService
	logger        *zap.Logger
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(searchService *services.SearchService, logger *zap.Logger) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
		logger:        logger,
	}
}

// RegisterRoutes registers all search-related routes
func (h *SearchHandler) RegisterRoutes(router *gin.RouterGroup) {
	v1 := router.Group("/v1")
	{
		// Basic searches
		v1.GET("/search", h.Search)
		v1.POST("/search", h.AdvancedSearch)
		v1.POST("/multi-search", h.MultiSearch)
		
		// Suggestions and autocomplete
		v1.GET("/suggest", h.Suggest)
		v1.POST("/autocomplete", h.Autocomplete)
		
		// Query building and optimization
		v1.POST("/query/build", h.BuildQuery)
		v1.POST("/query/optimize", h.OptimizeQuery)
		v1.POST("/query/explain", h.ExplainQuery)
		v1.POST("/query/validate", h.ValidateQuery)
		
		// Templates and analytics
		v1.GET("/templates", h.ListTemplates)
		v1.POST("/templates", h.CreateTemplate)
		v1.GET("/templates/:id", h.GetTemplate)
		v1.POST("/templates/:id/search", h.SearchWithTemplate)
		
		// Analytics
		v1.GET("/analytics/search-stats", h.GetSearchStats)
		v1.GET("/analytics/performance", h.GetPerformanceMetrics)
	}
}

// Search handles basic search requests (GET /search)
func (h *SearchHandler) Search(c *gin.Context) {
	req := &models.SearchRequest{
		RequestID: uuid.New().String(),
		Size:      10, // default
		From:      0,  // default
	}

	// Bind query parameters
	if err := c.ShouldBindQuery(req); err != nil {
		h.logger.Error("Failed to bind query parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "invalid_parameters",
			Message:   err.Error(),
			RequestID: req.RequestID,
			Timestamp: time.Now(),
		})
		return
	}

	// Validate required fields
	if req.Index == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "missing_index",
			Message:   "Index parameter is required",
			RequestID: req.RequestID,
			Timestamp: time.Now(),
		})
		return
	}

	// Set defaults
	if req.Size == 0 {
		req.Size = 10
	}
	if req.Size > 100 {
		req.Size = 100 // limit for performance
	}

	// Apply A/B test modifications if available
	if assignment, exists := middleware.GetABTestAssignment(c); exists {
		middleware.ApplyVariantModifications(req, assignment)
		req.ABTestVariant = assignment.VariantID
	}

	// Perform search
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	response, err := h.searchService.Search(ctx, req)
	if err != nil {
		h.logger.Error("Search failed", zap.Error(err), zap.String("request_id", req.RequestID))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:     "search_failed",
			Message:   err.Error(),
			RequestID: req.RequestID,
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// AdvancedSearch handles complex search requests (POST /search)
func (h *SearchHandler) AdvancedSearch(c *gin.Context) {
	req := &models.SearchRequest{
		RequestID: uuid.New().String(),
	}

	// Bind JSON body
	if err := c.ShouldBindJSON(req); err != nil {
		h.logger.Error("Failed to bind JSON request", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "invalid_json",
			Message:   err.Error(),
			RequestID: req.RequestID,
			Timestamp: time.Now(),
		})
		return
	}

	// Validate required fields
	if req.Index == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "missing_index",
			Message:   "Index field is required",
			RequestID: req.RequestID,
			Timestamp: time.Now(),
		})
		return
	}

	// Set defaults
	if req.Size == 0 {
		req.Size = 10
	}
	if req.Size > 1000 {
		req.Size = 1000 // higher limit for advanced search
	}

	// Set timeout from request or default
	timeout := 30 * time.Second
	if req.Timeout != "" {
		if parsedTimeout, err := time.ParseDuration(req.Timeout); err == nil {
			timeout = parsedTimeout
		}
	}

	// Apply A/B test modifications if available
	if assignment, exists := middleware.GetABTestAssignment(c); exists {
		middleware.ApplyVariantModifications(req, assignment)
		req.ABTestVariant = assignment.VariantID
	}

	// Perform search
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()

	response, err := h.searchService.Search(ctx, req)
	if err != nil {
		h.logger.Error("Advanced search failed", zap.Error(err), zap.String("request_id", req.RequestID))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:     "search_failed",
			Message:   err.Error(),
			RequestID: req.RequestID,
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// MultiSearch handles multiple search requests in a single call
func (h *SearchHandler) MultiSearch(c *gin.Context) {
	var requests []models.SearchRequest

	if err := c.ShouldBindJSON(&requests); err != nil {
		h.logger.Error("Failed to bind multi-search JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "invalid_json",
			Message:   err.Error(),
			RequestID: uuid.New().String(),
			Timestamp: time.Now(),
		})
		return
	}

	if len(requests) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "no_requests",
			Message:   "At least one search request is required",
			RequestID: uuid.New().String(),
			Timestamp: time.Now(),
		})
		return
	}

	if len(requests) > 10 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "too_many_requests",
			Message:   "Maximum 10 search requests allowed",
			RequestID: uuid.New().String(),
			Timestamp: time.Now(),
		})
		return
	}

	// Process searches concurrently
	responses := make([]*models.SearchResponse, len(requests))
	errors := make([]error, len(requests))

	// Create channels for concurrent processing
	type result struct {
		index    int
		response *models.SearchResponse
		err      error
	}

	resultCh := make(chan result, len(requests))

	// Start searches concurrently
	for i, req := range requests {
		go func(idx int, searchReq models.SearchRequest) {
			if searchReq.RequestID == "" {
				searchReq.RequestID = uuid.New().String()
			}

			ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
			defer cancel()

			resp, err := h.searchService.Search(ctx, &searchReq)
			resultCh <- result{index: idx, response: resp, err: err}
		}(i, req)
	}

	// Collect results
	for i := 0; i < len(requests); i++ {
		res := <-resultCh
		responses[res.index] = res.response
		errors[res.index] = res.err
	}

	// Check for errors
	haseErrors := false
	for _, err := range errors {
		if err != nil {
			haseErrors = true
			break
		}
	}

	if haseErrors {
		h.logger.Error("Multi-search had errors", zap.Errors("errors", errors))
		c.JSON(http.StatusMultiStatus, gin.H{
			"responses": responses,
			"errors":    errors,
			"timestamp": time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"responses": responses,
		"timestamp": time.Now(),
	})
}

// Suggest handles search suggestions (GET /suggest)
func (h *SearchHandler) Suggest(c *gin.Context) {
	req := &models.SuggestRequest{}

	if err := c.ShouldBindQuery(req); err != nil {
		h.logger.Error("Failed to bind suggest parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "invalid_parameters",
			Message:   err.Error(),
			RequestID: uuid.New().String(),
			Timestamp: time.Now(),
		})
		return
	}

	if req.Text == "" || req.Index == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "missing_parameters",
			Message:   "text and index parameters are required",
			RequestID: uuid.New().String(),
			Timestamp: time.Now(),
		})
		return
	}

	if req.Size == 0 {
		req.Size = 5 // default suggestion count
	}
	if req.Size > 20 {
		req.Size = 20 // limit suggestions
	}

	// Build search request for suggestions
	searchReq := &models.SearchRequest{
		RequestID: uuid.New().String(),
		Index:     req.Index,
		Size:      0, // we only want suggestions
		Suggest: map[string]models.SuggesterConfig{
			"text_suggest": {
				Text:  req.Text,
				Field: req.Field,
				Size:  req.Size,
				Type:  "term",
			},
			"completion_suggest": {
				Text:      req.Text,
				Field:     req.Field + ".suggest", // assuming completion field
				Size:      req.Size,
				Type:      "completion",
				Fuzziness: "AUTO",
			},
		},
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	response, err := h.searchService.Search(ctx, searchReq)
	if err != nil {
		h.logger.Error("Suggest failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:     "suggest_failed",
			Message:   err.Error(),
			RequestID: searchReq.RequestID,
			Timestamp: time.Now(),
		})
		return
	}

	// Transform suggestions to response format
	suggestions := make([]models.Suggestion, 0)
	for _, options := range response.Suggest {
		for _, option := range options {
			suggestions = append(suggestions, models.Suggestion{
				Text:  option.Text,
				Score: option.Score,
			})
		}
	}

	c.JSON(http.StatusOK, models.SuggestResponse{
		Suggestions: suggestions,
		RequestID:   searchReq.RequestID,
		Timestamp:   time.Now(),
	})
}

// Autocomplete handles advanced autocomplete requests
func (h *SearchHandler) Autocomplete(c *gin.Context) {
	req := &models.SuggestRequest{}

	if err := c.ShouldBindJSON(req); err != nil {
		h.logger.Error("Failed to bind autocomplete JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "invalid_json",
			Message:   err.Error(),
			RequestID: uuid.New().String(),
			Timestamp: time.Now(),
		})
		return
	}

	// For now, delegate to suggest - can be enhanced later
	h.Suggest(c)
}

// BuildQuery helps build queries from visual components
func (h *SearchHandler) BuildQuery(c *gin.Context) {
	req := &models.QueryBuilder{}

	if err := c.ShouldBindJSON(req); err != nil {
		h.logger.Error("Failed to bind query builder JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "invalid_json",
			Message:   err.Error(),
			RequestID: uuid.New().String(),
			Timestamp: time.Now(),
		})
		return
	}

	// This would integrate with a visual query builder
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"message":   "Query builder functionality coming soon",
		"request":   req,
		"timestamp": time.Now(),
	})
}

// OptimizeQuery provides query optimization suggestions
func (h *SearchHandler) OptimizeQuery(c *gin.Context) {
	var queryData map[string]interface{}

	if err := c.ShouldBindJSON(&queryData); err != nil {
		h.logger.Error("Failed to bind optimization JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "invalid_json",
			Message:   err.Error(),
			RequestID: uuid.New().String(),
			Timestamp: time.Now(),
		})
		return
	}

	// Placeholder for query optimization logic
	suggestions := []models.QueryOptimizationSuggestion{
		{
			Type:         "performance",
			Priority:     "high",
			Description:  "Consider using filter context instead of query context for exact matches",
			Impact:       "performance_gain",
			Suggestion:   "Move term queries to filter context",
			EstimatedGain: 25.0,
		},
		{
			Type:         "accuracy",
			Priority:     "medium",
			Description:  "Add minimum_should_match to improve precision",
			Impact:       "accuracy_improvement",
			Suggestion:   "Set minimum_should_match to 75%",
			EstimatedGain: 15.0,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"suggestions": suggestions,
		"timestamp":   time.Now(),
	})
}

// ExplainQuery provides detailed query explanation
func (h *SearchHandler) ExplainQuery(c *gin.Context) {
	var queryData map[string]interface{}

	if err := c.ShouldBindJSON(&queryData); err != nil {
		h.logger.Error("Failed to bind explain JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "invalid_json",
			Message:   err.Error(),
			RequestID: uuid.New().String(),
			Timestamp: time.Now(),
		})
		return
	}

	// Placeholder for query explanation logic
	explanation := models.SearchExplain{
		QueryID: uuid.New().String(),
		Query:   "sample query",
		Explanation: models.QueryExplanation{
			QueryType:     "bool",
			ParsedQuery:   queryData,
			IndexesUsed:   []string{"sample_index"},
			ShardsQueried: []string{"shard_0", "shard_1"},
			FieldsSearched: []string{"title", "content"},
			Complexity:    "moderate",
			EstimatedCost: 1.5,
		},
	}

	c.JSON(http.StatusOK, explanation)
}

// ValidateQuery validates query syntax and structure
func (h *SearchHandler) ValidateQuery(c *gin.Context) {
	var queryData map[string]interface{}

	if err := c.ShouldBindJSON(&queryData); err != nil {
		h.logger.Error("Failed to bind validation JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "invalid_json",
			Message:   err.Error(),
			RequestID: uuid.New().String(),
			Timestamp: time.Now(),
		})
		return
	}

	// Basic validation - can be enhanced
	valid := true
	warnings := []string{}

	if _, hasQuery := queryData["query"]; !hasQuery {
		warnings = append(warnings, "No query specified - will match all documents")
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":     valid,
		"warnings":  warnings,
		"query":     queryData,
		"timestamp": time.Now(),
	})
}

// Template management handlers (placeholders)
func (h *SearchHandler) ListTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"templates": []models.SearchTemplate{},
		"timestamp": time.Now(),
	})
}

func (h *SearchHandler) CreateTemplate(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"message":   "Template creation coming soon",
		"timestamp": time.Now(),
	})
}

func (h *SearchHandler) GetTemplate(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"template_id": id,
		"message":     "Template retrieval coming soon",
		"timestamp":   time.Now(),
	})
}

func (h *SearchHandler) SearchWithTemplate(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"template_id": id,
		"message":     "Template search coming soon",
		"timestamp":   time.Now(),
	})
}

// Analytics handlers (placeholders)
func (h *SearchHandler) GetSearchStats(c *gin.Context) {
	// Parse optional query parameters for filtering
	from := c.Query("from")
	to := c.Query("to")
	index := c.Query("index")

	// Placeholder stats
	stats := gin.H{
		"total_searches":    1000,
		"avg_response_time": "45ms",
		"most_searched":     []string{"elasticsearch", "search", "api"},
		"filters": gin.H{
			"from":  from,
			"to":    to,
			"index": index,
		},
		"timestamp": time.Now(),
	}

	c.JSON(http.StatusOK, stats)
}

func (h *SearchHandler) GetPerformanceMetrics(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "24"))
	if limit > 100 {
		limit = 100
	}

	// Placeholder metrics
	metrics := gin.H{
		"response_times": []float64{45.2, 38.1, 52.3, 41.7, 39.8},
		"throughput":     []int{150, 180, 165, 172, 190},
		"error_rate":     []float64{0.1, 0.2, 0.0, 0.1, 0.0},
		"cache_hit_rate": []float64{85.2, 87.1, 83.5, 86.9, 88.2},
		"limit":          limit,
		"timestamp":      time.Now(),
	}

	c.JSON(http.StatusOK, metrics)
}
