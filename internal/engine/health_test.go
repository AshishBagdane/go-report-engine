package engine_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AshishBagdane/report-engine/internal/engine"
	"github.com/AshishBagdane/report-engine/internal/formatter"
	"github.com/AshishBagdane/report-engine/internal/health"
	"github.com/AshishBagdane/report-engine/internal/output"
	"github.com/AshishBagdane/report-engine/internal/processor"
	"github.com/AshishBagdane/report-engine/internal/provider"
)

// MockHealthyProvider implements health.Checker
type MockHealthyProvider struct {
	provider.MockProvider
}

func (m *MockHealthyProvider) CheckHealth(ctx context.Context) (health.Result, error) {
	return health.Result{Status: health.StatusUp}, nil
}

// MockUnhealthyOutput implements health.Checker
type MockUnhealthyOutput struct {
	output.ConsoleOutput
}

func (m *MockUnhealthyOutput) CheckHealth(ctx context.Context) (health.Result, error) {
	return health.Result{
		Status: health.StatusDown,
		Error:  "connection refused",
	}, errors.New("connection refused")
}

func TestEngineHealth(t *testing.T) {
	// Setup
	prov := &MockHealthyProvider{}
	out := &MockUnhealthyOutput{}
	proc := &processor.BaseProcessor{} // Does not implement Checker
	fmttr := formatter.NewJSONFormatter("")

	eng, err := engine.NewEngineBuilder().
		WithProvider(prov).
		WithProcessor(proc).
		WithFormatter(fmttr).
		WithOutput(out).
		Build()

	if err != nil {
		t.Fatalf("Failed to build engine: %v", err)
	}

	// Calculate Health
	results := eng.Health(context.Background())

	// Verify Provider (Up)
	if res, ok := results["provider"]; !ok {
		t.Error("Missing provider health result")
	} else if res.Status != health.StatusUp {
		t.Errorf("Expected provider UP, got %v", res.Status)
	}

	// Verify Output (Down)
	if res, ok := results["output"]; !ok {
		t.Error("Missing output health result")
	} else if res.Status != health.StatusDown {
		t.Errorf("Expected output DOWN, got %v", res.Status)
	} else if res.Error != "connection refused" {
		t.Errorf("Expected output error 'connection refused', got '%v'", res.Error)
	}

	// Verify Processor (Not implemented -> should not be present or default? Logic says if checker, add it. else nothing?)
	// Engine logic:
	// Processor is checked ONLY if it implements Checker. BaseProcessor does not.
	// So results["processor"] should be missing.
	if _, ok := results["processor"]; ok {
		t.Error("Unexpected processor health result (BaseProcessor does not implement Checker)")
	}
}
