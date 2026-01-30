package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/AshishBagdane/go-report-engine/internal/config"
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
		return formatter.NewJSONFormatter("") // Placeholder
	})

	registry.RegisterOutput("console", func() output.OutputStrategy {
		return output.NewConsoleOutput()
	})

	registry.RegisterOutput("file", func() output.OutputStrategy {
		return output.NewConsoleOutput() // Placeholder
	})
}

func main() {
	fmt.Println("=== Integration Patterns Demo ===")
	fmt.Println()

	// Example 1: One-step load and build
	example1()

	// Example 2: Load with environment overrides
	example2()

	// Example 3: Build from preset configs
	example3()

	// Example 4: Build from raw bytes
	example4()

	// Example 5: Load with fallback to default
	example5()

	// Example 6: Must variants for initialization
	example6()

	// Example 7: Explicit validation
	example7()

	// Example 8: Config or file pattern
	example8()
}

// example1 demonstrates one-step load and build
func example1() {
	fmt.Println("--- Example 1: LoadAndBuild (One-Step) ---")

	// Single function call to load config and build engine
	engine, err := config.LoadAndBuild("examples/configs/config.minimal.yaml")
	if err != nil {
		log.Printf("Failed: %v", err)
		return
	}

	ctx := context.Background()
	if err := engine.RunWithContext(ctx); err != nil {
		log.Printf("Execution failed: %v", err)
		return
	}

	fmt.Println("✓ One-step load and build successful")
}

// example2 demonstrates load with environment overrides
func example2() {
	fmt.Println("--- Example 2: LoadAndBuildWithEnv ---")

	// Set environment variables
	os.Setenv("ENGINE_PROVIDER_TYPE", "mock")
	os.Setenv("ENGINE_FORMATTER_PARAM_INDENT", "4")
	defer func() {
		os.Unsetenv("ENGINE_PROVIDER_TYPE")
		os.Unsetenv("ENGINE_FORMATTER_PARAM_INDENT")
	}()

	// Load config with env overrides and build engine
	engine, err := config.LoadAndBuildWithEnv("examples/configs/config.minimal.yaml")
	if err != nil {
		log.Printf("Failed: %v", err)
		return
	}

	ctx := context.Background()
	if err := engine.RunWithContext(ctx); err != nil {
		log.Printf("Execution failed: %v", err)
		return
	}

	fmt.Println("✓ Load with env overrides successful")
}

// example3 demonstrates building from preset configs
func example3() {
	fmt.Println("--- Example 3: Build from Presets ---")

	// Build from default
	engine1, err := config.BuildFromDefault()
	if err != nil {
		log.Printf("BuildFromDefault failed: %v", err)
		return
	}
	fmt.Println("✓ Built from default config")

	// Build from development preset
	engine2, err := config.BuildFromDevelopment()
	if err != nil {
		log.Printf("BuildFromDevelopment failed: %v", err)
		return
	}
	fmt.Println("✓ Built from development config")

	// Build from production preset
	engine3, err := config.BuildFromProduction()
	if err != nil {
		log.Printf("BuildFromProduction failed: %v", err)
		return
	}
	fmt.Println("✓ Built from production config")

	// Build from testing preset
	engine4, err := config.BuildFromTesting()
	if err != nil {
		log.Printf("BuildFromTesting failed: %v", err)
		return
	}
	fmt.Println("✓ Built from testing config")

	// Run one of them
	ctx := context.Background()
	if err := engine1.RunWithContext(ctx); err != nil {
		log.Printf("Execution failed: %v", err)
		return
	}

	// Suppress unused variable warnings
	_ = engine2
	_ = engine3
	_ = engine4

	fmt.Println()
}

// example4 demonstrates building from raw bytes
func example4() {
	fmt.Println("--- Example 4: Build from Raw Bytes ---")

	// YAML configuration as bytes
	yamlConfig := []byte(`
provider:
  type: mock
processors: []
formatter:
  type: json
  params:
    indent: "2"
output:
  type: console
`)

	// Build directly from bytes
	engine, err := config.BuildFromBytes(yamlConfig, "yaml")
	if err != nil {
		log.Printf("Failed: %v", err)
		return
	}

	ctx := context.Background()
	if err := engine.RunWithContext(ctx); err != nil {
		log.Printf("Execution failed: %v", err)
		return
	}

	fmt.Println("✓ Built from raw YAML bytes")
}

