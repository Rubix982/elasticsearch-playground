package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/saif-islam/es-playground/projects/index-explorer/internal/models"
	"github.com/saif-islam/es-playground/projects/index-explorer/internal/services"
)

// DocumentHandler handles HTTP requests for document operations
type DocumentHandler struct {
	documentService *services.DocumentService
	logger          *zap.Logger
}

// NewDocumentHandler creates a new document handler
func NewDocumentHandler(documentService *services.DocumentService, logger *zap.Logger) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
		logger:          logger,
	}
}

// BulkIndex handles POST /api/v1/indices/:index/bulk
func (h *DocumentHandler) BulkIndex(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 300*time.Second) // 5 minutes for bulk operations
	defer cancel()

	indexName := c.Param("index")
	if indexName == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "Missing index name",
			Message:   "Index name is required",
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
		})
		return
	}

	var req models.BulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid bulk request", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "Invalid request",
			Message:   err.Error(),
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
		})
		return
	}

	// Set index name from URL if not provided in request
	if req.IndexName == "" {
		req.IndexName = indexName
	}

	h.logger.Info("Processing bulk index request",
		zap.String("index", req.IndexName),
		zap.Int("operations", len(req.Operations)),
		zap.String("optimize_for", req.OptimizeFor))

	response, err := h.documentService.BulkIndex(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to process bulk index",
			zap.String("index", req.IndexName),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:     "Failed to process bulk index",
			Message:   err.Error(),
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// BulkImportNDJSON handles POST /api/v1/indices/:index/import/ndjson
func (h *DocumentHandler) BulkImportNDJSON(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 600*time.Second) // 10 minutes for large imports
	defer cancel()

	indexName := c.Param("index")
	if indexName == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "Missing index name",
			Message:   "Index name is required",
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
		})
		return
	}

	// Parse query parameters for import options
	options := &services.BulkImportOptions{
		BatchSize:       1000, // Default
		ParallelWorkers: 8,    // Default
		ErrorTolerance:  "medium",
		GenerateIDs:     true,
	}

	if batchSizeStr := c.Query("batch_size"); batchSizeStr != "" {
		if batchSize, err := strconv.Atoi(batchSizeStr); err == nil && batchSize > 0 {
			options.BatchSize = batchSize
		}
	}

	if workersStr := c.Query("workers"); workersStr != "" {
		if workers, err := strconv.Atoi(workersStr); err == nil && workers > 0 {
			options.ParallelWorkers = workers
		}
	}

	if tolerance := c.Query("error_tolerance"); tolerance != "" {
		options.ErrorTolerance = tolerance
	}

	if generateIDs := c.Query("generate_ids"); generateIDs == "false" {
		options.GenerateIDs = false
	}

	h.logger.Info("Processing NDJSON bulk import",
		zap.String("index", indexName),
		zap.Int("batch_size", options.BatchSize),
		zap.Int("workers", options.ParallelWorkers))

	// Get request body as NDJSON
	body := c.Request.Body
	defer body.Close()

	response, err := h.documentService.BulkImportFromNDJSON(ctx, indexName, body, options)
	if err != nil {
		h.logger.Error("Failed to import NDJSON",
			zap.String("index", indexName),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:     "Failed to import NDJSON",
			Message:   err.Error(),
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "NDJSON import completed successfully",
		"index_name": indexName,
		"summary":    response.Summary,
		"request_id": c.GetString("request_id"),
		"timestamp":  time.Now(),
	})
}

