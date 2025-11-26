package factory

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/AshishBagdane/report-engine/internal/engine"
	"github.com/AshishBagdane/report-engine/internal/registry"
	"github.com/AshishBagdane/report-engine/pkg/api"
)

// Mock processor strategies for testing

type mockFilterStrategy struct {
	threshold int
}

func (m *mockFilterStrategy) Keep(row map[string]interface{}) bool {
	if val, ok := row["value"].(int); ok {
		return val >= m.threshold
	}
	return false
}

type mockConfigurableFilter struct {
	threshold  int
	configured bool
}

func (m *mockConfigurableFilter) Keep(row map[string]interface{}) bool {
	if val, ok := row["value"].(int); ok {
		return val >= m.threshold
	}
	return false
}

func (m *mockConfigurableFilter) Configure(params map[string]string) error {
	m.configured = true
	thresholdStr, ok := params["threshold"]
	if !ok {
		return api.ErrMissingParam("threshold")
	}
	// Simple parsing for test
	if thresholdStr == "50" {
		m.threshold = 50
	} else {
		m.threshold = 10
	}
	return nil
}

type mockValidatorStrategy struct{}

func (m *mockValidatorStrategy) Validate(row map[string]interface{}) error {
	if _, ok := row["required"]; !ok {
		return fmt.Errorf("missing required field")
	}
	return nil
}

type mockTransformerStrategy struct{}

func (m *mockTransformerStrategy) Transform(row map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range row {
		result[k] = v
	}
	result["transformed"] = true
	return result
}

// setupProcessorRegistries initializes processor registries for testing
func setupProcessorRegistries() {
	registry.ClearProcessors()

	registry.RegisterFilter("test_filter", &mockFilterStrategy{threshold: 50})
	registry.RegisterFilter("configurable_filter", &mockConfigurableFilter{})
	registry.RegisterValidator("test_validator", &mockValidatorStrategy{})
	registry.RegisterTransformer("test_transformer", &mockTransformerStrategy{})
}

// TestBuildProcessorChainEmpty tests building empty chain
func TestBuildProcessorChainEmpty(t *testing.T) {
	setupProcessorRegistries()

	configs := []engine.ProcessorConfig{}

	chain, err := BuildProcessorChain(configs)

	if err != nil {
		t.Fatalf("BuildProcessorChain() returned error: %v", err)
	}

	if chain == nil {
		t.Fatal("BuildProcessorChain() returned nil chain")
	}

	// Should be able to process data
	testData := []map[string]interface{}{
		{"id": 1},
	}

	ctx := context.Background()
	result, err := chain.Process(ctx, testData)
	if err != nil {
		t.Errorf("Empty chain Process() failed: %v", err)
	}
	if len(result) != len(testData) {
		t.Errorf("Empty chain modified data length")
	}
}

// TestBuildProcessorChainSingle tests building chain with single processor
func TestBuildProcessorChainSingle(t *testing.T) {
	setupProcessorRegistries()

	configs := []engine.ProcessorConfig{
		{Type: "test_filter", Params: map[string]string{}},
	}

	chain, err := BuildProcessorChain(configs)

	if err != nil {
		t.Fatalf("BuildProcessorChain() returned error: %v", err)
	}

	if chain == nil {
		t.Fatal("BuildProcessorChain() returned nil chain")
	}

	// Test that filter works
	testData := []map[string]interface{}{
		{"id": 1, "value": 30},
		{"id": 2, "value": 60},
	}

	ctx := context.Background()
	result, err := chain.Process(ctx, testData)
	if err != nil {
		t.Fatalf("Chain Process() failed: %v", err)
	}

	// Should filter out value < 50
	if len(result) != 1 {
		t.Errorf("Filter chain returned %d records, expected 1", len(result))
	}
}

// TestBuildProcessorChainMultiple tests building chain with multiple processors
func TestBuildProcessorChainMultiple(t *testing.T) {
	setupProcessorRegistries()

	configs := []engine.ProcessorConfig{
		{Type: "test_filter", Params: map[string]string{}},
		{Type: "test_validator", Params: map[string]string{}},
		{Type: "test_transformer", Params: map[string]string{}},
	}

	chain, err := BuildProcessorChain(configs)

	if err != nil {
		t.Fatalf("BuildProcessorChain() returned error: %v", err)
	}

	if chain == nil {
		t.Fatal("BuildProcessorChain() returned nil chain")
	}

	// Test full chain
	testData := []map[string]interface{}{
		{"id": 1, "value": 60, "required": "yes"},
	}

	ctx := context.Background()
	result, err := chain.Process(ctx, testData)
	if err != nil {
		t.Fatalf("Chain Process() failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Chain returned %d records, expected 1", len(result))
	}

	// Verify transformation occurred
	if _, ok := result[0]["transformed"]; !ok {
		t.Error("Transformation did not occur")
	}
}

