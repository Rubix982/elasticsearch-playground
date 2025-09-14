package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/saif-islam/es-playground/projects/cluster-explorer/internal/models"
	"github.com/saif-islam/es-playground/projects/cluster-explorer/internal/services"
)

// ClusterHandler handles HTTP requests for cluster operations
type ClusterHandler struct {
	clusterService *services.ClusterService
	logger         *zap.Logger
}

// NewClusterHandler creates a new cluster handler
func NewClusterHandler(clusterService *services.ClusterService, logger *zap.Logger) *ClusterHandler {
	return &ClusterHandler{
		clusterService: clusterService,
		logger:         logger,
	}
}

// GetClusterInfo handles GET /api/v1/cluster/info
func (h *ClusterHandler) GetClusterInfo(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	info, err := h.clusterService.GetClusterInfo(ctx)
	if err != nil {
		h.logger.Error("Failed to get cluster info", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to retrieve cluster information",
			"message":    err.Error(),
			"request_id": c.GetString("request_id"),
			"timestamp":  time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, info)
}

// GetClusterHealth handles GET /api/v1/cluster/health
func (h *ClusterHandler) GetClusterHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	health, err := h.clusterService.GetClusterHealth(ctx)
	if err != nil {
		h.logger.Error("Failed to get cluster health", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to retrieve cluster health",
			"message":    err.Error(),
			"request_id": c.GetString("request_id"),
			"timestamp":  time.Now(),
		})
		return
	}

	// Set appropriate HTTP status based on cluster health
	status := http.StatusOK
	switch health.Status {
	case "red":
		status = http.StatusServiceUnavailable
	case "yellow":
		status = http.StatusPartialContent
	}

	c.JSON(status, gin.H{
		"health":     health,
		"request_id": c.GetString("request_id"),
		"timestamp":  time.Now(),
	})
}

// GetClusterState handles GET /api/v1/cluster/state
func (h *ClusterHandler) GetClusterState(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	state, err := h.clusterService.GetClusterState(ctx)
	if err != nil {
		h.logger.Error("Failed to get cluster state", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to retrieve cluster state",
			"message":    err.Error(),
			"request_id": c.GetString("request_id"),
			"timestamp":  time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"state":      state,
		"request_id": c.GetString("request_id"),
		"timestamp":  time.Now(),
	})
}

// GetClusterStats handles GET /api/v1/cluster/stats
func (h *ClusterHandler) GetClusterStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	stats, err := h.clusterService.GetClusterStats(ctx)
	if err != nil {
		h.logger.Error("Failed to get cluster stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to retrieve cluster statistics",
			"message":    err.Error(),
			"request_id": c.GetString("request_id"),
			"timestamp":  time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats":      stats,
		"request_id": c.GetString("request_id"),
		"timestamp":  time.Now(),
	})
}

// GetNodes handles GET /api/v1/cluster/nodes
func (h *ClusterHandler) GetNodes(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	nodes, err := h.clusterService.GetNodesInfo(ctx)
	if err != nil {
		h.logger.Error("Failed to get nodes info", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to retrieve nodes information",
			"message":    err.Error(),
			"request_id": c.GetString("request_id"),
			"timestamp":  time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes":      nodes,
		"count":      len(nodes),
		"request_id": c.GetString("request_id"),
		"timestamp":  time.Now(),
	})
}

// GetIndices handles GET /api/v1/cluster/indices
func (h *ClusterHandler) GetIndices(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	indices, err := h.clusterService.GetIndicesInfo(ctx)
	if err != nil {
		h.logger.Error("Failed to get indices info", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to retrieve indices information",
			"message":    err.Error(),
			"request_id": c.GetString("request_id"),
			"timestamp":  time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"indices":    indices,
		"count":      len(indices),
		"request_id": c.GetString("request_id"),
		"timestamp":  time.Now(),
	})
}

// GetShardAllocation handles GET /api/v1/cluster/shards
func (h *ClusterHandler) GetShardAllocation(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	allocation, err := h.clusterService.GetShardAllocation(ctx)
	if err != nil {
		h.logger.Error("Failed to get shard allocation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to retrieve shard allocation",
			"message":    err.Error(),
			"request_id": c.GetString("request_id"),
			"timestamp":  time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"allocation": allocation,
		"request_id": c.GetString("request_id"),
		"timestamp":  time.Now(),
	})
}

// GetPerformanceMetrics handles GET /api/v1/cluster/performance
func (h *ClusterHandler) GetPerformanceMetrics(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	metrics, err := h.clusterService.GetPerformanceMetrics(ctx)
	if err != nil {
		h.logger.Error("Failed to get performance metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to retrieve performance metrics",
			"message":    err.Error(),
			"request_id": c.GetString("request_id"),
			"timestamp":  time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics":    metrics,
		"request_id": c.GetString("request_id"),
		"timestamp":  time.Now(),
	})
}

