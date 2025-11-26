package processor

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/AshishBagdane/report-engine/pkg/api"
)

// Mock implementations for testing

// mockFilter for testing FilterWrapper
type mockFilter struct {
	threshold int
}

func (m *mockFilter) Keep(row map[string]interface{}) bool {
	if val, ok := row["value"].(int); ok {
		return val >= m.threshold
	}
	return false
}

// mockConfigurableFilter for testing Configure
type mockConfigurableFilter struct {
	mockFilter
	configured bool
}

func (m *mockConfigurableFilter) Configure(params map[string]string) error {
	m.configured = true
	return nil
}

// mockValidator for testing ValidatorWrapper
type mockValidator struct{}

func (m *mockValidator) Validate(row map[string]interface{}) error {
	if _, ok := row["required_field"]; !ok {
		return fmt.Errorf("missing required_field")
	}
	return nil
}

// mockConfigurableValidator for testing Configure
type mockConfigurableValidator struct {
	mockValidator
	minValue   int
	configured bool
}

func (m *mockConfigurableValidator) Configure(params map[string]string) error {
	m.configured = true
	m.minValue = 10
	return nil
}

// mockTransformer for testing TransformWrapper
type mockTransformer struct{}

func (m *mockTransformer) Transform(row map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range row {
		if str, ok := v.(string); ok {
			result[k] = strings.ToUpper(str)
		} else {
			result[k] = v
		}
	}
	return result
}

// mockConfigurableTransformer for testing Configure
type mockConfigurableTransformer struct {
	mockTransformer
	multiplier int
	configured bool
}

func (m *mockConfigurableTransformer) Configure(params map[string]string) error {
	m.configured = true
	m.multiplier = 2
	return nil
}

// --- FilterWrapper Tests ---

func TestNewFilterWrapper(t *testing.T) {
	strategy := &mockFilter{threshold: 10}
	wrapper := NewFilterWrapper(strategy)

	if wrapper == nil {
		t.Fatal("NewFilterWrapper returned nil")
	}

	if wrapper.strategy != strategy {
		t.Error("Strategy not set correctly")
	}
}

func TestFilterWrapperProcess(t *testing.T) {
	strategy := &mockFilter{threshold: 50}
	wrapper := NewFilterWrapper(strategy)

	testData := []map[string]interface{}{
		{"id": 1, "value": 30},
		{"id": 2, "value": 60},
		{"id": 3, "value": 45},
		{"id": 4, "value": 70},
	}

	ctx := context.Background()
	result, err := wrapper.Process(ctx, testData)

	if err != nil {
		t.Fatalf("Process() returned error: %v", err)
	}

	expectedCount := 2 // Only records with value >= 50
	if len(result) != expectedCount {
		t.Errorf("Process() returned %d records, expected %d", len(result), expectedCount)
	}

	// Verify only high-value records remain
	for _, row := range result {
		if val := row["value"].(int); val < 50 {
			t.Errorf("Found record with value %d, should have been filtered", val)
		}
	}
}

func TestFilterWrapperProcessEmpty(t *testing.T) {
	strategy := &mockFilter{threshold: 100}
	wrapper := NewFilterWrapper(strategy)

	testData := []map[string]interface{}{
		{"id": 1, "value": 10},
		{"id": 2, "value": 20},
	}

	ctx := context.Background()
	result, err := wrapper.Process(ctx, testData)

	if err != nil {
		t.Fatalf("Process() returned error: %v", err)
	}

	// All records should be filtered out
	if len(result) != 0 {
		t.Errorf("Process() returned %d records, expected 0", len(result))
	}
}

func TestFilterWrapperProcessWithNext(t *testing.T) {
	strategy := &mockFilter{threshold: 50}
	wrapper := NewFilterWrapper(strategy)

	next := &BaseProcessor{}
	wrapper.SetNext(next)

	testData := []map[string]interface{}{
		{"id": 1, "value": 60},
		{"id": 2, "value": 40},
	}

	ctx := context.Background()
	result, err := wrapper.Process(ctx, testData)

	if err != nil {
		t.Fatalf("Process() with next failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Process() returned %d records, expected 1", len(result))
	}
}

func TestFilterWrapperConfigure(t *testing.T) {
	tests := []struct {
		name            string
		strategy        api.FilterStrategy
		shouldConfigure bool
	}{
		{
			name:            "configurable filter",
			strategy:        &mockConfigurableFilter{},
			shouldConfigure: true,
		},
		{
			name:            "non-configurable filter",
			strategy:        &mockFilter{threshold: 10},
			shouldConfigure: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapper := NewFilterWrapper(tt.strategy)

			params := map[string]string{"key": "value"}
			err := wrapper.Configure(params)

			if err != nil {
				t.Errorf("Configure() returned error: %v", err)
			}

			if tt.shouldConfigure {
				if configurable, ok := tt.strategy.(*mockConfigurableFilter); ok {
					if !configurable.configured {
						t.Error("Configurable filter was not configured")
					}
				}
			}
		})
	}
}

// --- ValidatorWrapper Tests ---