// AdaptiveBulkIndex handles POST /api/v1/bulk/adaptive
func (h *DocumentHandler) AdaptiveBulkIndex(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 600*time.Second)
	defer cancel()

	var req struct {
		IndexName         string                   `json:"index_name" binding:"required"`
		Documents         []map[string]interface{} `json:"documents" binding:"required"`
		AutoBatchSize     bool                     `json:"auto_batch_size,omitempty"`
		TargetThroughput  string                   `json:"target_throughput,omitempty"` // max, high, medium, low
		ErrorTolerance    string                   `json:"error_tolerance,omitempty"`   // low, medium, high
		OptimizeFor       string                   `json:"optimize_for,omitempty"`      // write_throughput, consistency
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid adaptive bulk request", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "Invalid request",
			Message:   err.Error(),
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
		})
		return
	}

	h.logger.Info("Processing adaptive bulk index request",
		zap.String("index", req.IndexName),
		zap.Int("documents", len(req.Documents)),
		zap.Bool("auto_batch_size", req.AutoBatchSize),
		zap.String("target_throughput", req.TargetThroughput))

	// Convert documents to bulk operations
	operations := make([]models.BulkOperation, len(req.Documents))
	for i, doc := range req.Documents {
		operations[i] = models.BulkOperation{
			Action:   "index",
			Document: doc,
		}
	}

	// Create adaptive bulk request
	bulkReq := &models.BulkRequest{
		IndexName:      req.IndexName,
		Operations:     operations,
		OptimizeFor:    req.OptimizeFor,
		ErrorTolerance: req.ErrorTolerance,
	}

	// Set defaults
	if bulkReq.OptimizeFor == "" {
		bulkReq.OptimizeFor = "write_throughput"
	}
	if bulkReq.ErrorTolerance == "" {
		bulkReq.ErrorTolerance = "medium"
	}

	// Adaptive batch sizing based on target throughput
	if req.AutoBatchSize {
		bulkReq.BatchSize = h.calculateAdaptiveBatchSize(req.Documents, req.TargetThroughput)
		bulkReq.ParallelWorkers = h.calculateAdaptiveWorkers(len(req.Documents), req.TargetThroughput)
	}

	response, err := h.documentService.BulkIndex(ctx, bulkReq)
	if err != nil {
		h.logger.Error("Failed to process adaptive bulk index",
			zap.String("index", req.IndexName),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:     "Failed to process adaptive bulk index",
			Message:   err.Error(),
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
		})
		return
	}

	// Add adaptive parameters to response
	adaptiveResponse := gin.H{
		"bulk_response": response,
		"adaptive_settings": gin.H{
			"batch_size":       bulkReq.BatchSize,
			"parallel_workers": bulkReq.ParallelWorkers,
			"target_throughput": req.TargetThroughput,
		},
		"request_id": c.GetString("request_id"),
		"timestamp":  time.Now(),
	}

	c.JSON(http.StatusOK, adaptiveResponse)
}

// calculateAdaptiveBatchSize calculates optimal batch size based on target throughput
func (h *DocumentHandler) calculateAdaptiveBatchSize(documents []map[string]interface{}, targetThroughput string) int {
	// Estimate average document size from sample
	avgSize := h.estimateDocumentSize(documents)
	
	switch targetThroughput {
	case "max":
		if avgSize < 1024 { // < 1KB
			return 10000
		} else if avgSize < 10*1024 { // < 10KB
			return 2000
		} else if avgSize < 100*1024 { // < 100KB
			return 500
		} else {
			return 100
		}
	case "high":
		if avgSize < 1024 {
			return 5000
		} else if avgSize < 10*1024 {
			return 1000
		} else if avgSize < 100*1024 {
			return 300
		} else {
			return 50
		}
	case "medium":
		if avgSize < 1024 {
			return 2000
		} else if avgSize < 10*1024 {
			return 500
		} else if avgSize < 100*1024 {
			return 200
		} else {
			return 30
		}
	default: // low
		if avgSize < 1024 {
			return 1000
		} else if avgSize < 10*1024 {
			return 200
		} else if avgSize < 100*1024 {
			return 100
		} else {
			return 20
		}
	}
}

// calculateAdaptiveWorkers calculates optimal worker count
func (h *DocumentHandler) calculateAdaptiveWorkers(docCount int, targetThroughput string) int {
	baseWorkers := 4
	
	switch targetThroughput {
	case "max":
		baseWorkers = 16
	case "high":
		baseWorkers = 12
	case "medium":
		baseWorkers = 8
	default: // low
		baseWorkers = 4
	}

	// Adjust based on document count
	if docCount < 1000 {
		return baseWorkers / 2
	} else if docCount > 100000 {
		return baseWorkers * 2
	}

	return baseWorkers
}

// estimateDocumentSize estimates average document size in bytes
func (h *DocumentHandler) estimateDocumentSize(documents []map[string]interface{}) int {
	if len(documents) == 0 {
		return 1024 // Default assumption
	}

	sampleSize := 10
	if len(documents) < sampleSize {
		sampleSize = len(documents)
	}

	totalSize := 0
	for i := 0; i < sampleSize; i++ {
		// Rough JSON size estimation
		fieldCount := len(documents[i])
		avgFieldSize := 50 // Rough estimate
		totalSize += fieldCount * avgFieldSize
	}

	return totalSize / sampleSize
}

