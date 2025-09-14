package abtesting

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/saif-islam/es-playground/projects/search-api/internal/models"
)

// ABTestFramework manages A/B testing experiments for search queries
type ABTestFramework struct {
	experiments map[string]*Experiment
	mu          sync.RWMutex
	logger      *zap.Logger
	
	// Traffic splitting configuration
	defaultTrafficSplit float64
	minSampleSize      int
	maxExperimentAge   time.Duration
}

// Experiment represents an A/B test experiment
type Experiment struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Status      ExperimentStatus       `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	EndedAt     *time.Time             `json:"ended_at,omitempty"`
	
	// Traffic allocation
	TrafficAllocation float64 `json:"traffic_allocation"` // 0.0 to 1.0
	
	// Variants
	ControlVariant    *Variant            `json:"control_variant"`
	TreatmentVariants map[string]*Variant `json:"treatment_variants"`
	
	// Targeting
	Targeting ExperimentTargeting `json:"targeting"`
	
	// Metrics
	PrimaryMetric   string   `json:"primary_metric"`
	SecondaryMetrics []string `json:"secondary_metrics"`
	
	// Results
	Results ExperimentResults `json:"results"`
	
	// Configuration
	MinSampleSize    int           `json:"min_sample_size"`
	MaxDuration      time.Duration `json:"max_duration"`
	SignificanceLevel float64      `json:"significance_level"`
	
	mu sync.RWMutex
}

// Variant represents a test variant (control or treatment)
type Variant struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Weight      float64                `json:"weight"` // Traffic weight (0.0 to 1.0)
	
	// Query modifications
	QueryModifications QueryModifications `json:"query_modifications"`
	
	// Performance metrics
	Metrics VariantMetrics `json:"metrics"`
	
	mu sync.RWMutex
}

// QueryModifications defines how to modify queries for this variant
type QueryModifications struct {
	// Query DSL modifications
	QueryType     string                 `json:"query_type,omitempty"`
	BoostFactors  map[string]float64     `json:"boost_factors,omitempty"`
	Fuzziness     string                 `json:"fuzziness,omitempty"`
	MinShouldMatch string                `json:"min_should_match,omitempty"`
	
	// Search parameters
	Size          int                    `json:"size,omitempty"`
	Timeout       string                 `json:"timeout,omitempty"`
	
	// Advanced features
	Rescore       []models.RescoreConfig `json:"rescore,omitempty"`
	Highlighting  *models.HighlightConfig `json:"highlighting,omitempty"`
	
	// Custom query template
	CustomQuery   string                 `json:"custom_query,omitempty"`
	
	// Feature flags
	EnableCaching     bool `json:"enable_caching"`
	EnablePrefetch    bool `json:"enable_prefetch"`
	EnablePersonalization bool `json:"enable_personalization"`
}

// VariantMetrics tracks performance metrics for a variant
type VariantMetrics struct {
	// Sample size
	TotalRequests int64 `json:"total_requests"`
	
	// Performance metrics
	AvgResponseTime    float64 `json:"avg_response_time_ms"`
	P95ResponseTime    float64 `json:"p95_response_time_ms"`
	P99ResponseTime    float64 `json:"p99_response_time_ms"`
	
	// Success metrics
	SuccessRate        float64 `json:"success_rate"`
	ErrorRate          float64 `json:"error_rate"`
	
	// Search quality metrics
	AvgResultCount     float64 `json:"avg_result_count"`
	ZeroResultsRate    float64 `json:"zero_results_rate"`
	
	// User engagement metrics (if available)
	ClickThroughRate   float64 `json:"click_through_rate,omitempty"`
	ConversionRate     float64 `json:"conversion_rate,omitempty"`
	UserSatisfaction   float64 `json:"user_satisfaction,omitempty"`
	
	// Statistical data
	ResponseTimes      []float64 `json:"-"` // Raw data for statistical analysis
	ResultCounts       []int64   `json:"-"`
	
	LastUpdated        time.Time `json:"last_updated"`
}

// ExperimentTargeting defines targeting rules for experiments
type ExperimentTargeting struct {
	// User-based targeting
	UserSegments    []string `json:"user_segments,omitempty"`
	UserProperties  map[string]interface{} `json:"user_properties,omitempty"`
	
	// Query-based targeting
	QueryPatterns   []string `json:"query_patterns,omitempty"`
	IndexPatterns   []string `json:"index_patterns,omitempty"`
	
	// Context-based targeting
	TimeOfDay       []string `json:"time_of_day,omitempty"`
	DaysOfWeek      []string `json:"days_of_week,omitempty"`
	
	// Geographic targeting
	Countries       []string `json:"countries,omitempty"`
	Regions         []string `json:"regions,omitempty"`
}

// ExperimentResults contains statistical analysis results
type ExperimentResults struct {
	Status           ResultStatus             `json:"status"`
	Winner           string                   `json:"winner,omitempty"`
	Confidence       float64                  `json:"confidence"`
	StatisticalPower float64                  `json:"statistical_power"`
	
	// Variant comparisons
	VariantResults   map[string]VariantResult `json:"variant_results"`
	
	// Recommendations
	Recommendations  []string                 `json:"recommendations"`
	
	UpdatedAt        time.Time                `json:"updated_at"`
}

// VariantResult contains results for a specific variant
type VariantResult struct {
	Variant         string  `json:"variant"`
	SampleSize      int64   `json:"sample_size"`
	ConversionRate  float64 `json:"conversion_rate"`
	ConfidenceInterval ConfidenceInterval `json:"confidence_interval"`
	PValue          float64 `json:"p_value"`
	Effect          float64 `json:"effect"` // % improvement over control
}

// ConfidenceInterval represents a statistical confidence interval
type ConfidenceInterval struct {
	Lower float64 `json:"lower"`
	Upper float64 `json:"upper"`
	Level float64 `json:"level"` // e.g., 0.95 for 95%
}

// ExperimentStatus represents the current status of an experiment
type ExperimentStatus string

const (
	StatusDraft    ExperimentStatus = "draft"
	StatusRunning  ExperimentStatus = "running"
	StatusPaused   ExperimentStatus = "paused"
	StatusComplete ExperimentStatus = "complete"
	StatusArchived ExperimentStatus = "archived"
)

// ResultStatus represents the statistical significance status
type ResultStatus string

const (
	ResultStatusInconclusive ResultStatus = "inconclusive"
	ResultStatusSignificant  ResultStatus = "significant"
	ResultStatusInsufficient ResultStatus = "insufficient_data"
)

// NewABTestFramework creates a new A/B testing framework
func NewABTestFramework(logger *zap.Logger) *ABTestFramework {
	return &ABTestFramework{
		experiments:         make(map[string]*Experiment),
		logger:              logger,
		defaultTrafficSplit: 0.1, // 10% of traffic by default
		minSampleSize:       100,  // Minimum samples per variant
		maxExperimentAge:    30 * 24 * time.Hour, // 30 days
	}
}

// CreateExperiment creates a new A/B test experiment
func (f *ABTestFramework) CreateExperiment(name, description string, config ExperimentConfig) (*Experiment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	experimentID := f.generateExperimentID(name)
	
	experiment := &Experiment{
		ID:                experimentID,
		Name:              name,
		Description:       description,
		Status:            StatusDraft,
		CreatedAt:         time.Now(),
		TrafficAllocation: config.TrafficAllocation,
		PrimaryMetric:     config.PrimaryMetric,
		SecondaryMetrics:  config.SecondaryMetrics,
		Targeting:         config.Targeting,
		MinSampleSize:     config.MinSampleSize,
		MaxDuration:       config.MaxDuration,
		SignificanceLevel: config.SignificanceLevel,
		TreatmentVariants: make(map[string]*Variant),
		Results: ExperimentResults{
			Status:         ResultStatusInsufficient,
			VariantResults: make(map[string]VariantResult),
		},
	}
	
	// Set defaults
	if experiment.TrafficAllocation == 0 {
		experiment.TrafficAllocation = f.defaultTrafficSplit
	}
	if experiment.MinSampleSize == 0 {
		experiment.MinSampleSize = f.minSampleSize
	}
	if experiment.MaxDuration == 0 {
		experiment.MaxDuration = f.maxExperimentAge
	}
	if experiment.SignificanceLevel == 0 {
		experiment.SignificanceLevel = 0.05 // 95% confidence
	}
	
	// Create control variant
	experiment.ControlVariant = &Variant{
		ID:          "control",
		Name:        "Control",
		Description: "Original query behavior",
		Weight:      0.5, // 50% of allocated traffic
		Metrics:     VariantMetrics{ResponseTimes: make([]float64, 0), ResultCounts: make([]int64, 0)},
	}
	
	f.experiments[experimentID] = experiment
	
	f.logger.Info("Created new A/B test experiment",
		zap.String("experiment_id", experimentID),
		zap.String("name", name),
		zap.Float64("traffic_allocation", experiment.TrafficAllocation))
	
	return experiment, nil
}

// AddTreatmentVariant adds a treatment variant to an experiment
func (f *ABTestFramework) AddTreatmentVariant(experimentID string, variant *Variant) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	experiment, exists := f.experiments[experimentID]
	if !exists {
		return fmt.Errorf("experiment %s not found", experimentID)
	}
	
	if experiment.Status != StatusDraft {
		return fmt.Errorf("cannot modify experiment %s: status is %s", experimentID, experiment.Status)
	}
	
	experiment.mu.Lock()
	defer experiment.mu.Unlock()
	
	variant.Metrics = VariantMetrics{
		ResponseTimes: make([]float64, 0),
		ResultCounts:  make([]int64, 0),
	}
	
	experiment.TreatmentVariants[variant.ID] = variant
	
	f.logger.Info("Added treatment variant to experiment",
		zap.String("experiment_id", experimentID),
		zap.String("variant_id", variant.ID),
		zap.String("variant_name", variant.Name))
	
	return nil
}

// StartExperiment starts an A/B test experiment
func (f *ABTestFramework) StartExperiment(experimentID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	experiment, exists := f.experiments[experimentID]
	if !exists {
		return fmt.Errorf("experiment %s not found", experimentID)
	}
	
	if experiment.Status != StatusDraft {
		return fmt.Errorf("experiment %s cannot be started: status is %s", experimentID, experiment.Status)
	}
	
	// Validate experiment configuration
	if len(experiment.TreatmentVariants) == 0 {
		return fmt.Errorf("experiment %s has no treatment variants", experimentID)
	}
	
	// Normalize variant weights
	totalWeight := experiment.ControlVariant.Weight
	for _, variant := range experiment.TreatmentVariants {
		totalWeight += variant.Weight
	}
	
	if totalWeight == 0 {
		return fmt.Errorf("experiment %s has zero total weight", experimentID)
	}
	
	// Start the experiment
	now := time.Now()
	experiment.Status = StatusRunning
	experiment.StartedAt = &now
	
	f.logger.Info("Started A/B test experiment",
		zap.String("experiment_id", experimentID),
		zap.String("name", experiment.Name),
		zap.Int("treatment_variants", len(experiment.TreatmentVariants)))
	
	return nil
}

// GetVariantForRequest determines which variant a request should use
func (f *ABTestFramework) GetVariantForRequest(request ABTestRequest) (*ExperimentAssignment, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	// Find applicable experiments
	for _, experiment := range f.experiments {
		if experiment.Status != StatusRunning {
			continue
		}
		
		// Check if request matches targeting criteria
		if !f.matchesTargeting(request, experiment.Targeting) {
			continue
		}
		
		// Check traffic allocation
		if !f.shouldParticipate(request, experiment.TrafficAllocation) {
			continue
		}
		
		// Determine variant assignment
		variant := f.assignVariant(request, experiment)
		
		return &ExperimentAssignment{
			ExperimentID: experiment.ID,
			VariantID:    variant.ID,
			VariantName:  variant.Name,
			Experiment:   experiment,
			Variant:      variant,
		}, nil
	}
	
	// No applicable experiments - return control
	return &ExperimentAssignment{
		ExperimentID: "control",
		VariantID:    "control",
		VariantName:  "Control",
	}, nil
}

// RecordExperimentResult records the result of an experiment request
func (f *ABTestFramework) RecordExperimentResult(assignment *ExperimentAssignment, result ExperimentResult) {
	if assignment.Experiment == nil {
		return // Control group
	}
	
	f.mu.RLock()
	experiment := f.experiments[assignment.ExperimentID]
	f.mu.RUnlock()
	
	if experiment == nil {
		return
	}
	
	experiment.mu.Lock()
	defer experiment.mu.Unlock()
	
	var variant *Variant
	if assignment.VariantID == "control" {
		variant = experiment.ControlVariant
	} else {
		variant = experiment.TreatmentVariants[assignment.VariantID]
	}
	
	if variant == nil {
		return
	}
	
	variant.mu.Lock()
	defer variant.mu.Unlock()
	
	// Update metrics
	variant.Metrics.TotalRequests++
	
	// Update response time metrics
	responseTime := float64(result.ResponseTime.Milliseconds())
	variant.Metrics.ResponseTimes = append(variant.Metrics.ResponseTimes, responseTime)
	variant.Metrics.AvgResponseTime = f.updateAverage(variant.Metrics.AvgResponseTime, responseTime, variant.Metrics.TotalRequests)
	
	// Update result count metrics
	variant.Metrics.ResultCounts = append(variant.Metrics.ResultCounts, result.ResultCount)
	variant.Metrics.AvgResultCount = f.updateAverage(variant.Metrics.AvgResultCount, float64(result.ResultCount), variant.Metrics.TotalRequests)
	
	// Update success/error rates
	if result.Success {
		variant.Metrics.SuccessRate = f.updateAverage(variant.Metrics.SuccessRate, 1.0, variant.Metrics.TotalRequests)
	} else {
		variant.Metrics.ErrorRate = f.updateAverage(variant.Metrics.ErrorRate, 1.0, variant.Metrics.TotalRequests)
	}
	
	// Update zero results rate
	if result.ResultCount == 0 {
		variant.Metrics.ZeroResultsRate = f.updateAverage(variant.Metrics.ZeroResultsRate, 1.0, variant.Metrics.TotalRequests)
	}
	
	variant.Metrics.LastUpdated = time.Now()
	
	// Trigger statistical analysis if we have enough data
	if variant.Metrics.TotalRequests >= int64(experiment.MinSampleSize) {
		go f.analyzeExperiment(experiment.ID)
	}
}

// GetExperimentResults returns the current results of an experiment
func (f *ABTestFramework) GetExperimentResults(experimentID string) (*ExperimentResults, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	experiment, exists := f.experiments[experimentID]
	if !exists {
		return nil, fmt.Errorf("experiment %s not found", experimentID)
	}
	
	experiment.mu.RLock()
	defer experiment.mu.RUnlock()
	
	// Create a copy to avoid concurrent access issues
	results := experiment.Results
	results.VariantResults = make(map[string]VariantResult)
	for k, v := range experiment.Results.VariantResults {
		results.VariantResults[k] = v
	}
	
	return &results, nil
}

// Helper methods

func (f *ABTestFramework) generateExperimentID(name string) string {
	hash := md5.Sum([]byte(name + time.Now().String()))
	return hex.EncodeToString(hash[:])[:8]
}

func (f *ABTestFramework) matchesTargeting(request ABTestRequest, targeting ExperimentTargeting) bool {
	// Query pattern matching
	if len(targeting.QueryPatterns) > 0 {
		matched := false
		for _, pattern := range targeting.QueryPatterns {
			// Simple pattern matching (in production, use regex)
			if contains(request.Query, pattern) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	
	// Index pattern matching
	if len(targeting.IndexPatterns) > 0 {
		matched := false
		for _, pattern := range targeting.IndexPatterns {
			if request.Index == pattern {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	
	// Time-based targeting
	if len(targeting.TimeOfDay) > 0 {
		currentHour := time.Now().Hour()
		matched := false
		for _, timeSlot := range targeting.TimeOfDay {
			// Simple time slot matching (format: "09-17" for 9 AM to 5 PM)
			if timeSlot == fmt.Sprintf("%02d", currentHour) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	
	return true
}

func (f *ABTestFramework) shouldParticipate(request ABTestRequest, trafficAllocation float64) bool {
	// Use consistent hashing based on user/session ID to ensure consistent experience
	hashInput := request.UserID
	if hashInput == "" {
		hashInput = request.SessionID
	}
	if hashInput == "" {
		hashInput = request.RequestID
	}
	
	hash := md5.Sum([]byte(hashInput))
	hashValue := float64(hash[0]) / 255.0
	
	return hashValue < trafficAllocation
}

func (f *ABTestFramework) assignVariant(request ABTestRequest, experiment *Experiment) *Variant {
	// Use consistent hashing for variant assignment
	hashInput := request.UserID + experiment.ID
	if request.UserID == "" {
		hashInput = request.SessionID + experiment.ID
	}
	
	hash := md5.Sum([]byte(hashInput))
	hashValue := float64(hash[1]) / 255.0
	
	// Calculate cumulative weights
	totalWeight := experiment.ControlVariant.Weight
	for _, variant := range experiment.TreatmentVariants {
		totalWeight += variant.Weight
	}
	
	// Assign based on weights
	threshold := 0.0
	normalizedHash := hashValue * totalWeight
	
	// Check control first
	threshold += experiment.ControlVariant.Weight
	if normalizedHash < threshold {
		return experiment.ControlVariant
	}
	
	// Check treatment variants
	for _, variant := range experiment.TreatmentVariants {
		threshold += variant.Weight
		if normalizedHash < threshold {
			return variant
		}
	}
	
	// Fallback to control
	return experiment.ControlVariant
}

func (f *ABTestFramework) updateAverage(currentAvg, newValue float64, count int64) float64 {
	if count <= 1 {
		return newValue
	}
	return (currentAvg*float64(count-1) + newValue) / float64(count)
}

func (f *ABTestFramework) analyzeExperiment(experimentID string) {
	f.mu.RLock()
	experiment := f.experiments[experimentID]
	f.mu.RUnlock()
	
	if experiment == nil {
		return
	}
	
	experiment.mu.Lock()
	defer experiment.mu.Unlock()
	
	// Perform statistical analysis
	controlMetrics := experiment.ControlVariant.Metrics
	
	// Check if we have sufficient data
	if controlMetrics.TotalRequests < int64(experiment.MinSampleSize) {
		experiment.Results.Status = ResultStatusInsufficient
		return
	}
	
	// Analyze each treatment variant against control
	bestVariant := "control"
	bestEffect := 0.0
	
	for variantID, variant := range experiment.TreatmentVariants {
		if variant.Metrics.TotalRequests < int64(experiment.MinSampleSize) {
			continue
		}
		
		// Simple statistical analysis (in production, use proper statistical tests)
		treatmentConversion := variant.Metrics.SuccessRate
		controlConversion := controlMetrics.SuccessRate
		
		if controlConversion > 0 {
			effect := (treatmentConversion - controlConversion) / controlConversion * 100
			
			// Simple significance test (in production, use t-test or chi-square)
			pValue := f.calculatePValue(controlMetrics, variant.Metrics)
			
			experiment.Results.VariantResults[variantID] = VariantResult{
				Variant:        variantID,
				SampleSize:     variant.Metrics.TotalRequests,
				ConversionRate: treatmentConversion,
				PValue:         pValue,
				Effect:         effect,
				ConfidenceInterval: ConfidenceInterval{
					Lower: effect - 5.0, // Simplified CI
					Upper: effect + 5.0,
					Level: 0.95,
				},
			}
			
			if pValue < experiment.SignificanceLevel && effect > bestEffect {
				bestVariant = variantID
				bestEffect = effect
				experiment.Results.Status = ResultStatusSignificant
			}
		}
	}
	
	if experiment.Results.Status == ResultStatusSignificant {
		experiment.Results.Winner = bestVariant
		experiment.Results.Confidence = (1.0 - experiment.SignificanceLevel) * 100
	} else {
		experiment.Results.Status = ResultStatusInconclusive
	}
	
	experiment.Results.UpdatedAt = time.Now()
	
	f.logger.Info("Updated experiment analysis",
		zap.String("experiment_id", experimentID),
		zap.String("status", string(experiment.Results.Status)),
		zap.String("winner", experiment.Results.Winner))
}

func (f *ABTestFramework) calculatePValue(control, treatment VariantMetrics) float64 {
	// Simplified p-value calculation
	// In production, use proper statistical libraries
	
	if control.TotalRequests == 0 || treatment.TotalRequests == 0 {
		return 1.0
	}
	
	// Simple effect size calculation
	effectSize := math.Abs(treatment.SuccessRate - control.SuccessRate)
	pooledStdDev := math.Sqrt(((control.SuccessRate*(1-control.SuccessRate))/float64(control.TotalRequests)) +
		((treatment.SuccessRate*(1-treatment.SuccessRate))/float64(treatment.TotalRequests)))
	
	if pooledStdDev == 0 {
		return 1.0
	}
	
	zScore := effectSize / pooledStdDev
	
	// Convert z-score to p-value (simplified)
	if zScore > 1.96 {
		return 0.01 // Significant
	} else if zScore > 1.64 {
		return 0.05 // Borderline
	}
	return 0.1 // Not significant
}

func contains(text, pattern string) bool {
	return len(text) >= len(pattern) && (text == pattern || 
		(len(pattern) > 0 && text[:len(pattern)] == pattern))
}

// GetAllExperiments returns all experiments
func (f *ABTestFramework) GetAllExperiments() []*Experiment {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	experiments := make([]*Experiment, 0, len(f.experiments))
	for _, experiment := range f.experiments {
		experiments = append(experiments, experiment)
	}
	
	return experiments
}

// GetExperiment returns a specific experiment
func (f *ABTestFramework) GetExperiment(experimentID string) *Experiment {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	return f.experiments[experimentID]
}

// GetExperimentAnalytics returns analytics for an experiment
func (f *ABTestFramework) GetExperimentAnalytics(experimentID string) *ExperimentAnalytics {
	f.mu.RLock()
	experiment := f.experiments[experimentID]
	f.mu.RUnlock()
	
	if experiment == nil {
		return nil
	}
	
	experiment.mu.RLock()
	defer experiment.mu.RUnlock()
	
	analytics := &ExperimentAnalytics{
		ExperimentID: experimentID,
		Status:       experiment.Status,
		StartedAt:    experiment.StartedAt,
		TotalRequests: 0,
		Variants:     make(map[string]VariantAnalytics),
	}
	
	// Control variant analytics
	if experiment.ControlVariant != nil {
		analytics.TotalRequests += experiment.ControlVariant.Metrics.TotalRequests
		analytics.Variants["control"] = VariantAnalytics{
			VariantID:     "control",
			Name:          experiment.ControlVariant.Name,
			TotalRequests: experiment.ControlVariant.Metrics.TotalRequests,
			Metrics:       experiment.ControlVariant.Metrics,
		}
	}
	
	// Treatment variants analytics
	for variantID, variant := range experiment.TreatmentVariants {
		analytics.TotalRequests += variant.Metrics.TotalRequests
		analytics.Variants[variantID] = VariantAnalytics{
			VariantID:     variantID,
			Name:          variant.Name,
			TotalRequests: variant.Metrics.TotalRequests,
			Metrics:       variant.Metrics,
		}
	}
	
	return analytics
}

// GetExperimentsOverview returns overview of all experiments
func (f *ABTestFramework) GetExperimentsOverview() *ExperimentsOverview {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	overview := &ExperimentsOverview{
		TotalExperiments: len(f.experiments),
		StatusCounts:     make(map[string]int),
	}
	
	for _, experiment := range f.experiments {
		overview.StatusCounts[string(experiment.Status)]++
		
		if experiment.Status == StatusRunning {
			overview.RunningExperiments++
		}
	}
	
	return overview
}

// Additional supporting types
type ExperimentAnalytics struct {
	ExperimentID  string                     `json:"experiment_id"`
	Status        ExperimentStatus           `json:"status"`
	StartedAt     *time.Time                 `json:"started_at,omitempty"`
	TotalRequests int64                      `json:"total_requests"`
	Variants      map[string]VariantAnalytics `json:"variants"`
}

type VariantAnalytics struct {
	VariantID     string         `json:"variant_id"`
	Name          string         `json:"name"`
	TotalRequests int64          `json:"total_requests"`
	Metrics       VariantMetrics `json:"metrics"`
}

type ExperimentsOverview struct {
	TotalExperiments    int            `json:"total_experiments"`
	RunningExperiments  int            `json:"running_experiments"`
	StatusCounts        map[string]int `json:"status_counts"`
}

// Supporting types

type ExperimentConfig struct {
	TrafficAllocation float64
	PrimaryMetric     string
	SecondaryMetrics  []string
	Targeting         ExperimentTargeting
	MinSampleSize     int
	MaxDuration       time.Duration
	SignificanceLevel float64
}

type ABTestRequest struct {
	RequestID string
	UserID    string
	SessionID string
	Query     string
	Index     string
	Context   map[string]interface{}
}

type ExperimentAssignment struct {
	ExperimentID string
	VariantID    string
	VariantName  string
	Experiment   *Experiment
	Variant      *Variant
}

type ExperimentResult struct {
	Success      bool
	ResponseTime time.Duration
	ResultCount  int64
	ClickThrough bool
	Conversion   bool
	UserRating   float64
}