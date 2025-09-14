package services

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/saif-islam/es-playground/projects/index-explorer/internal/models"
)

func TestDocumentService_BulkIndex(t *testing.T) {
	logger := zap.NewNop()
	esClient := newMockESClient()
	service := NewDocumentService(esClient, logger)

	ctx := context.Background()
	
	request := &models.BulkRequest{
		IndexName: "test-index",
		Operations: []models.BulkOperation{
			{
				Action: "index",
				Document: map[string]interface{}{
					"title":   "Test Document 1",
					"content": "This is test content for write optimization",
					"timestamp": time.Now(),
				},
			},
			{
				Action: "index",
				Document: map[string]interface{}{
					"title":   "Test Document 2", 
					"content": "More test content for bulk operations",
					"timestamp": time.Now(),
				},
			},
		},
		OptimizeFor:      "write_throughput",
		BatchSize:        1000,
		ParallelWorkers:  4,
		ErrorTolerance:   "medium",
	}
	
	response, err := service.BulkIndex(ctx, request)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	
	if response == nil {
		t.Errorf("Expected response but got nil")
		return
	}
	
	// Verify response structure
	if response.IndexName != request.IndexName {
		t.Errorf("Expected index name %s, got %s", request.IndexName, response.IndexName)
	}
	
	if response.Summary.TotalOperations != len(request.Operations) {
		t.Errorf("Expected %d operations, got %d", len(request.Operations), response.Summary.TotalOperations)
	}
}

func TestDocumentService_CalculateOptimalBatchSize(t *testing.T) {
	logger := zap.NewNop()
	esClient := newMockESClient()
	service := NewDocumentService(esClient, logger)

	testCases := []struct {
		name               string
		operations         []models.BulkOperation
		expectedBatchSize  int
	}{
		{
			name: "small documents",
			operations: []models.BulkOperation{
				{
					Action: "index",
					Document: map[string]interface{}{
						"id": 1,
						"name": "small",
					},
				},
			},
			expectedBatchSize: 5000,
		},
		{
			name: "medium documents", 
			operations: []models.BulkOperation{
				{
					Action: "index",
					Document: map[string]interface{}{
						"id": 1,
						"title": "Medium document with more content",
						"content": strings.Repeat("This is medium content ", 100),
						"metadata": map[string]interface{}{
							"category": "test",
							"tags": []string{"medium", "content", "test"},
						},
					},
				},
			},
			expectedBatchSize: 1000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := &models.BulkRequest{
				Operations: tc.operations,
			}
			
			batchSize := service.calculateOptimalBatchSize(request)
			if batchSize != tc.expectedBatchSize {
				t.Errorf("Expected batch size %d, got %d", tc.expectedBatchSize, batchSize)
			}
		})
	}
}

func TestDocumentService_BulkImportFromNDJSON(t *testing.T) {
	logger := zap.NewNop()
	esClient := newMockESClient()
	service := NewDocumentService(esClient, logger)

	ctx := context.Background()
	
	ndjsonData := `{"title": "Document 1", "content": "First document content"}
{"title": "Document 2", "content": "Second document content"}
{"title": "Document 3", "content": "Third document content"}`
	
	reader := bytes.NewReader([]byte(ndjsonData))
	
	options := &BulkImportOptions{
		BatchSize:       1000,
		ParallelWorkers: 4,
		ErrorTolerance:  "medium",
		GenerateIDs:     true,
	}
	
	response, err := service.BulkImportFromNDJSON(ctx, "test-index", reader, options)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	
	if response == nil {
		t.Errorf("Expected response but got nil")
		return
	}
	
	// Verify import summary
	if response.Summary.TotalOperations != 3 {
		t.Errorf("Expected 3 operations, got %d", response.Summary.TotalOperations)
	}
}

func TestDocumentService_GetWritePerformanceMetrics(t *testing.T) {
	logger := zap.NewNop()
	esClient := newMockESClient()
	service := NewDocumentService(esClient, logger)

	ctx := context.Background()
	
	metrics, err := service.GetWritePerformanceMetrics(ctx, "test-index")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	
	if metrics == nil {
		t.Errorf("Expected metrics but got nil")
		return
	}
	
	// Verify metrics structure
	if metrics.IndexName != "test-index" {
		t.Errorf("Expected index name test-index, got %s", metrics.IndexName)
	}
	
	if metrics.OptimizationScore < 0 || metrics.OptimizationScore > 100 {
		t.Errorf("Expected optimization score between 0-100, got %d", metrics.OptimizationScore)
	}
}

