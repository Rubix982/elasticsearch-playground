package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultAPIURL     = "http://localhost:8082"
	defaultDocCount   = 1000
	defaultWorkers    = 8
	defaultBatchSize  = 500
)

type PerformanceTest struct {
	APIURL    string
	DocCount  int
	Workers   int
	BatchSize int
	IndexName string
}

type TestResult struct {
	TestName        string        `json:"test_name"`
	DocumentCount   int           `json:"document_count"`
	TotalTime       time.Duration `json:"total_time"`
	DocsPerSecond   float64       `json:"docs_per_second"`
	AvgLatency      time.Duration `json:"avg_latency"`
	BatchSize       int           `json:"batch_size"`
	Workers         int           `json:"workers"`
	ErrorCount      int           `json:"error_count"`
	OptimizationScore int         `json:"optimization_score"`
}

func main() {
	// Parse command line arguments
	apiURL := getEnv("API_URL", defaultAPIURL)
	docCount, _ := strconv.Atoi(getEnv("DOC_COUNT", strconv.Itoa(defaultDocCount)))
	workers, _ := strconv.Atoi(getEnv("WORKERS", strconv.Itoa(defaultWorkers)))
	batchSize, _ := strconv.Atoi(getEnv("BATCH_SIZE", strconv.Itoa(defaultBatchSize)))
	
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "quick":
			docCount = 100
			workers = 4
			batchSize = 50
		case "medium":
			docCount = 1000
			workers = 8
			batchSize = 500
		case "heavy":
			docCount = 10000
			workers = 16
			batchSize = 1000
		case "extreme":
			docCount = 100000
			workers = 32
			batchSize = 2000
		}
	}

	perfTest := &PerformanceTest{
		APIURL:    apiURL,
		DocCount:  docCount,
		Workers:   workers,
		BatchSize: batchSize,
		IndexName: fmt.Sprintf("perf-test-%d", time.Now().Unix()),
	}

	fmt.Printf("🚀 Starting Write Performance Test\n")
	fmt.Printf("📊 Configuration:\n")
	fmt.Printf("   • API URL: %s\n", perfTest.APIURL)
	fmt.Printf("   • Documents: %d\n", perfTest.DocCount)
	fmt.Printf("   • Workers: %d\n", perfTest.Workers)
	fmt.Printf("   • Batch Size: %d\n", perfTest.BatchSize)
	fmt.Printf("   • Index: %s\n", perfTest.IndexName)
	fmt.Println()

	// Run performance tests
	results := runPerformanceTests(perfTest)
	
	// Display results
	displayResults(results)
	
	// Cleanup
	cleanup(perfTest)
}

func runPerformanceTests(perfTest *PerformanceTest) []TestResult {
	var results []TestResult
	
	// Test 1: Create write-optimized index
	fmt.Printf("📋 Test 1: Creating write-optimized index...\n")
	start := time.Now()
	err := createWriteOptimizedIndex(perfTest)
	if err != nil {
		log.Printf("❌ Failed to create index: %v", err)
		return results
	}
	indexCreationTime := time.Since(start)
	fmt.Printf("✅ Index created in %v\n\n", indexCreationTime)
	
	// Test 2: Small documents bulk test
	fmt.Printf("📋 Test 2: Small documents bulk indexing...\n")
	smallDocResult := bulkIndexTest(perfTest, "small", "Small Documents Test")
	results = append(results, smallDocResult)
	
	// Test 3: Medium documents bulk test  
	fmt.Printf("📋 Test 3: Medium documents bulk indexing...\n")
	mediumDocResult := bulkIndexTest(perfTest, "medium", "Medium Documents Test")
	results = append(results, mediumDocResult)
	
	// Test 4: Large documents bulk test
	fmt.Printf("📋 Test 4: Large documents bulk indexing...\n")
	largeDocResult := bulkIndexTest(perfTest, "large", "Large Documents Test")
	results = append(results, largeDocResult)
	
	// Test 5: Adaptive bulk test
	fmt.Printf("📋 Test 5: Adaptive bulk indexing...\n")
	adaptiveResult := adaptiveBulkTest(perfTest)
	results = append(results, adaptiveResult)
	
	// Test 6: NDJSON import test
	fmt.Printf("📋 Test 6: NDJSON import test...\n")
	ndjsonResult := ndjsonImportTest(perfTest)
	results = append(results, ndjsonResult)
	
	return results
}

