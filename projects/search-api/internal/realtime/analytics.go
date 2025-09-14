package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/saif-islam/es-playground/projects/search-api/internal/metrics"
	"github.com/saif-islam/es-playground/projects/search-api/internal/models"
)

// AnalyticsHub manages real-time analytics connections and data streams
type AnalyticsHub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	logger     *zap.Logger
	mu         sync.RWMutex
	
	// Analytics data
	searchMetrics    *SearchMetricsBuffer
	queryPatterns    *QueryPatternTracker
	performanceStats *PerformanceStatsTracker
}

// SearchEvent represents a real-time search event
type SearchEvent struct {
	Timestamp     time.Time              `json:"timestamp"`
	QueryID       string                 `json:"query_id"`
	Index         string                 `json:"index"`
	Query         string                 `json:"query"`
	QueryType     string                 `json:"query_type"`
	ResponseTime  time.Duration          `json:"response_time_ms"`
	ResultCount   int64                  `json:"result_count"`
	UserID        string                 `json:"user_id,omitempty"`
	SessionID     string                 `json:"session_id,omitempty"`
	Success       bool                   `json:"success"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	CacheHit      bool                   `json:"cache_hit"`
	ABTestVariant string                 `json:"ab_test_variant,omitempty"`
	TraceID       string                 `json:"trace_id,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// RealTimeMetrics represents aggregated real-time metrics
type RealTimeMetrics struct {
	Timestamp        time.Time            `json:"timestamp"`
	TotalSearches    int64                `json:"total_searches"`
	SearchesPerSec   float64              `json:"searches_per_sec"`
	AvgResponseTime  float64              `json:"avg_response_time_ms"`
	ErrorRate        float64              `json:"error_rate_percent"`
	CacheHitRate     float64              `json:"cache_hit_rate_percent"`
	TopQueries       []QueryStats         `json:"top_queries"`
	TopIndices       []IndexStats         `json:"top_indices"`
	PerformanceAlerts []PerformanceAlert  `json:"performance_alerts"`
	ABTestResults    map[string]ABMetrics `json:"ab_test_results"`
}

// QueryStats represents query performance statistics
type QueryStats struct {
	Query         string  `json:"query"`
	Count         int64   `json:"count"`
	AvgTime       float64 `json:"avg_time_ms"`
	ErrorRate     float64 `json:"error_rate"`
	Trend         string  `json:"trend"` // "up", "down", "stable"
}

// IndexStats represents index usage statistics
type IndexStats struct {
	Index         string  `json:"index"`
	SearchCount   int64   `json:"search_count"`
	AvgTime       float64 `json:"avg_time_ms"`
	ErrorRate     float64 `json:"error_rate"`
}

// PerformanceAlert represents a performance issue
type PerformanceAlert struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	QueryID     string    `json:"query_id,omitempty"`
	Metric      string    `json:"metric"`
	Value       float64   `json:"value"`
	Threshold   float64   `json:"threshold"`
}

