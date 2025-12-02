package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/AshishBagdane/report-engine/internal/processor"
	"github.com/AshishBagdane/report-engine/internal/registry"
)

func main() {
	fmt.Println("=== Go Report Engine - Parallel Processing Example ===")

	// Setup processor registrations
	setupProcessors()

	// Run examples
	example1_BasicParallelProcessing()
	example2_ConfiguredParallelProcessing()
	example3_ParallelFilter()
	example4_ParallelValidator()
	example5_ParallelTransformer()
	example6_PerformanceComparison()

	fmt.Println("\n=== All Examples Complete ===")
}

// setupProcessors registers test processors
func setupProcessors() {
	// Register a CPU-intensive filter
	registry.RegisterFilter("expensive_filter", &ExpensiveFilter{threshold: 50})

	// Register parallel version
	registry.RegisterParallelFilter("parallel_expensive_filter", &ExpensiveFilter{threshold: 50})

	// Register a validator
	registry.RegisterValidator("data_validator", &DataValidator{})

	// Register parallel version
	registry.RegisterParallelValidator("parallel_data_validator", &DataValidator{})

	// Register a transformer
	registry.RegisterTransformer("data_transformer", &DataTransformer{})

	// Register parallel version with custom config
	config := processor.ParallelConfig{
		Workers:      8,
		ChunkSize:    100,
		MinChunkSize: 50,
	}
	registry.RegisterParallelProcessorWithConfig("parallel_data_transformer_custom", config, func() processor.ProcessorHandler {
		return processor.NewTransformWrapper(&DataTransformer{})
	})
}

// Example 1: Basic parallel processing
func example1_BasicParallelProcessing() {
	fmt.Println("--- Example 1: Basic Parallel Processing ---")

	// Create a base processor
	baseProcessor := &processor.BaseProcessor{}

	// Wrap it in parallel processor with default config
	parallelProc := processor.NewParallelProcessor(baseProcessor)
	defer parallelProc.Close()

	// Process data
	data := generateTestData(1000)
	ctx := context.Background()

	start := time.Now()
	result, err := parallelProc.Process(ctx, data)
	duration := time.Since(start)

	if err != nil {
		log.Fatalf("Processing failed: %v", err)
	}

	fmt.Printf("✓ Processed %d records in %v\n", len(result), duration)
	fmt.Printf("  Workers: %d (default: runtime.NumCPU)\n", parallelProc.Workers())
	fmt.Printf("  Chunk Size: auto-calculated\n\n")
}

// Example 2: Configured parallel processing
func example2_ConfiguredParallelProcessing() {
	fmt.Println("--- Example 2: Configured Parallel Processing ---")

	// Create base processor
	baseProcessor := &processor.BaseProcessor{}

	// Configure parallel processing
	config := processor.ParallelConfig{
		Workers:      16,
		ChunkSize:    200,
		MinChunkSize: 50,
	}

	parallelProc := processor.NewParallelProcessorWithConfig(baseProcessor, config)
	defer parallelProc.Close()

	// Process data
	data := generateTestData(2000)
	ctx := context.Background()

	start := time.Now()
	result, err := parallelProc.Process(ctx, data)
	duration := time.Since(start)

	if err != nil {
		log.Fatalf("Processing failed: %v", err)
	}

	fmt.Printf("✓ Processed %d records in %v\n", len(result), duration)
	fmt.Printf("  Workers: %d\n", parallelProc.Workers())
	fmt.Printf("  Chunk Size: %d\n", parallelProc.ChunkSize())
	fmt.Printf("  Min Chunk Size: 50\n\n")
}

// Example 3: Parallel filter processing
func example3_ParallelFilter() {
	fmt.Println("--- Example 3: Parallel Filter Processing ---")

	// Get parallel filter from registry
	parallelFilter, err := registry.GetProcessor("parallel_expensive_filter")
	if err != nil {
		log.Fatalf("Failed to get processor: %v", err)
	}
	defer closeProcessor(parallelFilter)

	// Process data
	data := generateTestData(5000)
	ctx := context.Background()

	start := time.Now()
	result, err := parallelFilter.Process(ctx, data)
	duration := time.Since(start)

	if err != nil {
		log.Fatalf("Processing failed: %v", err)
	}

	fmt.Printf("✓ Filtered %d → %d records in %v\n", len(data), len(result), duration)
	fmt.Printf("  Filter: ExpensiveFilter (threshold: 50)\n")
	fmt.Printf("  Mode: Parallel\n\n")
}

// Example 4: Parallel validator processing
func example4_ParallelValidator() {
	fmt.Println("--- Example 4: Parallel Validator Processing ---")

	// Get parallel validator from registry
	parallelValidator, err := registry.GetProcessor("parallel_data_validator")
	if err != nil {
		log.Fatalf("Failed to get processor: %v", err)
	}
	defer closeProcessor(parallelValidator)

	// Generate valid data
	data := generateValidData(3000)
	ctx := context.Background()

	start := time.Now()
	result, err := parallelValidator.Process(ctx, data)
	duration := time.Since(start)

	if err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	fmt.Printf("✓ Validated %d records in %v\n", len(result), duration)
	fmt.Printf("  Validator: DataValidator\n")
	fmt.Printf("  Mode: Parallel\n\n")
}

