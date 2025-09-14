package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/saif-islam/es-playground/projects/search-api/internal/abtesting"
	"github.com/saif-islam/es-playground/projects/search-api/internal/models"
)

// ExperimentHandler handles A/B testing experiment management
type ExperimentHandler struct {
	framework *abtesting.ABTestFramework
	logger    *zap.Logger
}

// NewExperimentHandler creates a new experiment handler
func NewExperimentHandler(framework *abtesting.ABTestFramework, logger *zap.Logger) *ExperimentHandler {
	return &ExperimentHandler{
		framework: framework,
		logger:    logger,
	}
}

// RegisterRoutes registers experiment management routes
func (h *ExperimentHandler) RegisterRoutes(router *gin.RouterGroup) {
	experiments := router.Group("/experiments")
	{
		experiments.GET("", h.ListExperiments)
		experiments.POST("", h.CreateExperiment)
		experiments.GET("/:id", h.GetExperiment)
		experiments.PUT("/:id", h.UpdateExperiment)
		experiments.DELETE("/:id", h.DeleteExperiment)
		
		// Experiment control
		experiments.POST("/:id/start", h.StartExperiment)
		experiments.POST("/:id/pause", h.PauseExperiment)
		experiments.POST("/:id/stop", h.StopExperiment)
		
		// Variants
		experiments.POST("/:id/variants", h.AddVariant)
		experiments.PUT("/:id/variants/:variant_id", h.UpdateVariant)
		experiments.DELETE("/:id/variants/:variant_id", h.DeleteVariant)
		
		// Results
		experiments.GET("/:id/results", h.GetResults)
		experiments.GET("/:id/results/export", h.ExportResults)
		
		// Analytics
		experiments.GET("/:id/analytics", h.GetAnalytics)
		experiments.GET("/analytics/overview", h.GetOverview)
	}
	
	// Quick experiment creation templates
	templates := router.Group("/experiment-templates")
	{
		templates.GET("", h.ListTemplates)
		templates.GET("/:template", h.GetTemplate)
		templates.POST("/:template/create", h.CreateFromTemplate)
	}
}

// CreateExperiment creates a new A/B test experiment
func (h *ExperimentHandler) CreateExperiment(c *gin.Context) {
	var req CreateExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind create experiment request", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "invalid_request",
			Message:   err.Error(),
			RequestID: uuid.New().String(),
		})
		return
	}
	
	// Validate request
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "missing_name",
			Message: "Experiment name is required",
		})
		return
	}
	
	// Create experiment configuration
	config := abtesting.ExperimentConfig{
		TrafficAllocation: req.TrafficAllocation,
		PrimaryMetric:     req.PrimaryMetric,
		SecondaryMetrics:  req.SecondaryMetrics,
		Targeting:         req.Targeting,
		MinSampleSize:     req.MinSampleSize,
		MaxDuration:       req.MaxDuration,
		SignificanceLevel: req.SignificanceLevel,
	}
	
	// Set defaults
	if config.TrafficAllocation == 0 {
		config.TrafficAllocation = 0.1 // 10%
	}
	if config.PrimaryMetric == "" {
		config.PrimaryMetric = "success_rate"
	}
	if config.MinSampleSize == 0 {
		config.MinSampleSize = 100
	}
	if config.SignificanceLevel == 0 {
		config.SignificanceLevel = 0.05 // 95% confidence
	}
	
	experiment, err := h.framework.CreateExperiment(req.Name, req.Description, config)
	if err != nil {
		h.logger.Error("Failed to create experiment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
		return
	}
	
	// Add treatment variants if provided
	for _, variantReq := range req.TreatmentVariants {
		variant := &abtesting.Variant{
			ID:                 variantReq.ID,
			Name:               variantReq.Name,
			Description:        variantReq.Description,
			Weight:             variantReq.Weight,
			QueryModifications: variantReq.QueryModifications,
		}
		
		if err := h.framework.AddTreatmentVariant(experiment.ID, variant); err != nil {
			h.logger.Error("Failed to add treatment variant", zap.Error(err))
			// Continue with other variants
		}
	}
	
	h.logger.Info("Created A/B test experiment",
		zap.String("experiment_id", experiment.ID),
		zap.String("name", experiment.Name),
		zap.Int("treatment_variants", len(req.TreatmentVariants)))
	
	c.JSON(http.StatusCreated, experiment)
}