func createWriteOptimizedIndex(perfTest *PerformanceTest) error {
	payload := map[string]interface{}{
		"index_name":        perfTest.IndexName,
		"expected_volume":   "high",
		"expected_doc_size": "large",
		"ingestion_rate":    "high",
		"text_heavy":        true,
	}
	
	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(
		perfTest.APIURL+"/api/v1/indices/write-optimized",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create index: status %d", resp.StatusCode)
	}
	
	return nil
}

func bulkIndexTest(perfTest *PerformanceTest, docSize, testName string) TestResult {
	start := time.Now()
	errorCount := 0
	
	// Generate documents
	documents := generateDocuments(perfTest.DocCount, docSize)
	
	// Create bulk operations
	operations := make([]map[string]interface{}, len(documents))
	for i, doc := range documents {
		operations[i] = map[string]interface{}{
			"action":   "index",
			"document": doc,
		}
	}
	
	// Perform bulk index
	payload := map[string]interface{}{
		"operations":       operations,
		"optimize_for":     "write_throughput",
		"batch_size":       perfTest.BatchSize,
		"parallel_workers": perfTest.Workers,
		"error_tolerance":  "medium",
	}
	
	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(
		perfTest.APIURL+"/api/v1/indices/"+perfTest.IndexName+"/bulk",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	
	if err != nil {
		errorCount++
		log.Printf("❌ Bulk index failed: %v", err)
	} else {
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			errorCount++
			log.Printf("❌ Bulk index failed: status %d", resp.StatusCode)
		}
	}
	
	totalTime := time.Since(start)
	docsPerSecond := float64(perfTest.DocCount) / totalTime.Seconds()
	avgLatency := totalTime / time.Duration(perfTest.DocCount)
	
	result := TestResult{
		TestName:        testName,
		DocumentCount:   perfTest.DocCount,
		TotalTime:       totalTime,
		DocsPerSecond:   docsPerSecond,
		AvgLatency:      avgLatency,
		BatchSize:       perfTest.BatchSize,
		Workers:         perfTest.Workers,
		ErrorCount:      errorCount,
		OptimizationScore: calculateOptimizationScore(docsPerSecond, docSize),
	}
	
	fmt.Printf("✅ %s completed: %.2f docs/sec in %v\n\n", testName, docsPerSecond, totalTime)
	return result
}

func adaptiveBulkTest(perfTest *PerformanceTest) TestResult {
	start := time.Now()
	errorCount := 0
	
	// Generate mixed size documents
	documents := generateMixedDocuments(perfTest.DocCount)
	
	payload := map[string]interface{}{
		"index_name":         perfTest.IndexName + "-adaptive",
		"documents":          documents,
		"auto_batch_size":    true,
		"target_throughput":  "max",
		"error_tolerance":    "medium",
		"optimize_for":       "write_throughput",
	}
	
	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(
		perfTest.APIURL+"/api/v1/bulk/adaptive",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	
	if err != nil {
		errorCount++
		log.Printf("❌ Adaptive bulk failed: %v", err)
	} else {
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			errorCount++
			log.Printf("❌ Adaptive bulk failed: status %d", resp.StatusCode)
		}
	}
	
	totalTime := time.Since(start)
	docsPerSecond := float64(perfTest.DocCount) / totalTime.Seconds()
	avgLatency := totalTime / time.Duration(perfTest.DocCount)
	
	result := TestResult{
		TestName:        "Adaptive Bulk Test",
		DocumentCount:   perfTest.DocCount,
		TotalTime:       totalTime,
		DocsPerSecond:   docsPerSecond,
		AvgLatency:      avgLatency,
		BatchSize:       0, // Adaptive
		Workers:         0, // Adaptive
		ErrorCount:      errorCount,
		OptimizationScore: calculateOptimizationScore(docsPerSecond, "mixed"),
	}
	
	fmt.Printf("✅ Adaptive bulk completed: %.2f docs/sec in %v\n\n", docsPerSecond, totalTime)
	return result
}

