package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/saif-islam/es-playground/projects/index-explorer/internal/handlers"
	"github.com/saif-islam/es-playground/projects/index-explorer/internal/services"
	"github.com/saif-islam/es-playground/shared"
)

// Config represents the application configuration
type Config struct {
	Server        ServerConfig        `yaml:"server"`
	Elasticsearch ElasticsearchConfig `yaml:"elasticsearch"`
	Logging       LoggingConfig       `yaml:"logging"`
}

type ServerConfig struct {
	Port            int           `yaml:"port"`
	Host            string        `yaml:"host"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type ElasticsearchConfig struct {
	URLs      []string  `yaml:"urls"`
	Username  string    `yaml:"username"`
	Password  string    `yaml:"password"`
	APIKey    string    `yaml:"api_key"`
	TLSConfig TLSConfig `yaml:"tls"`
}

type TLSConfig struct {
	InsecureSkipVerify bool `yaml:"insecure_skip_verify"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

func main() {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := initLogger(config.Logging)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting Index & Document Explorer",
		zap.String("version", "1.0.0"),
		zap.Int("port", config.Server.Port))

	// Initialize Elasticsearch client
	esConfig := &shared.ESConfig{
		URLs:     config.Elasticsearch.URLs,
		Username: config.Elasticsearch.Username,
		Password: config.Elasticsearch.Password,
		APIKey:   config.Elasticsearch.APIKey,
		TLSConfig: &shared.TLSConfig{
			InsecureSkipVerify: config.Elasticsearch.TLSConfig.InsecureSkipVerify,
		},
	}

	esClient, err := shared.NewESClient(esConfig, logger)
	if err != nil {
		logger.Fatal("Failed to create Elasticsearch client", zap.Error(err))
	}

	// Wait for Elasticsearch to be ready
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := esClient.WaitForCluster(ctx, "yellow", 30*time.Second); err != nil {
		logger.Fatal("Elasticsearch cluster not ready", zap.Error(err))
	}

	// Initialize services
	indexService := services.NewIndexService(esClient, logger)
	documentService := services.NewDocumentService(esClient, logger)

	// Initialize handlers
	indexHandler := handlers.NewIndexHandler(indexService, documentService, logger)
	documentHandler := handlers.NewDocumentHandler(documentService, logger)

	// Setup HTTP server
	if config.Logging.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := setupRoutes(indexHandler, documentHandler, logger)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port),
		Handler:      router,
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Index & Document Explorer server starting",
			zap.String("address", server.Addr),
			zap.String("web_ui", fmt.Sprintf("http://localhost:%d", config.Server.Port)))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Index & Document Explorer...")

	// Graceful shutdown
	ctx, cancel = context.WithTimeout(context.Background(), config.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Index & Document Explorer exited")
}

func loadConfig() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	// Make path relative to project root
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join("projects/index-explorer", configPath)
	}

	// Default configuration
	config := &Config{
		Server: ServerConfig{
			Port:            8082,
			Host:            "0.0.0.0",
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			ShutdownTimeout: 10 * time.Second,
		},
		Elasticsearch: ElasticsearchConfig{
			URLs:     []string{"http://localhost:9200"},
			Username: "",
			Password: "",
			APIKey:   "",
			TLSConfig: TLSConfig{
				InsecureSkipVerify: false,
			},
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
	}

	// Try to load config file
	if data, err := os.ReadFile(configPath); err == nil {
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	} else {
		log.Printf("Config file not found at %s, using defaults", configPath)
	}

	return config, nil
}

