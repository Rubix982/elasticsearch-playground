package shared

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"go.uber.org/zap"
)

// ESConfig holds configuration for Elasticsearch client
type ESConfig struct {
	URLs      []string `yaml:"urls"`
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
	APIKey    string   `yaml:"api_key"`
	TLSConfig *TLSConfig `yaml:"tls"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	InsecureSkipVerify bool `yaml:"insecure_skip_verify"`
}

// ESClient wraps the Elasticsearch client with additional functionality
type ESClient struct {
	*elasticsearch.Client
	logger *zap.Logger
	config *ESConfig
}

// NewESClient creates a new Elasticsearch client with the given configuration
func NewESClient(config *ESConfig, logger *zap.Logger) (*ESClient, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	esConfig := elasticsearch.Config{
		Addresses: config.URLs,
		Username:  config.Username,
		Password:  config.Password,
		APIKey:    config.APIKey,
	}

	// Configure TLS if specified
	if config.TLSConfig != nil {
		esConfig.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.TLSConfig.InsecureSkipVerify,
			},
		}
	}

	client, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	esClient := &ESClient{
		Client: client,
		logger: logger,
		config: config,
	}

	// Test connection
	if err := esClient.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping Elasticsearch: %w", err)
	}

	logger.Info("Successfully connected to Elasticsearch", 
		zap.Strings("urls", config.URLs))

	return esClient, nil
}

// Ping tests the connection to Elasticsearch
func (c *ESClient) Ping(ctx context.Context) error {
	res, err := c.Client.Ping(
		c.Client.Ping.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("ping request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ping failed with status: %s", res.Status())
	}

	return nil
}

// Health checks the cluster health
func (c *ESClient) Health(ctx context.Context) (*ClusterHealth, error) {
	res, err := c.Client.Cluster.Health(
		c.Client.Cluster.Health.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("health request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("health check failed with status: %s", res.Status())
	}

	var health ClusterHealth
	if err := DecodeJSONResponse(res, &health); err != nil {
		return nil, fmt.Errorf("failed to decode health response: %w", err)
	}

	return &health, nil
}

// WaitForCluster waits for the cluster to be in the specified state
func (c *ESClient) WaitForCluster(ctx context.Context, status string, timeout time.Duration) error {
	c.logger.Info("Waiting for cluster status", 
		zap.String("status", status),
		zap.Duration("timeout", timeout))

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	res, err := c.Client.Cluster.Health(
		c.Client.Cluster.Health.WithContext(ctx),
		c.Client.Cluster.Health.WithWaitForStatus(status),
		c.Client.Cluster.Health.WithTimeout(timeout),
	)
	if err != nil {
		return fmt.Errorf("wait for cluster failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("cluster not ready, status: %s", res.Status())
	}

	c.logger.Info("Cluster is ready", zap.String("status", status))
	return nil
}

// ClusterHealth represents cluster health information
type ClusterHealth struct {
	ClusterName         string `json:"cluster_name"`
	Status              string `json:"status"`
	TimedOut            bool   `json:"timed_out"`
	NumberOfNodes       int    `json:"number_of_nodes"`
	NumberOfDataNodes   int    `json:"number_of_data_nodes"`
	ActivePrimaryShards int    `json:"active_primary_shards"`
	ActiveShards        int    `json:"active_shards"`
	RelocatingShards    int    `json:"relocating_shards"`
	InitializingShards  int    `json:"initializing_shards"`
	UnassignedShards    int    `json:"unassigned_shards"`
}

// DefaultESConfig returns a default Elasticsearch configuration
func DefaultESConfig() *ESConfig {
	return &ESConfig{
		URLs:     []string{"http://localhost:9200"},
		Username: "",
		Password: "",
		TLSConfig: &TLSConfig{
			InsecureSkipVerify: false,
		},
	}
}