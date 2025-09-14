module github.com/saif-islam/es-playground/projects/search-api

go 1.20

require (
	github.com/elastic/go-elasticsearch/v8 v8.11.1
	github.com/gin-gonic/gin v1.9.1
	github.com/google/uuid v1.4.0
	github.com/gorilla/websocket v1.5.0
	github.com/prometheus/client_golang v1.17.0
	github.com/redis/go-redis/v9 v9.3.0
	github.com/stretchr/testify v1.8.4
	go.opentelemetry.io/otel v1.20.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	go.opentelemetry.io/otel/sdk v1.20.0
	go.opentelemetry.io/otel/trace v1.20.0
	go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.46.1
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.46.1
	go.uber.org/zap v1.26.0
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/saif-islam/es-playground => ../../