func ndjsonImportTest(perfTest *PerformanceTest) TestResult {
	start := time.Now()
	errorCount := 0
	
	// Generate NDJSON data
	ndjsonData := generateNDJSONData(perfTest.DocCount)
	
	url := fmt.Sprintf("%s/api/v1/indices/%s-ndjson/import/ndjson?batch_size=%d&workers=%d",
		perfTest.APIURL, perfTest.IndexName, perfTest.BatchSize, perfTest.Workers)
	
	resp, err := http.Post(url, "application/x-ndjson", strings.NewReader(ndjsonData))
	
	if err != nil {
		errorCount++
		log.Printf("❌ NDJSON import failed: %v", err)
	} else {
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			errorCount++
			log.Printf("❌ NDJSON import failed: status %d", resp.StatusCode)
		}
	}
	
	totalTime := time.Since(start)
	docsPerSecond := float64(perfTest.DocCount) / totalTime.Seconds()
	avgLatency := totalTime / time.Duration(perfTest.DocCount)
	
	result := TestResult{
		TestName:        "NDJSON Import Test",
		DocumentCount:   perfTest.DocCount,
		TotalTime:       totalTime,
		DocsPerSecond:   docsPerSecond,
		AvgLatency:      avgLatency,
		BatchSize:       perfTest.BatchSize,
		Workers:         perfTest.Workers,
		ErrorCount:      errorCount,
		OptimizationScore: calculateOptimizationScore(docsPerSecond, "ndjson"),
	}
	
	fmt.Printf("✅ NDJSON import completed: %.2f docs/sec in %v\n\n", docsPerSecond, totalTime)
	return result
}

func generateDocuments(count int, size string) []map[string]interface{} {
	documents := make([]map[string]interface{}, count)
	
	for i := 0; i < count; i++ {
		var content string
		switch size {
		case "small":
			content = fmt.Sprintf("Small document content %d", i)
		case "medium":
			content = fmt.Sprintf("Medium document with more content %d. %s", i, strings.Repeat("Additional text ", 50))
		case "large":
			content = fmt.Sprintf("Large document with extensive content %d. %s", i, strings.Repeat("Lots of text content ", 200))
		}
		
		documents[i] = map[string]interface{}{
			"id":        fmt.Sprintf("doc_%d", i),
			"title":     fmt.Sprintf("Performance Test Document %d", i),
			"content":   content,
			"size":      size,
			"timestamp": time.Now().Format(time.RFC3339),
			"metadata": map[string]interface{}{
				"test_type": "performance",
				"doc_size":  size,
				"batch_id":  i / 100, // Group docs into batches of 100
			},
		}
	}
	
	return documents
}

func generateMixedDocuments(count int) []map[string]interface{} {
	documents := make([]map[string]interface{}, count)
	sizes := []string{"small", "medium", "large"}
	
	for i := 0; i < count; i++ {
		size := sizes[i%len(sizes)]
		var content string
		
		switch size {
		case "small":
			content = fmt.Sprintf("Small mixed document %d", i)
		case "medium":
			content = fmt.Sprintf("Medium mixed document %d. %s", i, strings.Repeat("Mixed content ", 30))
		case "large":
			content = fmt.Sprintf("Large mixed document %d. %s", i, strings.Repeat("Extensive mixed content ", 100))
		}
		
		documents[i] = map[string]interface{}{
			"id":        fmt.Sprintf("mixed_%d", i),
			"title":     fmt.Sprintf("Mixed Document %d", i),
			"content":   content,
			"size":      size,
			"timestamp": time.Now().Format(time.RFC3339),
		}
	}
	
	return documents
}

