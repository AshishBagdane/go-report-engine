package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/AshishBagdane/go-report-engine/internal/engine"
	"github.com/AshishBagdane/go-report-engine/internal/factory"
	"github.com/AshishBagdane/go-report-engine/internal/formatter"
	"github.com/AshishBagdane/go-report-engine/internal/output"
	"github.com/AshishBagdane/go-report-engine/internal/provider"
	"github.com/AshishBagdane/go-report-engine/internal/registry"
	"github.com/AshishBagdane/go-report-engine/pkg/api"
)

// --- Sample User Logic (Configurable Filter) ---

// MinScoreFilter implements the user's business logic
type MinScoreFilter struct {
	MinScore int
}

// Keep implements api.FilterStrategy
func (f *MinScoreFilter) Keep(row map[string]interface{}) bool {
	if score, ok := row["score"].(int); ok {
		return score >= f.MinScore
	}
	return false
}

// Configure implements api.Configurable, setting the threshold from the config file.
func (f *MinScoreFilter) Configure(params map[string]string) error {
	minScoreStr, ok := params["min_score"]
	if !ok {
		return api.ErrMissingParam("min_score")
	}

	score, err := strconv.Atoi(minScoreStr)
	if err != nil {
		return fmt.Errorf("min_score must be an integer: %w", err)
	}
	f.MinScore = score
	return nil
}

// init registers all available core components and the sample user component.
func init() {
	// Core Registrations
	// Provider requires data parameter
	registry.RegisterProvider("mock", func() provider.ProviderStrategy {
		return provider.NewMockProvider([]map[string]interface{}{
			{"id": 1, "name": "Alice", "score": 95},
			{"id": 2, "name": "Bob", "score": 88},
			{"id": 3, "name": "Charlie", "score": 92},
		})
	})

	// Formatter requires indent parameter
	registry.RegisterFormatter("json", func() formatter.FormatStrategy {
		return formatter.NewJSONFormatter("  ")
	})

	// Output needs to return OutputStrategy interface
	registry.RegisterOutput("console", func() output.OutputStrategy {
		return output.NewConsoleOutput()
	})

	// User Logic Registration (Uses the new type-safe helper function)
	registry.RegisterFilter("min_score_filter", &MinScoreFilter{})
}

func main() {
	// 1. Simulate loading config from file (e.g., config.yaml)
	appConfig := engine.Config{
		Provider: engine.ProviderConfig{Type: "mock"},
		Processors: []engine.ProcessorConfig{
			{
				Type: "min_score_filter", // Found via registry.RegisterFilter
				Params: map[string]string{
					"min_score": "90", // Passed to MinScoreFilter.Configure()
				},
			},
		},
		Formatter: engine.FormatterConfig{Type: "json"},
		Output:    engine.OutputConfig{Type: "console"},
	}

	// 2. Use the Factory to build the engine (New clean approach)
	re, err := factory.NewEngineFromConfig(appConfig)
	if err != nil {
		log.Fatalf("Failed to initialize engine: %v", err)
	}

	// 3. Run with context
	ctx := context.Background()
	if err := re.RunWithContext(ctx); err != nil {
		fmt.Println("Error during execution:", err)
	}
}