// Example 5: Parallel transformer processing
func example5_ParallelTransformer() {
	fmt.Println("--- Example 5: Parallel Transformer Processing ---")

	// Get custom configured parallel transformer
	parallelTransformer, err := registry.GetProcessor("parallel_data_transformer_custom")
	if err != nil {
		log.Fatalf("Failed to get processor: %v", err)
	}
	defer closeProcessor(parallelTransformer)

	// Process data
	data := generateTestData(4000)
	ctx := context.Background()

	start := time.Now()
	result, err := parallelTransformer.Process(ctx, data)
	duration := time.Since(start)

	if err != nil {
		log.Fatalf("Processing failed: %v", err)
	}

	fmt.Printf("✓ Transformed %d records in %v\n", len(result), duration)
	fmt.Printf("  Transformer: DataTransformer\n")
	fmt.Printf("  Workers: 8 (custom config)\n")
	fmt.Printf("  Chunk Size: 100 (custom config)\n\n")
}

// Example 6: Performance comparison
func example6_PerformanceComparison() {
	fmt.Println("--- Example 6: Sequential vs Parallel Performance ---")

	data := generateTestData(10000)
	ctx := context.Background()

	// Sequential processing
	seqProcessor, _ := registry.GetProcessor("expensive_filter")
	start := time.Now()
	seqResult, err := seqProcessor.Process(ctx, data)
	seqDuration := time.Since(start)
	if err != nil {
		log.Fatalf("Sequential processing failed: %v", err)
	}

	// Parallel processing
	parProcessor, _ := registry.GetProcessor("parallel_expensive_filter")
	start = time.Now()
	parResult, err := parProcessor.Process(ctx, data)
	parDuration := time.Since(start)
	if err != nil {
		log.Fatalf("Parallel processing failed: %v", err)
	}
	closeProcessor(parProcessor)

	// Compare results
	speedup := float64(seqDuration) / float64(parDuration)

	fmt.Printf("Dataset: %d records\n", len(data))
	fmt.Printf("\nSequential Processing:\n")
	fmt.Printf("  Duration: %v\n", seqDuration)
	fmt.Printf("  Result: %d records\n", len(seqResult))
	fmt.Printf("\nParallel Processing:\n")
	fmt.Printf("  Duration: %v\n", parDuration)
	fmt.Printf("  Result: %d records\n", len(parResult))
	fmt.Printf("  Workers: %d\n", runtime.NumCPU())
	fmt.Printf("\nSpeedup: %.2fx faster\n\n", speedup)
}

// --- Helper Functions ---

func generateTestData(count int) []map[string]interface{} {
	data := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		data[i] = map[string]interface{}{
			"id":    i,
			"value": i % 100,
			"name":  fmt.Sprintf("record_%d", i),
		}
	}
	return data
}

func generateValidData(count int) []map[string]interface{} {
	data := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		data[i] = map[string]interface{}{
			"id":             i,
			"name":           fmt.Sprintf("record_%d", i),
			"email":          fmt.Sprintf("user%d@example.com", i),
			"required_field": "present",
		}
	}
	return data
}

func closeProcessor(proc processor.ProcessorHandler) {
	if closer, ok := proc.(interface{ Close() error }); ok {
		closer.Close()
	}
}

// --- Mock Processor Strategies ---

// ExpensiveFilter simulates CPU-intensive filtering
type ExpensiveFilter struct {
	threshold int
}

func (f *ExpensiveFilter) Keep(row map[string]interface{}) bool {
	// Simulate expensive computation
	value, ok := row["value"].(int)
	if !ok {
		return false
	}

	// Artificial delay to simulate CPU work
	result := 0
	for i := 0; i < 1000; i++ {
		result += i % (value + 1)
	}

	return value >= f.threshold
}

// DataValidator simulates validation logic
type DataValidator struct{}

func (v *DataValidator) Validate(row map[string]interface{}) error {
	// Simulate expensive validation
	time.Sleep(100 * time.Microsecond)

	if _, ok := row["required_field"]; !ok {
		return fmt.Errorf("missing required_field")
	}

	if _, ok := row["email"]; !ok {
		return fmt.Errorf("missing email")
	}

	return nil
}

// DataTransformer simulates transformation logic
type DataTransformer struct{}

func (t *DataTransformer) Transform(row map[string]interface{}) map[string]interface{} {
	// Simulate expensive transformation
	result := make(map[string]interface{})

	for k, v := range row {
		result[k] = v
	}

	// Add computed field
	if id, ok := row["id"].(int); ok {
		result["computed"] = id * 2
	}

	// Simulate work
	for i := 0; i < 500; i++ {
		_ = i * i
	}

	return result
}
