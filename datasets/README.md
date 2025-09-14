# Sample Datasets for Write-Optimization Testing

This directory contains various sample datasets designed to test different write-optimization scenarios in Elasticsearch.

## Dataset Categories

### 📄 Document Corpus Datasets

- **Small Documents** (< 1KB): Logs, events, metrics
- **Medium Documents** (1-10KB): Articles, emails, product descriptions
- **Large Documents** (10-100KB): Research papers, documentation, reports
- **Huge Documents** (> 100KB): Books, manuals, comprehensive documents

### 🏭 Industry-Specific Datasets

- **E-commerce**: Product catalogs, reviews, inventory
- **News & Media**: Articles, comments, metadata
- **Financial**: Transactions, market data, reports
- **Healthcare**: Patient records, research data (anonymized)
- **Logs & Analytics**: Application logs, metrics, events

### ⚡ Performance Testing Datasets

- **High Volume**: Millions of small documents for throughput testing
- **Mixed Load**: Varied document sizes for realistic workload simulation
- **Bulk Import**: Pre-formatted NDJSON files for bulk import testing

## Usage Examples

### Quick Start with Small Dataset

```bash
# Generate 1,000 small documents
make generate-dataset TYPE=small COUNT=1000

# Import using Index Explorer
curl -X POST "http://localhost:8082/api/v1/indices/test-small/import/ndjson?batch_size=500" \
  --data-binary @datasets/small-documents.ndjson
```

### Performance Testing

```bash
# Generate large performance dataset
make generate-dataset TYPE=performance COUNT=100000

# Run comprehensive performance test
cd projects/index-explorer && go run cmd/perf-test/main.go heavy
```

### Custom Dataset Generation

```bash
# Create custom dataset
python3 datasets/generators/custom_generator.py \
  --type=mixed \
  --count=10000 \
  --output=datasets/custom-mixed.ndjson
```

## Dataset Specifications

| Dataset               | Document Count | Avg Size | Total Size | Use Case              |
| --------------------- | -------------- | -------- | ---------- | --------------------- |
| **sample-small**      | 1,000          | 500B     | ~500KB     | Quick testing         |
| **sample-medium**     | 5,000          | 5KB      | ~25MB      | Standard testing      |
| **sample-large**      | 1,000          | 50KB     | ~50MB      | Large doc testing     |
| **performance-high**  | 100,000        | 2KB      | ~200MB     | Throughput testing    |
| **ecommerce-catalog** | 10,000         | 3KB      | ~30MB      | E-commerce simulation |
| **news-articles**     | 5,000          | 8KB      | ~40MB      | Content management    |
| **log-events**        | 50,000         | 200B     | ~10MB      | Log analysis          |

## Files Structure

```
datasets/
├── README.md                    # This file
├── generators/                  # Dataset generation scripts
│   ├── document_generator.py    # Generic document generator
│   ├── ecommerce_generator.py   # E-commerce specific
│   ├── news_generator.py        # News articles
│   ├── logs_generator.py        # Log events
│   └── performance_generator.py # Performance testing
├── samples/                     # Pre-generated sample datasets
│   ├── small-documents.ndjson   # 1K small docs
│   ├── medium-documents.ndjson  # 5K medium docs
│   ├── large-documents.ndjson   # 1K large docs
│   └── mixed-workload.ndjson    # Mixed sizes
└── schemas/                     # Document schemas
    ├── ecommerce.json          # Product schema
    ├── news.json               # Article schema
    ├── logs.json               # Log event schema
    └── generic.json            # Generic document schema
```

## Performance Benchmarks

Based on testing with different document sizes:

### Throughput Expectations

- **Small documents** (< 1KB): 5,000-10,000 docs/sec
- **Medium documents** (1-10KB): 1,000-3,000 docs/sec
- **Large documents** (10-100KB): 200-800 docs/sec
- **Huge documents** (> 100KB): 50-200 docs/sec

### Optimization Recommendations

- **Small documents**: Use batch sizes of 3,000-5,000
- **Medium documents**: Use batch sizes of 500-1,500
- **Large documents**: Use batch sizes of 100-500
- **Huge documents**: Use batch sizes of 10-100

## Getting Started

1. **Generate your first dataset:**

   ```bash
   cd datasets/generators
   python3 document_generator.py --help
   ```

2. **Import into Elasticsearch:**

   ```bash
   make playground-setup  # Start services
   make run-index-explorer  # Start API
   # Import via CLI or API
   ```

3. **Run performance tests:**
   ```bash
   cd projects/index-explorer
   go run cmd/perf-test/main.go quick
   ```

## Contributing

To add new datasets:

1. Create generator script in `generators/`
2. Add schema definition in `schemas/`
3. Generate sample dataset in `samples/`
4. Update this README with specifications
5. Add performance benchmarks

## Tips for Write Optimization

### Document Structure

- **Minimize nested objects** for better indexing speed
- **Use appropriate field types** (text vs keyword)
- **Avoid deeply nested arrays**
- **Consider field exclusions** for non-searchable data

### Indexing Strategy

- **Start with 0 replicas** during bulk loading
- **Increase refresh interval** for write-heavy workloads
- **Use bulk operations** - never single document operations
- **Monitor segment count** and merge policy

### Performance Testing

- **Test with realistic data sizes**
- **Simulate actual query patterns**
- **Monitor resource usage** (CPU, memory, disk)
- **Test different batch sizes** to find optimal settings
