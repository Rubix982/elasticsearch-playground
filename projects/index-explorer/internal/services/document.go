package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/saif-islam/es-playground/shared"
	"github.com/saif-islam/es-playground/projects/index-explorer/internal/models"
)

// DocumentService provides write-optimized document operations
type DocumentService struct {
	esClient *shared.ESClient
	logger   *zap.Logger
}

// NewDocumentService creates a new document service instance
func NewDocumentService(esClient *shared.ESClient, logger *zap.Logger) *DocumentService {
	return &DocumentService{
		esClient: esClient,
		logger:   logger,
	}
}

// BulkIndex performs high-performance bulk indexing operations
func (s *DocumentService) BulkIndex(ctx context.Context, req *models.BulkRequest) (*models.BulkResponse, error) {
	s.logger.Info("Starting bulk index operation",
		zap.String("index", req.IndexName),
		zap.Int("operations", len(req.Operations)),
		zap.Int("batch_size", req.BatchSize),
		zap.Int("workers", req.ParallelWorkers))

	startTime := time.Now()

	// Validate and set defaults
	if err := s.validateBulkRequest(req); err != nil {
		return nil, fmt.Errorf("invalid bulk request: %w", err)
	}

	// Process operations in optimized batches
	response, err := s.processBulkOperations(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to process bulk operations: %w", err)
	}

	// Calculate performance metrics
	processingTime := time.Since(startTime)
	response.Summary = s.calculateBulkSummary(response, processingTime)
	response.RequestID = s.generateRequestID()
	response.Timestamp = time.Now()

	s.logger.Info("Completed bulk index operation",
		zap.String("index", req.IndexName),
		zap.Int64("successful", response.Summary.SuccessfulOperations),
		zap.Int64("failed", response.Summary.FailedOperations),
		zap.Float64("throughput", response.Summary.ThroughputPerSecond),
		zap.Duration("duration", processingTime))

	return response, nil
}

// validateBulkRequest validates and sets defaults for bulk request
func (s *DocumentService) validateBulkRequest(req *models.BulkRequest) error {
	if req.IndexName == "" {
		return fmt.Errorf("index name is required")
	}

	if len(req.Operations) == 0 {
		return fmt.Errorf("no operations provided")
	}

	// Set intelligent defaults based on optimization strategy
	if req.BatchSize == 0 {
		req.BatchSize = s.calculateOptimalBatchSize(req)
	}

	if req.ParallelWorkers == 0 {
		req.ParallelWorkers = s.calculateOptimalWorkerCount(req)
	}

	if req.OptimizeFor == "" {
		req.OptimizeFor = "write_throughput" // Default to write optimization
	}

	if req.ErrorTolerance == "" {
		req.ErrorTolerance = "medium"
	}

	if req.Settings == nil {
		req.Settings = s.getDefaultBulkSettings(req)
	}

	return nil
}

// calculateOptimalBatchSize determines the best batch size based on document characteristics
func (s *DocumentService) calculateOptimalBatchSize(req *models.BulkRequest) int {
	// Estimate average document size
	avgDocSize := s.estimateAverageDocumentSize(req.Operations)
	
	switch {
	case avgDocSize < 1024: // < 1KB - small documents
		return 5000
	case avgDocSize < 10*1024: // < 10KB - medium documents  
		return 1000
	case avgDocSize < 100*1024: // < 100KB - large documents
		return 500
	default: // > 100KB - huge documents
		return 100
	}
}

// calculateOptimalWorkerCount determines the best number of parallel workers
func (s *DocumentService) calculateOptimalWorkerCount(req *models.BulkRequest) int {
	totalOps := len(req.Operations)
	
	switch {
	case totalOps < 1000:
		return 2
	case totalOps < 10000:
		return 4
	case totalOps < 100000:
		return 8
	default:
		return 16
	}
}