func TestNewValidatorWrapper(t *testing.T) {
	strategy := &mockValidator{}
	wrapper := NewValidatorWrapper(strategy)

	if wrapper == nil {
		t.Fatal("NewValidatorWrapper returned nil")
	}

	if wrapper.strategy != strategy {
		t.Error("Strategy not set correctly")
	}
}

func TestValidatorWrapperProcessSuccess(t *testing.T) {
	strategy := &mockValidator{}
	wrapper := NewValidatorWrapper(strategy)

	testData := []map[string]interface{}{
		{"id": 1, "required_field": "value1"},
		{"id": 2, "required_field": "value2"},
	}

	ctx := context.Background()
	result, err := wrapper.Process(ctx, testData)

	if err != nil {
		t.Fatalf("Process() returned error: %v", err)
	}

	if len(result) != len(testData) {
		t.Errorf("Process() returned %d records, expected %d", len(result), len(testData))
	}
}

func TestValidatorWrapperProcessFailure(t *testing.T) {
	strategy := &mockValidator{}
	wrapper := NewValidatorWrapper(strategy)

	testData := []map[string]interface{}{
		{"id": 1, "required_field": "value1"},
		{"id": 2}, // Missing required_field
		{"id": 3, "required_field": "value3"},
	}

	ctx := context.Background()
	result, err := wrapper.Process(ctx, testData)

	if err == nil {
		t.Fatal("Process() should have returned error for invalid data")
	}

	if result != nil {
		t.Error("Process() should return nil on validation failure")
	}

	expectedMsg := "missing required_field"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error message %q should contain %q", err.Error(), expectedMsg)
	}
}

func TestValidatorWrapperProcessWithNext(t *testing.T) {
	strategy := &mockValidator{}
	wrapper := NewValidatorWrapper(strategy)

	next := &BaseProcessor{}
	wrapper.SetNext(next)

	testData := []map[string]interface{}{
		{"id": 1, "required_field": "value1"},
	}

	ctx := context.Background()
	result, err := wrapper.Process(ctx, testData)

	if err != nil {
		t.Fatalf("Process() with next failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Process() returned %d records, expected 1", len(result))
	}
}

func TestValidatorWrapperConfigure(t *testing.T) {
	tests := []struct {
		name            string
		strategy        api.ValidatorStrategy
		shouldConfigure bool
	}{
		{
			name:            "configurable validator",
			strategy:        &mockConfigurableValidator{},
			shouldConfigure: true,
		},
		{
			name:            "non-configurable validator",
			strategy:        &mockValidator{},
			shouldConfigure: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapper := NewValidatorWrapper(tt.strategy)

			params := map[string]string{"min_value": "10"}
			err := wrapper.Configure(params)

			if err != nil {
				t.Errorf("Configure() returned error: %v", err)
			}

			if tt.shouldConfigure {
				if configurable, ok := tt.strategy.(*mockConfigurableValidator); ok {
					if !configurable.configured {
						t.Error("Configurable validator was not configured")
					}
				}
			}
		})
	}
}

// --- TransformWrapper Tests ---

func TestNewTransformWrapper(t *testing.T) {
	strategy := &mockTransformer{}
	wrapper := NewTransformWrapper(strategy)

	if wrapper == nil {
		t.Fatal("NewTransformWrapper returned nil")
	}

	if wrapper.strategy != strategy {
		t.Error("Strategy not set correctly")
	}
}

func TestTransformWrapperProcess(t *testing.T) {
	strategy := &mockTransformer{}
	wrapper := NewTransformWrapper(strategy)

	testData := []map[string]interface{}{
		{"id": 1, "name": "alice"},
		{"id": 2, "name": "bob"},
	}

	ctx := context.Background()
	result, err := wrapper.Process(ctx, testData)

	if err != nil {
		t.Fatalf("Process() returned error: %v", err)
	}

	if len(result) != len(testData) {
		t.Errorf("Process() returned %d records, expected %d", len(result), len(testData))
	}

	// Verify transformation
	for i, row := range result {
		originalName := testData[i]["name"].(string)
		transformedName := row["name"].(string)

		if transformedName != strings.ToUpper(originalName) {
			t.Errorf("Record %d: name not transformed correctly, got %q", i, transformedName)
		}

		// Verify non-string fields are preserved
		if row["id"] != testData[i]["id"] {
			t.Errorf("Record %d: id field was modified", i)
		}
	}
}

func TestTransformWrapperProcessEmpty(t *testing.T) {
	strategy := &mockTransformer{}
	wrapper := NewTransformWrapper(strategy)

	testData := []map[string]interface{}{}

	ctx := context.Background()
	result, err := wrapper.Process(ctx, testData)

	if err != nil {
		t.Fatalf("Process() returned error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Process() returned %d records, expected 0", len(result))
	}
}

func TestTransformWrapperProcessWithNext(t *testing.T) {
	strategy := &mockTransformer{}
	wrapper := NewTransformWrapper(strategy)

	next := &BaseProcessor{}
	wrapper.SetNext(next)

	testData := []map[string]interface{}{
		{"id": 1, "name": "test"},
	}

	ctx := context.Background()
	result, err := wrapper.Process(ctx, testData)

	if err != nil {
		t.Fatalf("Process() with next failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Process() returned %d records, expected 1", len(result))
	}
}

