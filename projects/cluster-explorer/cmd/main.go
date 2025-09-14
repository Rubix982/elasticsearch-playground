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

	"github.com/saif-islam/es-playground/projects/cluster-explorer/internal/handlers"
	"github.com/saif-islam/es-playground/projects/cluster-explorer/internal/models"
	"github.com/saif-islam/es-playground/projects/cluster-explorer/internal/services"
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

	logger.Info("Starting Cluster Explorer",
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
	clusterService := services.NewClusterService(esClient, logger)

	// Initialize handlers
	clusterHandler := handlers.NewClusterHandler(clusterService, logger)

	// Setup HTTP server
	if config.Logging.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := setupRoutes(clusterHandler, logger)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port),
		Handler:      router,
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Cluster Explorer server starting",
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

	logger.Info("Shutting down Cluster Explorer...")

	// Graceful shutdown
	ctx, cancel = context.WithTimeout(context.Background(), config.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Cluster Explorer exited")
}

func loadConfig() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	// Make path relative to project root
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join("projects/cluster-explorer", configPath)
	}

	// Default configuration
	config := &Config{
		Server: ServerConfig{
			Port:            8081,
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

func setupRoutes(clusterHandler *handlers.ClusterHandler, logger *zap.Logger) *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Add request ID middleware
	router.Use(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("cluster-%d", time.Now().UnixNano())
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	})

	// Serve static files (for the web UI)
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")

	// Web UI routes
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Elasticsearch Cluster Explorer",
		})
	})

	router.GET("/dashboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title": "Cluster Dashboard",
		})
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":     "healthy",
			"service":    "cluster-explorer",
			"version":    "1.0.0",
			"request_id": c.GetString("request_id"),
			"timestamp":  time.Now(),
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		cluster := v1.Group("/cluster")
		{
			// Comprehensive cluster information
			cluster.GET("/info", clusterHandler.GetClusterInfo)
			cluster.GET("/overview", clusterHandler.GetClusterOverview)

			// Individual cluster components
			cluster.GET("/health", clusterHandler.GetClusterHealth)
			cluster.GET("/state", clusterHandler.GetClusterState)
			cluster.GET("/stats", clusterHandler.GetClusterStats)

			// Node management
			cluster.GET("/nodes", clusterHandler.GetNodes)
			cluster.GET("/nodes/:nodeId/hot-threads", clusterHandler.GetHotThreads)

			// Index management
			cluster.GET("/indices", clusterHandler.GetIndices)

			// Shard management
			cluster.GET("/shards", clusterHandler.GetShardAllocation)

			// Performance monitoring
			cluster.GET("/performance", clusterHandler.GetPerformanceMetrics)

			// Real-time monitoring
			cluster.GET("/monitor/health", clusterHandler.MonitorHealth)

			// Settings management
			cluster.GET("/settings", clusterHandler.GetClusterSettings)
			cluster.PUT("/settings", clusterHandler.UpdateClusterSettings)
		}
	}

	// Documentation routes
	docs := router.Group("/docs")
	{
		docs.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "docs.html", gin.H{
				"title": "API Documentation",
			})
		})

		docs.GET("/examples", func(c *gin.Context) {
			c.HTML(http.StatusOK, "examples.html", gin.H{
				"title": "Usage Examples",
			})
		})
	}

	return router
}