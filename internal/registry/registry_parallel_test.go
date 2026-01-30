package registry

import (
	"context"
	"errors"
	"runtime"
	"testing"

	"github.com/AshishBagdane/go-report-engine/internal/processor"
)

// --- RegisterParallelProcessor Tests ---

func TestRegisterParallelProcessor(t *testing.T) {
	ClearProcessors()

	factory := func() processor.ProcessorHandler {
		return &processor.BaseProcessor{}
	}

	RegisterParallelProcessor("test_parallel", factory)

	if !IsProcessorRegistered("test_parallel") {
		t.Error("Parallel processor was not registered")
	}

	// Retrieve and verify it's a ParallelProcessor
	proc, err := GetProcessor("test_parallel")
	if err != nil {
		t.Fatalf("Failed to get parallel processor: %v", err)
	}

	parallel, ok := proc.(*processor.ParallelProcessor)
	if !ok {
		t.Fatal("Retrieved processor is not a ParallelProcessor")
	}

	// Verify default configuration
	if parallel.Workers() != runtime.NumCPU() {
		t.Errorf("Workers = %d, expected %d", parallel.Workers(), runtime.NumCPU())
	}
}

func TestRegisterParallelProcessor_NilFactoryPanics(t *testing.T) {
	ClearProcessors()

	defer func() {
		if r := recover(); r == nil {
			t.Error("RegisterParallelProcessor with nil factory should panic")
		}
	}()

	RegisterParallelProcessor("test", nil)
}

func TestRegisterParallelProcessor_EmptyNamePanics(t *testing.T) {
	ClearProcessors()

	defer func() {
		if r := recover(); r == nil {
			t.Error("RegisterParallelProcessor with empty name should panic")
		}
	}()

	factory := func() processor.ProcessorHandler {
		return &processor.BaseProcessor{}
	}

	RegisterParallelProcessor("", factory)
}

func TestRegisterParallelProcessor_Functional(t *testing.T) {
	ClearProcessors()

	// Register a simple processor wrapped in parallel
	factory := func() processor.ProcessorHandler {
		return &processor.BaseProcessor{}
	}

	RegisterParallelProcessor("test_parallel", factory)

	proc, err := GetProcessor("test_parallel")
	if err != nil {
		t.Fatalf("Failed to get processor: %v", err)
	}

	// Test that it works
	ctx := context.Background()
	data := make([]map[string]interface{}, 200)
	for i := range data {
		data[i] = map[string]interface{}{"id": i}
	}

	result, err := proc.Process(ctx, data)
	if err != nil {
		t.Fatalf("Process() failed: %v", err)
	}

	if len(result) != 200 {
		t.Errorf("Process() returned %d records, expected 200", len(result))
	}
}

// --- RegisterParallelProcessorWithConfig Tests ---

func TestRegisterParallelProcessorWithConfig(t *testing.T) {
	ClearProcessors()

	factory := func() processor.ProcessorHandler {
		return &processor.BaseProcessor{}
	}

	config := processor.ParallelConfig{
		Workers:      16,
		ChunkSize:    500,
		MinChunkSize: 50,
	}

	RegisterParallelProcessorWithConfig("test_configured", config, factory)

	if !IsProcessorRegistered("test_configured") {
		t.Error("Configured parallel processor was not registered")
	}

	// Retrieve and verify configuration
	proc, err := GetProcessor("test_configured")
	if err != nil {
		t.Fatalf("Failed to get processor: %v", err)
	}

	parallel, ok := proc.(*processor.ParallelProcessor)
	if !ok {
		t.Fatal("Retrieved processor is not a ParallelProcessor")
	}

	if parallel.Workers() != 16 {
		t.Errorf("Workers = %d, expected 16", parallel.Workers())
	}

	if parallel.ChunkSize() != 500 {
		t.Errorf("ChunkSize = %d, expected 500", parallel.ChunkSize())
	}
}

func TestRegisterParallelProcessorWithConfig_NilFactoryPanics(t *testing.T) {
	ClearProcessors()

	defer func() {
		if r := recover(); r == nil {
			t.Error("RegisterParallelProcessorWithConfig with nil factory should panic")
		}
	}()

	config := processor.ParallelConfig{Workers: 4}
	RegisterParallelProcessorWithConfig("test", config, nil)
}

func TestRegisterParallelProcessorWithConfig_EmptyNamePanics(t *testing.T) {
	ClearProcessors()

	defer func() {
		if r := recover(); r == nil {
			t.Error("RegisterParallelProcessorWithConfig with empty name should panic")
		}
	}()

	factory := func() processor.ProcessorHandler {
		return &processor.BaseProcessor{}
	}
	config := processor.ParallelConfig{Workers: 4}

	RegisterParallelProcessorWithConfig("", config, factory)
}

