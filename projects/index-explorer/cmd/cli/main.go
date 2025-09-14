package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAPIURL = "http://localhost:8082"
	version       = "1.0.0"
)

type CLI struct {
	APIURL  string
	client  *http.Client
	scanner *bufio.Scanner
}

type APIResponse struct {
	Success bool        `json:"success,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

func main() {
	fmt.Printf("🚀 Elasticsearch Index Explorer CLI v%s\n", version)
	fmt.Printf("Write-Optimized Operations Interface\n")
	fmt.Println(strings.Repeat("=", 50))

	apiURL := getEnv("API_URL", defaultAPIURL)
	
	cli := &CLI{
		APIURL:  apiURL,
		client:  &http.Client{Timeout: 30 * time.Second},
		scanner: bufio.NewScanner(os.Stdin),
	}

	// Check API connectivity
	if !cli.checkConnection() {
		fmt.Printf("❌ Cannot connect to API at %s\n", apiURL)
		fmt.Printf("💡 Make sure the Index Explorer is running with: make run-index-explorer\n")
		os.Exit(1)
	}

	fmt.Printf("✅ Connected to Index Explorer at %s\n\n", apiURL)
	
	// Start interactive session
	cli.runInteractiveSession()
}

func (c *CLI) checkConnection() bool {
	resp, err := c.client.Get(c.APIURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (c *CLI) runInteractiveSession() {
	c.showMainMenu()
	
	for {
		fmt.Print("📍 Enter command (or 'help'): ")
		if !c.scanner.Scan() {
			break
		}
		
		input := strings.TrimSpace(c.scanner.Text())
		if input == "" {
			continue
		}
		
		switch strings.ToLower(input) {
		case "help", "h":
			c.showMainMenu()
		case "quit", "exit", "q":
			fmt.Println("👋 Goodbye!")
			return
		case "status", "s":
			c.showStatus()
		case "create", "c":
			c.createIndex()
		case "list", "l":
			c.listIndices()
		case "optimize", "o":
			c.optimizeIndex()
		case "bulk", "b":
			c.bulkIndex()
		case "adaptive", "a":
			c.adaptiveBulk()
		case "ndjson", "n":
			c.ndjsonImport()
		case "metrics", "m":
			c.showMetrics()
		case "recommendations", "r":
			c.getRecommendations()
		case "perf", "p":
			c.performanceTest()
		case "examples", "e":
			c.showExamples()
		case "clear":
			c.clearScreen()
		default:
			fmt.Printf("❓ Unknown command: %s\n", input)
			fmt.Println("💡 Type 'help' to see available commands")
		}
		
		fmt.Println()
	}
}

func (c *CLI) showMainMenu() {
	fmt.Println("📋 Available Commands:")
	fmt.Println("  📊 status (s)         - Show service status")
	fmt.Println("  🏗️  create (c)         - Create write-optimized index")
	fmt.Println("  📑 list (l)           - List all indices")
	fmt.Println("  ⚡ optimize (o)       - Optimize existing index")
	fmt.Println("  📦 bulk (b)           - Bulk index documents")
	fmt.Println("  🤖 adaptive (a)       - Adaptive bulk indexing")
	fmt.Println("  📄 ndjson (n)         - Import NDJSON data")
	fmt.Println("  📈 metrics (m)        - Show write performance metrics")
	fmt.Println("  💡 recommendations (r) - Get optimization recommendations")
	fmt.Println("  🏃 perf (p)           - Run performance test")
	fmt.Println("  📖 examples (e)       - Show API examples")
	fmt.Println("  🧹 clear             - Clear screen")
	fmt.Println("  ❓ help (h)          - Show this menu")
	fmt.Println("  👋 quit (q)          - Exit CLI")
	fmt.Println()
}

func (c *CLI) showStatus() {
	fmt.Println("🔍 Checking service status...")
	
	// Check Elasticsearch
	esResp, err := c.client.Get("http://localhost:9200/_cluster/health")
	if err != nil {
		fmt.Printf("❌ Elasticsearch: Not responding (%v)\n", err)
	} else {
		defer esResp.Body.Close()
		if esResp.StatusCode == http.StatusOK {
			fmt.Printf("✅ Elasticsearch: Running (status %d)\n", esResp.StatusCode)
		} else {
			fmt.Printf("⚠️  Elasticsearch: Issues detected (status %d)\n", esResp.StatusCode)
		}
	}
	
	// Check Index Explorer
	indexResp, err := c.client.Get(c.APIURL + "/health")
	if err != nil {
		fmt.Printf("❌ Index Explorer: Not responding (%v)\n", err)
	} else {
		defer indexResp.Body.Close()
		if indexResp.StatusCode == http.StatusOK {
			fmt.Printf("✅ Index Explorer: Running (status %d)\n", indexResp.StatusCode)
		} else {
			fmt.Printf("⚠️  Index Explorer: Issues detected (status %d)\n", indexResp.StatusCode)
		}
	}
}

func (c *CLI) createIndex() {
	fmt.Println("🏗️  Create Write-Optimized Index")
	fmt.Println(strings.Repeat("-", 40))
	
	indexName := c.prompt("Index name")
	if indexName == "" {
		fmt.Println("❌ Index name is required")
		return
	}
	
	volume := c.promptWithOptions("Expected volume", []string{"low", "medium", "high"}, "medium")
	docSize := c.promptWithOptions("Expected document size", []string{"small", "medium", "large", "huge"}, "medium")
	ingestionRate := c.promptWithOptions("Ingestion rate", []string{"low", "medium", "high"}, "medium")
	textHeavy := c.promptBool("Text-heavy content", true)
	
	payload := map[string]interface{}{
		"index_name":        indexName,
		"expected_volume":   volume,
		"expected_doc_size": docSize,
		"ingestion_rate":    ingestionRate,
		"text_heavy":        textHeavy,
		"write_optimized":   true,
	}
	
	fmt.Printf("🚀 Creating index '%s'...\n", indexName)
	
	resp, err := c.makeRequest("POST", "/api/v1/indices/write-optimized", payload)
	if err != nil {
		fmt.Printf("❌ Failed to create index: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Index '%s' created successfully!\n", indexName)
	c.prettyPrintJSON(resp)
}

func (c *CLI) listIndices() {
	fmt.Println("📑 Listing all indices...")
	
	resp, err := c.makeRequest("GET", "/api/v1/indices", nil)
	if err != nil {
		fmt.Printf("❌ Failed to list indices: %v\n", err)
		return
	}
	
	c.prettyPrintJSON(resp)
}

func (c *CLI) optimizeIndex() {
	fmt.Println("⚡ Optimize Index for Write Performance")
	fmt.Println(strings.Repeat("-", 40))
	
	indexName := c.prompt("Index name to optimize")
	if indexName == "" {
		fmt.Println("❌ Index name is required")
		return
	}
	
	workload := c.promptWithOptions("Workload type", []string{"bulk_write", "mixed", "read_heavy"}, "bulk_write")
	corpusSize := c.promptWithOptions("Corpus size", []string{"small", "medium", "large", "huge"}, "medium")
	priority := c.promptWithOptions("Priority", []string{"write_throughput", "consistency", "balanced"}, "write_throughput")
	applyChanges := c.promptBool("Apply changes immediately", false)
	
	payload := map[string]interface{}{
		"optimize_for":  "write_throughput",
		"workload":      workload,
		"corpus_size":   corpusSize,
		"priority":      priority,
		"apply_changes": applyChanges,
	}
	
	fmt.Printf("⚡ Optimizing index '%s'...\n", indexName)
	
	endpoint := fmt.Sprintf("/api/v1/indices/%s/optimize", indexName)
	resp, err := c.makeRequest("POST", endpoint, payload)
	if err != nil {
		fmt.Printf("❌ Failed to optimize index: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Index optimization completed!\n")
	c.prettyPrintJSON(resp)
}

func (c *CLI) bulkIndex() {
	fmt.Println("📦 Bulk Index Documents")
	fmt.Println(strings.Repeat("-", 40))
	
	indexName := c.prompt("Index name")
	if indexName == "" {
		fmt.Println("❌ Index name is required")
		return
	}
	
	docCount, _ := strconv.Atoi(c.promptWithDefault("Number of documents", "100"))
	batchSize, _ := strconv.Atoi(c.promptWithDefault("Batch size", "500"))
	workers, _ := strconv.Atoi(c.promptWithDefault("Parallel workers", "4"))
	
	// Generate sample documents
	fmt.Printf("📝 Generating %d sample documents...\n", docCount)
	operations := c.generateSampleOperations(docCount)
	
	payload := map[string]interface{}{
		"operations":       operations,
		"optimize_for":     "write_throughput",
		"batch_size":       batchSize,
		"parallel_workers": workers,
		"error_tolerance":  "medium",
	}
	
	fmt.Printf("📦 Bulk indexing %d documents...\n", docCount)
	
	start := time.Now()
	endpoint := fmt.Sprintf("/api/v1/indices/%s/bulk", indexName)
	resp, err := c.makeRequest("POST", endpoint, payload)
	duration := time.Since(start)
	
	if err != nil {
		fmt.Printf("❌ Failed to bulk index: %v\n", err)
		return
	}
	
	docsPerSec := float64(docCount) / duration.Seconds()
	fmt.Printf("✅ Bulk indexing completed in %v (%.2f docs/sec)!\n", duration, docsPerSec)
	c.prettyPrintJSON(resp)
}

func (c *CLI) adaptiveBulk() {
	fmt.Println("🤖 Adaptive Bulk Indexing")
	fmt.Println(strings.Repeat("-", 40))
	
	indexName := c.prompt("Index name")
	if indexName == "" {
		fmt.Println("❌ Index name is required")
		return
	}
	
	docCount, _ := strconv.Atoi(c.promptWithDefault("Number of documents", "500"))
	targetThroughput := c.promptWithOptions("Target throughput", []string{"low", "medium", "high", "max"}, "high")
	
	// Generate mixed size documents
	fmt.Printf("📝 Generating %d mixed-size documents...\n", docCount)
	documents := c.generateMixedDocuments(docCount)
	
	payload := map[string]interface{}{
		"index_name":         indexName,
		"documents":          documents,
		"auto_batch_size":    true,
		"target_throughput":  targetThroughput,
		"error_tolerance":    "medium",
		"optimize_for":       "write_throughput",
	}
	
	fmt.Printf("🤖 Adaptive bulk indexing %d documents...\n", docCount)
	
	start := time.Now()
	resp, err := c.makeRequest("POST", "/api/v1/bulk/adaptive", payload)
	duration := time.Since(start)
	
	if err != nil {
		fmt.Printf("❌ Failed to adaptive bulk index: %v\n", err)
		return
	}
	
	docsPerSec := float64(docCount) / duration.Seconds()
	fmt.Printf("✅ Adaptive bulk indexing completed in %v (%.2f docs/sec)!\n", duration, docsPerSec)
	c.prettyPrintJSON(resp)
}

func (c *CLI) ndjsonImport() {
	fmt.Println("📄 NDJSON Import")
	fmt.Println(strings.Repeat("-", 40))
	
	indexName := c.prompt("Index name")
	if indexName == "" {
		fmt.Println("❌ Index name is required")
		return
	}
	
	docCount, _ := strconv.Atoi(c.promptWithDefault("Number of documents to generate", "200"))
	batchSize, _ := strconv.Atoi(c.promptWithDefault("Batch size", "500"))
	workers, _ := strconv.Atoi(c.promptWithDefault("Workers", "4"))
	
	// Generate NDJSON data
	fmt.Printf("📝 Generating NDJSON data with %d documents...\n", docCount)
	ndjsonData := c.generateNDJSONData(docCount)
	
	url := fmt.Sprintf("%s/api/v1/indices/%s/import/ndjson?batch_size=%d&workers=%d",
		c.APIURL, indexName, batchSize, workers)
	
	fmt.Printf("📄 Importing NDJSON data...\n")
	
	start := time.Now()
	resp, err := http.Post(url, "application/x-ndjson", strings.NewReader(ndjsonData))
	duration := time.Since(start)
	
	if err != nil {
		fmt.Printf("❌ Failed to import NDJSON: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	var result interface{}
	json.Unmarshal(body, &result)
	
	docsPerSec := float64(docCount) / duration.Seconds()
	fmt.Printf("✅ NDJSON import completed in %v (%.2f docs/sec)!\n", duration, docsPerSec)
	c.prettyPrintJSON(result)
}

func (c *CLI) showMetrics() {
	fmt.Println("📈 Write Performance Metrics")
	fmt.Println(strings.Repeat("-", 40))
	
	indexName := c.prompt("Index name")
	if indexName == "" {
		fmt.Println("❌ Index name is required")
		return
	}
	
	fmt.Printf("📊 Fetching metrics for '%s'...\n", indexName)
	
	endpoint := fmt.Sprintf("/api/v1/indices/%s/metrics/write-performance", indexName)
	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		fmt.Printf("❌ Failed to get metrics: %v\n", err)
		return
	}
	
	c.prettyPrintJSON(resp)
}

func (c *CLI) getRecommendations() {
	fmt.Println("💡 Optimization Recommendations")
	fmt.Println(strings.Repeat("-", 40))
	
	indexName := c.prompt("Index name")
	if indexName == "" {
		fmt.Println("❌ Index name is required")
		return
	}
	
	workload := c.promptWithOptions("Workload type", []string{"bulk_write", "mixed", "read_heavy"}, "bulk_write")
	corpusSize := c.promptWithOptions("Corpus size", []string{"small", "medium", "large", "huge"}, "medium")
	
	fmt.Printf("💡 Getting recommendations for '%s'...\n", indexName)
	
	endpoint := fmt.Sprintf("/api/v1/indices/%s/recommendations?workload=%s&corpus_size=%s", indexName, workload, corpusSize)
	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		fmt.Printf("❌ Failed to get recommendations: %v\n", err)
		return
	}
	
	c.prettyPrintJSON(resp)
}

func (c *CLI) performanceTest() {
	fmt.Println("🏃 Performance Test")
	fmt.Println(strings.Repeat("-", 40))
	
	fmt.Println("Select test type:")
	fmt.Println("  1. Quick (100 docs)")
	fmt.Println("  2. Medium (1,000 docs)")
	fmt.Println("  3. Heavy (10,000 docs)")
	fmt.Println("  4. Custom")
	
	choice := c.promptWithDefault("Choice", "1")
	
	var docCount, workers, batchSize int
	
	switch choice {
	case "1":
		docCount, workers, batchSize = 100, 4, 50
	case "2":
		docCount, workers, batchSize = 1000, 8, 500
	case "3":
		docCount, workers, batchSize = 10000, 16, 1000
	case "4":
		docCount, _ = strconv.Atoi(c.promptWithDefault("Document count", "500"))
		workers, _ = strconv.Atoi(c.promptWithDefault("Workers", "8"))
		batchSize, _ = strconv.Atoi(c.promptWithDefault("Batch size", "500"))
	default:
		fmt.Println("❌ Invalid choice")
		return
	}
	
	fmt.Printf("🚀 Running performance test with %d documents...\n", docCount)
	fmt.Printf("⚙️  Configuration: %d workers, batch size %d\n", workers, batchSize)
	
	// This would ideally call the separate performance test binary
	// For now, we'll show a simplified version
	fmt.Println("💡 For comprehensive performance testing, use:")
	fmt.Printf("   cd projects/index-explorer && go run cmd/perf-test/main.go\n")
	fmt.Printf("   Or: make perf-test\n")
}

func (c *CLI) showExamples() {
	fmt.Println("📖 API Examples")
	fmt.Println(strings.Repeat("-", 40))
	
	examples := map[string]string{
		"Create write-optimized index": `curl -X POST "http://localhost:8082/api/v1/indices/write-optimized" \
  -H "Content-Type: application/json" \
  -d '{"index_name":"my-corpus","expected_volume":"high","text_heavy":true}'`,
		
		"Bulk index documents": `curl -X POST "http://localhost:8082/api/v1/indices/my-corpus/bulk" \
  -H "Content-Type: application/json" \
  -d '{"operations":[{"action":"index","document":{"title":"Doc","content":"Content"}}]}'`,
		
		"Adaptive bulk indexing": `curl -X POST "http://localhost:8082/api/v1/bulk/adaptive" \
  -H "Content-Type: application/json" \
  -d '{"index_name":"my-corpus","documents":[{"title":"Doc","content":"Content"}]}'`,
		
		"Get performance metrics": `curl "http://localhost:8082/api/v1/indices/my-corpus/metrics/write-performance"`,
		
		"Import NDJSON": `curl -X POST "http://localhost:8082/api/v1/indices/my-corpus/import/ndjson?batch_size=1000" \
  -H "Content-Type: application/x-ndjson" \
  --data-binary @data.ndjson`,
	}
	
	for title, example := range examples {
		fmt.Printf("🔹 %s:\n", title)
		fmt.Printf("%s\n\n", example)
	}
}

func (c *CLI) clearScreen() {
	fmt.Print("\033[H\033[2J")
	fmt.Printf("🚀 Elasticsearch Index Explorer CLI v%s\n", version)
	fmt.Printf("Write-Optimized Operations Interface\n")
	fmt.Println(strings.Repeat("=", 50))
}

// Helper functions
func (c *CLI) prompt(message string) string {
	fmt.Printf("  %s: ", message)
	if c.scanner.Scan() {
		return strings.TrimSpace(c.scanner.Text())
	}
	return ""
}

func (c *CLI) promptWithDefault(message, defaultValue string) string {
	fmt.Printf("  %s [%s]: ", message, defaultValue)
	if c.scanner.Scan() {
		input := strings.TrimSpace(c.scanner.Text())
		if input == "" {
			return defaultValue
		}
		return input
	}
	return defaultValue
}

func (c *CLI) promptWithOptions(message string, options []string, defaultValue string) string {
	fmt.Printf("  %s (%s) [%s]: ", message, strings.Join(options, "/"), defaultValue)
	if c.scanner.Scan() {
		input := strings.TrimSpace(c.scanner.Text())
		if input == "" {
			return defaultValue
		}
		// Validate input
		for _, option := range options {
			if strings.EqualFold(input, option) {
				return strings.ToLower(input)
			}
		}
		fmt.Printf("  ⚠️  Invalid option. Using default: %s\n", defaultValue)
		return defaultValue
	}
	return defaultValue
}

func (c *CLI) promptBool(message string, defaultValue bool) bool {
	defaultStr := "N"
	if defaultValue {
		defaultStr = "Y"
	}
	
	fmt.Printf("  %s (Y/N) [%s]: ", message, defaultStr)
	if c.scanner.Scan() {
		input := strings.TrimSpace(strings.ToUpper(c.scanner.Text()))
		if input == "" {
			return defaultValue
		}
		return input == "Y" || input == "YES"
	}
	return defaultValue
}

func (c *CLI) makeRequest(method, endpoint string, payload interface{}) (interface{}, error) {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	}
	
	req, err := http.NewRequest(method, c.APIURL+endpoint, body)
	if err != nil {
		return nil, err
	}
	
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var result interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return string(respBody), nil
	}
	
	return result, nil
}

func (c *CLI) prettyPrintJSON(data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Error formatting response: %v\n", err)
		fmt.Printf("Raw response: %+v\n", data)
		return
	}
	fmt.Println(string(jsonData))
}

func (c *CLI) generateSampleOperations(count int) []map[string]interface{} {
	operations := make([]map[string]interface{}, count)
	
	for i := 0; i < count; i++ {
		operations[i] = map[string]interface{}{
			"action": "index",
			"document": map[string]interface{}{
				"id":        fmt.Sprintf("doc_%d", i),
				"title":     fmt.Sprintf("Sample Document %d", i),
				"content":   fmt.Sprintf("This is sample content for document %d. %s", i, strings.Repeat("Content ", 20)),
				"timestamp": time.Now().Format(time.RFC3339),
				"metadata": map[string]interface{}{
					"source":    "cli",
					"batch_id":  i / 100,
					"doc_type":  "sample",
				},
			},
		}
	}
	
	return operations
}

func (c *CLI) generateMixedDocuments(count int) []map[string]interface{} {
	documents := make([]map[string]interface{}, count)
	sizes := []string{"small", "medium", "large"}
	
	for i := 0; i < count; i++ {
		size := sizes[i%len(sizes)]
		var content string
		
		switch size {
		case "small":
			content = fmt.Sprintf("Small document %d", i)
		case "medium":
			content = fmt.Sprintf("Medium document %d. %s", i, strings.Repeat("Medium content ", 30))
		case "large":
			content = fmt.Sprintf("Large document %d. %s", i, strings.Repeat("Extensive content ", 100))
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

func (c *CLI) generateNDJSONData(count int) string {
	var builder strings.Builder
	
	for i := 0; i < count; i++ {
		content := fmt.Sprintf("NDJSON document %d. %s", i, strings.Repeat("NDJSON ", 20))
		
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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}