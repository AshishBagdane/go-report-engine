package engine

import (
	"strings"
	"testing"

	"github.com/AshishBagdane/report-engine/internal/formatter"
	"github.com/AshishBagdane/report-engine/internal/output"
	"github.com/AshishBagdane/report-engine/internal/processor"
	"github.com/AshishBagdane/report-engine/internal/provider"
)

// Mock implementations for builder tests
type builderMockProvider struct{}

func (m *builderMockProvider) Fetch() ([]map[string]interface{}, error) {
	return []map[string]interface{}{{"id": 1}}, nil
}

type builderMockProcessor struct {
	processor.BaseProcessor
}

func (m *builderMockProcessor) Process(data []map[string]interface{}) ([]map[string]interface{}, error) {
	return m.BaseProcessor.Process(data)
}

type builderMockFormatter struct{}

func (m *builderMockFormatter) Format(data []map[string]interface{}) ([]byte, error) {
	return []byte("formatted"), nil
}

type builderMockOutput struct{}

func (m *builderMockOutput) Send(data []byte) error {
	return nil
}

// Verify mock implementations at compile time
var (
	_ provider.ProviderStrategy  = (*builderMockProvider)(nil)
	_ processor.ProcessorHandler = (*builderMockProcessor)(nil)
	_ formatter.FormatStrategy   = (*builderMockFormatter)(nil)
	_ output.OutputStrategy      = (*builderMockOutput)(nil)
)

// TestNewEngineBuilder tests builder creation
func TestNewEngineBuilder(t *testing.T) {
	builder := NewEngineBuilder()

	if builder == nil {
		t.Fatal("NewEngineBuilder() returned nil")
	}

	// Should start with no components
	if builder.IsComplete() {
		t.Error("New builder should not be complete")
	}
}

// TestEngineBuilderWithProvider tests setting provider
func TestEngineBuilderWithProvider(t *testing.T) {
	builder := NewEngineBuilder()
	prov := &builderMockProvider{}

	result := builder.WithProvider(prov)

	// Should return builder for chaining
	if result != builder {
		t.Error("WithProvider should return builder for chaining")
	}

	// Provider should be set
	if builder.provider != prov {
		t.Error("Provider was not set")
	}
}

// TestEngineBuilderWithProcessor tests setting processor
func TestEngineBuilderWithProcessor(t *testing.T) {
	builder := NewEngineBuilder()
	proc := &builderMockProcessor{}

	result := builder.WithProcessor(proc)

	if result != builder {
		t.Error("WithProcessor should return builder for chaining")
	}

	if builder.processor != proc {
		t.Error("Processor was not set")
	}
}

// TestEngineBuilderWithFormatter tests setting formatter
func TestEngineBuilderWithFormatter(t *testing.T) {
	builder := NewEngineBuilder()
	fmt := &builderMockFormatter{}

	result := builder.WithFormatter(fmt)

	if result != builder {
		t.Error("WithFormatter should return builder for chaining")
	}

	if builder.formatter != fmt {
		t.Error("Formatter was not set")
	}
}

// TestEngineBuilderWithOutput tests setting output
func TestEngineBuilderWithOutput(t *testing.T) {
	builder := NewEngineBuilder()
	out := &builderMockOutput{}

	result := builder.WithOutput(out)

	if result != builder {
		t.Error("WithOutput should return builder for chaining")
	}

	if builder.output != out {
		t.Error("Output was not set")
	}
}

// TestEngineBuilderBuildSuccess tests successful build
func TestEngineBuilderBuildSuccess(t *testing.T) {
	builder := NewEngineBuilder().
		WithProvider(&builderMockProvider{}).
		WithProcessor(&builderMockProcessor{}).
		WithFormatter(&builderMockFormatter{}).
		WithOutput(&builderMockOutput{})

	engine, err := builder.Build()

	if err != nil {
		t.Fatalf("Build() should succeed, got error: %v", err)
	}

	if engine == nil {
		t.Fatal("Build() returned nil engine")
	}

	// Verify all components are set
	if engine.Provider == nil {
		t.Error("Engine provider is nil")
	}
	if engine.Processor == nil {
		t.Error("Engine processor is nil")
	}
	if engine.Formatter == nil {
		t.Error("Engine formatter is nil")
	}
	if engine.Output == nil {
		t.Error("Engine output is nil")
	}
}

