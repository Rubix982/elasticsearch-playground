package realtime

import (
	"container/ring"
	"sync"
	"time"
)

// SearchMetricsBuffer is a thread-safe circular buffer for search events
type SearchMetricsBuffer struct {
	buffer   *ring.Ring
	capacity int
	mu       sync.RWMutex
}

// NewSearchMetricsBuffer creates a new search metrics buffer
func NewSearchMetricsBuffer(capacity int) *SearchMetricsBuffer {
	return &SearchMetricsBuffer{
		buffer:   ring.New(capacity),
		capacity: capacity,
	}
}

// Add adds a search event to the buffer
func (b *SearchMetricsBuffer) Add(event SearchEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	b.buffer.Value = event
	b.buffer = b.buffer.Next()
}

// GetRecent returns all events from the last duration
func (b *SearchMetricsBuffer) GetRecent(duration time.Duration) []SearchEvent {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	var events []SearchEvent
	cutoff := time.Now().Add(-duration)
	
	b.buffer.Do(func(v interface{}) {
		if v != nil {
			if event, ok := v.(SearchEvent); ok {
				if event.Timestamp.After(cutoff) {
					events = append(events, event)
				}
			}
		}
	})
	
	return events
}

// GetAll returns all events in the buffer
func (b *SearchMetricsBuffer) GetAll() []SearchEvent {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	var events []SearchEvent
	
	b.buffer.Do(func(v interface{}) {
		if v != nil {
			if event, ok := v.(SearchEvent); ok {
				events = append(events, event)
			}
		}
	})
	
	return events
}

// QueryPatternTracker tracks query patterns and performance
type QueryPatternTracker struct {
	patterns map[string]*QueryPattern
	mu       sync.RWMutex
}

// QueryPattern represents a query pattern with statistics
type QueryPattern struct {
	Query         string
	Count         int64
	TotalTime     time.Duration
	ErrorCount    int64
	LastSeen      time.Time
	AvgTime       time.Duration
	ErrorRate     float64
	Trend         string
	Variants      map[string]*QueryVariant
}

// QueryVariant represents a variant of a query pattern
type QueryVariant struct {
	Query       string
	Count       int64
	TotalTime   time.Duration
	ErrorCount  int64
	Performance float64 // Performance score
}

// NewQueryPatternTracker creates a new query pattern tracker
func NewQueryPatternTracker() *QueryPatternTracker {
	return &QueryPatternTracker{
		patterns: make(map[string]*QueryPattern),
	}
}

// Track tracks a query execution
func (t *QueryPatternTracker) Track(query string, responseTime time.Duration, success bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	pattern, exists := t.patterns[query]
	if !exists {
		pattern = &QueryPattern{
			Query:    query,
			Count:    0,
			LastSeen: time.Now(),
			Variants: make(map[string]*QueryVariant),
		}
		t.patterns[query] = pattern
	}
	
	pattern.Count++
	pattern.TotalTime += responseTime
	pattern.LastSeen = time.Now()
	pattern.AvgTime = time.Duration(int64(pattern.TotalTime) / pattern.Count)
	
	if !success {
		pattern.ErrorCount++
	}
	
	pattern.ErrorRate = float64(pattern.ErrorCount) / float64(pattern.Count)
	
	// Calculate trend (simplified)
	if pattern.Count > 10 {
		recentAvg := float64(responseTime.Milliseconds())
		historicalAvg := float64(pattern.AvgTime.Milliseconds())
		
		if recentAvg > historicalAvg*1.2 {
			pattern.Trend = "degrading"
		} else if recentAvg < historicalAvg*0.8 {
			pattern.Trend = "improving"
		} else {
			pattern.Trend = "stable"
		}
	}
}

// GetTopPatterns returns the most frequent query patterns
func (t *QueryPatternTracker) GetTopPatterns(limit int) []*QueryPattern {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	patterns := make([]*QueryPattern, 0, len(t.patterns))
	for _, pattern := range t.patterns {
		patterns = append(patterns, pattern)
	}
	
	// Simple sort by count (in production, use proper sorting)
	for i := 0; i < len(patterns)-1; i++ {
		for j := i + 1; j < len(patterns); j++ {
			if patterns[i].Count < patterns[j].Count {
				patterns[i], patterns[j] = patterns[j], patterns[i]
			}
		}
	}
	
	if limit > len(patterns) {
		limit = len(patterns)
	}
	
	return patterns[:limit]
}

