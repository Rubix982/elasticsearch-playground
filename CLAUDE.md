# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an Elasticsearch playground repository for building practical, production-ready projects that explore Elasticsearch features in depth. The goal is to create complete, reusable projects that demonstrate real-world Elasticsearch usage patterns, from basic operations to advanced features like analytics, search, and data processing.

### Project Goals
- Build comprehensive Elasticsearch examples and tutorials
- Create production-ready projects that people can use and learn from
- Explore advanced ES features: search, aggregations, analytics, machine learning
- Demonstrate best practices for ES integration with Go applications
- Provide complete project templates for common ES use cases

## Development Commands

Based on the project configuration:

### Core Commands

- **Start Elasticsearch**: `docker-compose up elasticsearch`
- **Start Full Stack**: `docker-compose up` (ES + Kibana + apps)
- **Build All Projects**: `make build-all`
- **Test All Projects**: `make test-all`
- **Build Specific Project**: `make build PROJECT=search-api`
- **Test Specific Project**: `make test PROJECT=search-api`
- **Run Project**: `make run PROJECT=search-api`
- **Linting**: `golangci-lint run ./...`
- **Formatting**: `go fmt ./...`

### Git Workflow

- **Main branch**: `dev` (not `main`)
- **Protected branches**: `main`, `dev`, `production`
- **Commit format**: `PRIV-{ticketId}: {description}`
- **Pre-commit hooks**: `go fmt`, `golangci-lint run`, `go test ./...`
- **Pre-push hooks**: `make test-integration`

## Architecture

### Technology Stack

- **Language**: Go 1.21+
- **Framework**: Temporal
- **Test Framework**: testify
- **Build Tool**: make
- **Package Manager**: go mod
- **Containerization**: docker
- **Orchestration**: kubernetes
- **Databases**: PostgreSQL, Redis, Elasticsearch

### Project Structure

The repository is organized as an Elasticsearch playground with multiple project examples:

- `projects/` - Individual Elasticsearch project examples
  - `search-api/` - RESTful search API with advanced queries
  - `log-analyzer/` - Log analysis and visualization system  
  - `ecommerce-search/` - E-commerce product search engine
  - `real-time-analytics/` - Real-time data analytics dashboard
  - `document-store/` - Document management system
- `shared/` - Common utilities and ES client configurations
- `docker/` - Docker compositions for different ES setups
- `docs/` - Documentation and tutorials for each project

### Key Patterns

- **Config files**: `*.yaml`, `*.yml`, `*.json`, `Makefile`, `Dockerfile`
- **Test files**: `*_test.go`
- **Migration files**: `*_migration.go`
- **Spec files**: `*/sync/scan.yaml`, `*/api/apis.yaml`

## Code Quality Requirements

### Testing

- **Minimum coverage**: 70%
- **Target coverage**: 85%
- **Mock generation**: Use mockery
- **Parallel execution**: Enabled
- **Benchmarks and fuzz testing**: Enabled

### Code Standards

- **Naming conventions**:
  - Functions/variables: camelCase
  - Constants: UPPER_SNAKE_CASE
  - Types/interfaces: PascalCase
  - Packages: lowercase
- **Max function length**: 100 lines
- **Cyclomatic complexity threshold**: 10
- **Documentation**: Required (godoc format)

## Security

- **Secret scanning**: Enabled
- **Dependency scanning**: Enabled
- **Allowed secret patterns**: `CRYPT:*`
- **Blocked patterns**: `password=*`, `secret=*`, `token=*`, `key=*`, AWS credentials

## Development Environment

- **Environment**: local
- **Debug mode**: Enabled
- **Log level**: debug
- **Timezone**: Asia/Karachi
- **Terminal**: bash
- **Hot reload**: Enabled

## Database

- **Migration directory**: `migrations/`
- **Connection timeout**: 30s
- **Query timeout**: 60s
- **Slow query detection**: Enabled

## Monitoring & Deployment

- **Health check endpoints**: `/health`, `/ready`
- **Monitoring tools**: Prometheus, Grafana, Elasticsearch
- **Container registry**: docker.io
- **Kubernetes namespaces**: `default`, `priv-dev`, `priv-staging`, `priv-prod`
- **Helm charts**: Enabled