// estimateAverageDocumentSize estimates the average size of documents in bytes
func (s *DocumentService) estimateAverageDocumentSize(operations []models.BulkOperation) int {
	if len(operations) == 0 {
		return 1024 // Default assumption
	}

	// Sample first 10 operations to estimate size
	sampleSize := int(math.Min(float64(len(operations)), 10))
	totalSize := 0

	for i := 0; i < sampleSize; i++ {
		if operations[i].Document != nil {
			docBytes, _ := json.Marshal(operations[i].Document)
			totalSize += len(docBytes)
		} else if operations[i].Source != nil {
			srcBytes, _ := json.Marshal(operations[i].Source)
			totalSize += len(srcBytes)
		}
	}

	if totalSize == 0 {
		return 1024 // Default
	}

	return totalSize / sampleSize
}

// getDefaultBulkSettings returns default settings for bulk operations
func (s *DocumentService) getDefaultBulkSettings(req *models.BulkRequest) *models.BulkSettings {
	settings := &models.BulkSettings{
		Timeout: 60 * time.Second,
	}

	switch req.OptimizeFor {
	case "write_throughput":
		settings.RefreshPolicy = "false" // Don't refresh immediately for max throughput
		settings.WaitForActiveShards = "1" // Only wait for primary shard
	case "consistency":
		settings.RefreshPolicy = "wait_for" // Wait for refresh for consistency
		settings.WaitForActiveShards = "all" // Wait for all shards
	default:
		settings.RefreshPolicy = "false"
		settings.WaitForActiveShards = "1"
	}

	return settings
}

