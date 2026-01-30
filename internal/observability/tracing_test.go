package observability_test

import (
	"context"
	"testing"

	"github.com/AshishBagdane/go-report-engine/internal/engine"
	"github.com/AshishBagdane/go-report-engine/internal/formatter"
	"github.com/AshishBagdane/go-report-engine/internal/observability"
	"github.com/AshishBagdane/go-report-engine/internal/output"
	"github.com/AshishBagdane/go-report-engine/internal/processor"
	"github.com/AshishBagdane/go-report-engine/internal/provider"
)

// MockTracer for testing
type MockTracer struct {
	Spans []*MockSpan
}

func (m *MockTracer) StartSpan(ctx context.Context, name string) (context.Context, observability.Span) {
	span := &MockSpan{Name: name}
	m.Spans = append(m.Spans, span)
	return ctx, span
}

// MockSpan captures values
type MockSpan struct {
	Name  string
	Tags  map[string]string
	Ended bool
	Err   error
}

func (m *MockSpan) End() {
	m.Ended = true
}

func (m *MockSpan) SetTag(key, value string) {
	if m.Tags == nil {
		m.Tags = make(map[string]string)
	}
	m.Tags[key] = value
}

func (m *MockSpan) RecordError(err error) {
	m.Err = err
}

func TestTracingIntegration(t *testing.T) {
	// Setup
	tracer := &MockTracer{}

	prov := provider.NewMockProvider([]map[string]interface{}{
		{"id": 1, "value": "test"},
	})

	proc := &processor.BaseProcessor{}
	fmttr := formatter.NewJSONFormatter("")
	out := &output.ConsoleOutput{} // Or mock output

	// Build Engine with Tracing
	eng, err := engine.NewEngineBuilder().
		WithProvider(prov).
		WithProcessor(proc).
		WithFormatter(fmttr).
		WithOutput(out).
		WithTracer(tracer).
		Build()

	if err != nil {
		t.Fatalf("Failed to build engine: %v", err)
	}

	// Execute
	err = eng.Run()
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// Verify Spans
	// Expected: provider.fetch, processor.process, output.send
	expectedSpans := map[string]bool{
		"provider.fetch":    false,
		"processor.process": false,
		"output.send":       false,
	}

	for _, span := range tracer.Spans {
		if !span.Ended {
			t.Errorf("Span %s not ended", span.Name)
		}
		if _, ok := expectedSpans[span.Name]; ok {
			expectedSpans[span.Name] = true
		}
		// Basic tag verification
		if span.Name == "provider.fetch" && span.Tags["record_count"] != "1" {
			t.Errorf("Provider span missing record_count tag")
		}
	}

	for name, found := range expectedSpans {
		if !found {
			t.Errorf("Expected span %s not found", name)
		}
	}
}
