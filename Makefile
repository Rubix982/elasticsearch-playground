.PHONY: help build test clean run docker-up docker-down write-test perf-test quick-demo

# Default project if not specified - focus on write-optimized Index Explorer
PROJECT ?= index-explorer

help: ## Show this help message
	@echo '🚀 Elasticsearch Playground - Write-Optimized Operations'
	@echo ''
	@echo 'Usage: make [target] [PROJECT=project-name]'
	@echo 'Default PROJECT: index-explorer (write-optimized focus)'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build specific project (default: search-api)
	@echo "Building $(PROJECT)..."
	@cd projects/$(PROJECT) && go build -v -o ../../bin/$(PROJECT) ./cmd/

build-all: ## Build all projects
	@echo "Building all projects..."
	@for project in $$(ls projects/); do \
		echo "Building $$project..."; \
		cd projects/$$project && go build -v -o ../../bin/$$project ./cmd/ && cd ../..; \
	done

test: ## Run tests for specific project
	@echo "Testing $(PROJECT)..."
	@cd projects/$(PROJECT) && go test -v ./...

test-all: ## Run tests for all projects
	@echo "Running tests for all projects..."
	@go test -v ./...

coverage: ## Generate coverage report for specific project
	@echo "Generating coverage for $(PROJECT)..."
	@cd projects/$(PROJECT) && go test -coverprofile=../../coverage/$(PROJECT).out ./...
	@go tool cover -html=coverage/$(PROJECT).out -o coverage/$(PROJECT).html

run: ## Run specific project
	@echo "Running $(PROJECT)..."
	@./bin/$(PROJECT)

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/ coverage/
	@go clean -cache

docker-up: ## Start Elasticsearch and Kibana
	@echo "Starting Elasticsearch stack..."
	@docker-compose up -d

docker-down: ## Stop Elasticsearch and Kibana
	@echo "Stopping Elasticsearch stack..."
	@docker-compose down

docker-logs: ## Show Elasticsearch logs
	@docker-compose logs -f elasticsearch

# Monitoring Commands
monitoring-up: ## Start full monitoring stack (Prometheus + Grafana + AlertManager)
	@echo "🔧 Starting monitoring stack..."
	@docker-compose --profile monitoring up -d
	@echo "✅ Monitoring services started!"
	@echo "📊 Grafana: http://localhost:3000 (admin/playground123)"
	@echo "🎯 Prometheus: http://localhost:9090"
	@echo "🚨 AlertManager: http://localhost:9093"

monitoring-down: ## Stop monitoring stack
	@echo "Stopping monitoring stack..."
	@docker-compose --profile monitoring down

monitoring-status: ## Check monitoring services status
	@echo "📊 Monitoring Services Status:"
	@echo ""
	@echo "🔍 Prometheus:"
	@curl -s http://localhost:9090/api/v1/query?query=up | jq '.data.result[] | {job: .metric.job, status: .value[1]}' 2>/dev/null || echo "❌ Prometheus not responding"
	@echo ""
	@echo "🔍 Grafana:"
	@curl -s http://localhost:3000/api/health | jq '.' 2>/dev/null || echo "❌ Grafana not responding"
	@echo ""
	@echo "🔍 Application Metrics:"
	@echo "  Index Explorer: $$(curl -s http://localhost:8082/metrics | grep -c '^http_requests_total' || echo '0') metrics available"
	@echo "  Search API: $$(curl -s http://localhost:8083/metrics | grep -c '^http_requests_total' || echo '0') metrics available"
	@echo "  Cluster Explorer: $$(curl -s http://localhost:8081/metrics | grep -c '^http_requests_total' || echo '0') metrics available"

monitoring-logs: ## Show monitoring services logs
	@docker-compose logs -f prometheus grafana alertmanager

performance-stack: ## Start performance monitoring stack
	@echo "⚡ Starting performance monitoring stack..."
	@docker-compose --profile performance --profile monitoring up -d
	@echo "✅ Performance stack ready!"