// ABMetrics represents A/B test performance metrics
type ABMetrics struct {
	Variant       string  `json:"variant"`
	RequestCount  int64   `json:"request_count"`
	AvgTime       float64 `json:"avg_time_ms"`
	ErrorRate     float64 `json:"error_rate"`
	SuccessRate   float64 `json:"success_rate"`
	Confidence    float64 `json:"confidence"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// NewAnalyticsHub creates a new analytics hub
func NewAnalyticsHub(logger *zap.Logger) *AnalyticsHub {
	hub := &AnalyticsHub{
		clients:          make(map[*websocket.Conn]bool),
		broadcast:        make(chan []byte, 256),
		register:         make(chan *websocket.Conn),
		unregister:       make(chan *websocket.Conn),
		logger:           logger,
		searchMetrics:    NewSearchMetricsBuffer(1000), // Keep last 1000 searches
		queryPatterns:    NewQueryPatternTracker(),
		performanceStats: NewPerformanceStatsTracker(),
	}
	
	go hub.run()
	go hub.generateMetrics()
	
	return hub
}

// run handles the main hub loop
func (h *AnalyticsHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.logger.Info("Client connected to analytics stream",
				zap.Int("total_clients", len(h.clients)))
			
			// Send initial metrics to new client
			go h.sendInitialMetrics(client)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()
			h.logger.Info("Client disconnected from analytics stream",
				zap.Int("total_clients", len(h.clients)))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.WriteMessage(websocket.TextMessage, message):
				default:
					delete(h.clients, client)
					client.Close()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// RecordSearchEvent records a search event for real-time analytics
func (h *AnalyticsHub) RecordSearchEvent(event SearchEvent) {
	// Add to metrics buffer
	h.searchMetrics.Add(event)
	
	// Update query patterns
	h.queryPatterns.Track(event.Query, event.ResponseTime, event.Success)
	
	// Update performance stats
	h.performanceStats.Update(event)
	
	// Check for performance alerts
	alerts := h.checkPerformanceAlerts(event)
	if len(alerts) > 0 {
		for _, alert := range alerts {
			h.broadcastAlert(alert)
		}
	}
	
	// Broadcast event to all connected clients
	eventJSON, err := json.Marshal(map[string]interface{}{
		"type": "search_event",
		"data": event,
	})
	if err != nil {
		h.logger.Error("Failed to marshal search event", zap.Error(err))
		return
	}
	
	select {
	case h.broadcast <- eventJSON:
	default:
		h.logger.Warn("Broadcast channel full, dropping search event")
	}
}

// generateMetrics generates and broadcasts aggregated metrics every second
func (h *AnalyticsHub) generateMetrics() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		metrics := h.generateRealTimeMetrics()
		
		metricsJSON, err := json.Marshal(map[string]interface{}{
			"type": "metrics_update",
			"data": metrics,
		})
		if err != nil {
			h.logger.Error("Failed to marshal metrics", zap.Error(err))
			continue
		}
		
		select {
		case h.broadcast <- metricsJSON:
		default:
			// Channel full, skip this update
		}
	}
}

// generateRealTimeMetrics creates current aggregated metrics
func (h *AnalyticsHub) generateRealTimeMetrics() RealTimeMetrics {
	now := time.Now()
	
	// Get metrics from buffer (last 60 seconds)
	recentEvents := h.searchMetrics.GetRecent(60 * time.Second)
	
	totalSearches := int64(len(recentEvents))
	searchesPerSec := float64(totalSearches) / 60.0
	
	var totalTime time.Duration
	var errorCount int64
	var cacheHits int64
	
	queryStats := make(map[string]*QueryStats)
	indexStats := make(map[string]*IndexStats)
	abStats := make(map[string]*ABMetrics)
	
	for _, event := range recentEvents {
		totalTime += event.ResponseTime
		
		if !event.Success {
			errorCount++
		}
		
		if event.CacheHit {
			cacheHits++
		}
		
		// Track query stats
		if stat, exists := queryStats[event.Query]; exists {
			stat.Count++
			stat.AvgTime = (stat.AvgTime*float64(stat.Count-1) + float64(event.ResponseTime.Milliseconds())) / float64(stat.Count)
			if !event.Success {
				stat.ErrorRate = (stat.ErrorRate*float64(stat.Count-1) + 1) / float64(stat.Count)
			}
		} else {
			errorRate := 0.0
			if !event.Success {
				errorRate = 1.0
			}
			queryStats[event.Query] = &QueryStats{
				Query:     event.Query,
				Count:     1,
				AvgTime:   float64(event.ResponseTime.Milliseconds()),
				ErrorRate: errorRate,
				Trend:     "stable",
			}
		}
		
		// Track index stats
		if stat, exists := indexStats[event.Index]; exists {
			stat.SearchCount++
			stat.AvgTime = (stat.AvgTime*float64(stat.SearchCount-1) + float64(event.ResponseTime.Milliseconds())) / float64(stat.SearchCount)
		} else {
			indexStats[event.Index] = &IndexStats{
				Index:       event.Index,
				SearchCount: 1,
				AvgTime:     float64(event.ResponseTime.Milliseconds()),
				ErrorRate:   0.0,
			}
		}
		
		// Track A/B test stats
		if event.ABTestVariant != "" {
			if stat, exists := abStats[event.ABTestVariant]; exists {
				stat.RequestCount++
				stat.AvgTime = (stat.AvgTime*float64(stat.RequestCount-1) + float64(event.ResponseTime.Milliseconds())) / float64(stat.RequestCount)
				if event.Success {
					stat.SuccessRate = (stat.SuccessRate*float64(stat.RequestCount-1) + 1) / float64(stat.RequestCount)
				}
			} else {
				successRate := 0.0
				if event.Success {
					successRate = 1.0
				}
				abStats[event.ABTestVariant] = &ABMetrics{
					Variant:      event.ABTestVariant,
					RequestCount: 1,
					AvgTime:      float64(event.ResponseTime.Milliseconds()),
					ErrorRate:    0.0,
					SuccessRate:  successRate,
					Confidence:   0.0, // Would need statistical calculation
				}
			}
		}
	}
	
	// Calculate averages
	avgResponseTime := 0.0
	if totalSearches > 0 {
		avgResponseTime = float64(totalTime.Milliseconds()) / float64(totalSearches)
	}
	
	errorRate := 0.0
	if totalSearches > 0 {
		errorRate = float64(errorCount) / float64(totalSearches) * 100
	}
	
	cacheHitRate := 0.0
	if totalSearches > 0 {
		cacheHitRate = float64(cacheHits) / float64(totalSearches) * 100
	}
	
	// Convert maps to slices (top 10)
	topQueries := make([]QueryStats, 0, 10)
	for _, stat := range queryStats {
		topQueries = append(topQueries, *stat)
	}
	
	topIndices := make([]IndexStats, 0, 10)
	for _, stat := range indexStats {
		topIndices = append(topIndices, *stat)
	}
	
	// Get performance alerts
	alerts := h.performanceStats.GetRecentAlerts(5 * time.Minute)
	
	return RealTimeMetrics{
		Timestamp:        now,
		TotalSearches:    totalSearches,
		SearchesPerSec:   searchesPerSec,
		AvgResponseTime:  avgResponseTime,
		ErrorRate:        errorRate,
		CacheHitRate:     cacheHitRate,
		TopQueries:       topQueries,
		TopIndices:       topIndices,
		PerformanceAlerts: alerts,
		ABTestResults:    abStats,
	}
}

// checkPerformanceAlerts checks for performance issues
func (h *AnalyticsHub) checkPerformanceAlerts(event SearchEvent) []PerformanceAlert {
	var alerts []PerformanceAlert
	
	// Slow query alert
	if event.ResponseTime > 2*time.Second {
		alerts = append(alerts, PerformanceAlert{
			Type:      "slow_query",
			Severity:  "warning",
			Message:   fmt.Sprintf("Slow query detected: %s", event.Query),
			Timestamp: event.Timestamp,
			QueryID:   event.QueryID,
			Metric:    "response_time",
			Value:     float64(event.ResponseTime.Milliseconds()),
			Threshold: 2000,
		})
	}
	
	// Error alert
	if !event.Success {
		alerts = append(alerts, PerformanceAlert{
			Type:      "query_error",
			Severity:  "error",
			Message:   fmt.Sprintf("Query failed: %s", event.ErrorMessage),
			Timestamp: event.Timestamp,
			QueryID:   event.QueryID,
			Metric:    "success_rate",
			Value:     0,
			Threshold: 1,
		})
	}
	
	return alerts
}

// broadcastAlert sends a performance alert to all clients
func (h *AnalyticsHub) broadcastAlert(alert PerformanceAlert) {
	alertJSON, err := json.Marshal(map[string]interface{}{
		"type": "performance_alert",
		"data": alert,
	})
	if err != nil {
		h.logger.Error("Failed to marshal alert", zap.Error(err))
		return
	}
	
	select {
	case h.broadcast <- alertJSON:
	default:
		h.logger.Warn("Broadcast channel full, dropping alert")
	}
}

// sendInitialMetrics sends initial metrics to a newly connected client
func (h *AnalyticsHub) sendInitialMetrics(client *websocket.Conn) {
	metrics := h.generateRealTimeMetrics()
	
	initialData, err := json.Marshal(map[string]interface{}{
		"type": "initial_metrics",
		"data": metrics,
	})
	if err != nil {
		h.logger.Error("Failed to marshal initial metrics", zap.Error(err))
		return
	}
	
	if err := client.WriteMessage(websocket.TextMessage, initialData); err != nil {
		h.logger.Error("Failed to send initial metrics", zap.Error(err))
	}
}

// HandleWebSocket handles WebSocket connections for real-time analytics
func (h *AnalyticsHub) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade WebSocket connection", zap.Error(err))
		return
	}
	
	h.register <- conn
	
	// Keep connection alive
	defer func() {
		h.unregister <- conn
		conn.Close()
	}()
	
	// Read messages from client (for ping/pong)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("WebSocket error", zap.Error(err))
			}
			break
		}
	}
}

// GetConnectedClients returns the number of connected clients
func (h *AnalyticsHub) GetConnectedClients() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}