// TestBuildProcessorChainUnknownType tests with unknown processor type
func TestBuildProcessorChainUnknownType(t *testing.T) {
	setupProcessorRegistries()

	configs := []engine.ProcessorConfig{
		{Type: "nonexistent_processor", Params: map[string]string{}},
	}

	chain, err := BuildProcessorChain(configs)

	if err == nil {
		t.Fatal("Expected error for unknown processor type")
	}

	if chain != nil {
		t.Error("Should return nil chain on error")
	}

	if !strings.Contains(err.Error(), "factory failed") {
		t.Errorf("Error should mention factory failure, got: %v", err)
	}
}

// TestBuildProcessorChainConfigurable tests with configurable processor
func TestBuildProcessorChainConfigurable(t *testing.T) {
	setupProcessorRegistries()

	configs := []engine.ProcessorConfig{
		{
			Type: "configurable_filter",
			Params: map[string]string{
				"threshold": "50",
			},
		},
	}

	chain, err := BuildProcessorChain(configs)

	if err != nil {
		t.Fatalf("BuildProcessorChain() with configurable returned error: %v", err)
	}

	// Test that configured threshold is used
	testData := []map[string]interface{}{
		{"id": 1, "value": 30},
		{"id": 2, "value": 60},
	}

	ctx := context.Background()
	result, err := chain.Process(ctx, testData)
	if err != nil {
		t.Fatalf("Chain Process() failed: %v", err)
	}

	// With threshold=50, should keep only value >= 50
	if len(result) != 1 {
		t.Errorf("Configured filter returned %d records, expected 1", len(result))
	}
}

// failingConfigFilter is a test filter that fails configuration
type failingConfigFilter struct {
	mockFilterStrategy
}

func (f *failingConfigFilter) Configure(params map[string]string) error {
	return fmt.Errorf("configuration failed")
}

// TestBuildProcessorChainConfigurationError tests config error handling
func TestBuildProcessorChainConfigurationError(t *testing.T) {
	setupProcessorRegistries()

	// Register a processor that fails configuration
	registry.RegisterFilter("failing_filter", &failingConfigFilter{})

	configs := []engine.ProcessorConfig{
		{
			Type:   "failing_filter",
			Params: map[string]string{},
		},
	}

	chain, err := BuildProcessorChain(configs)

	if err == nil {
		t.Fatal("Expected error for configuration failure")
	}

	if chain != nil {
		t.Error("Should return nil chain on error")
	}

	if !strings.Contains(err.Error(), "configuration failed") {
		t.Errorf("Error should mention configuration failure, got: %v", err)
	}
}

// TestBuildProcessorChainMissingParams tests with missing required params
func TestBuildProcessorChainMissingParams(t *testing.T) {
	setupProcessorRegistries()

	// configurable_filter requires "threshold" param
	configs := []engine.ProcessorConfig{
		{
			Type:   "configurable_filter",
			Params: map[string]string{}, // Missing threshold
		},
	}

	chain, err := BuildProcessorChain(configs)

	if err == nil {
		t.Fatal("Expected error for missing params")
	}

	if chain != nil {
		t.Error("Should return nil chain on error")
	}
}

// TestBuildProcessorChainOrdering tests that processors are chained in order
func TestBuildProcessorChainOrdering(t *testing.T) {
	setupProcessorRegistries()

	// Order: Transform first, then validate
	// This should add "transformed" field before validation
	configs := []engine.ProcessorConfig{
		{Type: "test_transformer", Params: map[string]string{}},
		{Type: "test_validator", Params: map[string]string{}},
	}

	chain, err := BuildProcessorChain(configs)
	if err != nil {
		t.Fatalf("BuildProcessorChain() failed: %v", err)
	}

	testData := []map[string]interface{}{
		{"id": 1, "required": "yes"},
	}

	ctx := context.Background()
	result, err := chain.Process(ctx, testData)
	if err != nil {
		t.Fatalf("Chain Process() failed: %v", err)
	}

	// Verify both transformation and validation occurred
	if _, ok := result[0]["transformed"]; !ok {
		t.Error("Transformation did not occur")
	}
}

// TestBuildProcessorChainErrorPropagation tests error propagation through chain
func TestBuildProcessorChainErrorPropagation(t *testing.T) {
	setupProcessorRegistries()

	configs := []engine.ProcessorConfig{
		{Type: "test_filter", Params: map[string]string{}},
		{Type: "test_validator", Params: map[string]string{}}, // Will fail if "required" missing
		{Type: "test_transformer", Params: map[string]string{}},
	}

	chain, err := BuildProcessorChain(configs)
	if err != nil {
		t.Fatalf("BuildProcessorChain() failed: %v", err)
	}

	// Data missing "required" field - should fail validation
	testData := []map[string]interface{}{
		{"id": 1, "value": 60}, // Missing "required"
	}

	ctx := context.Background()
	result, err := chain.Process(ctx, testData)

	if err == nil {
		t.Fatal("Expected validation error")
	}

	if result != nil {
		t.Error("Should return nil on validation failure")
	}

	if !strings.Contains(err.Error(), "missing required field") {
		t.Errorf("Error should mention validation failure, got: %v", err)
	}
}

