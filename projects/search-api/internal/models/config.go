package models

import (
	"time"
	
	"github.com/saif-islam/es-playground/projects/search-api/internal/tracing"
)

// Config represents the application configuration
type Config struct {
	Server        ServerConfig        `yaml:"server"`
	Elasticsearch ElasticsearchConfig `yaml:"elasticsearch"`
	Redis         RedisConfig         `yaml:"redis"`
	Logging       LoggingConfig       `yaml:"logging"`
	Search        SearchConfig        `yaml:"search"`
	Cache         CacheConfig         `yaml:"cache"`
	Tracing       tracing.TracingConfig `yaml:"tracing"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port            int           `yaml:"port"`
	Host            string        `yaml:"host"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

// ElasticsearchConfig holds Elasticsearch connection settings
type ElasticsearchConfig struct {
	URLs      []string  `yaml:"urls"`
	Username  string    `yaml:"username"`
	Password  string    `yaml:"password"`
	APIKey    string    `yaml:"api_key"`
	TLSConfig TLSConfig `yaml:"tls"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	InsecureSkipVerify bool `yaml:"insecure_skip_verify"`
}

// RedisConfig holds Redis connection settings
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

// SearchConfig holds search-specific configuration
type SearchConfig struct {
	DefaultSize int               `yaml:"default_size"`
	MaxSize     int               `yaml:"max_size"`
	Timeout     time.Duration     `yaml:"timeout"`
	Indices     map[string]string `yaml:"indices"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Enabled         bool          `yaml:"enabled"`
	TTL             time.Duration `yaml:"ttl"`
	Prefix          string        `yaml:"prefix"`
	MaxKeyLength    int           `yaml:"max_key_length"`
	MaxValueSize    int           `yaml:"max_value_size"`
	CompressionEnabled bool       `yaml:"compression_enabled"`
	
	// Smart caching features
	AdaptiveTTL     bool          `yaml:"adaptive_ttl"`
	PopularityBoost bool          `yaml:"popularity_boost"`
	PreemptiveRefresh bool        `yaml:"preemptive_refresh"`
	
	// Performance settings
	Pipeline        bool          `yaml:"pipeline"`
	MaxConnections  int           `yaml:"max_connections"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
}

