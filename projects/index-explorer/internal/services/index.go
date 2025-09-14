package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/saif-islam/es-playground/shared"
	"github.com/saif-islam/es-playground/projects/index-explorer/internal/models"
)

// IndexService provides write-optimized index management functionality
type IndexService struct {
	esClient *shared.ESClient
	logger   *zap.Logger
}

// NewIndexService creates a new index service instance
func NewIndexService(esClient *shared.ESClient, logger *zap.Logger) *IndexService {
	return &IndexService{
		esClient: esClient,
		logger:   logger,
	}
}

// CreateIndex creates a new index with write-optimized settings
func (s *IndexService) CreateIndex(ctx context.Context, req *models.IndexRequest) (*models.IndexResponse, error) {
	s.logger.Info("Creating write-optimized index",
		zap.String("index_name", req.IndexName),
		zap.Bool("write_optimized", req.WriteOptimized),
		zap.Bool("text_heavy", req.TextHeavy),
		zap.String("expected_volume", req.ExpectedVolume))

	// Build optimized settings based on request parameters
	settings := s.buildOptimizedSettings(req)
	
	// Prepare the index creation request
	indexBody := map[string]interface{}{}
	
	if settings != nil {
		indexBody["settings"] = settings
	}
	
	if req.Mappings != nil {
		indexBody["mappings"] = req.Mappings
	}
	
	if req.Aliases != nil {
		indexBody["aliases"] = req.Aliases
	}

	bodyBytes, err := json.Marshal(indexBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal index body: %w", err)
	}

	// Create the index
	res, err := s.esClient.Indices.Create(
		req.IndexName,
		s.esClient.Indices.Create.WithContext(ctx),
		s.esClient.Indices.Create.WithBody(strings.NewReader(string(bodyBytes))),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, shared.ParseESError(res)
	}

	var createResponse struct {
		Acknowledged bool `json:"acknowledged"`
		Index        string `json:"index"`
	}
	if err := shared.DecodeJSONResponse(res, &createResponse); err != nil {
		return nil, fmt.Errorf("failed to decode create response: %w", err)
	}

	// Get applied optimizations
	optimizations := s.getAppliedOptimizations(req)

	response := &models.IndexResponse{
		IndexName:     req.IndexName,
		Acknowledged:  createResponse.Acknowledged,
		Created:       true,
		Settings:      settings,
		Optimizations: optimizations,
		RequestID:     s.generateRequestID(),
		Timestamp:     time.Now(),
	}

	s.logger.Info("Successfully created write-optimized index",
		zap.String("index_name", req.IndexName),
		zap.Strings("optimizations", optimizations))

	return response, nil
}

// buildOptimizedSettings creates write-optimized settings based on request parameters
func (s *IndexService) buildOptimizedSettings(req *models.IndexRequest) *models.IndexSettings {
	settings := &models.IndexSettings{
		Additional: make(map[string]interface{}),
	}

	// Base settings from request
	if req.Settings != nil {
		*settings = *req.Settings
		if settings.Additional == nil {
			settings.Additional = make(map[string]interface{})
		}
	}

	// Apply write optimizations
	if req.WriteOptimized {
		s.applyWriteOptimizations(settings, req)
	}

	// Apply text-heavy optimizations
	if req.TextHeavy {
		s.applyTextOptimizations(settings, req)
	}

	// Apply volume-specific optimizations
	s.applyVolumeOptimizations(settings, req.ExpectedVolume)

	// Apply document size optimizations
	s.applyDocSizeOptimizations(settings, req.ExpectedDocSize)

	return settings
}