// TestEngineBuilderBuildFailures tests build validation failures
func TestEngineBuilderBuildFailures(t *testing.T) {
	tests := []struct {
		name          string
		setupBuilder  func() *EngineBuilder
		expectedError string
	}{
		{
			name: "missing provider",
			setupBuilder: func() *EngineBuilder {
				return NewEngineBuilder().
					WithProcessor(&builderMockProcessor{}).
					WithFormatter(&builderMockFormatter{}).
					WithOutput(&builderMockOutput{})
			},
			expectedError: "provider is required",
		},
		{
			name: "missing processor",
			setupBuilder: func() *EngineBuilder {
				return NewEngineBuilder().
					WithProvider(&builderMockProvider{}).
					WithFormatter(&builderMockFormatter{}).
					WithOutput(&builderMockOutput{})
			},
			expectedError: "processor is required",
		},
		{
			name: "missing formatter",
			setupBuilder: func() *EngineBuilder {
				return NewEngineBuilder().
					WithProvider(&builderMockProvider{}).
					WithProcessor(&builderMockProcessor{}).
					WithOutput(&builderMockOutput{})
			},
			expectedError: "formatter is required",
		},
		{
			name: "missing output",
			setupBuilder: func() *EngineBuilder {
				return NewEngineBuilder().
					WithProvider(&builderMockProvider{}).
					WithProcessor(&builderMockProcessor{}).
					WithFormatter(&builderMockFormatter{})
			},
			expectedError: "output is required",
		},
		{
			name: "all components missing",
			setupBuilder: func() *EngineBuilder {
				return NewEngineBuilder()
			},
			expectedError: "builder validation failed",
		},
		{
			name: "multiple components missing",
			setupBuilder: func() *EngineBuilder {
				return NewEngineBuilder().
					WithProvider(&builderMockProvider{})
			},
			expectedError: "builder validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.setupBuilder()
			engine, err := builder.Build()

			if err == nil {
				t.Fatal("Build() should return error")
			}

			if engine != nil {
				t.Error("Build() should return nil engine on error")
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Error should contain %q, got: %v", tt.expectedError, err)
			}

			// Verify it's a BuilderValidationError
			if _, ok := err.(*BuilderValidationError); !ok {
				t.Error("Error should be BuilderValidationError type")
			}
		})
	}
}

// TestEngineBuilderValidate tests the Validate method
func TestEngineBuilderValidate(t *testing.T) {
	tests := []struct {
		name        string
		builder     *EngineBuilder
		shouldError bool
	}{
		{
			name: "complete builder",
			builder: NewEngineBuilder().
				WithProvider(&builderMockProvider{}).
				WithProcessor(&builderMockProcessor{}).
				WithFormatter(&builderMockFormatter{}).
				WithOutput(&builderMockOutput{}),
			shouldError: false,
		},
		{
			name:        "empty builder",
			builder:     NewEngineBuilder(),
			shouldError: true,
		},
		{
			name: "partial builder",
			builder: NewEngineBuilder().
				WithProvider(&builderMockProvider{}).
				WithFormatter(&builderMockFormatter{}),
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.builder.Validate()

			if (err != nil) != tt.shouldError {
				t.Errorf("Validate() error = %v, shouldError = %v", err, tt.shouldError)
			}
		})
	}
}

