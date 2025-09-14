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

	"github.com/redis/go-redis/v9"

	"github.com/saif-islam/es-playground/projects/search-api/internal/abtesting"
	"github.com/saif-islam/es-playground/projects/search-api/internal/cache"
	"github.com/saif-islam/es-playground/projects/search-api/internal/handlers"
	"github.com/saif-islam/es-playground/projects/search-api/internal/metrics"
	"github.com/saif-islam/es-playground/projects/search-api/internal/middleware"
	"github.com/saif-islam/es-playground/projects/search-api/internal/models"
	"github.com/saif-islam/es-playground/projects/search-api/internal/realtime"
	"github.com/saif-islam/es-playground/projects/search-api/internal/services"
	"github.com/saif-islam/es-playground/projects/search-api/internal/tracing"
	"github.com/saif-islam/es-playground/shared"
)

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

	logger.Info("Starting Search API server",
		zap.String("version", "1.0.0"),
		zap.Int("port", config.Server.Port))

	// Initialize metrics
	metrics.SetApplicationInfo("1.0.0", "search-api", "development")

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

	// Initialize real-time analytics hub
	analyticsHub := realtime.NewAnalyticsHub(logger)

	// Initialize tracing
	tracingProvider, err := tracing.NewTracingProvider(config.Tracing, logger)
	if err != nil {
		logger.Fatal("Failed to initialize tracing", zap.Error(err))
	}
	defer func() {
		if err := tracingProvider.Shutdown(context.Background()); err != nil {
			logger.Error("Failed to shutdown tracing", zap.Error(err))
		}
	}()

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:         config.Redis.Addr,
		Password:     config.Redis.Password,
		DB:           config.Redis.DB,
		PoolSize:     config.Redis.PoolSize,
		ReadTimeout:  config.Cache.ReadTimeout,
		WriteTimeout: config.Cache.WriteTimeout,
	})

	// Test Redis connection
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Warn("Redis connection failed, caching will be disabled", zap.Error(err))
		config.Cache.Enabled = false
	}

	// Initialize cache
	redisCache := cache.NewRedisCache(redisClient, config.Cache, logger)
	cacheManager := cache.NewCacheManager(redisCache, logger)

	// Initialize search operation tracer
	searchTracer := tracing.NewSearchOperationTracer(tracingProvider)

	// Initialize A/B testing framework
	abTestFramework := abtesting.NewABTestFramework(logger)

	// Initialize services
	searchService := services.NewSearchService(esClient, logger, analyticsHub, searchTracer, cacheManager)

	// Initialize handlers
	searchHandler := handlers.NewSearchHandler(searchService, logger)
	experimentHandler := handlers.NewExperimentHandler(abTestFramework, logger)

	// Setup HTTP server
	if config.Logging.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := setupRoutes(searchHandler, experimentHandler, analyticsHub, abTestFramework, tracingProvider, logger)
	
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port),
		Handler:      router,
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server starting", 
			zap.String("address", server.Addr))
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel = context.WithTimeout(context.Background(), config.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

func loadConfig() (*models.Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	// Make path relative to project root
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join("projects/search-api", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config models.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func initLogger(config models.LoggingConfig) (*zap.Logger, error) {
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

func setupRoutes(searchHandler *handlers.SearchHandler, experimentHandler *handlers.ExperimentHandler, analyticsHub *realtime.AnalyticsHub, abTestFramework *abtesting.ABTestFramework, tracingProvider *tracing.TracingProvider, logger *zap.Logger) *gin.Engine {
	router := gin.New()
	
	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.PrometheusMiddleware())
	router.Use(tracing.TracingMiddleware(tracingProvider, logger))
	router.Use(middleware.ABTestingMiddleware(abTestFramework, logger))
	router.Use(tracing.SearchTracingMiddleware(tracingProvider))
	
	// Add request ID middleware
	router.Use(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("%d", time.Now().UnixNano())
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "search-api",
			"version":   "1.0.0",
			"timestamp": time.Now(),
		})
	})

	// Metrics endpoint
	router.GET("/metrics", middleware.PrometheusHandler())

	// Real-time analytics WebSocket endpoint
	router.GET("/ws/analytics", analyticsHub.HandleWebSocket)

	// Analytics dashboard endpoint
	router.GET("/analytics", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"connected_clients": analyticsHub.GetConnectedClients(),
			"status":           "active",
			"websocket_url":    "/ws/analytics",
		})
	})

	// Tracing dashboard endpoint
	router.GET("/tracing", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.File("web/tracing-dashboard.html")
	})

	// API routes
	api := router.Group("/api")
	{
		// Add experiment tracing middleware for experiment routes
		experiments := api.Group("/experiments")
		experiments.Use(tracing.ExperimentTracingMiddleware(tracingProvider))
		experimentHandler.RegisterRoutes(experiments)
		
		// Search routes with search-specific tracing
		searchHandler.RegisterRoutes(api)
	}

	return router
}