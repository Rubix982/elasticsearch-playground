module github.com/saif-islam/es-playground/projects/index-explorer

go 1.21

require (
	github.com/elastic/go-elasticsearch/v8 v8.11.1
	github.com/gin-gonic/gin v1.9.1
	go.uber.org/zap v1.26.0
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/saif-islam/es-playground => ../../