func TestTransformWrapperConfigure(t *testing.T) {
	tests := []struct {
		name            string
		strategy        api.TransformerStrategy
		shouldConfigure bool
	}{
		{
			name:            "configurable transformer",
			strategy:        &mockConfigurableTransformer{},
			shouldConfigure: true,
		},
		{
			name:            "non-configurable transformer",
			strategy:        &mockTransformer{},
			shouldConfigure: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapper := NewTransformWrapper(tt.strategy)

			params := map[string]string{"multiplier": "2"}
			err := wrapper.Configure(params)

			if err != nil {
				t.Errorf("Configure() returned error: %v", err)
			}

			if tt.shouldConfigure {
				if configurable, ok := tt.strategy.(*mockConfigurableTransformer); ok {
					if !configurable.configured {
						t.Error("Configurable transformer was not configured")
					}
				}
			}
		})
	}
}

// --- Context Tests ---

func TestFilterWrapperContextCancellation(t *testing.T) {
	strategy := &mockFilter{threshold: 0} // Keep all records
	wrapper := NewFilterWrapper(strategy)

	// Create large dataset
	testData := make([]map[string]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		testData[i] = map[string]interface{}{"id": i, "value": i}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result, err := wrapper.Process(ctx, testData)

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}

	// May have partial results
	_ = result
}

func TestValidatorWrapperContextCancellation(t *testing.T) {
	strategy := &mockValidator{}
	wrapper := NewValidatorWrapper(strategy)

	testData := []map[string]interface{}{
		{"id": 1, "required_field": "value1"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := wrapper.Process(ctx, testData)

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}
}

func TestTransformWrapperContextCancellation(t *testing.T) {
	strategy := &mockTransformer{}
	wrapper := NewTransformWrapper(strategy)

	// Create large dataset
	testData := make([]map[string]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		testData[i] = map[string]interface{}{"id": i, "name": "test"}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result, err := wrapper.Process(ctx, testData)

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}

	// May have partial results
	_ = result
}

// --- Integration Tests ---

func TestWrapperChaining(t *testing.T) {
	// Create a chain: Filter -> Validator -> Transformer
	filter := NewFilterWrapper(&mockFilter{threshold: 50})
	validator := NewValidatorWrapper(&mockValidator{})
	transformer := NewTransformWrapper(&mockTransformer{})

	filter.SetNext(validator)
	validator.SetNext(transformer)

	testData := []map[string]interface{}{
		{"id": 1, "value": 30, "required_field": "a", "name": "alice"},
		{"id": 2, "value": 60, "required_field": "b", "name": "bob"},
		{"id": 3, "value": 70, "required_field": "c", "name": "charlie"},
	}

	ctx := context.Background()
	result, err := filter.Process(ctx, testData)

	if err != nil {
		t.Fatalf("Chained Process() failed: %v", err)
	}

	// Should have 2 records (value >= 50)
	expectedCount := 2
	if len(result) != expectedCount {
		t.Errorf("Process() returned %d records, expected %d", len(result), expectedCount)
	}

	// Verify transformation occurred
	for _, row := range result {
		name := row["name"].(string)
		if name != strings.ToUpper(name) {
			t.Errorf("Name %q was not transformed to uppercase", name)
		}
	}
}

// --- Benchmark Tests ---

func BenchmarkFilterWrapperProcess(b *testing.B) {
	strategy := &mockFilter{threshold: 50}
	wrapper := NewFilterWrapper(strategy)

	testData := []map[string]interface{}{
		{"id": 1, "value": 30},
		{"id": 2, "value": 60},
		{"id": 3, "value": 45},
		{"id": 4, "value": 70},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wrapper.Process(ctx, testData)
	}
}

func BenchmarkValidatorWrapperProcess(b *testing.B) {
	strategy := &mockValidator{}
	wrapper := NewValidatorWrapper(strategy)

	testData := []map[string]interface{}{
		{"id": 1, "required_field": "value1"},
		{"id": 2, "required_field": "value2"},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wrapper.Process(ctx, testData)
	}
}

func BenchmarkTransformWrapperProcess(b *testing.B) {
	strategy := &mockTransformer{}
	wrapper := NewTransformWrapper(strategy)

	testData := []map[string]interface{}{
		{"id": 1, "name": "alice"},
		{"id": 2, "name": "bob"},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wrapper.Process(ctx, testData)
	}
}

func BenchmarkWrapperChaining(b *testing.B) {
	filter := NewFilterWrapper(&mockFilter{threshold: 50})
	validator := NewValidatorWrapper(&mockValidator{})
	transformer := NewTransformWrapper(&mockTransformer{})

	filter.SetNext(validator)
	validator.SetNext(transformer)

	testData := []map[string]interface{}{
		{"id": 1, "value": 60, "required_field": "a", "name": "alice"},
		{"id": 2, "value": 70, "required_field": "b", "name": "bob"},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.Process(ctx, testData)
	}
}
