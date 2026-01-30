package main

import (
	"context"
	"fmt"
	"log"

	"github.com/AshishBagdane/go-report-engine/internal/config"
	"github.com/AshishBagdane/go-report-engine/internal/factory"
	"github.com/AshishBagdane/go-report-engine/internal/formatter"
	"github.com/AshishBagdane/go-report-engine/internal/output"
	"github.com/AshishBagdane/go-report-engine/internal/provider"
	"github.com/AshishBagdane/go-report-engine/internal/registry"
)

func init() {
	// Register core components
	registry.RegisterProvider("mock", func() provider.ProviderStrategy {
		return provider.NewMockProvider([]map[string]interface{}{
			{"id": 1, "name": "Alice", "score": 95},
			{"id": 2, "name": "Bob", "score": 88},
			{"id": 3, "name": "Charlie", "score": 92},
		})
	})

	registry.RegisterFormatter("json", func() formatter.FormatStrategy {
		return formatter.NewJSONFormatter("  ")
	})

	registry.RegisterOutput("console", func() output.OutputStrategy {
		return output.NewConsoleOutput()
	})
}

func main() {
	// Example 1: Load from YAML file
	fmt.Println("=== Example 1: Load from YAML ===")
	runWithConfigFile("examples/config.yaml")

	// Example 2: Load from JSON file
	fmt.Println("\n=== Example 2: Load from JSON ===")
	runWithConfigFile("examples/config.json")

	// Example 3: Load with environment variable overrides
	fmt.Println("\n=== Example 3: Load with Environment Overrides ===")
	runWithEnvOverrides("examples/config.minimal.yaml")
}

// runWithConfigFile loads config from file and runs the engine
func runWithConfigFile(configPath string) {
	// Load configuration from file
	cfg, err := config.LoadFromFile(configPath)
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		return
	}

	// Create engine from config
	engine, err := factory.NewEngineFromConfig(*cfg)
	if err != nil {
		log.Printf("Failed to create engine: %v", err)
		return
	}

	// Run the report engine
	ctx := context.Background()
	if err := engine.RunWithContext(ctx); err != nil {
		log.Printf("Engine execution failed: %v", err)
		return
	}

	fmt.Println("✓ Report generated successfully")
}

// runWithEnvOverrides demonstrates environment variable overrides
func runWithEnvOverrides(configPath string) {
	// Load configuration with environment overrides
	// Set environment variables before running:
	//   export ENGINE_PROVIDER_TYPE=mock
	//   export ENGINE_FORMATTER_TYPE=json
	//   export ENGINE_OUTPUT_TYPE=console
	cfg, err := config.LoadFromFileWithEnv(configPath)
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		return
	}

	// Create engine from config
	engine, err := factory.NewEngineFromConfig(*cfg)
	if err != nil {
		log.Printf("Failed to create engine: %v", err)
		return
	}

	// Run the report engine
	ctx := context.Background()
	if err := engine.RunWithContext(ctx); err != nil {
		log.Printf("Engine execution failed: %v", err)
		return
	}

	fmt.Println("✓ Report generated with env overrides")
}