// applyWriteOptimizations applies settings for write-heavy workloads
func (s *IndexService) applyWriteOptimizations(settings *models.IndexSettings, req *models.IndexRequest) {
	// Optimize refresh interval for write performance
	if settings.RefreshInterval == "" {
		switch req.IngestionRate {
		case "high":
			settings.RefreshInterval = "30s" // Reduce refresh frequency for high write loads
		case "medium":
			settings.RefreshInterval = "5s"
		default:
			settings.RefreshInterval = "1s"
		}
	}

	// Optimize replica count for write performance (can be adjusted later)
	if settings.NumberOfReplicas == 0 && req.ExpectedVolume == "high" {
		settings.NumberOfReplicas = 0 // Start with 0 replicas for maximum write speed
	}

	// Translog optimizations for write performance
	if settings.TranslogFlushThresholdSize == "" {
		settings.TranslogFlushThresholdSize = "1gb" // Larger threshold for better write performance
	}
	
	if settings.TranslogSyncInterval == "" {
		settings.TranslogSyncInterval = "5s" // Less frequent sync for better performance
	}

	// For high-volume scenarios, optimize durability vs performance
	if req.ExpectedVolume == "high" && settings.TranslogDurability == "" {
		settings.TranslogDurability = "async" // Async for maximum write performance
	}

	// Index buffer size optimization
	if settings.IndexBufferSize == "" {
		switch req.ExpectedDocSize {
		case "large", "huge":
			settings.IndexBufferSize = "20%" // Larger buffer for large documents
		default:
			settings.IndexBufferSize = "10%" // Standard buffer size
		}
	}

	// Merge policy optimizations for write-heavy workloads
	if settings.MergePolicyMaxMergeSize == "" {
		settings.MergePolicyMaxMergeSize = "5gb" // Larger merges, less frequent
	}
	
	if settings.MergePolicySegmentsPerTier == 0 {
		settings.MergePolicySegmentsPerTier = 20 // More segments per tier for write performance
	}

	if settings.MergeSchedulerMaxThreadCount == 0 {
		settings.MergeSchedulerMaxThreadCount = 1 // Conservative merge threads to not interfere with writes
	}
}

// applyTextOptimizations applies settings optimized for text-heavy content
func (s *IndexService) applyTextOptimizations(settings *models.IndexSettings, req *models.IndexRequest) {
	// Use best compression for text-heavy indices
	if settings.Codec == "" {
		settings.Codec = "best_compression" // Better compression for text content
	}

	// Optimize for large text documents
	if req.ExpectedDocSize == "large" || req.ExpectedDocSize == "huge" {
		// Increase max merged segment size for large text documents
		if settings.MergePolicyMaxMergedSegmentMB == 0 {
			settings.MergePolicyMaxMergedSegmentMB = 10240 // 10GB segments for large text
		}
	}

	// Set source compression for text-heavy content
	settings.Additional["index.mapping.source.compress"] = true
	settings.Additional["index.mapping.source.compress_threshold"] = "1kb"
}

// applyVolumeOptimizations applies settings based on expected volume
func (s *IndexService) applyVolumeOptimizations(settings *models.IndexSettings, volume string) {
	switch volume {
	case "high":
		// High volume optimizations
		if settings.NumberOfShards == 0 {
			settings.NumberOfShards = 5 // More shards for parallel writes
		}
		
		// More aggressive merge settings for high volume
		if settings.MergePolicySegmentsPerTier == 0 {
			settings.MergePolicySegmentsPerTier = 30
		}
		
	case "medium":
		if settings.NumberOfShards == 0 {
			settings.NumberOfShards = 3 // Moderate shard count
		}
		
	default: // low volume
		if settings.NumberOfShards == 0 {
			settings.NumberOfShards = 1 // Single shard for low volume
		}
	}
}

// applyDocSizeOptimizations applies settings based on expected document size
func (s *IndexService) applyDocSizeOptimizations(settings *models.IndexSettings, docSize string) {
	switch docSize {
	case "huge": // > 100KB
		// Optimize for very large documents
		settings.Additional["index.mapping.total_fields.limit"] = 2000
		settings.Additional["index.mapping.depth.limit"] = 40
		settings.Additional["index.mapping.nested_fields.limit"] = 100
		
	case "large": // 10-100KB
		// Optimize for large documents
		settings.Additional["index.mapping.total_fields.limit"] = 1500
		settings.Additional["index.mapping.depth.limit"] = 30
		
	case "medium": // 1-10KB
		// Standard settings for medium documents
		settings.Additional["index.mapping.total_fields.limit"] = 1000
		settings.Additional["index.mapping.depth.limit"] = 20
		
	default: // small < 1KB
		// Optimize for small, numerous documents
		settings.Additional["index.mapping.total_fields.limit"] = 500
		settings.Additional["index.mapping.depth.limit"] = 10
	}
}

