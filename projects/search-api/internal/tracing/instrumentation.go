package tracing

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// SearchOperationTracer provides tracing for search operations
type SearchOperationTracer struct {
	provider *TracingProvider
}

// NewSearchOperationTracer creates a new search operation tracer
func NewSearchOperationTracer(provider *TracingProvider) *SearchOperationTracer {
	return &SearchOperationTracer{
		provider: provider,
	}
}

// TraceSearchOperation creates a span for search operations
func (s *SearchOperationTracer) TraceSearchOperation(ctx context.Context, operationName string, searchRequest interface{}) (context.Context, trace.Span) {
	ctx, span := s.provider.StartSpan(ctx, fmt.Sprintf("search.%s", operationName),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("operation.type", "search"),
			attribute.String("operation.name", operationName),
		),
	)

	// Add search request details if available
	if searchRequest != nil {
		s.addSearchRequestAttributes(span, searchRequest)
	}

	return ctx, span
}

// TraceElasticsearchOperation creates a span for Elasticsearch operations
func (s *SearchOperationTracer) TraceElasticsearchOperation(ctx context.Context, method, endpoint string, requestBody interface{}) (context.Context, trace.Span) {
	ctx, span := s.provider.StartSpan(ctx, fmt.Sprintf("elasticsearch.%s", method),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "elasticsearch"),
			attribute.String("db.operation", method),
			attribute.String("http.method", method),
			attribute.String("http.url", endpoint),
			attribute.String("component", "elasticsearch-client"),
		),
	)

	// Add request body size and type
	if requestBody != nil {
		if bodyBytes, err := json.Marshal(requestBody); err == nil {
			span.SetAttributes(
				attribute.Int("http.request.body.size", len(bodyBytes)),
				attribute.String("db.statement.type", "json"),
			)
		}
	}

	return ctx, span
}

// TraceABTestOperation creates a span for A/B testing operations
func (s *SearchOperationTracer) TraceABTestOperation(ctx context.Context, operationType, experimentID string) (context.Context, trace.Span) {
	ctx, span := s.provider.StartSpan(ctx, fmt.Sprintf("experiment.%s", operationType),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("operation.type", "ab_test"),
			attribute.String("operation.name", operationType),
			attribute.String("experiment.id", experimentID),
		),
	)

	return ctx, span
}

// TraceAnalyticsOperation creates a span for analytics operations
func (s *SearchOperationTracer) TraceAnalyticsOperation(ctx context.Context, operationType string) (context.Context, trace.Span) {
	ctx, span := s.provider.StartSpan(ctx, fmt.Sprintf("analytics.%s", operationType),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("operation.type", "analytics"),
			attribute.String("operation.name", operationType),
		),
	)

	return ctx, span
}

// RecordSearchResult records search result metrics in the span
func (s *SearchOperationTracer) RecordSearchResult(ctx context.Context, resultCount int64, took time.Duration, successful bool) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(
			attribute.Int64("search.result_count", resultCount),
			attribute.Int64("search.took_ms", took.Milliseconds()),
			attribute.Bool("search.successful", successful),
		)

		if successful {
			span.SetStatus(codes.Ok, "Search completed successfully")
		} else {
			span.SetStatus(codes.Error, "Search failed")
		}
	}
}

// RecordElasticsearchResult records Elasticsearch response metrics
func (s *SearchOperationTracer) RecordElasticsearchResult(ctx context.Context, statusCode int, responseSize int, took time.Duration) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(
			attribute.Int("http.response.status_code", statusCode),
			attribute.Int("http.response.size", responseSize),
			attribute.Int64("elasticsearch.took_ms", took.Milliseconds()),
		)

		if statusCode >= 200 && statusCode < 300 {
			span.SetStatus(codes.Ok, "Elasticsearch request successful")
		} else {
			span.SetStatus(codes.Error, fmt.Sprintf("Elasticsearch request failed with status %d", statusCode))
		}
	}
}