full-stack: ## Start complete stack (ES + Apps + Monitoring + Performance)
	@echo "🚀 Starting complete Elasticsearch playground stack..."
	@docker-compose --profile monitoring --profile performance up -d
	@echo "✅ Full stack ready!"
	@echo ""
	@echo "🎯 Services Available:"
	@echo "  • Elasticsearch: http://localhost:9200"
	@echo "  • Kibana: http://localhost:5601"
	@echo "  • Grafana: http://localhost:3000 (admin/playground123)"
	@echo "  • Prometheus: http://localhost:9090"
	@echo "  • AlertManager: http://localhost:9093"
	@echo ""
	@echo "🔧 Next Steps:"
	@echo "  make build-all              # Build all applications"
	@echo "  make monitoring-status      # Check monitoring health"

lint: ## Run linter
	@golangci-lint run ./...

fmt: ## Format code
	@go fmt ./...

mod-tidy: ## Tidy go modules
	@go mod tidy

deps: ## Download dependencies
	@go mod download

init-project: ## Initialize new project (usage: make init-project PROJECT=new-project-name)
	@echo "Creating new project: $(PROJECT)"
	@mkdir -p projects/$(PROJECT)/cmd projects/$(PROJECT)/internal projects/$(PROJECT)/pkg
	@mkdir -p bin coverage

# Development helpers
dev-setup: docker-up deps ## Setup development environment
	@echo "Development environment ready!"

dev-reset: docker-down clean docker-up ## Reset development environment
	@echo "Development environment reset!"

# Write-Optimization Focused Commands
write-test: ## Quick write performance test with Index Explorer
	@echo "🚀 Testing write-optimized operations..."
	@echo "1. Starting Index Explorer..."
	@make build PROJECT=index-explorer
	@echo "2. Creating write-optimized index..."
	@curl -s -X POST "http://localhost:8082/api/v1/indices/write-optimized" \
		-H "Content-Type: application/json" \
		-d '{"index_name":"perf-test","expected_volume":"high","text_heavy":true,"ingestion_rate":"high"}' || echo "Service not running - start with 'make run'"
	@echo ""
	@echo "3. Testing bulk operations..."
	@curl -s -X POST "http://localhost:8082/api/v1/indices/perf-test/bulk" \
		-H "Content-Type: application/json" \
		-d '{"operations":[{"action":"index","document":{"title":"Performance Test","content":"Large text content for write optimization testing..."}}],"optimize_for":"write_throughput"}' || echo "Create index first"

perf-test: ## Run comprehensive write performance benchmarks
	@echo "📊 Running write performance benchmarks..."
	@echo "Building index-explorer for benchmarks..."
	@make build PROJECT=index-explorer
	@echo "Run this after starting the service: make run PROJECT=index-explorer"
	@echo "Then in another terminal:"
	@echo "  curl -X POST http://localhost:8082/api/v1/indices/write-optimized -d '{\"index_name\":\"benchmark\",\"expected_volume\":\"high\"}'"
	@echo "  curl -X POST http://localhost:8082/api/v1/indices/benchmark/bulk -d '{\"operations\":[...],\"optimize_for\":\"write_throughput\"}'"

quick-demo: docker-up ## Quick demo of write-optimized features
	@echo "🎯 Quick Write-Optimization Demo"
	@echo "⏳ Waiting for Elasticsearch to start..."
	@sleep 10
	@echo "🔨 Building index-explorer..."
	@make build PROJECT=index-explorer
	@echo ""
	@echo "🚀 Start the Index Explorer with: make run PROJECT=index-explorer"
	@echo "🌐 Then visit: http://localhost:8082/debug/examples"
	@echo ""
	@echo "💡 Key endpoints to try:"
	@echo "  • POST http://localhost:8082/api/v1/indices/write-optimized"
	@echo "  • POST http://localhost:8082/api/v1/bulk/adaptive"
	@echo "  • GET  http://localhost:8082/api/v1/metrics/write-performance"

status: ## Show service status
	@echo "📊 Service Status:"
	@echo ""
	@echo "🔍 Elasticsearch:"
	@curl -s http://localhost:9200/_cluster/health | jq '.' 2>/dev/null || echo "❌ Elasticsearch not responding"
	@echo ""
	@echo "🔍 Index Explorer:"
	@curl -s http://localhost:8082/health | jq '.' 2>/dev/null || echo "❌ Index Explorer not responding"

