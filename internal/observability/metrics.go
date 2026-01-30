package observability

// MetricsCollector defines the interface for collecting metrics.
// It abstracts away the underlying metrics system (Prometheus, StatsD, OpenTelemetry, etc.).
type MetricsCollector interface {
	// Count increments a counter metric.
	// Used for: total records, error counts, request counts.
	Count(name string, value int, tags map[string]string)

	// Gauge sets a gauge metric.
	// Used for: current memory usage, active goroutines, queue size.
	Gauge(name string, value float64, tags map[string]string)

	// Histogram records a value in a histogram.
	// Used for: request duration, payload size distribution.
	Histogram(name string, value float64, tags map[string]string)
}

// NoopCollector is a default implementation that does nothing.
// Useful for tests or when metrics are disabled.
type NoopCollector struct{}

func (n *NoopCollector) Count(name string, value int, tags map[string]string)         {}
func (n *NoopCollector) Gauge(name string, value float64, tags map[string]string)     {}
func (n *NoopCollector) Histogram(name string, value float64, tags map[string]string) {}

// NewNoopCollector creates a new NoopCollector.
func NewNoopCollector() MetricsCollector {
	return &NoopCollector{}
}