// example5 demonstrates load with fallback to default
func example5() {
	fmt.Println("--- Example 5: Load with Fallback ---")

	// Try to load config, fall back to default if not found
	cfg, err := config.LoadOrDefault("nonexistent-config.yaml")
	if err != nil {
		fmt.Printf("Config file not found, using default: %v\n", err)
	} else {
		fmt.Println("Config loaded from file")
	}

	// Build engine from config (whether from file or default)
	engine, err := config.ValidateAndBuild(*cfg)
	if err != nil {
		log.Printf("Failed to build engine: %v", err)
		return
	}

	ctx := context.Background()
	if err := engine.RunWithContext(ctx); err != nil {
		log.Printf("Execution failed: %v", err)
		return
	}

	fmt.Println("✓ Engine built and executed successfully")
}

// example6 demonstrates Must variants for initialization
func example6() {
	fmt.Println("--- Example 6: Must Variants (Panic on Error) ---")

	// Must variants are useful in init() or main() where failure should stop the program
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Caught panic (expected): %v\n", r)
		}
	}()

	// This would panic if it failed, but we have a valid config
	engine := config.MustBuildFromDefault()
	fmt.Println("✓ MustBuildFromDefault succeeded")

	ctx := context.Background()
	if err := engine.RunWithContext(ctx); err != nil {
		log.Printf("Execution failed: %v", err)
		return
	}

	// Demonstrate panic behavior with invalid path
	fmt.Println("Testing panic behavior with invalid path...")
	config.MustLoadAndBuild("/invalid/path/config.yaml")

	// This line won't be reached due to panic above
	fmt.Println("This won't print")
}

// example7 demonstrates explicit validation
func example7() {
	fmt.Println("\n--- Example 7: Explicit Validation ---")

	// Start with a config
	cfg := config.DefaultConfig()

	// Modify it
	cfg.Provider.Type = "mock"
	cfg = config.ConfigWithProcessor(cfg, "validator", map[string]string{
		"strict": "true",
	})

	// Explicitly validate before building
	engine, err := config.ValidateAndBuild(cfg)
	if err != nil {
		log.Printf("Validation or build failed: %v", err)
		return
	}

	fmt.Println("✓ Config validated and engine built")

	ctx := context.Background()
	if err := engine.RunWithContext(ctx); err != nil {
		log.Printf("Execution failed: %v", err)
		return
	}

	fmt.Println()
}

// example8 demonstrates config or file pattern
func example8() {
	fmt.Println("--- Example 8: Config or File Pattern ---")

	// Try with nil config (will load from file)
	engine1, err := config.BuildFromConfigOrFile(nil, "examples/configs/config.minimal.yaml")
	if err != nil {
		log.Printf("Failed to build from file: %v", err)
	} else {
		fmt.Println("✓ Built from file (nil config provided)")
	}

	// Try with provided config (will use config, ignore file)
	cfg := config.DefaultConfig()
	engine2, err := config.BuildFromConfigOrFile(&cfg, "nonexistent.yaml")
	if err != nil {
		log.Printf("Failed to build from config: %v", err)
	} else {
		fmt.Println("✓ Built from provided config (file ignored)")
	}

	// Run one of them
	if engine1 != nil {
		ctx := context.Background()
		if err := engine1.RunWithContext(ctx); err != nil {
			log.Printf("Execution failed: %v", err)
			return
		}
	}

	// Suppress unused variable warnings
	_ = engine2

	fmt.Println()
}

// Production usage pattern example
func productionUsageExample() {
	fmt.Println("=== Production Usage Pattern ===")

	// In production, you might do:
	engine, err := config.LoadAndBuildWithEnv("/etc/myapp/config.yaml")
	if err != nil {
		log.Fatalf("Failed to initialize engine: %v", err)
	}

	// Run reports
	ctx := context.Background()
	if err := engine.RunWithContext(ctx); err != nil {
		log.Printf("Report generation failed: %v", err)
		// Handle error, maybe retry, send alert, etc.
	}
}

// Development usage pattern example
func developmentUsageExample() {
	fmt.Println("=== Development Usage Pattern ===")

	// In development, you might do:
	engine, err := config.LoadOrDefault("config.yaml")
	if err != nil {
		fmt.Printf("Using default config: %v\n", err)
	}

	validatedEngine, err := config.ValidateAndBuild(*engine)
	if err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	// Run with verbose output
	ctx := context.Background()
	if err := validatedEngine.RunWithContext(ctx); err != nil {
		log.Printf("Execution failed: %v", err)
	}
}

// Testing usage pattern example
func testingUsageExample() {
	fmt.Println("=== Testing Usage Pattern ===")

	// In tests, you might do:
	engine := config.MustBuildFromTesting()

	// Run test
	ctx := context.Background()
	if err := engine.RunWithContext(ctx); err != nil {
		log.Fatalf("Test failed: %v", err)
	}
}