// processBulkOperations processes bulk operations with optimal performance
func (s *DocumentService) processBulkOperations(ctx context.Context, req *models.BulkRequest) (*models.BulkResponse, error) {
	totalOps := len(req.Operations)
	batchSize := req.BatchSize
	workerCount := req.ParallelWorkers

	// Calculate number of batches
	numBatches := int(math.Ceil(float64(totalOps) / float64(batchSize)))

	s.logger.Info("Processing bulk operations",
		zap.Int("total_operations", totalOps),
		zap.Int("batch_size", batchSize),
		zap.Int("num_batches", numBatches),
		zap.Int("workers", workerCount))

	// Create channels for work distribution
	batchChan := make(chan batchWork, numBatches)
	resultChan := make(chan batchResult, numBatches)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go s.bulkWorker(ctx, req, batchChan, resultChan, &wg)
	}

	// Send batches to workers
	go func() {
		defer close(batchChan)
		for i := 0; i < numBatches; i++ {
			start := i * batchSize
			end := int(math.Min(float64(start+batchSize), float64(totalOps)))
			
			batch := batchWork{
				id:         i,
				operations: req.Operations[start:end],
			}
			
			select {
			case batchChan <- batch:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Close result channel when all workers are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var allItems []models.BulkResponseItem
	totalTook := int64(0)
	hasErrors := false

	for result := range resultChan {
		if result.err != nil {
			s.logger.Error("Batch processing failed",
				zap.Int("batch_id", result.id),
				zap.Error(result.err))
			// Continue processing other batches
			continue
		}

		allItems = append(allItems, result.items...)
		totalTook += result.took
		if result.hasErrors {
			hasErrors = true
		}
	}

	return &models.BulkResponse{
		Took:   totalTook / int64(numBatches), // Average took time
		Errors: hasErrors,
		Items:  allItems,
	}, nil
}

// batchWork represents work for a single batch
type batchWork struct {
	id         int
	operations []models.BulkOperation
}

// batchResult represents the result of processing a batch
type batchResult struct {
	id        int
	items     []models.BulkResponseItem
	took      int64
	hasErrors bool
	err       error
}

// bulkWorker processes batches of bulk operations
func (s *DocumentService) bulkWorker(ctx context.Context, req *models.BulkRequest, 
	batchChan <-chan batchWork, resultChan chan<- batchResult, wg *sync.WaitGroup) {
	
	defer wg.Done()

	for batch := range batchChan {
		select {
		case <-ctx.Done():
			return
		default:
			result := s.processBatch(ctx, req, batch)
			resultChan <- result
		}
	}
}

// processBatch processes a single batch of operations
func (s *DocumentService) processBatch(ctx context.Context, req *models.BulkRequest, batch batchWork) batchResult {
	// Build bulk request body
	var buf bytes.Buffer
	for _, op := range batch.operations {
		// Action line
		actionLine := s.buildActionLine(op, req.IndexName)
		buf.WriteString(actionLine)
		buf.WriteByte('\n')

		// Document line (if needed)
		if op.Action != "delete" {
			var doc interface{}
			if op.Document != nil {
				doc = op.Document
			} else if op.Source != nil {
				doc = op.Source
			}

			if doc != nil {
				docBytes, _ := json.Marshal(doc)
				buf.Write(docBytes)
				buf.WriteByte('\n')
			}
		}
	}

	// Execute bulk request
	res, err := s.esClient.Bulk(
		s.esClient.Bulk.WithContext(ctx),
		s.esClient.Bulk.WithBody(&buf),
		s.esClient.Bulk.WithIndex(req.IndexName),
		s.esClient.Bulk.WithRefresh(req.Settings.RefreshPolicy),
		s.esClient.Bulk.WithTimeout(req.Settings.Timeout),
	)

	if err != nil {
		return batchResult{
			id:  batch.id,
			err: fmt.Errorf("bulk request failed: %w", err),
		}
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return batchResult{
			id:  batch.id,
			err: fmt.Errorf("bulk request error: %s - %s", res.Status(), string(body)),
		}
	}

	// Parse response
	var bulkResp struct {
		Took   int64 `json:"took"`
		Errors bool  `json:"errors"`
		Items  []models.BulkResponseItem `json:"items"`
	}

	if err := json.NewDecoder(res.Body).Decode(&bulkResp); err != nil {
		return batchResult{
			id:  batch.id,
			err: fmt.Errorf("failed to decode bulk response: %w", err),
		}
	}

	return batchResult{
		id:        batch.id,
		items:     bulkResp.Items,
		took:      bulkResp.Took,
		hasErrors: bulkResp.Errors,
	}
}

// buildActionLine builds the action line for bulk operations
func (s *DocumentService) buildActionLine(op models.BulkOperation, defaultIndex string) string {
	action := map[string]interface{}{}

	indexName := op.Index
	if indexName == "" {
		indexName = defaultIndex
	}

	actionBody := map[string]interface{}{
		"_index": indexName,
	}

	if op.ID != "" {
		actionBody["_id"] = op.ID
	}

	if op.Version != nil {
		actionBody["_version"] = *op.Version
	}

	if op.Routing != "" {
		actionBody["_routing"] = op.Routing
	}

	action[op.Action] = actionBody

	actionBytes, _ := json.Marshal(action)
	return string(actionBytes)
}

// calculateBulkSummary calculates summary statistics for bulk operations
func (s *DocumentService) calculateBulkSummary(response *models.BulkResponse, processingTime time.Duration) *models.BulkSummary {
	summary := &models.BulkSummary{
		TotalOperations:  int64(len(response.Items)),
		ProcessingTime:   processingTime,
		AverageLatency:   time.Duration(response.Took) * time.Millisecond,
	}

	// Count operation results
	for _, item := range response.Items {
		var itemResponse *models.BulkItemResponse
		
		// Find the actual response (could be index, create, update, or delete)
		if item.Index != nil {
			itemResponse = item.Index
		} else if item.Create != nil {
			itemResponse = item.Create
		} else if item.Update != nil {
			itemResponse = item.Update
		} else if item.Delete != nil {
			itemResponse = item.Delete
		}

		if itemResponse != nil {
			if itemResponse.Error != nil {
				summary.FailedOperations++
			} else {
				summary.SuccessfulOperations++
				
				// Count by operation type
				switch itemResponse.Result {
				case "created":
					summary.IndexedDocuments++
				case "updated":
					summary.UpdatedDocuments++
				case "deleted":
					summary.DeletedDocuments++
				default:
					summary.IndexedDocuments++ // Default to indexed
				}
			}
		}
	}

	// Calculate rates
	if processingTime.Seconds() > 0 {
		summary.ThroughputPerSecond = float64(summary.SuccessfulOperations) / processingTime.Seconds()
	}

	if summary.TotalOperations > 0 {
		summary.ErrorRate = float64(summary.FailedOperations) / float64(summary.TotalOperations) * 100.0
	}

	return summary
}

// IndexDocument indexes a single document (wrapper around bulk for consistency)
func (s *DocumentService) IndexDocument(ctx context.Context, indexName, docID string, document map[string]interface{}) (*models.BulkResponse, error) {
	bulkReq := &models.BulkRequest{
		IndexName: indexName,
		Operations: []models.BulkOperation{
			{
				Action:   "index",
				ID:       docID,
				Document: document,
			},
		},
		BatchSize:       1,
		ParallelWorkers: 1,
		OptimizeFor:     "consistency", // Single doc operations prioritize consistency
	}

	return s.BulkIndex(ctx, bulkReq)
}

// GetDocument retrieves a single document by ID
func (s *DocumentService) GetDocument(ctx context.Context, indexName, docID string) (map[string]interface{}, error) {
	s.logger.Debug("Getting document",
		zap.String("index", indexName),
		zap.String("id", docID))

	res, err := s.esClient.Get(
		indexName,
		docID,
		s.esClient.Get.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("document not found")
	}

	if res.IsError() {
		return nil, shared.ParseESError(res)
	}

	var response struct {
		Found  bool                   `json:"found"`
		Source map[string]interface{} `json:"_source"`
	}

	if err := shared.DecodeJSONResponse(res, &response); err != nil {
		return nil, fmt.Errorf("failed to decode get response: %w", err)
	}

	if !response.Found {
		return nil, fmt.Errorf("document not found")
	}

	return response.Source, nil
}

// UpdateDocument updates a single document
func (s *DocumentService) UpdateDocument(ctx context.Context, indexName, docID string, updates map[string]interface{}) (*models.BulkResponse, error) {
	bulkReq := &models.BulkRequest{
		IndexName: indexName,
		Operations: []models.BulkOperation{
			{
				Action:   "update",
				ID:       docID,
				Document: map[string]interface{}{"doc": updates},
			},
		},
		BatchSize:       1,
		ParallelWorkers: 1,
		OptimizeFor:     "consistency",
	}

	return s.BulkIndex(ctx, bulkReq)
}

// DeleteDocument deletes a single document
func (s *DocumentService) DeleteDocument(ctx context.Context, indexName, docID string) (*models.BulkResponse, error) {
	bulkReq := &models.BulkRequest{
		IndexName: indexName,
		Operations: []models.BulkOperation{
			{
				Action: "delete",
				ID:     docID,
			},
		},
		BatchSize:       1,
		ParallelWorkers: 1,
		OptimizeFor:     "consistency",
	}

	return s.BulkIndex(ctx, bulkReq)
}

// BulkImportFromNDJSON imports documents from NDJSON format with optimal performance
func (s *DocumentService) BulkImportFromNDJSON(ctx context.Context, indexName string, ndjsonData io.Reader, options *BulkImportOptions) (*models.BulkResponse, error) {
	if options == nil {
		options = s.getDefaultImportOptions()
	}

	s.logger.Info("Starting NDJSON bulk import",
		zap.String("index", indexName),
		zap.Int("batch_size", options.BatchSize),
		zap.Int("workers", options.ParallelWorkers))

	// Parse NDJSON into operations
	operations, err := s.parseNDJSON(ndjsonData, indexName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse NDJSON: %w", err)
	}

	// Create bulk request
	bulkReq := &models.BulkRequest{
		IndexName:       indexName,
		Operations:      operations,
		BatchSize:       options.BatchSize,
		ParallelWorkers: options.ParallelWorkers,
		OptimizeFor:     "write_throughput",
		ErrorTolerance:  options.ErrorTolerance,
	}

	return s.BulkIndex(ctx, bulkReq)
}

// BulkImportOptions defines options for bulk import operations
type BulkImportOptions struct {
	BatchSize       int
	ParallelWorkers int
	ErrorTolerance  string
	GenerateIDs     bool
}

// getDefaultImportOptions returns default options for bulk import
func (s *DocumentService) getDefaultImportOptions() *BulkImportOptions {
	return &BulkImportOptions{
		BatchSize:       1000,
		ParallelWorkers: 8,
		ErrorTolerance:  "medium",
		GenerateIDs:     true,
	}
}

// parseNDJSON parses NDJSON data into bulk operations
func (s *DocumentService) parseNDJSON(reader io.Reader, indexName string) ([]models.BulkOperation, error) {
	var operations []models.BulkOperation
	
	// Read all data
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// Split by lines
	lines := strings.Split(string(data), "\n")
	
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var document map[string]interface{}
		if err := json.Unmarshal([]byte(line), &document); err != nil {
			s.logger.Warn("Failed to parse JSON line",
				zap.Int("line", i+1),
				zap.Error(err))
			continue
		}

		// Extract ID if present
		var docID string
		if id, exists := document["_id"]; exists {
			docID = fmt.Sprintf("%v", id)
			delete(document, "_id") // Remove from document body
		}

		operation := models.BulkOperation{
			Action:   "index",
			Index:    indexName,
			ID:       docID,
			Document: document,
		}

		operations = append(operations, operation)
	}

	return operations, nil
}

// GetWritePerformanceMetrics calculates write performance metrics for an index
func (s *DocumentService) GetWritePerformanceMetrics(ctx context.Context, indexName string) (*models.WriteMetrics, error) {
	// Get index statistics
	res, err := s.esClient.Indices.Stats(
		s.esClient.Indices.Stats.WithContext(ctx),
		s.esClient.Indices.Stats.WithIndex(indexName),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get index stats: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, shared.ParseESError(res)
	}

	var statsResponse map[string]interface{}
	if err := shared.DecodeJSONResponse(res, &statsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode stats response: %w", err)
	}

	// Extract stats for the specific index
	indices, ok := statsResponse["indices"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid stats response format")
	}

	indexStats, ok := indices[indexName].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("index %s not found in stats", indexName)
	}

	// Parse into structured stats
	var stats models.IndexStats
	statsBytes, _ := json.Marshal(indexStats)
	json.Unmarshal(statsBytes, &stats)

	// Calculate write metrics
	metrics := s.calculateWriteMetrics(&stats)
	return metrics, nil
}