// GetHotThreads handles GET /api/v1/cluster/nodes/:nodeId/hot-threads
func (h *ClusterHandler) GetHotThreads(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	nodeID := c.Param("nodeId")
	
	hotThreads, err := h.clusterService.GetHotThreads(ctx, nodeID)
	if err != nil {
		h.logger.Error("Failed to get hot threads", 
			zap.String("node_id", nodeID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to retrieve hot threads",
			"message":    err.Error(),
			"request_id": c.GetString("request_id"),
			"timestamp":  time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"node_id":     nodeID,
		"hot_threads": hotThreads,
		"request_id":  c.GetString("request_id"),
		"timestamp":   time.Now(),
	})
}

// MonitorHealth handles GET /api/v1/cluster/monitor/health
func (h *ClusterHandler) MonitorHealth(c *gin.Context) {
	// Parse interval parameter
	intervalStr := c.DefaultQuery("interval", "5s")
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid interval format",
			"message":    "Use format like '5s', '1m', '10s'",
			"request_id": c.GetString("request_id"),
			"timestamp":  time.Now(),
		})
		return
	}

	// Set minimum interval to prevent too frequent requests
	if interval < time.Second {
		interval = time.Second
	}

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	healthCh, err := h.clusterService.MonitorClusterHealth(ctx, interval)
	if err != nil {
		h.logger.Error("Failed to start health monitoring", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to start health monitoring",
			"message":    err.Error(),
			"request_id": c.GetString("request_id"),
			"timestamp":  time.Now(),
		})
		return
	}

	// Set up Server-Sent Events
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Stream health updates
	for health := range healthCh {
		select {
		case <-c.Request.Context().Done():
			return
		default:
			c.SSEvent("health", gin.H{
				"health":     health,
				"request_id": c.GetString("request_id"),
				"timestamp":  time.Now(),
			})
			c.Writer.Flush()
		}
	}
}

// GetClusterSettings handles GET /api/v1/cluster/settings
func (h *ClusterHandler) GetClusterSettings(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	settings, err := h.clusterService.GetClusterSettings(ctx)
	if err != nil {
		h.logger.Error("Failed to get cluster settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to retrieve cluster settings",
			"message":    err.Error(),
			"request_id": c.GetString("request_id"),
			"timestamp":  time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"settings":   settings,
		"request_id": c.GetString("request_id"),
		"timestamp":  time.Now(),
	})
}

// UpdateClusterSettings handles PUT /api/v1/cluster/settings
func (h *ClusterHandler) UpdateClusterSettings(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Parse request body
	var request struct {
		Settings   map[string]interface{} `json:"settings" binding:"required"`
		Persistent bool                   `json:"persistent"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid request body",
			"message":    err.Error(),
			"request_id": c.GetString("request_id"),
			"timestamp":  time.Now(),
		})
		return
	}

	err := h.clusterService.UpdateClusterSettings(ctx, request.Settings, request.Persistent)
	if err != nil {
		h.logger.Error("Failed to update cluster settings", 
			zap.Any("settings", request.Settings),
			zap.Bool("persistent", request.Persistent),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to update cluster settings",
			"message":    err.Error(),
			"request_id": c.GetString("request_id"),
			"timestamp":  time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Cluster settings updated successfully",
		"settings":   request.Settings,
		"persistent": request.Persistent,
		"request_id": c.GetString("request_id"),
		"timestamp":  time.Now(),
	})
}

// GetClusterOverview handles GET /api/v1/cluster/overview
func (h *ClusterHandler) GetClusterOverview(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()

	// Get basic cluster information quickly
	health, err := h.clusterService.GetClusterHealth(ctx)
	if err != nil {
		h.logger.Error("Failed to get cluster health for overview", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to retrieve cluster overview",
			"message":    err.Error(),
			"request_id": c.GetString("request_id"),
			"timestamp":  time.Now(),
		})
		return
	}

	nodes, err := h.clusterService.GetNodesInfo(ctx)
	if err != nil {
		h.logger.Error("Failed to get nodes info for overview", zap.Error(err))
		// Continue with what we have
		nodes = []models.NodeInfo{}
	}

	indices, err := h.clusterService.GetIndicesInfo(ctx)
	if err != nil {
		h.logger.Error("Failed to get indices info for overview", zap.Error(err))
		// Continue with what we have
		indices = []models.IndexInfo{}
	}

	// Create overview summary
	overview := gin.H{
		"cluster": gin.H{
			"name":   health.ClusterName,
			"status": health.Status,
			"nodes": gin.H{
				"total": health.NumberOfNodes,
				"data":  health.NumberOfDataNodes,
			},
			"shards": gin.H{
				"active":        health.ActiveShards,
				"primary":       health.ActivePrimaryShards,
				"relocating":    health.RelocatingShards,
				"initializing":  health.InitializingShards,
				"unassigned":    health.UnassignedShards,
			},
		},
		"nodes": gin.H{
			"count": len(nodes),
			"roles": h.summarizeNodeRoles(nodes),
		},
		"indices": gin.H{
			"count":  len(indices),
			"health": h.summarizeIndexHealth(indices),
		},
		"request_id": c.GetString("request_id"),
		"timestamp":  time.Now(),
	}

	c.JSON(http.StatusOK, overview)
}

// Helper function to summarize node roles
func (h *ClusterHandler) summarizeNodeRoles(nodes []models.NodeInfo) map[string]int {
	roles := make(map[string]int)
	
	for _, node := range nodes {
		for _, role := range node.Roles {
			roles[role]++
		}
	}
	
	return roles
}

// Helper function to summarize index health
func (h *ClusterHandler) summarizeIndexHealth(indices []models.IndexInfo) map[string]int {
	health := map[string]int{
		"green":  0,
		"yellow": 0,
		"red":    0,
	}
	
	for _, index := range indices {
		if status, exists := health[index.Health]; exists {
			health[index.Health] = status + 1
		}
	}
	
	return health
}