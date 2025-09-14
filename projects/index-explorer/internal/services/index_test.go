package services

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"go.uber.org/zap"

	"github.com/saif-islam/es-playground/projects/index-explorer/internal/models"
	"github.com/saif-islam/es-playground/shared"
)

// Mock Elasticsearch client for testing
type mockESClient struct {
	responses map[string]string
}

func (m *mockESClient) Indices() *elasticsearch.Indices {
	return &elasticsearch.Indices{}
}

func (m *mockESClient) Info(o ...func(*esapi.InfoRequest)) (*esapi.Response, error) {
	return &esapi.Response{
		StatusCode: 200,
		Body:       strings.NewReader(`{"version":{"number":"8.11.1"}}`),
	}, nil
}

func (m *mockESClient) Search(o ...func(*esapi.SearchRequest)) (*esapi.Response, error) {
	return &esapi.Response{
		StatusCode: 200,
		Body:       strings.NewReader(`{"hits":{"total":{"value":0}}}`),
	}, nil
}

func (m *mockESClient) WaitForCluster(ctx context.Context, status string, timeout time.Duration) error {
	return nil
}

func (m *mockESClient) Ping() error {
	return nil
}

func (m *mockESClient) GetClusterHealth() (*shared.ClusterHealth, error) {
	return &shared.ClusterHealth{
		Status:               "green",
		NumberOfNodes:        1,
		ActivePrimaryShards:  1,
		ActiveShards:         1,
		RelocatingShards:     0,
		InitializingShards:   0,
		UnassignedShards:     0,
		DelayedUnassignedShards: 0,
	}, nil
}

func newMockESClient() shared.ESClientInterface {
	return &mockESClient{
		responses: make(map[string]string),
	}
}

func TestIndexService_CreateWriteOptimizedIndex(t *testing.T) {
	logger := zap.NewNop()
	esClient := newMockESClient()
	service := NewIndexService(esClient, logger)

	ctx := context.Background()
	
	testCases := []struct {
		name        string
		request     *models.IndexRequest
		expectError bool
	}{
		{
			name: "high volume text-heavy index",
			request: &models.IndexRequest{
				IndexName:        "test-high-volume",
				WriteOptimized:   true,
				TextHeavy:        true,
				ExpectedVolume:   "high",
				ExpectedDocSize:  "large",
				IngestionRate:    "high",
			},
			expectError: false,
		},
		{
			name: "medium volume small documents",
			request: &models.IndexRequest{
				IndexName:        "test-small-docs",
				WriteOptimized:   true,
				TextHeavy:        false,
				ExpectedVolume:   "medium",
				ExpectedDocSize:  "small",
				IngestionRate:    "medium",
			},
			expectError: false,
		},
		{
			name: "missing index name",
			request: &models.IndexRequest{
				WriteOptimized: true,
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response, err := service.CreateWriteOptimizedIndex(ctx, tc.request)
			
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if response == nil {
				t.Errorf("Expected response but got nil")
				return
			}
			
			// Verify response structure
			if response.IndexName != tc.request.IndexName {
				t.Errorf("Expected index name %s, got %s", tc.request.IndexName, response.IndexName)
			}
			
			if !response.WriteOptimized {
				t.Errorf("Expected write optimized to be true")
			}
		})
	}
}

func TestIndexService_OptimizeIndex(t *testing.T) {
	logger := zap.NewNop()
	esClient := newMockESClient()
	service := NewIndexService(esClient, logger)

	ctx := context.Background()
	
	request := &models.OptimizationRequest{
		IndexName:    "test-index",
		OptimizeFor:  "write_throughput",
		Workload:     "bulk_write",
		CorpusSize:   "large",
		Priority:     "write_throughput",
		ApplyChanges: false, // Don't apply in test
	}
	
	response, err := service.OptimizeIndex(ctx, request)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	
	if response == nil {
		t.Errorf("Expected response but got nil")
		return
	}
	
	// Verify optimization was calculated
	if len(response.Recommendations) == 0 {
		t.Errorf("Expected recommendations but got none")
	}
	
	if response.OptimizationScore == 0 {
		t.Errorf("Expected optimization score but got 0")
	}
}