// TestBuildProcessorChainNilConfigs tests with nil configs
func TestBuildProcessorChainNilConfigs(t *testing.T) {
	setupProcessorRegistries()

	chain, err := BuildProcessorChain(nil)

	if err != nil {
		t.Fatalf("BuildProcessorChain(nil) returned error: %v", err)
	}

	if chain == nil {
		t.Fatal("BuildProcessorChain(nil) returned nil chain")
	}

	// Should return base processor
	testData := []map[string]interface{}{
		{"id": 1},
	}

	ctx := context.Background()
	result, err := chain.Process(ctx, testData)
	if err != nil {
		t.Errorf("Chain Process() failed: %v", err)
	}
	if len(result) != len(testData) {
		t.Error("Nil config chain modified data")
	}
}

// TestBuildProcessorChainLongChain tests with many processors
func TestBuildProcessorChainLongChain(t *testing.T) {
	setupProcessorRegistries()

	// Build a long chain of processors
	configs := []engine.ProcessorConfig{
		{Type: "test_transformer", Params: map[string]string{}},
		{Type: "test_filter", Params: map[string]string{}},
		{Type: "test_transformer", Params: map[string]string{}},
		{Type: "test_validator", Params: map[string]string{}},
		{Type: "test_transformer", Params: map[string]string{}},
	}

	chain, err := BuildProcessorChain(configs)
	if err != nil {
		t.Fatalf("BuildProcessorChain() with long chain failed: %v", err)
	}

	testData := []map[string]interface{}{
		{"id": 1, "value": 60, "required": "yes"},
	}

	ctx := context.Background()
	result, err := chain.Process(ctx, testData)
	if err != nil {
		t.Fatalf("Long chain Process() failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Long chain returned %d records, expected 1", len(result))
	}
}

// TestBuildProcessorChainMultipleCalls tests creating multiple chains
func TestBuildProcessorChainMultipleCalls(t *testing.T) {
	setupProcessorRegistries()

	configs := []engine.ProcessorConfig{
		{Type: "test_filter", Params: map[string]string{}},
	}

	chain1, err1 := BuildProcessorChain(configs)
	if err1 != nil {
		t.Fatalf("First BuildProcessorChain() failed: %v", err1)
	}

	chain2, err2 := BuildProcessorChain(configs)
	if err2 != nil {
		t.Fatalf("Second BuildProcessorChain() failed: %v", err2)
	}

	// Should create separate instances
	if chain1 == chain2 {
		t.Error("BuildProcessorChain() should create new instances")
	}
}

// TestBuildProcessorChainConcurrent tests concurrent chain building
func TestBuildProcessorChainConcurrent(t *testing.T) {
	setupProcessorRegistries()

	configs := []engine.ProcessorConfig{
		{Type: "test_filter", Params: map[string]string{}},
	}

	const goroutines = 10
	errors := make(chan error, goroutines)
	chains := make(chan interface{}, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			chain, err := BuildProcessorChain(configs)
			if err != nil {
				errors <- err
				return
			}
			chains <- chain
		}()
	}

	for i := 0; i < goroutines; i++ {
		select {
		case err := <-errors:
			t.Errorf("Concurrent BuildProcessorChain() failed: %v", err)
		case chain := <-chains:
			if chain == nil {
				t.Error("Concurrent BuildProcessorChain() returned nil")
			}
		}
	}
}

// BenchmarkBuildProcessorChainSingle benchmarks single processor
func BenchmarkBuildProcessorChainSingle(b *testing.B) {
	setupProcessorRegistries()

	configs := []engine.ProcessorConfig{
		{Type: "test_filter", Params: map[string]string{}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildProcessorChain(configs)
	}
}

// BenchmarkBuildProcessorChainMultiple benchmarks multiple processors
func BenchmarkBuildProcessorChainMultiple(b *testing.B) {
	setupProcessorRegistries()

	configs := []engine.ProcessorConfig{
		{Type: "test_filter", Params: map[string]string{}},
		{Type: "test_validator", Params: map[string]string{}},
		{Type: "test_transformer", Params: map[string]string{}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildProcessorChain(configs)
	}
}

// BenchmarkBuildProcessorChainConfigurable benchmarks with configuration
func BenchmarkBuildProcessorChainConfigurable(b *testing.B) {
	setupProcessorRegistries()

	configs := []engine.ProcessorConfig{
		{
			Type: "configurable_filter",
			Params: map[string]string{
				"threshold": "50",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildProcessorChain(configs)
	}
}
