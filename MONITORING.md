# 📊 Monitoring & Observability Guide

Complete monitoring setup for the Elasticsearch Playground using **Prometheus** and **Grafana** with comprehensive alerting and observability.

## 🎯 Overview

This monitoring stack provides:
- **Real-time metrics** for all applications and infrastructure
- **Custom dashboards** for Elasticsearch, applications, and system resources
- **Intelligent alerting** with severity-based routing
- **Performance insights** for query optimization
- **Distributed tracing** capabilities (planned)

## 🚀 Quick Start

### Start the Full Monitoring Stack

```bash
# Start all services including monitoring
docker-compose --profile monitoring up -d

# Or start specific monitoring services
docker-compose up -d prometheus grafana alertmanager
```

### Access Monitoring Services

| Service | URL | Credentials |
|---------|-----|-------------|
| **Grafana** | http://localhost:3000 | admin / playground123 |
| **Prometheus** | http://localhost:9090 | No auth required |
| **AlertManager** | http://localhost:9093 | No auth required |

## 📈 Dashboards

### Pre-built Dashboards

1. **Elasticsearch Overview** (`/d/elasticsearch-overview`)
   - Cluster health and status
   - Search and indexing performance
   - Resource utilization
   - Response times and throughput

2. **Applications Overview** (`/d/applications-overview`)
   - HTTP request metrics
   - Response time percentiles
   - Error rates
   - Search performance analytics

### Key Metrics to Monitor

#### Elasticsearch Metrics
```
# Cluster Health
elasticsearch_cluster_health_status
elasticsearch_cluster_health_number_of_nodes
elasticsearch_cluster_health_number_of_indices

# Performance
elasticsearch_indices_search_query_total
elasticsearch_indices_search_query_time_seconds
elasticsearch_indices_indexing_index_total
elasticsearch_indices_indexing_index_time_seconds

# Resource Usage
elasticsearch_jvm_memory_used_bytes
elasticsearch_filesystem_data_size_bytes
elasticsearch_process_cpu_seconds_total
```

#### Application Metrics
```
# HTTP Performance
http_requests_total
http_request_duration_seconds

# Search Analytics
elasticsearch_search_requests_total
elasticsearch_search_duration_seconds
elasticsearch_search_results_count
slow_queries_total

# System Resources
process_resident_memory_bytes
process_cpu_seconds_total
```

## 🚨 Alerting

### Alert Categories

#### Critical Alerts (Immediate Action Required)
- Elasticsearch cluster RED status
- Application downtime
- Critical disk usage (>95%)
- Connection pool exhaustion

#### Warning Alerts (Monitoring Required)
- Elasticsearch cluster YELLOW status
- High response times (>1s)
- High memory usage (>85%)
- Slow queries detected

### Alert Routing

```yaml
# AlertManager routes alerts by severity
Critical → Immediate notification (email/Slack)
Warning → Standard notification
Info → Log only
```

### Customizing Alerts

Edit alert rules in:
- `docker/prometheus/alerts/elasticsearch.yml`
- `docker/prometheus/alerts/applications.yml`

Example custom alert:
```yaml
- alert: CustomSlowSearch
  expr: histogram_quantile(0.95, rate(elasticsearch_search_duration_seconds_bucket[5m])) > 3
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Custom search performance alert"
    description: "95th percentile search time is above 3 seconds"
```

## 🔧 Configuration

### Prometheus Configuration

Main config: `docker/prometheus/prometheus.yml`

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'search-api'
    static_configs:
      - targets: ['host.docker.internal:8083']
    metrics_path: '/metrics'
    scrape_interval: 30s
```

### Adding New Metrics to Applications

1. **Add Prometheus dependency** to `go.mod`:
```go
require github.com/prometheus/client_golang v1.17.0
```

2. **Create custom metrics**:
```go
var customCounter = promauto.NewCounter(
    prometheus.CounterOpts{
        Name: "custom_operations_total",
        Help: "Total custom operations",
    },
)
```

3. **Record metrics in code**:
```go
customCounter.Inc()
```

4. **Expose metrics endpoint**:
```go
router.GET("/metrics", promhttp.Handler())
```

## 📊 Performance Optimization

### Query Performance Insights

Monitor these key search metrics:
```
# Average search time by query type
rate(elasticsearch_search_duration_seconds_sum[5m]) / rate(elasticsearch_search_duration_seconds_count[5m])

# Slow query rate
rate(slow_queries_total[5m])

# Search result distribution
histogram_quantile(0.95, rate(elasticsearch_search_results_count_bucket[5m]))
```

### Optimization Recommendations

Based on metrics, the system provides automated suggestions:
- **Move to filter context** when exact matches dominate
- **Add minimum_should_match** for better precision
- **Implement caching** for repeated queries
- **Optimize index mappings** for frequently searched fields

## 🔍 Troubleshooting

### Common Issues

#### High Memory Usage
```bash
# Check JVM heap usage
elasticsearch_jvm_memory_used_bytes{area="heap"} / elasticsearch_jvm_memory_max_bytes{area="heap"} * 100

# Recommended actions:
# 1. Increase JVM heap size
# 2. Optimize field data usage
# 3. Implement field data circuit breaker
```

#### Slow Queries
```bash
# Identify slow query patterns
rate(slow_queries_total[5m])

# Analyze by query type
elasticsearch_search_duration_seconds_bucket{query_type="match"}

# Recommended actions:
# 1. Add query profiling
# 2. Optimize index structure
# 3. Use query caching
```

#### High Error Rates
```bash
# Track error patterns
rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) * 100

# Recommended actions:
# 1. Check Elasticsearch connectivity
# 2. Verify index availability
# 3. Monitor resource constraints
```

### Debug Mode

Enable detailed metrics collection:
```yaml
# In application config
search:
  enable_profiling: true
  cache_results: true
```

## 🎯 Best Practices

### Metrics Naming
- Use descriptive names: `elasticsearch_search_duration_seconds`
- Include units: `_seconds`, `_bytes`, `_total`
- Group related metrics: `elasticsearch_*`, `http_*`

### Dashboard Design
- **Start with overview** - high-level health indicators
- **Drill down capability** - detailed metrics on demand
- **Consistent time ranges** - align all panels
- **Meaningful alerts** - avoid alert fatigue

### Resource Management
- **Retention policies** - 30 days for detailed metrics
- **Sampling rates** - adjust based on volume
- **Storage planning** - ~1GB per million samples

## 🔮 Advanced Features

### Planned Enhancements

1. **Distributed Tracing** with OpenTelemetry
2. **Custom Elasticsearch Metrics** via plugin
3. **ML-based Anomaly Detection**
4. **Cost Analysis Dashboard**
5. **Performance Regression Detection**

### Integration Examples

#### Slack Notifications
```yaml
# In alertmanager.yml
slack_configs:
- api_url: 'YOUR_SLACK_WEBHOOK'
  channel: '#elasticsearch-alerts'
  title: '🚨 {{ .GroupLabels.alertname }}'
```

#### Custom Exporters
```go
// Custom exporter for business metrics
func (e *CustomExporter) Collect(ch chan<- prometheus.Metric) {
    // Collect custom business metrics
    ch <- prometheus.MustNewConstMetric(
        e.businessMetric, prometheus.GaugeValue, value,
    )
}
```

## 🤝 Contributing

To add new monitoring features:
1. Create metrics in `internal/metrics/`
2. Add dashboard panels in `docker/grafana/dashboards/`
3. Define alerts in `docker/prometheus/alerts/`
4. Update documentation

---

**Ready to monitor your Elasticsearch playground like a pro!** 📈✨