func TestIndexService_GetIndexRecommendations(t *testing.T) {
	logger := zap.NewNop()
	esClient := newMockESClient()
	service := NewIndexService(esClient, logger)

	ctx := context.Background()
	
	response, err := service.GetIndexRecommendations(ctx, "test-index", "bulk_write", "large")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	
	if response == nil {
		t.Errorf("Expected response but got nil")
		return
	}
	
	// Verify recommendations were generated
	if len(response.Recommendations) == 0 {
		t.Errorf("Expected recommendations but got none")
	}
}

// Benchmark tests for write optimization
func BenchmarkIndexService_CreateWriteOptimizedIndex(b *testing.B) {
	logger := zap.NewNop()
	esClient := newMockESClient()
	service := NewIndexService(esClient, logger)

	ctx := context.Background()
	request := &models.IndexRequest{
		IndexName:        "benchmark-index",
		WriteOptimized:   true,
		TextHeavy:        true,
		ExpectedVolume:   "high",
		ExpectedDocSize:  "large",
		IngestionRate:    "high",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.IndexName = "benchmark-index-" + string(rune(i))
		_, err := service.CreateWriteOptimizedIndex(ctx, request)
		if err != nil {
			b.Errorf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkIndexService_OptimizeIndex(b *testing.B) {
	logger := zap.NewNop()
	esClient := newMockESClient()
	service := NewIndexService(esClient, logger)

	ctx := context.Background()
	request := &models.OptimizationRequest{
		IndexName:    "benchmark-index",
		OptimizeFor:  "write_throughput",
		Workload:     "bulk_write",
		CorpusSize:   "large",
		Priority:     "write_throughput",
		ApplyChanges: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.OptimizeIndex(ctx, request)
		if err != nil {
			b.Errorf("Benchmark failed: %v", err)
		}
	}
}

// Test helper functions
func TestApplyWriteOptimizations(t *testing.T) {
	logger := zap.NewNop()
	esClient := newMockESClient()
	service := NewIndexService(esClient, logger)

	settings := &models.IndexSettings{
		NumberOfShards:   1,
		NumberOfReplicas: 1,
		RefreshInterval:  "1s",
	}

	testCases := []struct {
		name           string
		request        *models.IndexRequest
		expectedRefresh string
		expectedReplicas int
	}{
		{
			name: "high ingestion rate",
			request: &models.IndexRequest{
				IngestionRate: "high",
			},
			expectedRefresh: "30s",
			expectedReplicas: 0,
		},
		{
			name: "medium ingestion rate", 
			request: &models.IndexRequest{
				IngestionRate: "medium",
			},
			expectedRefresh: "5s",
			expectedReplicas: 0,
		},
		{
			name: "low ingestion rate",
			request: &models.IndexRequest{
				IngestionRate: "low",
			},
			expectedRefresh: "1s",
			expectedReplicas: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset settings
			settings.RefreshInterval = "1s"
			settings.NumberOfReplicas = 1
			
			service.applyWriteOptimizations(settings, tc.request)
			
			if settings.RefreshInterval != tc.expectedRefresh {
				t.Errorf("Expected refresh interval %s, got %s", tc.expectedRefresh, settings.RefreshInterval)
			}
			
			if settings.NumberOfReplicas != tc.expectedReplicas {
				t.Errorf("Expected replicas %d, got %d", tc.expectedReplicas, settings.NumberOfReplicas)
			}
		})
	}
}

// Performance test with various document sizes
func BenchmarkWriteOptimizations(b *testing.B) {
	logger := zap.NewNop()
	esClient := newMockESClient()
	service := NewIndexService(esClient, logger)

	ctx := context.Background()
	
	docSizes := []string{"small", "medium", "large", "huge"}
	volumes := []string{"low", "medium", "high"}
	
	for _, docSize := range docSizes {
		for _, volume := range volumes {
			b.Run("DocSize_"+docSize+"_Volume_"+volume, func(b *testing.B) {
				request := &models.IndexRequest{
					IndexName:        "perf-test",
					WriteOptimized:   true,
					TextHeavy:        true,
					ExpectedVolume:   volume,
					ExpectedDocSize:  docSize,
					IngestionRate:    "high",
				}
				
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, err := service.CreateWriteOptimizedIndex(ctx, request)
					if err != nil {
						b.Errorf("Benchmark failed: %v", err)
					}
				}
			})
		}
	}
}