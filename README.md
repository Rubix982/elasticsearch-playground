# 🚀 Elasticsearch Playground - Write-Optimized Operations

A comprehensive learning and experimentation platform for **write-heavy Elasticsearch workloads**, focusing on bulk operations, performance optimization, and real-world text corpus management.

> **Core Philosophy**: Elasticsearch excels when treated as a write-optimized database that prioritizes ingestion throughput for large text corpora.

[![CI/CD](https://github.com/saif-islam/es-playground/actions/workflows/ci.yml/badge.svg)](https://github.com/saif-islam/es-playground/actions)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)
[![Elasticsearch](https://img.shields.io/badge/Elasticsearch-8.11+-005571?logo=elasticsearch)](https://www.elastic.co/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## ✨ Key Features

### 🏗️ **Write-Optimized Index Management**
- Intelligent index creation with write-heavy optimizations
- Adaptive settings based on document size and volume
- Performance-first configuration recommendations

### 📦 **High-Performance Bulk Operations**
- Bulk-first approach (10-100x faster than individual operations)
- Adaptive batch sizing based on document characteristics
- Parallel processing with configurable worker pools
- NDJSON import with streaming support

### 📊 **Real-Time Performance Monitoring**
- Live write performance metrics and dashboards
- Optimization scoring (0-100 scale)
- Bottleneck identification and recommendations
- Resource utilization tracking

### 🛠️ **Complete Developer Experience**
- Interactive CLI tool for exploration
- Web-based monitoring dashboard
- Comprehensive test suite with benchmarks
- Sample datasets for different use cases

## 🏗️ Repository Structure

```
├── projects/                    # Production-ready Elasticsearch projects
│   ├── index-explorer/         # 🚀 Write-optimized index & document management
│   ├── cluster-explorer/       # 🔍 Cluster health and node management  
│   └── search-api/             # 📊 Advanced search patterns and queries
├── datasets/                   # Sample data generators and schemas
│   ├── generators/             # Python scripts for realistic test data
│   ├── samples/                # Generated datasets for testing
│   └── schemas/                # JSON schemas with ES mappings
├── shared/                     # Common utilities and ES client configurations
├── docker/                     # Write-optimized Docker configurations
└── .github/workflows/          # Comprehensive CI/CD pipeline
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Make (optional, for convenience commands)

### Setup

1. **Clone and navigate to the repository**

   ```bash
   git clone <repo-url>
   cd es-playground
   ```

2. **Start Elasticsearch stack**

   ```bash
   make docker-up
   # or manually: docker-compose up -d
   ```

3. **Wait for services to be ready**

   ```bash
   # Check Elasticsearch health
   curl http://localhost:9200/_cluster/health

   # Access Kibana (optional)
   open http://localhost:5601
   ```

4. **Download dependencies**
   ```bash
   make deps
   ```

## 📚 Available Projects

| Project | Focus | Port | Status | Key Features |
|---------|-------|------|--------|--------------|
| **[Index Explorer](projects/index-explorer/)** | **Write Optimization** | 8082 | ✅ **Production Ready** | Bulk operations, performance monitoring, CLI tools, web dashboard |
| [Cluster Explorer](projects/cluster-explorer/) | Cluster Management | 8081 | ✅ **Complete** | Node health, shard allocation, cluster monitoring |
| **[Search API](projects/search-api/)** | **Query Optimization** | 8083 | ✅ **Production Ready** | Advanced search, query optimization, analytics, suggestions |

### 🎯 **Philosophy: Quality Over Quantity**

Instead of many incomplete projects, we've focused on building **fewer, better examples** that thoroughly demonstrate Elasticsearch concepts with production-ready code, comprehensive documentation, and real-world usage patterns.

### 🚀 **Flagship: Index Explorer**

The **Index Explorer** is our flagship project, showcasing:

- **Write-First Philosophy**: Treating ES as a write-optimized database
- **10-100x Performance**: Through bulk operations and optimization
- **Real-Time Monitoring**: Live performance dashboards and metrics
- **Interactive Tools**: CLI for exploration, web UI for monitoring
- **Complete Testing**: Benchmarks, performance tests, sample datasets
- **Production Ready**: Docker deployment, CI/CD, comprehensive docs

## 🛠️ Development Commands

```bash
# Build specific project
make build PROJECT=search-api

# Run specific project
make run PROJECT=search-api

# Test specific project
make test PROJECT=search-api

# Build all projects
make build-all

# Run all tests
make test-all

# Start development environment
make dev-setup
```

## 🐳 Docker Services

The `docker-compose.yml` includes:

### Core Services
- **Elasticsearch 8.11.1** - Main search engine (port 9200)
- **Kibana 8.11.1** - Visualization and management (port 5601)
- **Redis** - Caching and session storage (port 6379)
- **PostgreSQL** - Relational data storage (port 5432)

### Monitoring Stack 📊
- **Prometheus** - Metrics collection (port 9090)
- **Grafana** - Visualization dashboards (port 3000)
- **AlertManager** - Alert routing and notifications (port 9093)
- **Node Exporter** - System metrics (port 9100)
- **Elasticsearch Exporter** - ES-specific metrics (port 9114)
- **Redis Exporter** - Redis metrics (port 9121)
- **Postgres Exporter** - PostgreSQL metrics (port 9187)

### Data Processing (Optional)
- **Filebeat** - Log shipping (profile: `monitoring`)
- **Logstash** - Data processing (profile: `monitoring`)

### Using Different Stacks

```bash
# Basic Elasticsearch stack
make docker-up

# Full monitoring stack
make monitoring-up

# Complete stack (ES + Monitoring + Performance)
make full-stack
```

## 🧪 Testing

Each project includes comprehensive tests:

```bash
# Unit tests
make test PROJECT=search-api

# Integration tests (requires running ES)
make test-integration PROJECT=search-api

# All tests with coverage
make coverage PROJECT=search-api
```

## 📖 Learning Path

**Beginner → Intermediate → Advanced**

### 🎯 **Recommended Learning Sequence:**

1. **Index Explorer** - Master write-optimized operations and bulk indexing
2. **Search API** - Learn advanced query patterns and optimization
3. **Cluster Explorer** - Understand cluster management and operations
4. **Monitoring & Observability** - Production monitoring with Prometheus + Grafana

### 📊 **Monitoring & Observability**

Learn production-ready monitoring:
- **Prometheus metrics** - Custom application metrics
- **Grafana dashboards** - Real-time visualization
- **Alerting** - Proactive issue detection
- **Performance analysis** - Query optimization insights

See [MONITORING.md](MONITORING.md) for detailed setup and best practices.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Follow the coding standards in CLAUDE.md
4. Add tests for your changes
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## 📝 Project Ideas

Looking to extend this playground? Consider adding:

- **Time Series Database** - Using ES for metrics storage
- **Geospatial Search** - Location-based applications
- **Machine Learning** - Anomaly detection and classification
- **Graph Analytics** - Relationship analysis
- **Security Analytics** - SIEM-like functionality

## 📞 Support

- 📚 [Elasticsearch Documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html)
- 🐛 [Report Issues](../../issues)
- 💬 [Discussions](../../discussions)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Happy Searching!** 🔍✨