// calculateWriteMetrics calculates write performance metrics from index stats
func (s *DocumentService) calculateWriteMetrics(stats *models.IndexStats) *models.WriteMetrics {
	if stats.Total == nil {
		return &models.WriteMetrics{}
	}

	total := stats.Total
	metrics := &models.WriteMetrics{}

	// Calculate indexing rate (docs per second)
	if total.Indexing.IndexTimeInMillis > 0 {
		timeSeconds := float64(total.Indexing.IndexTimeInMillis) / 1000.0
		metrics.IndexingRate = float64(total.Indexing.IndexTotal) / timeSeconds
	}

	// Calculate average document size
	if total.Docs.Count > 0 && total.Store.SizeInBytes > 0 {
		metrics.AverageDocSize = total.Store.SizeInBytes / total.Docs.Count
	}

	// Calculate write latency (average time per document)
	if total.Indexing.IndexTotal > 0 {
		metrics.WriteLatency = float64(total.Indexing.IndexTimeInMillis) / float64(total.Indexing.IndexTotal)
	}

	// Calculate bulk latency (this is an approximation)
	metrics.BulkLatency = metrics.WriteLatency * 1000 // Assuming 1000 docs per bulk

	// Get segment count
	metrics.SegmentCount = total.Segments.Count

	// Calculate merge rate
	if total.Merges.TotalTimeInMillis > 0 {
		mergeTimeSeconds := float64(total.Merges.TotalTimeInMillis) / 1000.0
		metrics.MergeRate = float64(total.Merges.Total) / mergeTimeSeconds
	}

	// Calculate refresh rate  
	if total.Refresh.TotalTimeInMillis > 0 {
		refreshTimeSeconds := float64(total.Refresh.TotalTimeInMillis) / 1000.0
		metrics.RefreshRate = float64(total.Refresh.Total) / refreshTimeSeconds
	}

	// Get translog size
	metrics.TranslogSize = total.Translog.SizeInBytes

	// Calculate write load (simplified)
	metrics.WriteLoad = float64(total.Indexing.IndexCurrent) / 10.0

	// Calculate optimization score
	metrics.OptimizationScore = s.calculateWriteOptimizationScore(total)

	// Generate recommendations
	metrics.Recommendations = s.generateWriteOptimizationRecommendations(total, metrics)
	metrics.LastOptimized = time.Now()

	return metrics
}

