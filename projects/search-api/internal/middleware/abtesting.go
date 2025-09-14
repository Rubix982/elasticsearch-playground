package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/saif-islam/es-playground/projects/search-api/internal/abtesting"
	"github.com/saif-islam/es-playground/projects/search-api/internal/models"
)

// ABTestingMiddleware adds A/B testing capabilities to search requests
func ABTestingMiddleware(framework *abtesting.ABTestFramework, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract request information for A/B testing
		requestID := c.GetString("request_id")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		// Get user/session information
		userID := c.GetHeader("X-User-ID")
		sessionID := c.GetHeader("X-Session-ID")
		if sessionID == "" {
			sessionID = c.GetHeader("X-Request-ID")
		}
		
		// Extract query parameters
		query := c.Query("q")
		if query == "" {
			query = c.Query("query")
		}
		index := c.Query("index")
		
		// Create A/B test request
		abTestRequest := abtesting.ABTestRequest{
			RequestID: requestID,
			UserID:    userID,
			SessionID: sessionID,
			Query:     query,
			Index:     index,
			Context:   make(map[string]interface{}),
		}
		
		// Add additional context
		abTestRequest.Context["path"] = c.Request.URL.Path
		abTestRequest.Context["method"] = c.Request.Method
		abTestRequest.Context["user_agent"] = c.GetHeader("User-Agent")
		abTestRequest.Context["timestamp"] = time.Now()
		
		// Get experiment assignment
		assignment, err := framework.GetVariantForRequest(abTestRequest)
		if err != nil {
			logger.Error("Failed to get A/B test assignment", 
				zap.Error(err),
				zap.String("request_id", requestID))
			// Continue without A/B testing
			c.Next()
			return
		}
		
		// Store assignment in context
		c.Set("ab_test_assignment", assignment)
		c.Set("ab_test_request", abTestRequest)
		
		// Add tracing attributes if span is available
		if span := trace.SpanFromContext(c.Request.Context()); span.IsRecording() {
			span.SetAttributes(
				attribute.String("ab_test.experiment_id", assignment.ExperimentID),
				attribute.String("ab_test.variant_id", assignment.VariantID),
				attribute.String("ab_test.variant_name", assignment.VariantName),
			)
		}
		
		// Log assignment
		logger.Debug("A/B test assignment",
			zap.String("request_id", requestID),
			zap.String("experiment_id", assignment.ExperimentID),
			zap.String("variant_id", assignment.VariantID),
			zap.String("variant_name", assignment.VariantName))
		
		// Add response headers for debugging
		c.Header("X-AB-Test-Experiment", assignment.ExperimentID)
		c.Header("X-AB-Test-Variant", assignment.VariantID)
		
		// Process the request
		startTime := time.Now()
		c.Next()
		responseTime := time.Since(startTime)
		
		// Record experiment result
		go func() {
			result := abtesting.ExperimentResult{
				Success:      c.Writer.Status() < 400,
				ResponseTime: responseTime,
				ResultCount:  0, // Will be updated by search handler
			}
			
			framework.RecordExperimentResult(assignment, result)
		}()
	}
}

// ApplyVariantModifications applies A/B test variant modifications to a search request
func ApplyVariantModifications(searchReq *models.SearchRequest, assignment *abtesting.ExperimentAssignment) {
	if assignment == nil || assignment.Variant == nil {
		return
	}
	
	modifications := assignment.Variant.QueryModifications
	
	// Apply query type modification
	if modifications.QueryType != "" {
		searchReq.QueryType = modifications.QueryType
	}
	
	// Apply fuzziness modification
	if modifications.Fuzziness != "" {
		searchReq.Fuzziness = modifications.Fuzziness
	}
	
	// Apply minimum should match
	if modifications.MinShouldMatch != "" {
		// Store in metadata for query builder to use
		if searchReq.Metadata == nil {
			searchReq.Metadata = make(map[string]interface{})
		}
		searchReq.Metadata["min_should_match"] = modifications.MinShouldMatch
	}
	
	// Apply size modification
	if modifications.Size > 0 {
		searchReq.Size = modifications.Size
	}
	
	// Apply timeout modification
	if modifications.Timeout != "" {
		searchReq.Timeout = modifications.Timeout
	}
	
	// Apply boost factors
	if len(modifications.BoostFactors) > 0 {
		if searchReq.Metadata == nil {
			searchReq.Metadata = make(map[string]interface{})
		}
		searchReq.Metadata["boost_factors"] = modifications.BoostFactors
	}
	
	// Apply rescore modifications
	if len(modifications.Rescore) > 0 {
		searchReq.Rescore = modifications.Rescore
	}
	
	// Apply highlighting modifications
	if modifications.Highlighting != nil {
		searchReq.Highlight = *modifications.Highlighting
	}
	
	// Apply custom query
	if modifications.CustomQuery != "" {
		if searchReq.Metadata == nil {
			searchReq.Metadata = make(map[string]interface{})
		}
		searchReq.Metadata["custom_query"] = modifications.CustomQuery
	}
	
	// Apply feature flags
	if searchReq.Metadata == nil {
		searchReq.Metadata = make(map[string]interface{})
	}
	searchReq.Metadata["enable_caching"] = modifications.EnableCaching
	searchReq.Metadata["enable_prefetch"] = modifications.EnablePrefetch
	searchReq.Metadata["enable_personalization"] = modifications.EnablePersonalization
}

// GetABTestAssignment retrieves A/B test assignment from Gin context
func GetABTestAssignment(c *gin.Context) (*abtesting.ExperimentAssignment, bool) {
	assignment, exists := c.Get("ab_test_assignment")
	if !exists {
		return nil, false
	}
	
	abAssignment, ok := assignment.(*abtesting.ExperimentAssignment)
	return abAssignment, ok
}

// GetABTestRequest retrieves A/B test request from Gin context
func GetABTestRequest(c *gin.Context) (*abtesting.ABTestRequest, bool) {
	request, exists := c.Get("ab_test_request")
	if !exists {
		return nil, false
	}
	
	abRequest, ok := request.(*abtesting.ABTestRequest)
	return abRequest, ok
}