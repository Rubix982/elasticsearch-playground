# Write-Heavy Optimization Patterns for Elasticsearch

A comprehensive guide to mastering write-optimized Elasticsearch operations using the Index & Document Explorer.

## 🎯 Core Philosophy: Bulk-First, Write-Optimized

Elasticsearch excels when you embrace its write-heavy, append-only architecture. This guide shows you how to leverage this for maximum performance with large text corpora.

## 🚀 Quick Start

### 1. Start the Index Explorer

```bash
# Start Elasticsearch (if not running)
make docker-up

# Build and run the index explorer
make build PROJECT=index-explorer
make run PROJECT=index-explorer
```

**Access Points:**
- **API**: http://localhost:8082/api/v1
- **Health**: http://localhost:8082/health
- **Examples**: http://localhost:8082/debug/examples

## 📖 Pattern 1: Write-Optimized Index Creation

### The Problem
Default Elasticsearch settings prioritize read performance and consistency over write throughput.

### The Solution: Write-Optimized Index Creation

```bash
# Create a write-optimized index for high-volume text ingestion
curl -X POST "http://localhost:8082/api/v1/indices/write-optimized" \
  -H "Content-Type: application/json" \
  -d '{
    "index_name": "large-text-corpus",
    "expected_volume": "high",
    "expected_doc_size": "large", 
    "ingestion_rate": "high",
    "text_heavy": true
  }'
```

**What happens behind the scenes:**
- **Refresh interval**: Extended to 30s (vs default 1s)
- **Replica count**: Set to 0 during bulk loading
- **Translog**: Optimized for async durability
- **Merge policy**: Configured for write performance
- **Compression**: Best compression for text content
- **Buffer sizes**: Optimized for large documents

### Verify the Optimizations

```bash
# Check what optimizations were applied
curl "http://localhost:8082/api/v1/indices/large-text-corpus" | jq '.Settings'
```

## 📖 Pattern 2: High-Performance Bulk Operations

### The Problem
Individual document operations are 10-100x slower than bulk operations in Elasticsearch.

### The Solution: Intelligent Bulk Processing

#### A. Standard Bulk Operation

```bash
# Bulk index with optimal settings
curl -X POST "http://localhost:8082/api/v1/indices/large-text-corpus/bulk" \
  -H "Content-Type: application/json" \
  -d '{
    "operations": [
      {
        "action": "index",
        "document": {
          "title": "Sample Document 1",
          "content": "Large text content here...",
          "timestamp": "2024-01-01T00:00:00Z"
        }
      },
      {
        "action": "index", 
        "document": {
          "title": "Sample Document 2",
          "content": "More large text content...",
          "timestamp": "2024-01-01T01:00:00Z"
        }
      }
    ],
    "batch_size": 1000,
    "parallel_workers": 8,
    "optimize_for": "write_throughput"
  }'
```

#### B. Adaptive Bulk Processing (Recommended)

```bash
# Let the system automatically optimize batch sizes and workers
curl -X POST "http://localhost:8082/api/v1/bulk/adaptive" \
  -H "Content-Type: application/json" \
  -d '{
    "index_name": "large-text-corpus",
    "documents": [
      {"title": "Doc 1", "content": "Large text..."},
      {"title": "Doc 2", "content": "More text..."}
    ],
    "auto_batch_size": true,
    "target_throughput": "max",
    "error_tolerance": "medium"
  }'
```

**Adaptive optimization automatically:**
- Analyzes document sizes to determine optimal batch size
- Calculates worker count based on volume and target throughput
- Adjusts settings for maximum write performance

#### C. NDJSON Import (Best for Large Datasets)

```bash
# Import large NDJSON files with optimal performance
curl -X POST "http://localhost:8082/api/v1/indices/large-text-corpus/import/ndjson?batch_size=1000&workers=8" \
  -H "Content-Type: application/x-ndjson" \
  --data-binary @large-dataset.ndjson
```

**Example NDJSON format:**
```ndjson
{"title": "Document 1", "content": "Large text content here...", "category": "news"}
{"title": "Document 2", "content": "More large text content...", "category": "articles"}
{"title": "Document 3", "content": "Even more text content...", "category": "blogs"}
```

## 📖 Pattern 3: Write Performance Monitoring