// calculateWriteOptimizationScore calculates optimization score for write performance
func (s *DocumentService) calculateWriteOptimizationScore(stats *models.IndexStatsDetails) float64 {
	score := 100.0

	// Penalize high segment count
	if stats.Segments.Count > 50 {
		score -= math.Min(20.0, float64(stats.Segments.Count-50)/5.0)
	}

	// Penalize high merge time ratio
	if stats.Indexing.IndexTimeInMillis > 0 {
		mergeRatio := float64(stats.Merges.TotalTimeInMillis) / float64(stats.Indexing.IndexTimeInMillis)
		if mergeRatio > 0.1 {
			score -= math.Min(15.0, (mergeRatio-0.1)*100.0)
		}
	}

	// Penalize large translog
	if stats.Translog.SizeInBytes > 100*1024*1024 { // > 100MB
		score -= math.Min(10.0, float64(stats.Translog.SizeInBytes)/(1024*1024*1000))
	}

	// Penalize throttling
	if stats.Indexing.IsThrottled {
		score -= 15.0
	}

	// Penalize low indexing rate (if we have enough data)
	if stats.Indexing.IndexTotal > 1000 {
		avgRate := float64(stats.Indexing.IndexTotal) / (float64(stats.Indexing.IndexTimeInMillis) / 1000.0)
		if avgRate < 100 { // Less than 100 docs/sec
			score -= 10.0
		}
	}

	return math.Max(0.0, score)
}