// IndexDocument handles POST /api/v1/indices/:index/documents (single document)
func (h *DocumentHandler) IndexDocument(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	indexName := c.Param("index")
	if indexName == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "Missing index name",
			Message:   "Index name is required",
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
		})
		return
	}

	// Get document ID from query parameter or generate one
	docID := c.Query("id")

	var document map[string]interface{}
	if err := c.ShouldBindJSON(&document); err != nil {
		h.logger.Error("Invalid document", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "Invalid document",
			Message:   err.Error(),
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
		})
		return
	}

	h.logger.Debug("Indexing single document",
		zap.String("index", indexName),
		zap.String("id", docID))

	response, err := h.documentService.IndexDocument(ctx, indexName, docID, document)
	if err != nil {
		h.logger.Error("Failed to index document",
			zap.String("index", indexName),
			zap.String("id", docID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:     "Failed to index document",
			Message:   err.Error(),
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetDocument handles GET /api/v1/indices/:index/documents/:id
func (h *DocumentHandler) GetDocument(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	indexName := c.Param("index")
	docID := c.Param("id")

	if indexName == "" || docID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "Missing parameters",
			Message:   "Index name and document ID are required",
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
		})
		return
	}

	document, err := h.documentService.GetDocument(ctx, indexName, docID)
	if err != nil {
		h.logger.Error("Failed to get document",
			zap.String("index", indexName),
			zap.String("id", docID),
			zap.Error(err))
		
		status := http.StatusInternalServerError
		if err.Error() == "document not found" {
			status = http.StatusNotFound
		}
		
		c.JSON(status, models.ErrorResponse{
			Error:     "Failed to get document",
			Message:   err.Error(),
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"index":      indexName,
		"id":         docID,
		"document":   document,
		"request_id": c.GetString("request_id"),
		"timestamp":  time.Now(),
	})
}

// UpdateDocument handles PUT /api/v1/indices/:index/documents/:id
func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	indexName := c.Param("index")
	docID := c.Param("id")

	if indexName == "" || docID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "Missing parameters",
			Message:   "Index name and document ID are required",
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
		})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid update document", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "Invalid update document",
			Message:   err.Error(),
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
		})
		return
	}

	response, err := h.documentService.UpdateDocument(ctx, indexName, docID, updates)
	if err != nil {
		h.logger.Error("Failed to update document",
			zap.String("index", indexName),
			zap.String("id", docID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:     "Failed to update document",
			Message:   err.Error(),
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteDocument handles DELETE /api/v1/indices/:index/documents/:id
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	indexName := c.Param("index")
	docID := c.Param("id")

	if indexName == "" || docID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "Missing parameters",
			Message:   "Index name and document ID are required",
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
		})
		return
	}

	response, err := h.documentService.DeleteDocument(ctx, indexName, docID)
	if err != nil {
		h.logger.Error("Failed to delete document",
			zap.String("index", indexName),
			zap.String("id", docID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:     "Failed to delete document",
			Message:   err.Error(),
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetBulkOperationStatus handles GET /api/v1/bulk/status
func (h *DocumentHandler) GetBulkOperationStatus(c *gin.Context) {
	// This would typically track ongoing bulk operations
	// For now, return a simple status
	c.JSON(http.StatusOK, gin.H{
		"message": "Bulk operations status endpoint",
		"status":  "operational",
		"active_operations": 0, // Would track actual operations
		"request_id": c.GetString("request_id"),
		"timestamp":  time.Now(),
	})
}

// GetWritePerformanceMetrics handles GET /api/v1/indices/:index/metrics/write-performance
func (h *DocumentHandler) GetWritePerformanceMetrics(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	indexName := c.Param("index")
	if indexName == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "Missing index name",
			Message:   "Index name is required",
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
		})
		return
	}

	metrics, err := h.documentService.GetWritePerformanceMetrics(ctx, indexName)
	if err != nil {
		h.logger.Error("Failed to get write performance metrics",
			zap.String("index", indexName),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:     "Failed to get write performance metrics",
			Message:   err.Error(),
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"index_name": indexName,
		"metrics":    metrics,
		"request_id": c.GetString("request_id"),
		"timestamp":  time.Now(),
	})
}