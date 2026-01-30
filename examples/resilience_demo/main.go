package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/AshishBagdane/go-report-engine/internal/engine"
	errs "github.com/AshishBagdane/go-report-engine/internal/errors"
	"github.com/AshishBagdane/go-report-engine/internal/formatter"
	"github.com/AshishBagdane/go-report-engine/internal/logging"
	"github.com/AshishBagdane/go-report-engine/internal/output"
	"github.com/AshishBagdane/go-report-engine/internal/processor"
	"github.com/AshishBagdane/go-report-engine/internal/provider"
	"github.com/AshishBagdane/go-report-engine/internal/resilience"
)

// FailingProvider simulates a provider that fails a few times before succeeding
type FailingProvider struct {
	attempts int
}

func (p *FailingProvider) Fetch(ctx context.Context) ([]map[string]interface{}, error) {
	p.attempts++
	fmt.Printf("Provider: Fetch attempt #%d\n", p.attempts)

	if p.attempts <= 2 {
		// Create a transient error that the engine will retry
		return nil, errs.NewErrorContext(errs.ComponentProvider, "fetch").
			WithType(errs.ErrorTypeTransient).
			New("simulated network error")
	}

	return []map[string]interface{}{
		{"id": 1, "message": "Success after retries!"},
	}, nil
}

func (p *FailingProvider) Stream(ctx context.Context) (provider.Iterator, error) {
	return nil, errors.New("not implemented")
}

func (p *FailingProvider) Close() error {
	return nil
}

func main() {
	// Configure logging
	logger := logging.NewLogger(logging.Config{
		Level:  logging.LevelInfo,
		Format: logging.FormatText,
	})

	// 1. Retry Policy: Retry up to 3 times
	retryPolicy := resilience.RetryPolicy{
		MaxRetries: 3,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   500 * time.Millisecond,
		Factor:     2,
	}

	// 2. Circuit Breaker: Open after 5 failures
	cb := resilience.NewCircuitBreaker("demo-breaker", 5, 2*time.Second)

	// Build Engine
	eng, err := engine.NewEngineBuilder().
		WithProvider(&FailingProvider{}).
		WithProcessor(&processor.BaseProcessor{}).
		WithFormatter(formatter.NewJSONFormatter("")).
		WithOutput(&output.ConsoleOutput{}).
		WithRetry(retryPolicy).
		WithCircuitBreaker(cb).
		Build()

	if err != nil {
		slog.Error("Failed to build engine", "error", err)
		os.Exit(1)
	}

	// Set Logger
	eng.WithLogger(logger)

	fmt.Println("--- Starting Resilience Demo ---")
	if err := eng.Run(); err != nil {
		slog.Error("Engine run failed", "error", err)
		os.Exit(1)
	}
	fmt.Println("--- Demo Completed Successfully ---")
}