// generateWriteOptimizationRecommendations generates recommendations for write optimization
func (s *DocumentService) generateWriteOptimizationRecommendations(stats *models.IndexStatsDetails, metrics *models.WriteMetrics) []string {
	var recommendations []string

	// High segment count
	if stats.Segments.Count > 50 {
		recommendations = append(recommendations,
			"High segment count detected - consider force merging or adjusting merge policy")
	}

	// Low indexing rate
	if metrics.IndexingRate < 100 && stats.Indexing.IndexTotal > 100 {
		recommendations = append(recommendations,
			"Low indexing rate - consider increasing bulk batch sizes or using parallel processing")
	}

	// High merge overhead
	if stats.Indexing.IndexTimeInMillis > 0 {
		mergeRatio := float64(stats.Merges.TotalTimeInMillis) / float64(stats.Indexing.IndexTimeInMillis)
		if mergeRatio > 0.15 {
			recommendations = append(recommendations,
				"High merge overhead - consider optimizing merge policy or increasing merge thread count")
		}
	}

	// Large translog
	if stats.Translog.SizeInBytes > 500*1024*1024 { // > 500MB
		recommendations = append(recommendations,
			"Large translog detected - consider reducing flush threshold or increasing flush frequency")
	}

	// Throttling
	if stats.Indexing.IsThrottled {
		recommendations = append(recommendations,
			"Indexing throttling detected - check disk I/O performance and merge settings")
	}

	// High write latency
	if metrics.WriteLatency > 10.0 { // > 10ms per document
		recommendations = append(recommendations,
			"High write latency - consider optimizing document structure or index settings")
	}

	// Many failed operations
	if stats.Indexing.IndexFailed > stats.Indexing.IndexTotal/10 { // > 10% failure rate
		recommendations = append(recommendations,
			"High failure rate detected - review document validation and error handling")
	}

	return recommendations
}

// generateRequestID generates a unique request ID
func (s *DocumentService) generateRequestID() string {
	return fmt.Sprintf("doc-%d", time.Now().UnixNano())
}