// PerformanceStatsTracker tracks performance statistics
type PerformanceStatsTracker struct {
	alerts       []PerformanceAlert
	alertsBuffer *ring.Ring
	mu           sync.RWMutex
}

// NewPerformanceStatsTracker creates a new performance stats tracker
func NewPerformanceStatsTracker() *PerformanceStatsTracker {
	return &PerformanceStatsTracker{
		alerts:       make([]PerformanceAlert, 0),
		alertsBuffer: ring.New(100), // Keep last 100 alerts
	}
}

// Update updates performance statistics with a new search event
func (p *PerformanceStatsTracker) Update(event SearchEvent) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Check for performance issues
	var alerts []PerformanceAlert
	
	// Slow query detection
	if event.ResponseTime > 1*time.Second {
		alert := PerformanceAlert{
			Type:      "slow_query",
			Severity:  determineSeverity(event.ResponseTime),
			Message:   "Slow query detected",
			Timestamp: event.Timestamp,
			QueryID:   event.QueryID,
			Metric:    "response_time",
			Value:     float64(event.ResponseTime.Milliseconds()),
			Threshold: 1000,
		}
		alerts = append(alerts, alert)
	}
	
	// Error detection
	if !event.Success {
		alert := PerformanceAlert{
			Type:      "query_error",
			Severity:  "error",
			Message:   "Query execution failed",
			Timestamp: event.Timestamp,
			QueryID:   event.QueryID,
			Metric:    "success_rate",
			Value:     0,
			Threshold: 1,
		}
		alerts = append(alerts, alert)
	}
	
	// Add alerts to buffer
	for _, alert := range alerts {
		p.alertsBuffer.Value = alert
		p.alertsBuffer = p.alertsBuffer.Next()
		p.alerts = append(p.alerts, alert)
	}
	
	// Keep only recent alerts in memory
	if len(p.alerts) > 1000 {
		p.alerts = p.alerts[len(p.alerts)-1000:]
	}
}

// GetRecentAlerts returns alerts from the last duration
func (p *PerformanceStatsTracker) GetRecentAlerts(duration time.Duration) []PerformanceAlert {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	var recentAlerts []PerformanceAlert
	cutoff := time.Now().Add(-duration)
	
	for _, alert := range p.alerts {
		if alert.Timestamp.After(cutoff) {
			recentAlerts = append(recentAlerts, alert)
		}
	}
	
	return recentAlerts
}

// determineSeverity determines alert severity based on response time
func determineSeverity(responseTime time.Duration) string {
	if responseTime > 5*time.Second {
		return "critical"
	} else if responseTime > 2*time.Second {
		return "warning"
	}
	return "info"
}

// TimeSeriesPoint represents a point in time series data
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// TimeSeriesBuffer maintains time series data
type TimeSeriesBuffer struct {
	points   []TimeSeriesPoint
	capacity int
	mu       sync.RWMutex
}

// NewTimeSeriesBuffer creates a new time series buffer
func NewTimeSeriesBuffer(capacity int) *TimeSeriesBuffer {
	return &TimeSeriesBuffer{
		points:   make([]TimeSeriesPoint, 0, capacity),
		capacity: capacity,
	}
}

// Add adds a point to the time series
func (t *TimeSeriesBuffer) Add(timestamp time.Time, value float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	point := TimeSeriesPoint{
		Timestamp: timestamp,
		Value:     value,
	}
	
	t.points = append(t.points, point)
	
	// Keep only the most recent points
	if len(t.points) > t.capacity {
		t.points = t.points[1:]
	}
}

// GetRecent returns points from the last duration
func (t *TimeSeriesBuffer) GetRecent(duration time.Duration) []TimeSeriesPoint {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	cutoff := time.Now().Add(-duration)
	var recent []TimeSeriesPoint
	
	for _, point := range t.points {
		if point.Timestamp.After(cutoff) {
			recent = append(recent, point)
		}
	}
	
	return recent
}

// GetAll returns all points in the buffer
func (t *TimeSeriesBuffer) GetAll() []TimeSeriesPoint {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	// Return a copy to avoid race conditions
	points := make([]TimeSeriesPoint, len(t.points))
	copy(points, t.points)
	return points
}