### The Problem
Without monitoring, you can't optimize what you can't measure.

### The Solution: Comprehensive Write Metrics

```bash
# Get detailed write performance metrics
curl "http://localhost:8082/api/v1/indices/large-text-corpus/metrics/write-performance"
```

**Key metrics to monitor:**
- **Indexing rate**: Documents per second
- **Write latency**: Time per document
- **Segment count**: Merge efficiency indicator
- **Translog size**: Write buffer usage
- **Optimization score**: Overall write health (0-100)

### Real-Time Performance Analysis

```bash
# Get comprehensive write performance analysis
curl "http://localhost:8082/api/v1/indices/large-text-corpus/analyze/write-performance"
```

**This provides:**
- Performance bottleneck identification
- Health assessment with specific issues
- Optimization recommendations
- Resource utilization analysis

## 📖 Pattern 4: Dynamic Index Optimization

### The Problem
Index settings that work during initial loading may not be optimal for ongoing operations.

### The Solution: Adaptive Index Tuning

#### A. Get Optimization Recommendations

```bash
# Get recommendations without applying changes
curl "http://localhost:8082/api/v1/indices/large-text-corpus/recommendations?workload=bulk_write&corpus_size=large"
```

#### B. Apply Write-Heavy Optimizations

```bash
# Automatically tune index for write-heavy workload
curl -X POST "http://localhost:8082/api/v1/indices/large-text-corpus/tune/write-heavy" \
  -H "Content-Type: application/json" \
  -d '{
    "workload": "bulk_write",
    "corpus_size": "large", 
    "priority": "write_throughput",
    "avg_doc_size": "100KB"
  }'
```

#### C. Manual Optimization with Full Control

```bash
# Apply specific optimizations with detailed control
curl -X POST "http://localhost:8082/api/v1/indices/large-text-corpus/optimize" \
  -H "Content-Type: application/json" \
  -d '{
    "optimize_for": "write_throughput",
    "workload": "bulk_write",
    "corpus_size": "huge",
    "priority": "write_throughput", 
    "apply_changes": true
  }'
```

## 📖 Pattern 5: Document Size-Aware Optimization

### The Problem
Different document sizes require different optimization strategies.

### The Solution: Size-Specific Tuning

#### Small Documents (< 1KB) - High Volume Strategy
```bash
curl -X POST "http://localhost:8082/api/v1/indices/write-optimized" \
  -d '{
    "index_name": "small-docs-index",
    "expected_doc_size": "small",
    "expected_volume": "high",
    "ingestion_rate": "high"
  }'

# Optimized settings: Large batch sizes (5000), many workers
curl -X POST "http://localhost:8082/api/v1/bulk/adaptive" \
  -d '{
    "index_name": "small-docs-index",
    "documents": [...],
    "target_throughput": "max"
  }'
```

#### Large Documents (10-100KB) - Balanced Strategy
```bash
curl -X POST "http://localhost:8082/api/v1/indices/write-optimized" \
  -d '{
    "index_name": "large-docs-index", 
    "expected_doc_size": "large",
    "expected_volume": "medium",
    "text_heavy": true
  }'

# Optimized settings: Medium batch sizes (500), compression enabled
```

#### Huge Documents (> 100KB) - Memory-Aware Strategy
```bash
curl -X POST "http://localhost:8082/api/v1/indices/write-optimized" \
  -d '{
    "index_name": "huge-docs-index",
    "expected_doc_size": "huge", 
    "expected_volume": "low",
    "text_heavy": true
  }'

# Optimized settings: Small batch sizes (100), high compression
```

## 📖 Pattern 6: Error Handling and Resilience

### The Problem
Bulk operations can partially fail, requiring sophisticated error handling.

### The Solution: Resilient Bulk Processing

```bash
# Bulk operation with error tolerance
curl -X POST "http://localhost:8082/api/v1/indices/large-text-corpus/bulk" \
  -d '{
    "operations": [...],
    "error_tolerance": "high",
    "optimize_for": "write_throughput"
  }'
```

**Error tolerance levels:**
- **Low**: Stop on first error
- **Medium**: Continue with warnings, report errors
- **High**: Maximum resilience, continue processing

### Monitor Bulk Operation Results

```bash
# Check bulk operation status
curl "http://localhost:8082/api/v1/bulk/status"
```

