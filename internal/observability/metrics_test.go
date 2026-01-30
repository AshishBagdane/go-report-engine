package observability_test

import (
	"testing"

	"github.com/AshishBagdane/report-engine/internal/engine"
	"github.com/AshishBagdane/report-engine/internal/formatter"
	"github.com/AshishBagdane/report-engine/internal/output"
	"github.com/AshishBagdane/report-engine/internal/processor"
	"github.com/AshishBagdane/report-engine/internal/provider"
)

// MockCollector captures metrics for verification
type MockCollector struct {
	Counts     map[string]int
	Histograms map[string]float64
}

func NewMockCollector() *MockCollector {
	return &MockCollector{
		Counts:     make(map[string]int),
		Histograms: make(map[string]float64),
	}
}

func (m *MockCollector) Count(name string, value int, tags map[string]string) {
	m.Counts[name] += value
}

func (m *MockCollector) Gauge(name string, value float64, tags map[string]string) {}

func (m *MockCollector) Histogram(name string, value float64, tags map[string]string) {
	m.Histograms[name] = value // Just store last value for simplicity
}

func TestMetricsIntegration(t *testing.T) {
	// Setup
	collector := NewMockCollector()

	prov := provider.NewMockProvider([]map[string]interface{}{
		{"id": 1, "value": "test"},
	})

	proc := &processor.BaseProcessor{}
	fmttr := formatter.NewJSONFormatter("")
	out := &output.ConsoleOutput{} // Or mock output to avoid stdout noise

	// Build Engine with Metrics
	eng, err := engine.NewEngineBuilder().
		WithProvider(prov).
		WithProcessor(proc).
		WithFormatter(fmttr).
		WithOutput(out).
		WithMetrics(collector).
		Build()

	if err != nil {
		t.Fatalf("Failed to build engine: %v", err)
	}

	// Execute
	err = eng.Run()
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// Verify Metrics
	expectedCounts := map[string]int{
		"report_engine_provider_records_count":         1,
		"report_engine_processor_input_records_count":  1,
		"report_engine_processor_output_records_count": 1,
		"report_engine_output_bytes_count":             15, // Approximate JSON length [{"id":1...}]
	}

	for name, expected := range expectedCounts {
		if got, ok := collector.Counts[name]; !ok || got == 0 {
			t.Errorf("Metric %s missing or 0", name)
		} else if name == "report_engine_output_bytes_count" && got < 10 {
			t.Errorf("Metric %s seems too low: %d", name, got)
		} else if name != "report_engine_output_bytes_count" && got != expected {
			t.Errorf("Metric %s: expected %d, got %d", name, expected, got)
		}
	}

	// Verify Histograms (just presence)
	expectedHistograms := []string{
		"report_engine_provider_fetch_duration_seconds",
		"report_engine_processor_duration_seconds",
		"report_engine_output_duration_seconds",
	}

	for _, name := range expectedHistograms {
		if _, ok := collector.Histograms[name]; !ok {
			t.Errorf("Histogram %s missing", name)
		}
	}
}
