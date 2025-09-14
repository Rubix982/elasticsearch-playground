package tracing

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TracingMiddleware creates a Gin middleware for distributed tracing
func TracingMiddleware(provider *TracingProvider, logger *zap.Logger) gin.HandlerFunc {
	if !provider.config.Enabled {
		// Return a no-op middleware if tracing is disabled
		return gin.HandlerFunc(func(c *gin.Context) {
			c.Next()
		})
	}

	// Use the otelgin middleware as base
	baseMiddleware := otelgin.Middleware(provider.config.ServiceName)

	return gin.HandlerFunc(func(c *gin.Context) {
		// Execute the base OpenTelemetry middleware
		baseMiddleware(c)

		// Add custom attributes and logic
		span := trace.SpanFromContext(c.Request.Context())
		if span.IsRecording() {
			// Add custom attributes
			span.SetAttributes(
				attribute.String("http.user_agent", c.Request.UserAgent()),
				attribute.String("http.remote_addr", c.ClientIP()),
				attribute.String("http.request_id", getRequestID(c)),
				attribute.String("gin.version", gin.Version),
			)

			// Add query parameters (be careful with sensitive data)
			if len(c.Request.URL.RawQuery) > 0 {
				span.SetAttributes(attribute.String("http.query", c.Request.URL.RawQuery))
			}

			// Record the start time
			startTime := time.Now()
			c.Set("trace.start_time", startTime)

			// Process request
			c.Next()

			// Add response attributes
			duration := time.Since(startTime)
			span.SetAttributes(
				attribute.Int("http.response.status_code", c.Writer.Status()),
				attribute.Int64("http.response.duration_ms", duration.Milliseconds()),
				attribute.Int("http.response.size", c.Writer.Size()),
			)

			// Set span status based on HTTP status code
			if c.Writer.Status() >= 400 {
				span.SetStatus(codes.Error, "HTTP "+strconv.Itoa(c.Writer.Status()))
			} else {
				span.SetStatus(codes.Ok, "")
			}

			// Log trace information
			logger.Debug("Request traced",
				zap.String("trace_id", span.SpanContext().TraceID().String()),
				zap.String("span_id", span.SpanContext().SpanID().String()),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", c.Writer.Status()),
				zap.Duration("duration", duration),
			)
		}
	})
}

// SearchTracingMiddleware adds search-specific tracing context
func SearchTracingMiddleware(provider *TracingProvider) gin.HandlerFunc {
	if !provider.config.Enabled {
		return gin.HandlerFunc(func(c *gin.Context) { c.Next() })
	}

	return gin.HandlerFunc(func(c *gin.Context) {
		span := trace.SpanFromContext(c.Request.Context())
		if span.IsRecording() {
			// Add search-specific attributes
			if query := c.Query("q"); query != "" {
				span.SetAttributes(attribute.String("search.query", query))
			}
			if index := c.Query("index"); index != "" {
				span.SetAttributes(attribute.String("search.index", index))
			}
			if size := c.Query("size"); size != "" {
				span.SetAttributes(attribute.String("search.size", size))
			}
			if from := c.Query("from"); from != "" {
				span.SetAttributes(attribute.String("search.from", from))
			}

			// Add A/B test information if available
			if variant := c.GetString("ab_test_variant"); variant != "" {
				span.SetAttributes(attribute.String("experiment.variant", variant))
			}
			if experimentID := c.GetString("experiment_id"); experimentID != "" {
				span.SetAttributes(attribute.String("experiment.id", experimentID))
			}
		}
		c.Next()
	})
}

// ExperimentTracingMiddleware adds A/B testing tracing context
func ExperimentTracingMiddleware(provider *TracingProvider) gin.HandlerFunc {
	if !provider.config.Enabled {
		return gin.HandlerFunc(func(c *gin.Context) { c.Next() })
	}

	return gin.HandlerFunc(func(c *gin.Context) {
		span := trace.SpanFromContext(c.Request.Context())
		if span.IsRecording() {
			// Add experiment management attributes
			if experimentID := c.Param("id"); experimentID != "" {
				span.SetAttributes(attribute.String("experiment.id", experimentID))
			}
			if variantID := c.Param("variant_id"); variantID != "" {
				span.SetAttributes(attribute.String("experiment.variant_id", variantID))
			}
			if template := c.Param("template"); template != "" {
				span.SetAttributes(attribute.String("experiment.template", template))
			}
		}
		c.Next()
	})
}

// getRequestID extracts request ID from various sources
func getRequestID(c *gin.Context) string {
	// Try to get from context first
	if reqID, exists := c.Get("request_id"); exists {
		if id, ok := reqID.(string); ok {
			return id
		}
	}

	// Try various headers
	headers := []string{"X-Request-ID", "X-Request-Id", "X-Correlation-ID", "X-Trace-ID"}
	for _, header := range headers {
		if id := c.GetHeader(header); id != "" {
			return id
		}
	}

	return ""
}