package tracing

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TracingConfig represents tracing configuration
type TracingConfig struct {
	Enabled     bool   `yaml:"enabled"`
	ServiceName string `yaml:"service_name"`
	Environment string `yaml:"environment"`
	
	// Jaeger configuration
	JaegerEndpoint string  `yaml:"jaeger_endpoint"`
	SamplingRate   float64 `yaml:"sampling_rate"`
	
	// Additional settings
	MaxTagLength    int           `yaml:"max_tag_length"`
	BatchTimeout    time.Duration `yaml:"batch_timeout"`
	MaxQueueSize    int           `yaml:"max_queue_size"`
	MaxPacketSize   int           `yaml:"max_packet_size"`
}

// TracingProvider manages OpenTelemetry tracing
type TracingProvider struct {
	config   TracingConfig
	logger   *zap.Logger
	provider trace.TracerProvider
	tracer   trace.Tracer
}

// NewTracingProvider creates a new tracing provider
func NewTracingProvider(config TracingConfig, logger *zap.Logger) (*TracingProvider, error) {
	if !config.Enabled {
		logger.Info("Distributed tracing is disabled")
		return &TracingProvider{
			config:   config,
			logger:   logger,
			provider: trace.NewNoopTracerProvider(),
			tracer:   trace.NewNoopTracerProvider().Tracer("noop"),
		}, nil
	}

	// Set defaults
	if config.ServiceName == "" {
		config.ServiceName = "search-api"
	}
	if config.Environment == "" {
		config.Environment = "development"
	}
	if config.SamplingRate == 0 {
		config.SamplingRate = 1.0 // Sample all traces in development
	}
	if config.JaegerEndpoint == "" {
		config.JaegerEndpoint = "http://localhost:14268/api/traces"
	}
	if config.MaxTagLength == 0 {
		config.MaxTagLength = 1024
	}
	if config.BatchTimeout == 0 {
		config.BatchTimeout = 1 * time.Second
	}
	if config.MaxQueueSize == 0 {
		config.MaxQueueSize = 2048
	}

	// Create Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.JaegerEndpoint)))
	if err != nil {
		return nil, err
	}

	// Create resource with service information
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion("1.0.0"),
			semconv.DeploymentEnvironment(config.Environment),
			semconv.ServiceInstanceID("search-api-1"),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp,
			sdktrace.WithBatchTimeout(config.BatchTimeout),
			sdktrace.WithMaxQueueSize(config.MaxQueueSize),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(config.SamplingRate)),
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)
	
	// Set global propagator for trace context propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create tracer
	tracer := tp.Tracer(config.ServiceName)

	logger.Info("Distributed tracing initialized",
		zap.String("service", config.ServiceName),
		zap.String("environment", config.Environment),
		zap.String("jaeger_endpoint", config.JaegerEndpoint),
		zap.Float64("sampling_rate", config.SamplingRate),
	)

	return &TracingProvider{
		config:   config,
		logger:   logger,
		provider: tp,
		tracer:   tracer,
	}, nil
}

// GetTracer returns the tracer instance
func (tp *TracingProvider) GetTracer() trace.Tracer {
	return tp.tracer
}

// GetTracerProvider returns the tracer provider
func (tp *TracingProvider) GetTracerProvider() trace.TracerProvider {
	return tp.provider
}

// Shutdown gracefully shuts down the tracing provider
func (tp *TracingProvider) Shutdown(ctx context.Context) error {
	if !tp.config.Enabled {
		return nil
	}

	if provider, ok := tp.provider.(*sdktrace.TracerProvider); ok {
		return provider.Shutdown(ctx)
	}
	return nil
}

// StartSpan starts a new span with the given name and options
func (tp *TracingProvider) StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tp.tracer.Start(ctx, spanName, opts...)
}

// GetSpanFromContext returns the span from the context
func (tp *TracingProvider) GetSpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddSpanEvent adds an event to the current span
func (tp *TracingProvider) AddSpanEvent(ctx context.Context, name string, attributes map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		attrs := make([]trace.EventOption, 0, len(attributes))
		for k, v := range attributes {
			attrs = append(attrs, trace.WithAttributes(convertToAttribute(k, v)))
		}
		span.AddEvent(name, attrs...)
	}
}

// SetSpanAttributes sets attributes on the current span
func (tp *TracingProvider) SetSpanAttributes(ctx context.Context, attributes map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		attrs := make([]trace.Attribute, 0, len(attributes))
		for k, v := range attributes {
			attrs = append(attrs, convertToAttribute(k, v))
		}
		span.SetAttributes(attrs...)
	}
}

// RecordError records an error on the current span
func (tp *TracingProvider) RecordError(ctx context.Context, err error, attributes map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		opts := []trace.EventOption{trace.WithStackTrace(true)}
		if attributes != nil {
			attrs := make([]trace.Attribute, 0, len(attributes))
			for k, v := range attributes {
				attrs = append(attrs, convertToAttribute(k, v))
			}
			opts = append(opts, trace.WithAttributes(attrs...))
		}
		span.RecordError(err, opts...)
	}
}

// Helper function to convert interface{} to trace.Attribute
func convertToAttribute(key string, value interface{}) trace.Attribute {
	switch v := value.(type) {
	case string:
		return trace.StringAttribute(key, v)
	case int:
		return trace.IntAttribute(key, v)
	case int64:
		return trace.Int64Attribute(key, v)
	case float64:
		return trace.Float64Attribute(key, v)
	case bool:
		return trace.BoolAttribute(key, v)
	default:
		return trace.StringAttribute(key, "unsupported_type")
	}
}