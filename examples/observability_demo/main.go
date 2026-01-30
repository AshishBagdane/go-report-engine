package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/AshishBagdane/go-report-engine/internal/engine"
	"github.com/AshishBagdane/go-report-engine/internal/formatter"
	"github.com/AshishBagdane/go-report-engine/internal/logging"
	"github.com/AshishBagdane/go-report-engine/internal/observability"
	"github.com/AshishBagdane/go-report-engine/internal/output"
	"github.com/AshishBagdane/go-report-engine/internal/processor"
	"github.com/AshishBagdane/go-report-engine/internal/provider"
)

// ConsoleMetricsCollector prints metrics to stdout
type ConsoleMetricsCollector struct{}

func (c *ConsoleMetricsCollector) Count(name string, value int, tags map[string]string) {
	fmt.Printf("[METRIC] Count: %s = %d %v\n", name, value, tags)
}

func (c *ConsoleMetricsCollector) Gauge(name string, value float64, tags map[string]string) {
	fmt.Printf("[METRIC] Gauge: %s = %f %v\n", name, value, tags)
}

func (c *ConsoleMetricsCollector) Histogram(name string, value float64, tags map[string]string) {
	fmt.Printf("[METRIC] Histogram: %s = %f %v\n", name, value, tags)
}

// ConsoleTracer prints spans to stdout
type ConsoleTracer struct{}

func (t *ConsoleTracer) StartSpan(ctx context.Context, name string) (context.Context, observability.Span) {
	fmt.Printf("[TRACE] Start: %s\n", name)
	return ctx, &ConsoleSpan{name: name, start: time.Now()}
}

type ConsoleSpan struct {
	name  string
	start time.Time
}

func (s *ConsoleSpan) End() {
	fmt.Printf("[TRACE] End:   %s (Duration: %v)\n", s.name, time.Since(s.start))
}

func (s *ConsoleSpan) SetTag(key, value string) {
	fmt.Printf("[TRACE] Tag:   %s -> %s=%s\n", s.name, key, value)
}

func (s *ConsoleSpan) RecordError(err error) {
	fmt.Printf("[TRACE] Error: %s -> %v\n", s.name, err)
}

func main() {
	// Configure logging
	logger := logging.NewLogger(logging.Config{
		Level:  logging.LevelInfo,
		Format: logging.FormatText,
	})

	// Sample Data Provider
	prov := provider.NewMockProvider([]map[string]interface{}{
		{"id": 1, "item": "Widget A"},
		{"id": 2, "item": "Widget B"},
		{"id": 3, "item": "Widget C"},
	})

	// Build Engine
	eng, err := engine.NewEngineBuilder().
		WithProvider(prov).
		WithProcessor(&processor.BaseProcessor{}).
		WithFormatter(formatter.NewJSONFormatter("")).
		WithOutput(&output.ConsoleOutput{}).
		WithMetrics(&ConsoleMetricsCollector{}).
		WithTracer(&ConsoleTracer{}).
		Build()

	if err != nil {
		slog.Error("Failed to build engine", "error", err)
		os.Exit(1)
	}

	// Set Logger
	eng.WithLogger(logger)

	fmt.Println("--- Starting Observability Demo ---")
	if err := eng.Run(); err != nil {
		slog.Error("Engine run failed", "error", err)
		os.Exit(1)
	}
	fmt.Println("--- Demo Completed Successfully ---")
}