func generateNDJSONData(count int) string {
	var builder strings.Builder
	
	for i := 0; i < count; i++ {
		content := fmt.Sprintf("NDJSON document content %d. %s", i, strings.Repeat("NDJSON text ", 20))
		
		doc := map[string]interface{}{
			"id":        fmt.Sprintf("ndjson_%d", i),
			"title":     fmt.Sprintf("NDJSON Document %d", i),
			"content":   content,
			"timestamp": time.Now().Format(time.RFC3339),
		}
		
		jsonBytes, _ := json.Marshal(doc)
		builder.Write(jsonBytes)
		
		if i < count-1 {
			builder.WriteString("\n")
		}
	}
	
	return builder.String()
}

func calculateOptimizationScore(docsPerSecond float64, docType string) int {
	// Base scoring on docs per second with adjustments for document type
	var baseTarget float64
	
	switch docType {
	case "small":
		baseTarget = 5000  // Expected 5000 docs/sec for small docs
	case "medium":
		baseTarget = 1000  // Expected 1000 docs/sec for medium docs
	case "large":
		baseTarget = 200   // Expected 200 docs/sec for large docs
	case "mixed":
		baseTarget = 1500  // Expected 1500 docs/sec for mixed docs
	case "ndjson":
		baseTarget = 2000  // Expected 2000 docs/sec for NDJSON
	default:
		baseTarget = 1000
	}
	
	score := int((docsPerSecond / baseTarget) * 100)
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}
	
	return score
}

func displayResults(results []TestResult) {
	fmt.Printf("📊 Performance Test Results\n")
	fmt.Printf("=" + strings.Repeat("=", 80) + "\n")
	
	for _, result := range results {
		fmt.Printf("🔥 %s\n", result.TestName)
		fmt.Printf("   Documents: %d\n", result.DocumentCount)
		fmt.Printf("   Total Time: %v\n", result.TotalTime)
		fmt.Printf("   Throughput: %.2f docs/sec\n", result.DocsPerSecond)
		fmt.Printf("   Avg Latency: %v\n", result.AvgLatency)
		fmt.Printf("   Batch Size: %d\n", result.BatchSize)
		fmt.Printf("   Workers: %d\n", result.Workers)
		fmt.Printf("   Errors: %d\n", result.ErrorCount)
		fmt.Printf("   Optimization Score: %d/100\n", result.OptimizationScore)
		fmt.Println()
	}
	
	// Calculate and display summary
	totalDocs := 0
	totalTime := time.Duration(0)
	totalErrors := 0
	avgScore := 0
	
	for _, result := range results {
		totalDocs += result.DocumentCount
		totalTime += result.TotalTime
		totalErrors += result.ErrorCount
		avgScore += result.OptimizationScore
	}
	
	if len(results) > 0 {
		avgScore = avgScore / len(results)
		overallThroughput := float64(totalDocs) / totalTime.Seconds()
		
		fmt.Printf("📈 Summary\n")
		fmt.Printf("   Total Documents: %d\n", totalDocs)
		fmt.Printf("   Total Time: %v\n", totalTime)
		fmt.Printf("   Overall Throughput: %.2f docs/sec\n", overallThroughput)
		fmt.Printf("   Total Errors: %d\n", totalErrors)
		fmt.Printf("   Average Optimization Score: %d/100\n", avgScore)
		
		// Performance assessment
		if avgScore >= 90 {
			fmt.Printf("🏆 Excellent - Production ready write performance!\n")
		} else if avgScore >= 70 {
			fmt.Printf("👍 Good - Minor optimizations could improve performance\n")
		} else if avgScore >= 50 {
			fmt.Printf("⚠️  Fair - Significant optimizations needed\n")
		} else {
			fmt.Printf("❌ Poor - Major performance issues detected\n")
		}
	}
}

func cleanup(perfTest *PerformanceTest) {
	fmt.Printf("🧹 Cleaning up test indices...\n")
	
	// Delete test indices (implementation would depend on API availability)  
	indices := []string{
		perfTest.IndexName,
		perfTest.IndexName + "-adaptive",
		perfTest.IndexName + "-ndjson",
	}
	
	for _, index := range indices {
		req, _ := http.NewRequest("DELETE", perfTest.APIURL+"/api/v1/indices/"+index, nil)
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
		}
	}
	
	fmt.Printf("✅ Cleanup completed\n")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}