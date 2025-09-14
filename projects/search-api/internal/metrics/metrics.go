package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP request metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Elasticsearch search metrics
	ElasticsearchSearchTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "elasticsearch_search_requests_total",
			Help: "Total number of Elasticsearch search requests",
		},
		[]string{"index", "query_type"},
	)

	ElasticsearchSearchDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "elasticsearch_search_duration_seconds",
			Help:    "Elasticsearch search request duration in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"index", "query_type"},
	)

	ElasticsearchSearchResults = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "elasticsearch_search_results_count",
			Help:    "Number of results returned by Elasticsearch searches",
			Buckets: []float64{0, 1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
		},
		[]string{"index", "query_type"},
	)

	// Elasticsearch bulk operations metrics
	ElasticsearchBulkTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "elasticsearch_bulk_requests_total",
			Help: "Total number of Elasticsearch bulk requests",
		},
		[]string{"index", "operation"},
	)

	ElasticsearchBulkDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "elasticsearch_bulk_duration_seconds",
			Help:    "Elasticsearch bulk operation duration in seconds",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60},
		},
		[]string{"index", "operation"},
	)

	ElasticsearchBulkDocuments = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "elasticsearch_bulk_documents_count",
			Help:    "Number of documents in bulk operations",
			Buckets: []float64{1, 10, 50, 100, 250, 500, 1000, 2500, 5000, 10000},
		},
		[]string{"index", "operation"},
	)

	// Elasticsearch errors
	ElasticsearchErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "elasticsearch_errors_total",
			Help: "Total number of Elasticsearch errors",
		},
		[]string{"index", "operation", "error_type"},
	)

	// Cache metrics
	CacheOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_operations_total",
			Help: "Total number of cache operations",
		},
		[]string{"operation", "result"},
	)

	CacheDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cache_operation_duration_seconds",
			Help:    "Cache operation duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"operation"},
	)

	// Connection pool metrics
	ElasticsearchConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "elasticsearch_client_connections_active",
			Help: "Number of active Elasticsearch connections",
		},
	)

	ElasticsearchConnectionsMax = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "elasticsearch_client_connections_max",
			Help: "Maximum number of Elasticsearch connections",
		},
	)

	// Query performance insights
	SlowQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "slow_queries_total",
			Help: "Total number of slow queries (>1s)",
		},
		[]string{"index", "query_type"},
	)

	QueryOptimizationSuggestions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "query_optimization_suggestions_total",
			Help: "Total number of query optimization suggestions generated",
		},
		[]string{"suggestion_type"},
	)

	// Application health metrics
	ApplicationInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "application_info",
			Help: "Application information",
		},
		[]string{"version", "service", "environment"},
	)

	ApplicationUptime = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "application_uptime_seconds_total",
			Help: "Total application uptime in seconds",
		},
	)
)

// Timer is a helper for measuring durations
type Timer struct {
	histogram prometheus.Observer
	start     time.Time
}

// NewTimer creates a new timer for measuring durations
func NewTimer(histogram prometheus.Observer) *Timer {
	return &Timer{
		histogram: histogram,
		start:     time.Now(),
	}
}

// ObserveDuration records the duration since the timer was created
func (t *Timer) ObserveDuration() {
	if t.histogram != nil {
		t.histogram.Observe(time.Since(t.start).Seconds())
	}
}

// RecordHTTPRequest records HTTP request metrics
func RecordHTTPRequest(method, endpoint, status string, duration time.Duration) {
	HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordElasticsearchSearch records Elasticsearch search metrics
func RecordElasticsearchSearch(index, queryType string, duration time.Duration, resultCount int64) {
	ElasticsearchSearchTotal.WithLabelValues(index, queryType).Inc()
	ElasticsearchSearchDuration.WithLabelValues(index, queryType).Observe(duration.Seconds())
	ElasticsearchSearchResults.WithLabelValues(index, queryType).Observe(float64(resultCount))
	
	// Track slow queries
	if duration.Seconds() > 1.0 {
		SlowQueriesTotal.WithLabelValues(index, queryType).Inc()
	}
}

// RecordElasticsearchBulk records Elasticsearch bulk operation metrics
func RecordElasticsearchBulk(index, operation string, duration time.Duration, docCount int) {
	ElasticsearchBulkTotal.WithLabelValues(index, operation).Inc()
	ElasticsearchBulkDuration.WithLabelValues(index, operation).Observe(duration.Seconds())
	ElasticsearchBulkDocuments.WithLabelValues(index, operation).Observe(float64(docCount))
}

// RecordElasticsearchError records Elasticsearch error metrics
func RecordElasticsearchError(index, operation, errorType string) {
	ElasticsearchErrors.WithLabelValues(index, operation, errorType).Inc()
}

// RecordCacheOperation records cache operation metrics
func RecordCacheOperation(operation, result string, duration time.Duration) {
	CacheOperations.WithLabelValues(operation, result).Inc()
	CacheDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// UpdateConnectionMetrics updates connection pool metrics
func UpdateConnectionMetrics(active, max int) {
	ElasticsearchConnectionsActive.Set(float64(active))
	ElasticsearchConnectionsMax.Set(float64(max))
}

// RecordOptimizationSuggestion records query optimization suggestion metrics
func RecordOptimizationSuggestion(suggestionType string) {
	QueryOptimizationSuggestions.WithLabelValues(suggestionType).Inc()
}

// SetApplicationInfo sets application information metrics
func SetApplicationInfo(version, service, environment string) {
	ApplicationInfo.WithLabelValues(version, service, environment).Set(1)
}

// IncrementUptime increments the application uptime counter
func IncrementUptime(seconds float64) {
	ApplicationUptime.Add(seconds)
}