// --- RegisterParallelFilter Tests ---

func TestRegisterParallelFilter(t *testing.T) {
	ClearProcessors()

	filter := &mockParallelFilterStrategy{threshold: 50}
	RegisterParallelFilter("parallel_filter", filter)

	if !IsProcessorRegistered("parallel_filter") {
		t.Error("Parallel filter was not registered")
	}

	// Retrieve and verify it's a ParallelProcessor wrapping FilterWrapper
	proc, err := GetProcessor("parallel_filter")
	if err != nil {
		t.Fatalf("Failed to get processor: %v", err)
	}

	parallel, ok := proc.(*processor.ParallelProcessor)
	if !ok {
		t.Fatal("Retrieved processor is not a ParallelProcessor")
	}

	// Test functionality
	ctx := context.Background()
	data := make([]map[string]interface{}, 200)
	for i := range data {
		data[i] = map[string]interface{}{"value": i}
	}

	result, err := parallel.Process(ctx, data)
	if err != nil {
		t.Fatalf("Process() failed: %v", err)
	}

	// Should have filtered out values < 50
	expectedCount := 150
	if len(result) != expectedCount {
		t.Errorf("Process() returned %d records, expected %d", len(result), expectedCount)
	}
}

func TestRegisterParallelFilter_NilStrategyPanics(t *testing.T) {
	ClearProcessors()

	defer func() {
		if r := recover(); r == nil {
			t.Error("RegisterParallelFilter with nil strategy should panic")
		}
	}()

	RegisterParallelFilter("test", nil)
}

// --- RegisterParallelValidator Tests ---

func TestRegisterParallelValidator(t *testing.T) {
	ClearProcessors()

	validator := &mockParallelValidatorStrategy{}
	RegisterParallelValidator("parallel_validator", validator)

	if !IsProcessorRegistered("parallel_validator") {
		t.Error("Parallel validator was not registered")
	}

	// Retrieve and verify
	proc, err := GetProcessor("parallel_validator")
	if err != nil {
		t.Fatalf("Failed to get processor: %v", err)
	}

	parallel, ok := proc.(*processor.ParallelProcessor)
	if !ok {
		t.Fatal("Retrieved processor is not a ParallelProcessor")
	}

	// Test with valid data
	ctx := context.Background()
	validData := make([]map[string]interface{}, 200)
	for i := range validData {
		validData[i] = map[string]interface{}{
			"id":             i,
			"required_field": "present",
		}
	}

	result, err := parallel.Process(ctx, validData)
	if err != nil {
		t.Fatalf("Process() with valid data failed: %v", err)
	}

	if len(result) != 200 {
		t.Errorf("Process() returned %d records, expected 200", len(result))
	}

	// Test with invalid data
	invalidData := make([]map[string]interface{}, 200)
	for i := range invalidData {
		if i == 100 {
			// Missing required_field
			invalidData[i] = map[string]interface{}{"id": i}
		} else {
			invalidData[i] = map[string]interface{}{
				"id":             i,
				"required_field": "present",
			}
		}
	}

	_, err = parallel.Process(ctx, invalidData)
	if err == nil {
		t.Fatal("Process() should fail with invalid data")
	}
}

func TestRegisterParallelValidator_NilStrategyPanics(t *testing.T) {
	ClearProcessors()

	defer func() {
		if r := recover(); r == nil {
			t.Error("RegisterParallelValidator with nil strategy should panic")
		}
	}()

	RegisterParallelValidator("test", nil)
}

// --- RegisterParallelTransformer Tests ---

func TestRegisterParallelTransformer(t *testing.T) {
	ClearProcessors()

	transformer := &mockParallelTransformerStrategy{}
	RegisterParallelTransformer("parallel_transformer", transformer)

	if !IsProcessorRegistered("parallel_transformer") {
		t.Error("Parallel transformer was not registered")
	}

	// Retrieve and verify
	proc, err := GetProcessor("parallel_transformer")
	if err != nil {
		t.Fatalf("Failed to get processor: %v", err)
	}

	parallel, ok := proc.(*processor.ParallelProcessor)
	if !ok {
		t.Fatal("Retrieved processor is not a ParallelProcessor")
	}

	// Test functionality
	ctx := context.Background()
	data := make([]map[string]interface{}, 200)
	for i := range data {
		data[i] = map[string]interface{}{"name": "alice"}
	}

	result, err := parallel.Process(ctx, data)
	if err != nil {
		t.Fatalf("Process() failed: %v", err)
	}

	// Verify transformation occurred
	for i, row := range result {
		name := row["name"].(string)
		if name != "ALICE" {
			t.Errorf("Record %d: name = %q, expected ALICE", i, name)
		}
	}
}