// StartExperiment starts an experiment
func (h *ExperimentHandler) StartExperiment(c *gin.Context) {
	experimentID := c.Param("id")
	
	if err := h.framework.StartExperiment(experimentID); err != nil {
		h.logger.Error("Failed to start experiment", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "start_failed",
			Message: err.Error(),
		})
		return
	}
	
	h.logger.Info("Started experiment", zap.String("experiment_id", experimentID))
	c.JSON(http.StatusOK, gin.H{"status": "started"})
}

// GetResults returns experiment results
func (h *ExperimentHandler) GetResults(c *gin.Context) {
	experimentID := c.Param("id")
	
	results, err := h.framework.GetExperimentResults(experimentID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "experiment_not_found",
			Message: err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, results)
}

// ListExperiments returns all experiments
func (h *ExperimentHandler) ListExperiments(c *gin.Context) {
	// Get query parameters for filtering
	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	
	experiments := h.framework.GetAllExperiments()
	
	// Filter by status if provided
	if status != "" {
		filtered := make([]*abtesting.Experiment, 0)
		for _, exp := range experiments {
			if string(exp.Status) == status {
				filtered = append(filtered, exp)
			}
		}
		experiments = filtered
	}
	
	// Limit results
	if limit > 0 && len(experiments) > limit {
		experiments = experiments[:limit]
	}
	
	c.JSON(http.StatusOK, gin.H{
		"experiments": experiments,
		"total":       len(experiments),
	})
}

// GetExperiment returns a specific experiment
func (h *ExperimentHandler) GetExperiment(c *gin.Context) {
	experimentID := c.Param("id")
	
	experiment := h.framework.GetExperiment(experimentID)
	if experiment == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "experiment_not_found",
			Message: "Experiment not found",
		})
		return
	}
	
	c.JSON(http.StatusOK, experiment)
}

// GetAnalytics returns experiment analytics
func (h *ExperimentHandler) GetAnalytics(c *gin.Context) {
	experimentID := c.Param("id")
	
	analytics := h.framework.GetExperimentAnalytics(experimentID)
	if analytics == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "experiment_not_found",
			Message: "Experiment not found",
		})
		return
	}
	
	c.JSON(http.StatusOK, analytics)
}

// GetOverview returns overview of all experiments
func (h *ExperimentHandler) GetOverview(c *gin.Context) {
	overview := h.framework.GetExperimentsOverview()
	c.JSON(http.StatusOK, overview)
}

// ListTemplates returns available experiment templates
func (h *ExperimentHandler) ListTemplates(c *gin.Context) {
	templates := []ExperimentTemplate{
		{
			ID:          "query-optimization",
			Name:        "Query Optimization",
			Description: "Compare different query types and parameters",
			Variants: []TemplateVariant{
				{
					Name:        "Match Query",
					Description: "Use match query with default settings",
					QueryModifications: abtesting.QueryModifications{
						QueryType: "match",
					},
				},
				{
					Name:        "Multi-Match Query",
					Description: "Use multi_match query with boosting",
					QueryModifications: abtesting.QueryModifications{
						QueryType: "multi_match",
						BoostFactors: map[string]float64{
							"title": 2.0,
							"content": 1.0,
						},
					},
				},
			},
		},
		{
			ID:          "fuzzy-search",
			Name:        "Fuzzy Search Test",
			Description: "Test different fuzziness levels",
			Variants: []TemplateVariant{
				{
					Name:        "No Fuzziness",
					Description: "Exact matching only",
					QueryModifications: abtesting.QueryModifications{
						Fuzziness: "0",
					},
				},
				{
					Name:        "Auto Fuzziness",
					Description: "Automatic fuzziness based on term length",
					QueryModifications: abtesting.QueryModifications{
						Fuzziness: "AUTO",
					},
				},
			},
		},
		{
			ID:          "result-count",
			Name:        "Result Count Optimization",
			Description: "Test different result counts for user experience",
			Variants: []TemplateVariant{
				{
					Name:        "10 Results",
					Description: "Default result count",
					QueryModifications: abtesting.QueryModifications{
						Size: 10,
					},
				},
				{
					Name:        "20 Results",
					Description: "More results per page",
					QueryModifications: abtesting.QueryModifications{
						Size: 20,
					},
				},
			},
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
	})
}