## 📖 Pattern 7: Production Deployment Workflow

### The Complete Write-Optimized Workflow

#### Phase 1: Initial Setup
```bash
# 1. Create write-optimized index
curl -X POST "http://localhost:8082/api/v1/indices/write-optimized" \
  -d '{
    "index_name": "production-corpus",
    "expected_volume": "high",
    "expected_doc_size": "large",
    "text_heavy": true
  }'

# 2. Verify optimizations
curl "http://localhost:8082/api/v1/indices/production-corpus/recommendations"
```

#### Phase 2: Bulk Loading
```bash
# 3. High-performance bulk import
curl -X POST "http://localhost:8082/api/v1/indices/production-corpus/import/ndjson?batch_size=1000&workers=16" \
  --data-binary @production-data.ndjson

# 4. Monitor performance during loading
curl "http://localhost:8082/api/v1/indices/production-corpus/analyze/write-performance"
```

#### Phase 3: Post-Load Optimization
```bash
# 5. Add replicas after bulk loading
curl -X POST "http://localhost:8082/api/v1/indices/production-corpus/optimize" \
  -d '{
    "optimize_for": "read_performance",
    "apply_changes": true
  }'

# 6. Force merge to optimize segments
# (This would be done through direct ES API or added to the explorer)
```

## 📈 Performance Benchmarks

### Expected Throughput Targets

| Document Size | Batch Size | Workers | Expected Rate | Use Case |
|---------------|------------|---------|---------------|----------|
| < 1KB | 5000 | 8-16 | 10,000/sec | Logs, events |
| 1-10KB | 1000 | 4-8 | 5,000/sec | Articles, emails |
| 10-100KB | 500 | 2-4 | 1,000/sec | Documents, reports |
| > 100KB | 100 | 1-2 | 100/sec | Books, manuals |

### Optimization Score Targets

- **90-100**: Excellent - Production ready
- **70-89**: Good - Minor optimizations needed
- **50-69**: Fair - Significant improvements needed
- **< 50**: Poor - Major optimization required

## 🔧 Troubleshooting Common Issues

### Issue 1: Low Indexing Rate
```bash
# Diagnosis
curl "http://localhost:8082/api/v1/indices/my-index/analyze/write-performance"

# Common solutions:
# - Increase batch size
# - Add more parallel workers  
# - Reduce refresh frequency
# - Optimize document structure
```

### Issue 2: High Memory Usage
```bash
# Check recommendations
curl "http://localhost:8082/api/v1/indices/my-index/recommendations"

# Common solutions:
# - Reduce batch size
# - Enable compression
# - Increase flush threshold
# - Optimize field mappings
```

### Issue 3: Slow Merges
```bash
# Tune merge policy
curl -X POST "http://localhost:8082/api/v1/indices/my-index/tune/write-heavy" \
  -d '{
    "workload": "bulk_write",
    "priority": "write_throughput"
  }'
```

## 💡 Best Practices Summary

### 1. Index Design
- **Always start write-optimized** for bulk loading scenarios
- **Use appropriate shard counts** based on data volume
- **Enable compression** for text-heavy content
- **Optimize refresh intervals** for write workloads

### 2. Bulk Operations
- **Never use individual operations** for bulk data
- **Use adaptive bulk processing** for optimal performance
- **Monitor and adjust** batch sizes based on document size
- **Implement proper error handling** for resilience

### 3. Monitoring
- **Track write performance metrics** continuously
- **Monitor optimization scores** and act on recommendations
- **Watch for performance degradation** during high load
- **Regular performance analysis** to identify issues early

### 4. Lifecycle Management
- **Start with 0 replicas** during bulk loading
- **Add replicas** after initial data load
- **Force merge** periodically for optimal segment structure
- **Adjust settings** based on workload changes

## 🎉 Key Takeaways

**Elasticsearch is fundamentally designed for write-heavy workloads:**

1. **Bulk operations are 10-100x faster** than individual operations
2. **Write optimization can improve throughput by 300-500%**
3. **Document size matters** - optimize batch sizes accordingly
4. **Monitoring is essential** - you can't optimize what you don't measure
5. **Adaptive approaches work best** - let the system optimize itself

**Ready to achieve maximum write performance with your Elasticsearch deployment!** 🚀