func TestRegisterParallelTransformer_NilStrategyPanics(t *testing.T) {
	ClearProcessors()

	defer func() {
		if r := recover(); r == nil {
			t.Error("RegisterParallelTransformer with nil strategy should panic")
		}
	}()

	RegisterParallelTransformer("test", nil)
}

// --- Integration Tests ---

func TestRegisterParallelProcessor_WithExistingProcessor(t *testing.T) {
	ClearProcessors()

	// First register a regular processor
	RegisterProcessor("base_processor", func() processor.ProcessorHandler {
		return &processor.BaseProcessor{}
	})

	// Then wrap it in parallel
	RegisterParallelProcessor("parallel_base", func() processor.ProcessorHandler {
		proc, _ := GetProcessor("base_processor")
		return proc
	})

	// Verify both exist
	if !IsProcessorRegistered("base_processor") {
		t.Error("Base processor not registered")
	}

	if !IsProcessorRegistered("parallel_base") {
		t.Error("Parallel wrapper not registered")
	}

	// Verify parallel version works
	parallel, err := GetProcessor("parallel_base")
	if err != nil {
		t.Fatalf("Failed to get parallel processor: %v", err)
	}

	ctx := context.Background()
	data := make([]map[string]interface{}, 200)
	for i := range data {
		data[i] = map[string]interface{}{"id": i}
	}

	result, err := parallel.Process(ctx, data)
	if err != nil {
		t.Fatalf("Process() failed: %v", err)
	}

	if len(result) != 200 {
		t.Errorf("Process() returned %d records, expected 200", len(result))
	}
}

func TestRegisterParallelProcessor_MultipleRegistrations(t *testing.T) {
	ClearProcessors()

	// Register multiple parallel processors
	configs := []struct {
		name   string
		config processor.ParallelConfig
	}{
		{
			name: "parallel_4",
			config: processor.ParallelConfig{
				Workers:      4,
				ChunkSize:    100,
				MinChunkSize: 10,
			},
		},
		{
			name: "parallel_8",
			config: processor.ParallelConfig{
				Workers:      8,
				ChunkSize:    200,
				MinChunkSize: 20,
			},
		},
		{
			name: "parallel_16",
			config: processor.ParallelConfig{
				Workers:      16,
				ChunkSize:    500,
				MinChunkSize: 50,
			},
		},
	}

	for _, cfg := range configs {
		RegisterParallelProcessorWithConfig(cfg.name, cfg.config, func() processor.ProcessorHandler {
			return &processor.BaseProcessor{}
		})
	}

	// Verify all registered
	for _, cfg := range configs {
		if !IsProcessorRegistered(cfg.name) {
			t.Errorf("Processor %s was not registered", cfg.name)
		}

		proc, err := GetProcessor(cfg.name)
		if err != nil {
			t.Errorf("Failed to get processor %s: %v", cfg.name, err)
			continue
		}

		parallel, ok := proc.(*processor.ParallelProcessor)
		if !ok {
			t.Errorf("Processor %s is not a ParallelProcessor", cfg.name)
			continue
		}

		if parallel.Workers() != cfg.config.Workers {
			t.Errorf("Processor %s: workers = %d, expected %d", cfg.name, parallel.Workers(), cfg.config.Workers)
		}
	}
}

// --- Mock Strategies for Testing ---

type mockParallelFilterStrategy struct {
	threshold int
}

func (m *mockParallelFilterStrategy) Keep(row map[string]interface{}) bool {
	if val, ok := row["value"].(int); ok {
		return val >= m.threshold
	}
	return false
}

type mockParallelValidatorStrategy struct{}

func (m *mockParallelValidatorStrategy) Validate(row map[string]interface{}) error {
	if _, ok := row["required_field"]; !ok {
		return errors.New("missing required_field")
	}
	return nil
}

type mockParallelTransformerStrategy struct{}

func (m *mockParallelTransformerStrategy) Transform(row map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range row {
		if str, ok := v.(string); ok {
			result[k] = toUpperSimple(str)
		} else {
			result[k] = v
		}
	}
	return result
}

func toUpperSimple(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			result[i] = c - 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// --- Benchmark Tests ---

func BenchmarkRegisterParallelProcessor(b *testing.B) {
	ClearProcessors()

	factory := func() processor.ProcessorHandler {
		return &processor.BaseProcessor{}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RegisterParallelProcessor("bench_parallel", factory)
	}
}

func BenchmarkGetParallelProcessor(b *testing.B) {
	ClearProcessors()

	RegisterParallelProcessor("bench", func() processor.ProcessorHandler {
		return &processor.BaseProcessor{}
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetProcessor("bench")
	}
}

func BenchmarkRegisterParallelFilter(b *testing.B) {
	ClearProcessors()

	filter := &mockParallelFilterStrategy{threshold: 50}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RegisterParallelFilter("bench_filter", filter)
	}
}
