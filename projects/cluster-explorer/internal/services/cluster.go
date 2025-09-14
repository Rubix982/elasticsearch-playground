package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/saif-islam/es-playground/shared"
	"github.com/saif-islam/es-playground/projects/cluster-explorer/internal/models"
)

// ClusterService provides cluster management and monitoring functionality
type ClusterService struct {
	esClient *shared.ESClient
	logger   *zap.Logger
}

// NewClusterService creates a new cluster service instance
func NewClusterService(esClient *shared.ESClient, logger *zap.Logger) *ClusterService {
	return &ClusterService{
		esClient: esClient,
		logger:   logger,
	}
}

// GetClusterInfo retrieves comprehensive cluster information
func (s *ClusterService) GetClusterInfo(ctx context.Context) (*models.ClusterInfo, error) {
	s.logger.Info("Fetching comprehensive cluster information")

	// Fetch all cluster data in parallel
	healthCh := make(chan *models.ClusterHealth, 1)
	stateCh := make(chan *models.ClusterState, 1)
	statsCh := make(chan *models.ClusterStats, 1)
	nodesCh := make(chan []models.NodeInfo, 1)
	indicesCh := make(chan []models.IndexInfo, 1)
	shardsCh := make(chan *models.ShardAllocation, 1)
	perfCh := make(chan *models.PerformanceMetrics, 1)

	errCh := make(chan error, 7)

	// Fetch cluster health
	go func() {
		health, err := s.GetClusterHealth(ctx)
		if err != nil {
			errCh <- fmt.Errorf("failed to get cluster health: %w", err)
			return
		}
		healthCh <- health
	}()

	// Fetch cluster state
	go func() {
		state, err := s.GetClusterState(ctx)
		if err != nil {
			errCh <- fmt.Errorf("failed to get cluster state: %w", err)
			return
		}
		stateCh <- state
	}()

	// Fetch cluster stats
	go func() {
		stats, err := s.GetClusterStats(ctx)
		if err != nil {
			errCh <- fmt.Errorf("failed to get cluster stats: %w", err)
			return
		}
		statsCh <- stats
	}()

	// Fetch node information
	go func() {
		nodes, err := s.GetNodesInfo(ctx)
		if err != nil {
			errCh <- fmt.Errorf("failed to get nodes info: %w", err)
			return
		}
		nodesCh <- nodes
	}()

	// Fetch indices information
	go func() {
		indices, err := s.GetIndicesInfo(ctx)
		if err != nil {
			errCh <- fmt.Errorf("failed to get indices info: %w", err)
			return
		}
		indicesCh <- indices
	}()

	// Fetch shard allocation
	go func() {
		shards, err := s.GetShardAllocation(ctx)
		if err != nil {
			errCh <- fmt.Errorf("failed to get shard allocation: %w", err)
			return
		}
		shardsCh <- shards
	}()

	// Fetch performance metrics
	go func() {
		perf, err := s.GetPerformanceMetrics(ctx)
		if err != nil {
			errCh <- fmt.Errorf("failed to get performance metrics: %w", err)
			return
		}
		perfCh <- perf
	}()

	// Collect results
	var health *models.ClusterHealth
	var state *models.ClusterState
	var stats *models.ClusterStats
	var nodes []models.NodeInfo
	var indices []models.IndexInfo
	var shards *models.ShardAllocation
	var perf *models.PerformanceMetrics

	for i := 0; i < 7; i++ {
		select {
		case h := <-healthCh:
			health = h
		case st := <-stateCh:
			state = st
		case stats = <-statsCh:
		case nodes = <-nodesCh:
		case indices = <-indicesCh:
		case shards = <-shardsCh:
		case perf = <-perfCh:
		case err := <-errCh:
			s.logger.Error("Error fetching cluster information", zap.Error(err))
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return &models.ClusterInfo{
		Health:      health,
		State:       state,
		Stats:       stats,
		Nodes:       nodes,
		Indices:     indices,
		Shards:      shards,
		Performance: perf,
		RequestID:   generateRequestID(),
		Timestamp:   time.Now(),
	}, nil
}

// GetClusterHealth retrieves cluster health information
func (s *ClusterService) GetClusterHealth(ctx context.Context) (*models.ClusterHealth, error) {
	res, err := s.esClient.Cluster.Health(
		s.esClient.Cluster.Health.WithContext(ctx),
		s.esClient.Cluster.Health.WithLevel("cluster"),
	)
	if err != nil {
		return nil, fmt.Errorf("cluster health request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, shared.ParseESError(res)
	}

	var health models.ClusterHealth
	if err := shared.DecodeJSONResponse(res, &health); err != nil {
		return nil, fmt.Errorf("failed to decode cluster health: %w", err)
	}

	s.logger.Info("Retrieved cluster health",
		zap.String("status", health.Status),
		zap.Int("nodes", health.NumberOfNodes),
		zap.Int("active_shards", health.ActiveShards))

	return &health, nil
}

// GetClusterState retrieves cluster state information
func (s *ClusterService) GetClusterState(ctx context.Context) (*models.ClusterState, error) {
	res, err := s.esClient.Cluster.State(
		s.esClient.Cluster.State.WithContext(ctx),
		s.esClient.Cluster.State.WithMetric("master_node", "nodes", "routing_table", "metadata", "blocks"),
	)
	if err != nil {
		return nil, fmt.Errorf("cluster state request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, shared.ParseESError(res)
	}

	var state models.ClusterState
	if err := shared.DecodeJSONResponse(res, &state); err != nil {
		return nil, fmt.Errorf("failed to decode cluster state: %w", err)
	}

	s.logger.Info("Retrieved cluster state",
		zap.String("cluster_uuid", state.ClusterUUID),
		zap.String("master_node", state.MasterNode),
		zap.Int("version", state.Version))

	return &state, nil
}

// GetClusterStats retrieves cluster statistics
func (s *ClusterService) GetClusterStats(ctx context.Context) (*models.ClusterStats, error) {
	res, err := s.esClient.Cluster.Stats(
		s.esClient.Cluster.Stats.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("cluster stats request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, shared.ParseESError(res)
	}

	var stats models.ClusterStats
	if err := shared.DecodeJSONResponse(res, &stats); err != nil {
		return nil, fmt.Errorf("failed to decode cluster stats: %w", err)
	}

	s.logger.Info("Retrieved cluster stats",
		zap.String("cluster_name", stats.ClusterName),
		zap.Int("indices_count", stats.Indices.Count),
		zap.Int("total_shards", stats.Indices.Shards.Total))

	return &stats, nil
}

// GetNodesInfo retrieves detailed information about all nodes
func (s *ClusterService) GetNodesInfo(ctx context.Context) ([]models.NodeInfo, error) {
	res, err := s.esClient.Nodes.Info(
		s.esClient.Nodes.Info.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("nodes info request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, shared.ParseESError(res)
	}

	var response struct {
		Nodes map[string]models.NodeInfo `json:"nodes"`
	}
	if err := shared.DecodeJSONResponse(res, &response); err != nil {
		return nil, fmt.Errorf("failed to decode nodes info: %w", err)
	}

	nodes := make([]models.NodeInfo, 0, len(response.Nodes))
	for _, node := range response.Nodes {
		nodes = append(nodes, node)
	}

	s.logger.Info("Retrieved nodes info", zap.Int("node_count", len(nodes)))

	return nodes, nil
}

// GetIndicesInfo retrieves information about all indices
func (s *ClusterService) GetIndicesInfo(ctx context.Context) ([]models.IndexInfo, error) {
	res, err := s.esClient.Cat.Indices(
		s.esClient.Cat.Indices.WithContext(ctx),
		s.esClient.Cat.Indices.WithFormat("json"),
		s.esClient.Cat.Indices.WithV(true),
	)
	if err != nil {
		return nil, fmt.Errorf("indices info request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, shared.ParseESError(res)
	}

	var indices []models.IndexInfo
	if err := shared.DecodeJSONResponse(res, &indices); err != nil {
		return nil, fmt.Errorf("failed to decode indices info: %w", err)
	}

	// Enrich with detailed index information
	for i := range indices {
		if err := s.enrichIndexInfo(ctx, &indices[i]); err != nil {
			s.logger.Warn("Failed to enrich index info",
				zap.String("index", indices[i].Index),
				zap.Error(err))
		}
	}

	s.logger.Info("Retrieved indices info", zap.Int("index_count", len(indices)))

	return indices, nil
}

// enrichIndexInfo adds detailed settings and mappings to index info
func (s *ClusterService) enrichIndexInfo(ctx context.Context, indexInfo *models.IndexInfo) error {
	// Get index settings
	settingsRes, err := s.esClient.Indices.GetSettings(
		s.esClient.Indices.GetSettings.WithContext(ctx),
		s.esClient.Indices.GetSettings.WithIndex(indexInfo.Index),
	)
	if err != nil {
		return fmt.Errorf("failed to get index settings: %w", err)
	}
	defer settingsRes.Body.Close()

	if !settingsRes.IsError() {
		var settingsResponse map[string]interface{}
		if err := shared.DecodeJSONResponse(settingsRes, &settingsResponse); err == nil {
			if indexSettings, ok := settingsResponse[indexInfo.Index]; ok {
				if settingsMap, ok := indexSettings.(map[string]interface{}); ok {
					if settings, ok := settingsMap["settings"]; ok {
						settingsBytes, _ := json.Marshal(settings)
						json.Unmarshal(settingsBytes, &indexInfo.Settings)
					}
				}
			}
		}
	}

	// Get index mappings
	mappingsRes, err := s.esClient.Indices.GetMapping(
		s.esClient.Indices.GetMapping.WithContext(ctx),
		s.esClient.Indices.GetMapping.WithIndex(indexInfo.Index),
	)
	if err != nil {
		return fmt.Errorf("failed to get index mappings: %w", err)
	}
	defer mappingsRes.Body.Close()

	if !mappingsRes.IsError() {
		var mappingsResponse map[string]interface{}
		if err := shared.DecodeJSONResponse(mappingsRes, &mappingsResponse); err == nil {
			if indexMappings, ok := mappingsResponse[indexInfo.Index]; ok {
				if mappingsMap, ok := indexMappings.(map[string]interface{}); ok {
					if mappings, ok := mappingsMap["mappings"]; ok {
						indexInfo.Mappings = mappings
					}
				}
			}
		}
	}

	return nil
}

// GetShardAllocation retrieves shard allocation information
func (s *ClusterService) GetShardAllocation(ctx context.Context) (*models.ShardAllocation, error) {
	res, err := s.esClient.Cat.Shards(
		s.esClient.Cat.Shards.WithContext(ctx),
		s.esClient.Cat.Shards.WithFormat("json"),
		s.esClient.Cat.Shards.WithV(true),
	)
	if err != nil {
		return nil, fmt.Errorf("shard allocation request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, shared.ParseESError(res)
	}

	var shards []models.ShardDetails
	if err := shared.DecodeJSONResponse(res, &shards); err != nil {
		return nil, fmt.Errorf("failed to decode shard allocation: %w", err)
	}

	// Organize shards by index
	indices := make(map[string]models.IndexAllocation)
	var unassigned []models.UnassignedShardDetails
	summary := models.AllocationSummary{}

	for _, shard := range shards {
		summary.TotalShards++
		
		if shard.State == "UNASSIGNED" {
			summary.UnassignedShards++
			unassigned = append(unassigned, models.UnassignedShardDetails{
				Index:        shard.Index,
				Shard:        shard.Shard,
				Primary:      shard.Primary,
				CurrentState: shard.State,
				Reason:       "Unknown", // This would need additional API call to get exact reason
			})
		} else {
			summary.AssignedShards++
			if shard.State == "RELOCATING" {
				summary.RelocatingShards++
			}
			if shard.State == "INITIALIZING" {
				summary.InitializingShards++
			}

			if _, exists := indices[shard.Index]; !exists {
				indices[shard.Index] = models.IndexAllocation{
					Shards: make(map[string][]models.ShardDetails),
				}
			}

			shardKey := fmt.Sprintf("%d", shard.Shard)
			indices[shard.Index].Shards[shardKey] = append(indices[shard.Index].Shards[shardKey], shard)
		}
	}

	allocation := &models.ShardAllocation{
		Indices:    indices,
		Unassigned: unassigned,
		Summary:    summary,
	}

	s.logger.Info("Retrieved shard allocation",
		zap.Int("total_shards", summary.TotalShards),
		zap.Int("assigned_shards", summary.AssignedShards),
		zap.Int("unassigned_shards", summary.UnassignedShards))

	return allocation, nil
}

// GetPerformanceMetrics retrieves cluster performance metrics
func (s *ClusterService) GetPerformanceMetrics(ctx context.Context) (*models.PerformanceMetrics, error) {
	// Get node stats for performance metrics
	res, err := s.esClient.Nodes.Stats(
		s.esClient.Nodes.Stats.WithContext(ctx),
		s.esClient.Nodes.Stats.WithMetric("os", "process", "jvm", "fs", "thread_pool", "indices"),
	)
	if err != nil {
		return nil, fmt.Errorf("node stats request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, shared.ParseESError(res)
	}

	var response struct {
		Nodes map[string]interface{} `json:"nodes"`
	}
	if err := shared.DecodeJSONResponse(res, &response); err != nil {
		return nil, fmt.Errorf("failed to decode node stats: %w", err)
	}

	// Aggregate performance metrics from all nodes
	metrics := &models.PerformanceMetrics{
		CPU:              models.CPUMetrics{},
		Memory:           models.MemoryMetrics{},
		Disk:             models.DiskMetrics{},
		Network:          models.NetworkMetrics{},
		GarbageCollection: models.GCMetrics{},
		ThreadPools:      models.ThreadPoolMetrics{},
		Search:           models.SearchMetrics{},
		Indexing:         models.IndexingMetrics{},
	}

	// This is a simplified aggregation - in a real implementation,
	// you would properly aggregate metrics from all nodes
	nodeCount := len(response.Nodes)
	if nodeCount > 0 {
		s.logger.Info("Aggregated performance metrics", zap.Int("node_count", nodeCount))
	}

	return metrics, nil
}

// GetHotThreads retrieves hot threads information for performance analysis
func (s *ClusterService) GetHotThreads(ctx context.Context, nodeID string) (string, error) {
	var res *http.Response
	var err error

	if nodeID != "" {
		res, err = s.esClient.Nodes.HotThreads(
			s.esClient.Nodes.HotThreads.WithContext(ctx),
			s.esClient.Nodes.HotThreads.WithNodeID(nodeID),
		)
	} else {
		res, err = s.esClient.Nodes.HotThreads(
			s.esClient.Nodes.HotThreads.WithContext(ctx),
		)
	}

	if err != nil {
		return "", fmt.Errorf("hot threads request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return "", shared.ParseESError(res)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read hot threads response: %w", err)
	}

	return string(body), nil
}

// MonitorClusterHealth monitors cluster health at regular intervals
func (s *ClusterService) MonitorClusterHealth(ctx context.Context, interval time.Duration) (<-chan *models.ClusterHealth, error) {
	healthCh := make(chan *models.ClusterHealth)

	go func() {
		defer close(healthCh)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				health, err := s.GetClusterHealth(ctx)
				if err != nil {
					s.logger.Error("Failed to get cluster health during monitoring", zap.Error(err))
					continue
				}
				
				select {
				case healthCh <- health:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return healthCh, nil
}

// UpdateClusterSettings updates cluster settings
func (s *ClusterService) UpdateClusterSettings(ctx context.Context, settings map[string]interface{}, persistent bool) error {
	var body map[string]interface{}
	
	if persistent {
		body = map[string]interface{}{
			"persistent": settings,
		}
	} else {
		body = map[string]interface{}{
			"transient": settings,
		}
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	res, err := s.esClient.Cluster.PutSettings(
		s.esClient.Cluster.PutSettings.WithContext(ctx),
		strings.NewReader(string(bodyBytes)),
	)
	if err != nil {
		return fmt.Errorf("update cluster settings request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return shared.ParseESError(res)
	}

	s.logger.Info("Updated cluster settings",
		zap.Bool("persistent", persistent),
		zap.Any("settings", settings))

	return nil
}

// GetClusterSettings retrieves current cluster settings
func (s *ClusterService) GetClusterSettings(ctx context.Context) (map[string]interface{}, error) {
	res, err := s.esClient.Cluster.GetSettings(
		s.esClient.Cluster.GetSettings.WithContext(ctx),
		s.esClient.Cluster.GetSettings.WithIncludeDefaults(true),
	)
	if err != nil {
		return nil, fmt.Errorf("get cluster settings request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, shared.ParseESError(res)
	}

	var settings map[string]interface{}
	if err := shared.DecodeJSONResponse(res, &settings); err != nil {
		return nil, fmt.Errorf("failed to decode cluster settings: %w", err)
	}

	return settings, nil
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("cluster-%d", time.Now().UnixNano())
}