// RecordABTestResult records A/B test result metrics
func (s *SearchOperationTracer) RecordABTestResult(ctx context.Context, variantID string, assigned bool, sampleSize int64) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(
			attribute.String("experiment.variant_assigned", variantID),
			attribute.Bool("experiment.assigned", assigned),
			attribute.Int64("experiment.sample_size", sampleSize),
		)
	}
}

// RecordCacheOperation records cache operation metrics
func (s *SearchOperationTracer) RecordCacheOperation(ctx context.Context, operation string, hit bool, key string) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(
			attribute.String("cache.operation", operation),
			attribute.Bool("cache.hit", hit),
			attribute.String("cache.key", s.truncateString(key, 100)),
		)
	}
}

// AddCustomAttribute adds a custom attribute to the current span
func (s *SearchOperationTracer) AddCustomAttribute(ctx context.Context, key string, value interface{}) {
	s.provider.SetSpanAttributes(ctx, map[string]interface{}{key: value})
}

// AddEvent adds an event to the current span
func (s *SearchOperationTracer) AddEvent(ctx context.Context, name string, attributes map[string]interface{}) {
	s.provider.AddSpanEvent(ctx, name, attributes)
}

// RecordError records an error in the current span
func (s *SearchOperationTracer) RecordError(ctx context.Context, err error, attributes map[string]interface{}) {
	s.provider.RecordError(ctx, err, attributes)
	
	// Also set span status to error
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetStatus(codes.Error, err.Error())
	}
}

// addSearchRequestAttributes adds search request attributes to span
func (s *SearchOperationTracer) addSearchRequestAttributes(span trace.Span, searchRequest interface{}) {
	// Use reflection or type assertion to extract attributes
	// This is a simplified version - in production, you'd want more sophisticated handling
	if reqBytes, err := json.Marshal(searchRequest); err == nil {
		var reqMap map[string]interface{}
		if json.Unmarshal(reqBytes, &reqMap) == nil {
			if query, ok := reqMap["query"].(string); ok && query != "" {
				span.SetAttributes(attribute.String("search.query", s.truncateString(query, s.provider.config.MaxTagLength)))
			}
			if index, ok := reqMap["index"].(string); ok && index != "" {
				span.SetAttributes(attribute.String("search.index", index))
			}
			if size, ok := reqMap["size"].(float64); ok {
				span.SetAttributes(attribute.Int("search.size", int(size)))
			}
			if from, ok := reqMap["from"].(float64); ok {
				span.SetAttributes(attribute.Int("search.from", int(from)))
			}
			if queryType, ok := reqMap["query_type"].(string); ok && queryType != "" {
				span.SetAttributes(attribute.String("search.query_type", queryType))
			}
			if variant, ok := reqMap["ab_test_variant"].(string); ok && variant != "" {
				span.SetAttributes(attribute.String("experiment.variant", variant))
			}
		}
	}
}

// truncateString truncates a string to the specified length
func (s *SearchOperationTracer) truncateString(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen-3] + "..."
}

// ElasticsearchTransport wraps HTTP transport with tracing
type ElasticsearchTransport struct {
	base   interface{}
	tracer *SearchOperationTracer
}

// NewElasticsearchTransport creates a new traced Elasticsearch transport
func NewElasticsearchTransport(baseTransport interface{}, tracer *SearchOperationTracer) *ElasticsearchTransport {
	return &ElasticsearchTransport{
		base:   baseTransport,
		tracer: tracer,
	}
}

// TracedWebSocketUpgrade traces WebSocket upgrade operations
func (s *SearchOperationTracer) TracedWebSocketUpgrade(ctx context.Context, endpoint string) (context.Context, trace.Span) {
	ctx, span := s.provider.StartSpan(ctx, "websocket.upgrade",
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("websocket.endpoint", endpoint),
			attribute.String("component", "realtime-analytics"),
		),
	)

	return ctx, span
}

// RecordWebSocketMetrics records WebSocket connection metrics
func (s *SearchOperationTracer) RecordWebSocketMetrics(ctx context.Context, connectedClients int, messagesCount int64) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(
			attribute.Int("websocket.connected_clients", connectedClients),
			attribute.Int64("websocket.messages_sent", messagesCount),
		)
	}
}