func initLogger(config LoggingConfig) (*zap.Logger, error) {
	var zapConfig zap.Config

	if config.Level == "debug" {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	// Set log level
	level, err := zap.ParseAtomicLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	zapConfig.Level = level

	// Set encoding
	if config.Format == "console" {
		zapConfig.Encoding = "console"
	}

	return zapConfig.Build()
}

func setupRoutes(indexHandler *handlers.IndexHandler, documentHandler *handlers.DocumentHandler, logger *zap.Logger) *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Add request ID middleware
	router.Use(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("index-explorer-%d", time.Now().UnixNano())
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	})

	// CORS middleware for development
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":     "healthy",
			"service":    "index-explorer",
			"version":    "1.0.0",
			"focus":      "write-optimized Elasticsearch operations",
			"request_id": c.GetString("request_id"),
			"timestamp":  time.Now(),
		})
	})

	// Landing page - redirect to dashboard
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/dashboard")
	})

	// Serve monitoring dashboard
	router.Static("/static", "./web")
	router.GET("/dashboard", func(c *gin.Context) {
		c.File("./web/dashboard.html")
	})

	// API info endpoint
	router.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":     "Elasticsearch Index & Document Explorer",
			"version":     "1.0.0",
			"description": "Write-optimized Elasticsearch index and document management",
			"features": []string{
				"Write-optimized index creation",
				"High-performance bulk operations",
				"Adaptive batch processing",
				"Write performance monitoring",
				"Index optimization recommendations",
				"Large text corpus handling",
			},
			"endpoints": gin.H{
				"indices":   "/api/v1/indices",
				"documents": "/api/v1/indices/{index}/documents",
				"bulk":      "/api/v1/indices/{index}/bulk",
				"health":    "/health",
				"dashboard": "/dashboard",
			},
			"request_id": c.GetString("request_id"),
			"timestamp":  time.Now(),
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Index management routes
		indices := v1.Group("/indices")
		{
			// Core index operations
			indices.POST("/", indexHandler.CreateIndex)
			indices.GET("/", indexHandler.ListIndices)
			indices.GET("/:index", indexHandler.GetIndex)
			indices.DELETE("/:index", indexHandler.DeleteIndex)

			// Write-optimized index creation
			indices.POST("/write-optimized", indexHandler.CreateWriteOptimizedIndex)

			// Index optimization and tuning
			indices.POST("/:index/optimize", indexHandler.OptimizeIndex)
			indices.GET("/:index/recommendations", indexHandler.GetIndexRecommendations)
			indices.POST("/:index/tune/write-heavy", indexHandler.TuneIndexForWriteWorkload)

			// Performance analysis
			indices.GET("/:index/performance/write", indexHandler.GetIndexWritePerformance)
			indices.GET("/:index/analyze/write-performance", indexHandler.AnalyzeIndexWritePerformance)

			// Document operations within index context
			indices.POST("/:index/documents", documentHandler.IndexDocument)
			indices.GET("/:index/documents/:id", documentHandler.GetDocument)
			indices.PUT("/:index/documents/:id", documentHandler.UpdateDocument)
			indices.DELETE("/:index/documents/:id", documentHandler.DeleteDocument)

			// Bulk operations (the primary focus)
			indices.POST("/:index/bulk", documentHandler.BulkIndex)
			indices.POST("/:index/import/ndjson", documentHandler.BulkImportNDJSON)

			// Write performance metrics
			indices.GET("/:index/metrics/write-performance", documentHandler.GetWritePerformanceMetrics)
		}

		// Global bulk operations
		bulk := v1.Group("/bulk")
		{
			bulk.POST("/adaptive", documentHandler.AdaptiveBulkIndex)
			bulk.GET("/status", documentHandler.GetBulkOperationStatus)
		}

		// Metrics and monitoring
		metrics := v1.Group("/metrics")
		{
			metrics.GET("/write-performance", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "Global write performance metrics endpoint",
					"note":    "Use /api/v1/indices/{index}/metrics/write-performance for index-specific metrics",
					"request_id": c.GetString("request_id"),
					"timestamp":  time.Now(),
				})
			})
		}
	}

	// Development and debugging routes
	debug := router.Group("/debug")
	{
		debug.GET("/config", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"server_port": 8082,
				"focus":       "write-optimized operations",
				"features": []string{
					"bulk-first approach",
					"adaptive batch sizing",
					"write performance monitoring",
					"index optimization",
				},
				"request_id": c.GetString("request_id"),
				"timestamp":  time.Now(),
			})
		})

		debug.GET("/examples", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"write_optimized_index": gin.H{
					"url":    "POST /api/v1/indices/write-optimized",
					"example": gin.H{
						"index_name":        "my-text-corpus",
						"expected_volume":   "high",
						"expected_doc_size": "large",
						"ingestion_rate":    "high",
						"text_heavy":        true,
					},
				},
				"bulk_import": gin.H{
					"url":     "POST /api/v1/indices/{index}/import/ndjson",
					"params":  "?batch_size=1000&workers=8",
					"example": "Send NDJSON data in request body",
				},
				"adaptive_bulk": gin.H{
					"url":    "POST /api/v1/bulk/adaptive",
					"example": gin.H{
						"index_name":         "my-index",
						"documents":          "[]",
						"auto_batch_size":    true,
						"target_throughput":  "max",
						"error_tolerance":    "medium",
					},
				},
				"request_id": c.GetString("request_id"),
				"timestamp":  time.Now(),
			})
		})
	}

	return router
}