# Write-focused project commands
run-index-explorer: ## Run Index & Document Explorer specifically
	@echo "🚀 Starting Write-Optimized Index & Document Explorer..."
	@echo "📍 API: http://localhost:8082/api/v1"
	@echo "📍 Health: http://localhost:8082/health" 
	@echo "📍 Examples: http://localhost:8082/debug/examples"
	@make run PROJECT=index-explorer

playground-setup: ## Complete playground setup for write-optimization exploration
	@echo "🏗️  Setting up Write-Optimization Playground..."
	@make docker-up
	@echo "⏳ Waiting for services..."
	@sleep 15
	@make build PROJECT=index-explorer
	@echo ""
	@echo "✅ Setup complete! Next steps:"
	@echo "1. make run-index-explorer    # Start the API server"
	@echo "2. make write-test           # Test write operations"
	@echo "3. Open http://localhost:8082/dashboard for monitoring dashboard"
	@echo "4. Open http://localhost:8082/debug/examples for API examples"

# Dataset Generation Commands
generate-dataset: ## Generate sample datasets (usage: make generate-dataset TYPE=small COUNT=1000)
	@echo "📝 Generating $(TYPE) dataset with $(COUNT) documents..."
	@cd datasets/generators && python3 document_generator.py --type=$(TYPE) --count=$(COUNT) --output=../samples/$(TYPE)-$(COUNT).ndjson
	@echo "✅ Dataset generated: datasets/samples/$(TYPE)-$(COUNT).ndjson"

generate-ecommerce: ## Generate e-commerce product catalog
	@echo "🛒 Generating e-commerce catalog..."
	@cd datasets/generators && python3 ecommerce_generator.py --count=10000 --output=../samples/ecommerce-catalog.ndjson
	@echo "✅ E-commerce catalog generated!"

generate-news: ## Generate news articles dataset
	@echo "📰 Generating news articles..."
	@cd datasets/generators && python3 news_generator.py --count=5000 --output=../samples/news-articles.ndjson
	@echo "✅ News articles generated!"

generate-logs: ## Generate log events dataset
	@echo "📊 Generating log events..."
	@cd datasets/generators && python3 logs_generator.py --count=50000 --type=mixed --output=../samples/log-events.ndjson
	@echo "✅ Log events generated!"

generate-performance: ## Generate high-performance test dataset
	@echo "⚡ Generating performance test dataset..."
	@cd datasets/generators && python3 performance_generator.py --count=100000 --type=mixed --output=../samples/performance-test.ndjson
	@echo "✅ Performance dataset generated!"

generate-all-samples: ## Generate all sample datasets
	@echo "📦 Generating all sample datasets..."
	@make generate-dataset TYPE=small COUNT=1000
	@make generate-dataset TYPE=medium COUNT=5000  
	@make generate-dataset TYPE=large COUNT=1000
	@make generate-dataset TYPE=mixed COUNT=10000
	@make generate-ecommerce
	@make generate-news
	@make generate-logs
	@echo "✅ All sample datasets generated!"

dataset-info: ## Show information about available datasets
	@echo "📊 Available Sample Datasets:"
	@ls -lah datasets/samples/ 2>/dev/null || echo "No datasets found. Run 'make generate-all-samples' to create them."
	@echo ""
	@echo "💡 Dataset Generators Available:"
	@echo "  • make generate-dataset TYPE=small COUNT=1000    # Generic small docs"
	@echo "  • make generate-ecommerce                        # E-commerce products"
	@echo "  • make generate-news                             # News articles"
	@echo "  • make generate-logs                             # Application logs"
	@echo "  • make generate-performance                      # High-volume performance testing"

# CLI Tools
run-cli: ## Run interactive CLI tool
	@echo "🖥️  Starting Interactive CLI..."
	@cd projects/index-explorer && go run cmd/cli/main.go

run-perf-test: ## Run performance testing tool
	@echo "⚡ Starting Performance Test..."
	@cd projects/index-explorer && go run cmd/perf-test/main.go $(ARGS)