// TestEngineBuilderIsComplete tests the IsComplete method
func TestEngineBuilderIsComplete(t *testing.T) {
	tests := []struct {
		name     string
		builder  *EngineBuilder
		expected bool
	}{
		{
			name:     "empty builder",
			builder:  NewEngineBuilder(),
			expected: false,
		},
		{
			name: "partial builder",
			builder: NewEngineBuilder().
				WithProvider(&builderMockProvider{}).
				WithFormatter(&builderMockFormatter{}),
			expected: false,
		},
		{
			name: "complete builder",
			builder: NewEngineBuilder().
				WithProvider(&builderMockProvider{}).
				WithProcessor(&builderMockProcessor{}).
				WithFormatter(&builderMockFormatter{}).
				WithOutput(&builderMockOutput{}),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder.IsComplete()

			if result != tt.expected {
				t.Errorf("IsComplete() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestEngineBuilderReset tests the Reset method
func TestEngineBuilderReset(t *testing.T) {
	builder := NewEngineBuilder().
		WithProvider(&builderMockProvider{}).
		WithProcessor(&builderMockProcessor{}).
		WithFormatter(&builderMockFormatter{}).
		WithOutput(&builderMockOutput{})

	if !builder.IsComplete() {
		t.Fatal("Builder should be complete before reset")
	}

	result := builder.Reset()

	// Should return builder for chaining
	if result != builder {
		t.Error("Reset should return builder for chaining")
	}

	// Should not be complete after reset
	if builder.IsComplete() {
		t.Error("Builder should not be complete after reset")
	}

	// All components should be nil
	if builder.provider != nil {
		t.Error("Provider should be nil after reset")
	}
	if builder.processor != nil {
		t.Error("Processor should be nil after reset")
	}
	if builder.formatter != nil {
		t.Error("Formatter should be nil after reset")
	}
	if builder.output != nil {
		t.Error("Output should be nil after reset")
	}
}

// TestEngineBuilderChaining tests method chaining
func TestEngineBuilderChaining(t *testing.T) {
	builder := NewEngineBuilder().
		WithProvider(&builderMockProvider{}).
		WithProcessor(&builderMockProcessor{}).
		WithFormatter(&builderMockFormatter{}).
		WithOutput(&builderMockOutput{})

	if !builder.IsComplete() {
		t.Error("Chained builder should be complete")
	}

	engine, err := builder.Build()
	if err != nil {
		t.Fatalf("Chained build should succeed, got: %v", err)
	}

	if engine == nil {
		t.Error("Chained build should return engine")
	}
}

// TestEngineBuilderReuseAfterBuild tests reusing builder after build
func TestEngineBuilderReuseAfterBuild(t *testing.T) {
	builder := NewEngineBuilder().
		WithProvider(&builderMockProvider{}).
		WithProcessor(&builderMockProcessor{}).
		WithFormatter(&builderMockFormatter{}).
		WithOutput(&builderMockOutput{})

	engine1, err1 := builder.Build()
	if err1 != nil {
		t.Fatalf("First build failed: %v", err1)
	}

	// Build again with same builder
	engine2, err2 := builder.Build()
	if err2 != nil {
		t.Fatalf("Second build failed: %v", err2)
	}

	// Should create separate engine instances
	if engine1 == engine2 {
		t.Error("Builder should create new engine instances")
	}
}

// TestBuilderValidationErrorMessage tests error message formatting
func TestBuilderValidationErrorMessage(t *testing.T) {
	tests := []struct {
		name   string
		errors []string
		expect string
	}{
		{
			name:   "single error",
			errors: []string{"provider is required"},
			expect: "builder validation failed: provider is required",
		},
		{
			name:   "multiple errors",
			errors: []string{"provider is required", "formatter is required"},
			expect: "builder validation failed: 2 errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &BuilderValidationError{Errors: tt.errors}
			msg := err.Error()

			if !strings.Contains(msg, tt.expect) {
				t.Errorf("Error message should contain %q, got: %q", tt.expect, msg)
			}
		})
	}
}

// TestPredefinedBuilderErrors tests predefined error constants
func TestPredefinedBuilderErrors(t *testing.T) {
	if ErrBuilderIncomplete == nil {
		t.Error("ErrBuilderIncomplete should not be nil")
	}
	if ErrBuilderProviderNil == nil {
		t.Error("ErrBuilderProviderNil should not be nil")
	}
	if ErrBuilderProcessorNil == nil {
		t.Error("ErrBuilderProcessorNil should not be nil")
	}
	if ErrBuilderFormatterNil == nil {
		t.Error("ErrBuilderFormatterNil should not be nil")
	}
	if ErrBuilderOutputNil == nil {
		t.Error("ErrBuilderOutputNil should not be nil")
	}
}

// BenchmarkEngineBuilderBuild benchmarks engine building
func BenchmarkEngineBuilderBuild(b *testing.B) {
	prov := &builderMockProvider{}
	proc := &builderMockProcessor{}
	fmt := &builderMockFormatter{}
	out := &builderMockOutput{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := NewEngineBuilder().
			WithProvider(prov).
			WithProcessor(proc).
			WithFormatter(fmt).
			WithOutput(out)

		builder.Build()
	}
}

// BenchmarkEngineBuilderValidate benchmarks validation
func BenchmarkEngineBuilderValidate(b *testing.B) {
	builder := NewEngineBuilder().
		WithProvider(&builderMockProvider{}).
		WithProcessor(&builderMockProcessor{}).
		WithFormatter(&builderMockFormatter{}).
		WithOutput(&builderMockOutput{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.Validate()
	}
}