// CreateFromTemplate creates experiment from template
func (h *ExperimentHandler) CreateFromTemplate(c *gin.Context) {
	templateID := c.Param("template")
	
	var req CreateFromTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}
	
	// Find template
	var template *ExperimentTemplate
	templates := h.getAvailableTemplates()
	for _, t := range templates {
		if t.ID == templateID {
			template = &t
			break
		}
	}
	
	if template == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "template_not_found",
			Message: "Template not found",
		})
		return
	}
	
	// Create experiment from template
	config := abtesting.ExperimentConfig{
		TrafficAllocation: req.TrafficAllocation,
		PrimaryMetric:     "success_rate",
		MinSampleSize:     req.MinSampleSize,
		SignificanceLevel: 0.05,
		Targeting:         req.Targeting,
	}
	
	experiment, err := h.framework.CreateExperiment(req.Name, template.Description, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
		return
	}
	
	// Add variants from template
	for i, variantTemplate := range template.Variants {
		variant := &abtesting.Variant{
			ID:                 strconv.Itoa(i + 1),
			Name:               variantTemplate.Name,
			Description:        variantTemplate.Description,
			Weight:             1.0 / float64(len(template.Variants)), // Equal weights
			QueryModifications: variantTemplate.QueryModifications,
		}
		
		h.framework.AddTreatmentVariant(experiment.ID, variant)
	}
	
	c.JSON(http.StatusCreated, experiment)
}

// Placeholder implementations for remaining endpoints
func (h *ExperimentHandler) UpdateExperiment(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Update experiment - coming soon"})
}

func (h *ExperimentHandler) DeleteExperiment(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Delete experiment - coming soon"})
}

func (h *ExperimentHandler) PauseExperiment(c *gin.Context) {
	experimentID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Experiment paused", "experiment_id": experimentID})
}

func (h *ExperimentHandler) StopExperiment(c *gin.Context) {
	experimentID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Experiment stopped", "experiment_id": experimentID})
}

func (h *ExperimentHandler) AddVariant(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Add variant - coming soon"})
}

func (h *ExperimentHandler) UpdateVariant(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Update variant - coming soon"})
}

func (h *ExperimentHandler) DeleteVariant(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Delete variant - coming soon"})
}

func (h *ExperimentHandler) ExportResults(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Export results - coming soon"})
}

func (h *ExperimentHandler) GetTemplate(c *gin.Context) {
	templateID := c.Param("template")
	templates := h.getAvailableTemplates()
	
	for _, template := range templates {
		if template.ID == templateID {
			c.JSON(http.StatusOK, template)
			return
		}
	}
	
	c.JSON(http.StatusNotFound, models.ErrorResponse{
		Error:   "template_not_found",
		Message: "Template not found",
	})
}

func (h *ExperimentHandler) getAvailableTemplates() []ExperimentTemplate {
	return []ExperimentTemplate{
		{
			ID:          "query-optimization",
			Name:        "Query Optimization",
			Description: "Compare different query types and parameters",
		},
		{
			ID:          "fuzzy-search",
			Name:        "Fuzzy Search Test",
			Description: "Test different fuzziness levels",
		},
		{
			ID:          "result-count",
			Name:        "Result Count Optimization",
			Description: "Test different result counts for user experience",
		},
	}
}

// Request/Response types

type CreateExperimentRequest struct {
	Name               string                          `json:"name" binding:"required"`
	Description        string                          `json:"description"`
	TrafficAllocation  float64                         `json:"traffic_allocation"`
	PrimaryMetric      string                          `json:"primary_metric"`
	SecondaryMetrics   []string                        `json:"secondary_metrics"`
	Targeting          abtesting.ExperimentTargeting   `json:"targeting"`
	MinSampleSize      int                             `json:"min_sample_size"`
	MaxDuration        time.Duration                   `json:"max_duration"`
	SignificanceLevel  float64                         `json:"significance_level"`
	TreatmentVariants  []CreateVariantRequest          `json:"treatment_variants"`
}

type CreateVariantRequest struct {
	ID                 string                          `json:"id" binding:"required"`
	Name               string                          `json:"name" binding:"required"`
	Description        string                          `json:"description"`
	Weight             float64                         `json:"weight"`
	QueryModifications abtesting.QueryModifications    `json:"query_modifications"`
}

type CreateFromTemplateRequest struct {
	Name              string                        `json:"name" binding:"required"`
	TrafficAllocation float64                       `json:"traffic_allocation"`
	MinSampleSize     int                           `json:"min_sample_size"`
	Targeting         abtesting.ExperimentTargeting `json:"targeting"`
}

type ExperimentTemplate struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Variants    []TemplateVariant `json:"variants,omitempty"`
}

type TemplateVariant struct {
	Name               string                       `json:"name"`
	Description        string                       `json:"description"`
	QueryModifications abtesting.QueryModifications `json:"query_modifications"`
}