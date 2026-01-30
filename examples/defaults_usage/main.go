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

	registry.RegisterFormatter("csv", func() formatter.FormatStrategy {
		// Placeholder - actual CSV formatter would be implemented
		return formatter.NewJSONFormatter("")
	})

	registry.RegisterOutput("console", func() output.OutputStrategy {
		return output.NewConsoleOutput()
	})

	registry.RegisterOutput("file", func() output.OutputStrategy {
		// Placeholder - actual file output would be implemented
		return output.NewConsoleOutput()
	})
}

func main() {
	fmt.Println("=== Default Config Examples ===")
	fmt.Println()

	// Example 1: Simple default config
	example1()

	// Example 2: Development config
	example2()

	// Example 3: Production config
	example3()

	// Example 4: CSV config
	example4()

	// Example 5: Builder pattern
	example5()

	// Example 6: Combining defaults with file loading
	example6()
}

// example1 demonstrates using simple default config
func example1() {
	fmt.Println("--- Example 1: Simple Default Config ---")

	// Get default config - simplest way to start
	cfg := config.DefaultConfig()

	// Use it directly
	engine, err := factory.NewEngineFromConfig(cfg)
	if err != nil {
		log.Printf("Failed to create engine: %v", err)
		return
	}

	ctx := context.Background()
	if err := engine.RunWithContext(ctx); err != nil {
		log.Printf("Engine execution failed: %v", err)
		return
	}

	fmt.Println("✓ Default config executed successfully")
}

// example2 demonstrates using development config
func example2() {
	fmt.Println("--- Example 2: Development Config ---")

	// Get development-friendly config with verbose output
	cfg := config.DevelopmentConfig()

	engine, err := factory.NewEngineFromConfig(cfg)
	if err != nil {
		log.Printf("Failed to create engine: %v", err)
		return
	}

	ctx := context.Background()
	if err := engine.RunWithContext(ctx); err != nil {
		log.Printf("Engine execution failed: %v", err)
		return
	}

	fmt.Println("✓ Development config executed with pretty output")
}

// example3 demonstrates using production config
func example3() {
	fmt.Println("--- Example 3: Production Config ---")

	// Get production-ready config
	cfg := config.ProductionConfig()

	// Customize for your environment
	cfg.Provider.Type = "mock" // In production, use "postgres", "mysql", etc.

	engine, err := factory.NewEngineFromConfig(cfg)
	if err != nil {
		log.Printf("Failed to create engine: %v", err)
		return
	}

	ctx := context.Background()
	if err := engine.RunWithContext(ctx); err != nil {
		log.Printf("Engine execution failed: %v", err)
		return
	}

	fmt.Println("✓ Production config executed with validation")
}

// example4 demonstrates using CSV config
func example4() {
	fmt.Println("--- Example 4: CSV Config ---")

	// Get CSV-specific config
	cfg := config.CSVConfig()

	// Update output path
	cfg.Output.Params["path"] = "/tmp/report.csv"

	engine, err := factory.NewEngineFromConfig(cfg)
	if err != nil {
		log.Printf("Failed to create engine: %v", err)
		return
	}

	ctx := context.Background()
	if err := engine.RunWithContext(ctx); err != nil {
		log.Printf("Engine execution failed: %v", err)
		return
	}

	fmt.Println("✓ CSV config executed")
}

// example5 demonstrates builder pattern with helpers
func example5() {
	fmt.Println("--- Example 5: Builder Pattern ---")

	// Start with default
	cfg := config.DefaultConfig()

	// Add provider params
	cfg = config.ConfigWithProviderParams(cfg, map[string]string{
		"verbose": "true",
	})

	// Add processing pipeline
	cfg = config.ConfigWithProcessor(cfg, "filter", map[string]string{
		"min_score": "90",
	})
	cfg = config.ConfigWithProcessor(cfg, "validator", map[string]string{
		"strict": "true",
	})

	// Update formatter
	cfg = config.ConfigWithFormatterParams(cfg, map[string]string{
		"indent": "4",
		"pretty": "true",
	})

	// Use the customized config
	engine, err := factory.NewEngineFromConfig(cfg)
	if err != nil {
		log.Printf("Failed to create engine: %v", err)
		return
	}

	ctx := context.Background()
	if err := engine.RunWithContext(ctx); err != nil {
		log.Printf("Engine execution failed: %v", err)
		return
	}

	fmt.Println("✓ Custom built config executed")
}

// example6 demonstrates combining defaults with file loading
func example6() {
	fmt.Println("--- Example 6: Defaults + File Loading ---")

	// Load from file
	cfg, err := config.LoadFromFile("examples/config.minimal.yaml")
	if err != nil {
		// If file doesn't exist, use default
		fmt.Println("Config file not found, using defaults")
		defaultCfg := config.DefaultConfig()
		cfg = &defaultCfg
	}

	// Enhance with additional processors
	*cfg = config.ConfigWithProcessor(*cfg, "validator", map[string]string{
		"strict": "true",
	})

	// Use the config
	engine, err := factory.NewEngineFromConfig(*cfg)
	if err != nil {
		log.Printf("Failed to create engine: %v", err)
		return
	}

	ctx := context.Background()
	if err := engine.RunWithContext(ctx); err != nil {
		log.Printf("Engine execution failed: %v", err)
		return
	}

	fmt.Println("✓ File + defaults combination executed")
}