// Benchmark tests for document operations
func BenchmarkDocumentService_BulkIndex(b *testing.B) {
	logger := zap.NewNop()
	esClient := newMockESClient()
	service := NewDocumentService(esClient, logger)

	ctx := context.Background()
	
	// Create test documents of varying sizes
	smallDoc := map[string]interface{}{
		"id": 1,
		"title": "Small",
	}
	
	mediumDoc := map[string]interface{}{
		"id": 1,
		"title": "Medium document",
		"content": strings.Repeat("Medium content ", 50),
	}
	
	largeDoc := map[string]interface{}{
		"id": 1,
		"title": "Large document with extensive content",
		"content": strings.Repeat("Large content with lots of text ", 200),
		"metadata": map[string]interface{}{
			"category": "benchmark",
			"tags": []string{"large", "content", "benchmark", "test"},
			"properties": map[string]interface{}{
				"size": "large",
				"type": "text",
				"indexed": true,
			},
		},
	}

	testCases := []struct {
		name     string
		document map[string]interface{}
		batchSize int
	}{
		{"SmallDoc_Batch1000", smallDoc, 1000},
		{"SmallDoc_Batch5000", smallDoc, 5000},
		{"MediumDoc_Batch500", mediumDoc, 500},
		{"MediumDoc_Batch1000", mediumDoc, 1000},
		{"LargeDoc_Batch100", largeDoc, 100},
		{"LargeDoc_Batch500", largeDoc, 500},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			operations := make([]models.BulkOperation, tc.batchSize)
			for i := 0; i < tc.batchSize; i++ {
				operations[i] = models.BulkOperation{
					Action:   "index",
					Document: tc.document,
				}
			}
			
			request := &models.BulkRequest{
				IndexName:       "benchmark-index",
				Operations:      operations,
				OptimizeFor:     "write_throughput",
				BatchSize:       tc.batchSize,
				ParallelWorkers: 4,
				ErrorTolerance:  "medium",
			}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := service.BulkIndex(ctx, request)
				if err != nil {
					b.Errorf("Benchmark failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkDocumentService_CalculateOptimalBatchSize(b *testing.B) {
	logger := zap.NewNop()
	esClient := newMockESClient()
	service := NewDocumentService(esClient, logger)

	// Create operations with different document sizes
	operations := []models.BulkOperation{
		{
			Action: "index",
			Document: map[string]interface{}{
				"small": "content",
			},
		},
		{
			Action: "index", 
			Document: map[string]interface{}{
				"medium": strings.Repeat("content ", 100),
				"data": map[string]interface{}{
					"nested": "values",
				},
			},
		},
		{
			Action: "index",
			Document: map[string]interface{}{
				"large": strings.Repeat("extensive content ", 500),
				"metadata": map[string]interface{}{
					"complex": "structure",
					"array": []string{"item1", "item2", "item3"},
				},
			},
		},
	}

	request := &models.BulkRequest{
		Operations: operations,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.calculateOptimalBatchSize(request)
	}
}

// Test adaptive worker calculation
func TestDocumentService_CalculateOptimalWorkers(t *testing.T) {
	logger := zap.NewNop()
	esClient := newMockESClient()
	service := NewDocumentService(esClient, logger)

	testCases := []struct {
		name           string
		operations     int
		targetThroughput string
		expectedWorkers int
	}{
		{
			name:            "small operation count low throughput",
			operations:      100,
			targetThroughput: "low",
			expectedWorkers: 2,
		},
		{
			name:            "medium operation count high throughput", 
			operations:      10000,
			targetThroughput: "high",
			expectedWorkers: 12,
		},
		{
			name:            "large operation count max throughput",
			operations:      100000,
			targetThroughput: "max",
			expectedWorkers: 32,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			workers := service.calculateOptimalWorkers(tc.operations, tc.targetThroughput)
			if workers != tc.expectedWorkers {
				t.Errorf("Expected %d workers, got %d", tc.expectedWorkers, workers)
			}
		})
	}
}

// Performance test for NDJSON import
func BenchmarkDocumentService_BulkImportFromNDJSON(b *testing.B) {
	logger := zap.NewNop()
	esClient := newMockESClient()
	service := NewDocumentService(esClient, logger)

	ctx := context.Background()
	
	// Generate NDJSON data of different sizes
	smallNDJSON := generateNDJSONData(100, "small")
	mediumNDJSON := generateNDJSONData(1000, "medium")
	largeNDJSON := generateNDJSONData(10000, "large")

	testCases := []struct {
		name     string
		data     string
		batchSize int
	}{
		{"Small100_Batch100", smallNDJSON, 100},
		{"Medium1K_Batch500", mediumNDJSON, 500},
		{"Large10K_Batch1000", largeNDJSON, 1000},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			options := &BulkImportOptions{
				BatchSize:       tc.batchSize,
				ParallelWorkers: 4,
				ErrorTolerance:  "medium",
				GenerateIDs:     true,
			}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				reader := bytes.NewReader([]byte(tc.data))
				_, err := service.BulkImportFromNDJSON(ctx, "benchmark-index", reader, options)
				if err != nil {
					b.Errorf("Benchmark failed: %v", err)
				}
			}
		})
	}
}

// Helper function to generate NDJSON test data
func generateNDJSONData(count int, size string) string {
	var builder strings.Builder
	
	for i := 0; i < count; i++ {
		var content string
		switch size {
		case "small":
			content = "Small content"
		case "medium":
			content = strings.Repeat("Medium content ", 20)
		case "large":
			content = strings.Repeat("Large content with extensive text ", 100)
		}
		
		builder.WriteString(`{"id":`)
		builder.WriteString(`"doc_`)
		builder.WriteString(string(rune(i)))
		builder.WriteString(`","title":"Document `)
		builder.WriteString(string(rune(i)))
		builder.WriteString(`","content":"`)
		builder.WriteString(content)
		builder.WriteString(`"}`)
		
		if i < count-1 {
			builder.WriteString("\n")
		}
	}
	
	return builder.String()
}