// getAppliedOptimizations returns a list of optimizations that were applied
func (s *IndexService) getAppliedOptimizations(req *models.IndexRequest) []string {
	var optimizations []string

	if req.WriteOptimized {
		optimizations = append(optimizations, 
			"write-optimized refresh interval",
			"translog write performance tuning",
			"merge policy optimization for writes",
			"index buffer size optimization")
	}

	if req.TextHeavy {
		optimizations = append(optimizations,
			"best compression codec for text",
			"source compression enabled",
			"large segment optimization for text")
	}

	if req.ExpectedVolume == "high" {
		optimizations = append(optimizations,
			"increased shard count for parallel writes",
			"async translog durability for performance")
	}

	if req.ExpectedDocSize == "large" || req.ExpectedDocSize == "huge" {
		optimizations = append(optimizations,
			"large document field limits",
			"increased mapping depth limits")
	}

	return optimizations
}

// GetIndexInfo retrieves comprehensive information about an index
func (s *IndexService) GetIndexInfo(ctx context.Context, indexName string) (*models.IndexInfo, error) {
	s.logger.Info("Getting index information", zap.String("index_name", indexName))

	// Get basic index information
	catRes, err := s.esClient.Cat.Indices(
		s.esClient.Cat.Indices.WithContext(ctx),
		s.esClient.Cat.Indices.WithIndex(indexName),
		s.esClient.Cat.Indices.WithFormat("json"),
		s.esClient.Cat.Indices.WithV(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get cat indices: %w", err)
	}
	defer catRes.Body.Close()

	if catRes.IsError() {
		return nil, shared.ParseESError(catRes)
	}

	var catIndices []models.IndexInfo
	if err := shared.DecodeJSONResponse(catRes, &catIndices); err != nil {
		return nil, fmt.Errorf("failed to decode cat indices: %w", err)
	}

	if len(catIndices) == 0 {
		return nil, fmt.Errorf("index %s not found", indexName)
	}

	indexInfo := &catIndices[0]

	// Enrich with detailed settings
	if err := s.enrichIndexSettings(ctx, indexInfo); err != nil {
		s.logger.Warn("Failed to enrich index settings", zap.Error(err))
	}

	// Enrich with mappings
	if err := s.enrichIndexMappings(ctx, indexInfo); err != nil {
		s.logger.Warn("Failed to enrich index mappings", zap.Error(err))
	}

	// Enrich with statistics
	if err := s.enrichIndexStats(ctx, indexInfo); err != nil {
		s.logger.Warn("Failed to enrich index stats", zap.Error(err))
	}

	// Calculate write metrics
	if err := s.enrichWriteMetrics(ctx, indexInfo); err != nil {
		s.logger.Warn("Failed to enrich write metrics", zap.Error(err))
	}

	indexInfo.RequestID = s.generateRequestID()
	indexInfo.Timestamp = time.Now()

	return indexInfo, nil
}

// enrichIndexSettings adds detailed settings to index info
func (s *IndexService) enrichIndexSettings(ctx context.Context, indexInfo *models.IndexInfo) error {
	res, err := s.esClient.Indices.GetSettings(
		s.esClient.Indices.GetSettings.WithContext(ctx),
		s.esClient.Indices.GetSettings.WithIndex(indexInfo.IndexName),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return shared.ParseESError(res)
	}

	var settingsResponse map[string]interface{}
	if err := shared.DecodeJSONResponse(res, &settingsResponse); err != nil {
		return err
	}

	if indexSettings, ok := settingsResponse[indexInfo.IndexName]; ok {
		if settingsMap, ok := indexSettings.(map[string]interface{}); ok {
			if settings, ok := settingsMap["settings"]; ok {
				settingsBytes, _ := json.Marshal(settings)
				json.Unmarshal(settingsBytes, &indexInfo.Settings)
			}
		}
	}

	return nil
}

// enrichIndexMappings adds mappings to index info
func (s *IndexService) enrichIndexMappings(ctx context.Context, indexInfo *models.IndexInfo) error {
	res, err := s.esClient.Indices.GetMapping(
		s.esClient.Indices.GetMapping.WithContext(ctx),
		s.esClient.Indices.GetMapping.WithIndex(indexInfo.IndexName),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return shared.ParseESError(res)
	}

	var mappingsResponse map[string]interface{}
	if err := shared.DecodeJSONResponse(res, &mappingsResponse); err != nil {
		return err
	}

	if indexMappings, ok := mappingsResponse[indexInfo.IndexName]; ok {
		if mappingsMap, ok := indexMappings.(map[string]interface{}); ok {
			if mappings, ok := mappingsMap["mappings"]; ok {
				indexInfo.Mappings = mappings
			}
		}
	}

	return nil
}

// enrichIndexStats adds statistics to index info
func (s *IndexService) enrichIndexStats(ctx context.Context, indexInfo *models.IndexInfo) error {
	res, err := s.esClient.Indices.Stats(
		s.esClient.Indices.Stats.WithContext(ctx),
		s.esClient.Indices.Stats.WithIndex(indexInfo.IndexName),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return shared.ParseESError(res)
	}

	var statsResponse map[string]interface{}
	if err := shared.DecodeJSONResponse(res, &statsResponse); err != nil {
		return err
	}

	if indices, ok := statsResponse["indices"].(map[string]interface{}); ok {
		if indexStats, ok := indices[indexInfo.IndexName]; ok {
			statsBytes, _ := json.Marshal(indexStats)
			json.Unmarshal(statsBytes, &indexInfo.Stats)
		}
	}

	return nil
}

// enrichWriteMetrics calculates write-specific performance metrics
func (s *IndexService) enrichWriteMetrics(ctx context.Context, indexInfo *models.IndexInfo) error {
	if indexInfo.Stats == nil || indexInfo.Stats.Total == nil {
		return nil
	}

	stats := indexInfo.Stats.Total
	writeMetrics := &models.WriteMetrics{}

	// Calculate indexing rate (docs per second)
	if stats.Indexing.IndexTimeInMillis > 0 {
		timeSeconds := float64(stats.Indexing.IndexTimeInMillis) / 1000.0
		writeMetrics.IndexingRate = float64(stats.Indexing.IndexTotal) / timeSeconds
	}

	// Calculate average document size
	if stats.Docs.Count > 0 && stats.Store.SizeInBytes > 0 {
		writeMetrics.AverageDocSize = stats.Store.SizeInBytes / stats.Docs.Count
	}

	// Calculate write latency
	if stats.Indexing.IndexTotal > 0 {
		writeMetrics.WriteLatency = float64(stats.Indexing.IndexTimeInMillis) / float64(stats.Indexing.IndexTotal)
	}

	// Get segment count
	writeMetrics.SegmentCount = stats.Segments.Count

	// Calculate merge rate
	if stats.Merges.TotalTimeInMillis > 0 {
		mergeTimeSeconds := float64(stats.Merges.TotalTimeInMillis) / 1000.0
		writeMetrics.MergeRate = float64(stats.Merges.Total) / mergeTimeSeconds
	}

	// Calculate refresh rate
	if stats.Refresh.TotalTimeInMillis > 0 {
		refreshTimeSeconds := float64(stats.Refresh.TotalTimeInMillis) / 1000.0
		writeMetrics.RefreshRate = float64(stats.Refresh.Total) / refreshTimeSeconds
	}

	// Get translog size
	writeMetrics.TranslogSize = stats.Translog.SizeInBytes

	// Calculate write load (current indexing operations / max capacity estimate)
	writeMetrics.WriteLoad = float64(stats.Indexing.IndexCurrent) / 10.0 // Simplified calculation

	// Calculate optimization score (0-100)
	writeMetrics.OptimizationScore = s.calculateOptimizationScore(stats)

	// Generate recommendations
	writeMetrics.Recommendations = s.generateWriteRecommendations(stats, writeMetrics)
	writeMetrics.LastOptimized = time.Now()

	indexInfo.WriteMetrics = writeMetrics
	return nil
}

// calculateOptimizationScore calculates a score (0-100) for write optimization
func (s *IndexService) calculateOptimizationScore(stats *models.IndexStatsDetails) float64 {
	score := 100.0

	// Penalize high segment count (indicates merge inefficiency)
	if stats.Segments.Count > 50 {
		score -= math.Min(20.0, float64(stats.Segments.Count-50)/5.0)
	}

	// Penalize high merge time ratio
	if stats.Indexing.IndexTimeInMillis > 0 {
		mergeRatio := float64(stats.Merges.TotalTimeInMillis) / float64(stats.Indexing.IndexTimeInMillis)
		if mergeRatio > 0.1 { // More than 10% time spent merging
			score -= math.Min(15.0, (mergeRatio-0.1)*100.0)
		}
	}

	// Penalize large translog
	if stats.Translog.SizeInBytes > 1024*1024*100 { // > 100MB
		score -= math.Min(10.0, float64(stats.Translog.SizeInBytes)/(1024*1024*1000)) // Penalize per GB
	}

	// Penalize throttling
	if stats.Indexing.IsThrottled {
		score -= 15.0
	}

	return math.Max(0.0, score)
}

// generateWriteRecommendations generates optimization recommendations
func (s *IndexService) generateWriteRecommendations(stats *models.IndexStatsDetails, metrics *models.WriteMetrics) []string {
	var recommendations []string

	// High segment count
	if stats.Segments.Count > 50 {
		recommendations = append(recommendations, 
			"Consider force-merging to reduce segment count and improve performance")
	}

	// Low indexing rate
	if metrics.IndexingRate < 100 {
		recommendations = append(recommendations,
			"Indexing rate appears low - consider increasing bulk batch sizes or refresh interval")
	}

	// High merge overhead
	if stats.Indexing.IndexTimeInMillis > 0 {
		mergeRatio := float64(stats.Merges.TotalTimeInMillis) / float64(stats.Indexing.IndexTimeInMillis)
		if mergeRatio > 0.15 {
			recommendations = append(recommendations,
				"High merge overhead detected - consider tuning merge policy settings")
		}
	}

	// Large translog
	if stats.Translog.SizeInBytes > 1024*1024*500 { // > 500MB
		recommendations = append(recommendations,
			"Large translog detected - consider reducing flush threshold or increasing flush frequency")
	}

	// Throttling detected
	if stats.Indexing.IsThrottled {
		recommendations = append(recommendations,
			"Indexing throttling detected - consider increasing merge thread count or optimizing disk I/O")
	}

	// Low optimization score
	if metrics.OptimizationScore < 80 {
		recommendations = append(recommendations,
			"Overall optimization score is low - run index optimization analysis for detailed recommendations")
	}

	return recommendations
}

// OptimizeIndex analyzes and optimizes an index for write performance
func (s *IndexService) OptimizeIndex(ctx context.Context, req *models.OptimizationRequest) (*models.OptimizationResponse, error) {
	s.logger.Info("Optimizing index for write performance",
		zap.String("index_name", req.IndexName),
		zap.String("optimize_for", req.OptimizeFor),
		zap.String("workload", req.Workload))

	// Get current settings
	currentSettings, err := s.getCurrentIndexSettings(ctx, req.IndexName)
	if err != nil {
		return nil, fmt.Errorf("failed to get current settings: %w", err)
	}

	// Generate recommended settings
	recommendedSettings := s.generateOptimizedSettings(req, currentSettings)

	// Calculate the changes
	changes := s.calculateOptimizationChanges(currentSettings, recommendedSettings)

	response := &models.OptimizationResponse{
		IndexName:           req.IndexName,
		CurrentSettings:     currentSettings,
		RecommendedSettings: recommendedSettings,
		OptimizationsApplied: changes,
		PerformanceImpact:   s.estimatePerformanceImpact(changes),
		Applied:             false,
		RequestID:           s.generateRequestID(),
		Timestamp:           time.Now(),
	}

	// Apply changes if requested
	if req.ApplyChanges {
		if err := s.applyOptimizedSettings(ctx, req.IndexName, recommendedSettings); err != nil {
			return nil, fmt.Errorf("failed to apply optimizations: %w", err)
		}
		response.Applied = true
		s.logger.Info("Successfully applied index optimizations",
			zap.String("index_name", req.IndexName),
			zap.Int("changes_applied", len(changes)))
	}

	return response, nil
}

// getCurrentIndexSettings retrieves current index settings
func (s *IndexService) getCurrentIndexSettings(ctx context.Context, indexName string) (map[string]interface{}, error) {
	res, err := s.esClient.Indices.GetSettings(
		s.esClient.Indices.GetSettings.WithContext(ctx),
		s.esClient.Indices.GetSettings.WithIndex(indexName),
		s.esClient.Indices.GetSettings.WithIncludeDefaults(true),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, shared.ParseESError(res)
	}

	var response map[string]interface{}
	if err := shared.DecodeJSONResponse(res, &response); err != nil {
		return nil, err
	}

	if indexSettings, ok := response[indexName]; ok {
		if settingsMap, ok := indexSettings.(map[string]interface{}); ok {
			if settings, ok := settingsMap["settings"]; ok {
				return settings.(map[string]interface{}), nil
			}
		}
	}

	return make(map[string]interface{}), nil
}

// generateOptimizedSettings creates optimized settings based on request
func (s *IndexService) generateOptimizedSettings(req *models.OptimizationRequest, current map[string]interface{}) map[string]interface{} {
	optimized := make(map[string]interface{})

	switch req.OptimizeFor {
	case "write_throughput":
		s.addWriteThroughputSettings(optimized, req)
	case "read_performance":
		s.addReadPerformanceSettings(optimized, req)
	case "storage":
		s.addStorageSettings(optimized, req)
	default:
		s.addWriteThroughputSettings(optimized, req) // Default to write optimization
	}

	return optimized
}

// addWriteThroughputSettings adds settings optimized for write throughput
func (s *IndexService) addWriteThroughputSettings(settings map[string]interface{}, req *models.OptimizationRequest) {
	switch req.Workload {
	case "bulk_write":
		settings["index.refresh_interval"] = "30s"
		settings["index.translog.flush_threshold_size"] = "1gb"
		settings["index.translog.sync_interval"] = "5s"
		settings["index.translog.durability"] = "async"
		settings["index.merge.policy.segments_per_tier"] = 30
		
	case "real_time_write":
		settings["index.refresh_interval"] = "1s"
		settings["index.translog.flush_threshold_size"] = "512mb"
		settings["index.translog.sync_interval"] = "1s"
		settings["index.merge.policy.segments_per_tier"] = 20
		
	default:
		settings["index.refresh_interval"] = "5s"
		settings["index.translog.flush_threshold_size"] = "512mb"
		settings["index.merge.policy.segments_per_tier"] = 20
	}

	// Corpus size optimizations
	switch req.CorpusSize {
	case "huge":
		settings["index.merge.policy.max_merge_size"] = "10gb"
		settings["index.merge.policy.max_merged_segment_mb"] = 10240
	case "large":
		settings["index.merge.policy.max_merge_size"] = "5gb"
		settings["index.merge.policy.max_merged_segment_mb"] = 5120
	default:
		settings["index.merge.policy.max_merge_size"] = "2gb"
	}
}

// addReadPerformanceSettings adds settings optimized for read performance
func (s *IndexService) addReadPerformanceSettings(settings map[string]interface{}, req *models.OptimizationRequest) {
	settings["index.refresh_interval"] = "1s"
	settings["index.merge.policy.segments_per_tier"] = 10 // Fewer segments for better read performance
	settings["index.merge.policy.max_merge_size"] = "5gb"
}

// addStorageSettings adds settings optimized for storage efficiency
func (s *IndexService) addStorageSettings(settings map[string]interface{}, req *models.OptimizationRequest) {
	settings["index.codec"] = "best_compression"
	settings["index.merge.policy.segments_per_tier"] = 10 // Fewer segments for better compression
	settings["index.mapping.source.compress"] = true
}

// calculateOptimizationChanges compares current and recommended settings
func (s *IndexService) calculateOptimizationChanges(current, recommended map[string]interface{}) []models.OptimizationChange {
	var changes []models.OptimizationChange

	for key, newValue := range recommended {
		currentValue := current[key]
		if currentValue != newValue {
			change := models.OptimizationChange{
				Setting:  key,
				OldValue: currentValue,
				NewValue: newValue,
				Reason:   s.getOptimizationReason(key, newValue),
				Impact:   s.getOptimizationImpact(key),
				Category: s.getOptimizationCategory(key),
			}
			changes = append(changes, change)
		}
	}

	return changes
}

// getOptimizationReason returns the reason for a specific optimization
func (s *IndexService) getOptimizationReason(setting string, value interface{}) string {
	switch setting {
	case "index.refresh_interval":
		return "Optimize refresh frequency for write workload"
	case "index.translog.flush_threshold_size":
		return "Increase flush threshold for better write throughput"
	case "index.translog.durability":
		return "Use async durability for maximum write performance"
	case "index.merge.policy.segments_per_tier":
		return "Optimize segment count for workload characteristics"
	case "index.codec":
		return "Use best compression for storage efficiency"
	default:
		return "Performance optimization for write-heavy workload"
	}
}

// getOptimizationImpact returns the expected impact level
func (s *IndexService) getOptimizationImpact(setting string) string {
	switch setting {
	case "index.refresh_interval", "index.translog.durability":
		return "high"
	case "index.translog.flush_threshold_size", "index.merge.policy.segments_per_tier":
		return "medium"
	default:
		return "low"
	}
}

// getOptimizationCategory returns the optimization category
func (s *IndexService) getOptimizationCategory(setting string) string {
	if strings.Contains(setting, "refresh") || strings.Contains(setting, "translog") {
		return "write_performance"
	}
	if strings.Contains(setting, "merge") {
		return "storage"
	}
	if strings.Contains(setting, "codec") || strings.Contains(setting, "compress") {
		return "storage"
	}
	return "reliability"
}

// estimatePerformanceImpact estimates the performance impact of changes
func (s *IndexService) estimatePerformanceImpact(changes []models.OptimizationChange) *models.PerformanceImpact {
	impact := &models.PerformanceImpact{
		WritePerformance:  "neutral",
		ReadPerformance:   "neutral",
		StorageEfficiency: "neutral",
		ResourceUsage:     "neutral",
	}

	writeImprovements := 0
	storageImprovements := 0

	for _, change := range changes {
		switch change.Category {
		case "write_performance":
			if change.Impact == "high" {
				writeImprovements += 3
			} else if change.Impact == "medium" {
				writeImprovements += 2
			} else {
				writeImprovements += 1
			}
		case "storage":
			storageImprovements += 1
		}
	}

	// Estimate write performance impact
	if writeImprovements >= 5 {
		impact.WritePerformance = "improved"
		impact.EstimatedImprovementPercent = 20.0 + float64(writeImprovements-5)*5.0
	} else if writeImprovements >= 2 {
		impact.WritePerformance = "improved"
		impact.EstimatedImprovementPercent = float64(writeImprovements) * 5.0
	}

	// Estimate storage impact
	if storageImprovements >= 2 {
		impact.StorageEfficiency = "improved"
	}

	return impact
}

// applyOptimizedSettings applies the optimized settings to the index
func (s *IndexService) applyOptimizedSettings(ctx context.Context, indexName string, settings map[string]interface{}) error {
	settingsBody := map[string]interface{}{
		"index": settings,
	}

	bodyBytes, err := json.Marshal(settingsBody)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	res, err := s.esClient.Indices.PutSettings(
		s.esClient.Indices.PutSettings.WithContext(ctx),
		s.esClient.Indices.PutSettings.WithIndex(indexName),
		s.esClient.Indices.PutSettings.WithBody(strings.NewReader(string(bodyBytes))),
	)
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return shared.ParseESError(res)
	}

	return nil
}

// DeleteIndex deletes an index
func (s *IndexService) DeleteIndex(ctx context.Context, indexName string) error {
	s.logger.Info("Deleting index", zap.String("index_name", indexName))

	res, err := s.esClient.Indices.Delete(
		[]string{indexName},
		s.esClient.Indices.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to delete index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return shared.ParseESError(res)
	}

	s.logger.Info("Successfully deleted index", zap.String("index_name", indexName))
	return nil
}

// ListIndices lists all indices with basic information
func (s *IndexService) ListIndices(ctx context.Context) ([]models.IndexInfo, error) {
	res, err := s.esClient.Cat.Indices(
		s.esClient.Cat.Indices.WithContext(ctx),
		s.esClient.Cat.Indices.WithFormat("json"),
		s.esClient.Cat.Indices.WithV(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list indices: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, shared.ParseESError(res)
	}

	var indices []models.IndexInfo
	if err := shared.DecodeJSONResponse(res, &indices); err != nil {
		return nil, fmt.Errorf("failed to decode indices: %w", err)
	}

	return indices, nil
}

// generateRequestID generates a unique request ID
func (s *IndexService) generateRequestID() string {
	return fmt.Sprintf("index-%d", time.Now().UnixNano())
}