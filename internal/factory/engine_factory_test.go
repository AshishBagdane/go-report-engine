package factory

import (
	"context"
	"strings"
	"testing"

	"github.com/AshishBagdane/go-report-engine/internal/engine"
	"github.com/AshishBagdane/go-report-engine/internal/formatter"
	"github.com/AshishBagdane/go-report-engine/internal/output"
	"github.com/AshishBagdane/go-report-engine/internal/provider"
	"github.com/AshishBagdane/go-report-engine/internal/registry"
)

// setupRegistries initializes all required registries for testing
func setupRegistries() {
	registry.ClearProviders()
	registry.ClearProcessors()
	registry.ClearFormatters()
	registry.ClearOutputs()

	// Register core components
	registry.RegisterProvider("mock", func() provider.ProviderStrategy {
		return provider.NewMockProvider([]map[string]interface{}{
			{"id": 1, "name": "Alice", "score": 95},
			{"id": 2, "name": "Bob", "score": 88},
		})
	})
	registry.RegisterFormatter("json", func() formatter.FormatStrategy {
		return formatter.NewJSONFormatter("  ")
	})
	registry.RegisterOutput("console", func() output.OutputStrategy {
		return output.NewConsoleOutput()
	})
}

// TestNewEngineFromConfigSuccess tests successful engine creation
func TestNewEngineFromConfigSuccess(t *testing.T) {
	setupRegistries()

	config := engine.Config{
		Provider:   engine.ProviderConfig{Type: "mock"},
		Processors: []engine.ProcessorConfig{},
		Formatter:  engine.FormatterConfig{Type: "json"},
		Output:     engine.OutputConfig{Type: "console"},
	}

	eng, err := NewEngineFromConfig(config)

	if err != nil {
		t.Fatalf("NewEngineFromConfig() returned error: %v", err)
	}

	if eng == nil {
		t.Fatal("NewEngineFromConfig() returned nil engine")
	}

	// Verify components are set
	if eng.Provider == nil {
		t.Error("Engine Provider is nil")
	}
	if eng.Processor == nil {
		t.Error("Engine Processor is nil")
	}
	if eng.Formatter == nil {
		t.Error("Engine Formatter is nil")
	}
	if eng.Output == nil {
		t.Error("Engine Output is nil")
	}
}

// TestNewEngineFromConfigInvalidConfig tests config validation
func TestNewEngineFromConfigInvalidConfig(t *testing.T) {
	setupRegistries()

	tests := []struct {
		name        string
		config      engine.Config
		shouldError bool
		errorMsg    string
	}{
		{
			name: "missing provider",
			config: engine.Config{
				Provider:   engine.ProviderConfig{Type: ""},
				Processors: []engine.ProcessorConfig{},
				Formatter:  engine.FormatterConfig{Type: "json"},
				Output:     engine.OutputConfig{Type: "console"},
			},
			shouldError: true,
			errorMsg:    "provider",
		},
		{
			name: "missing formatter",
			config: engine.Config{
				Provider:   engine.ProviderConfig{Type: "mock"},
				Processors: []engine.ProcessorConfig{},
				Formatter:  engine.FormatterConfig{Type: ""},
				Output:     engine.OutputConfig{Type: "console"},
			},
			shouldError: true,
			errorMsg:    "formatter",
		},
		{
			name: "missing output",
			config: engine.Config{
				Provider:   engine.ProviderConfig{Type: "mock"},
				Processors: []engine.ProcessorConfig{},
				Formatter:  engine.FormatterConfig{Type: "json"},
				Output:     engine.OutputConfig{Type: ""},
			},
			shouldError: true,
			errorMsg:    "output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eng, err := NewEngineFromConfig(tt.config)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for %s, got nil", tt.name)
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Error should contain %q, got: %v", tt.errorMsg, err)
				}
				if eng != nil {
					t.Error("Should return nil engine on error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if eng == nil {
					t.Error("Should return non-nil engine on success")
				}
			}
		})
	}
}

// TestNewEngineFromConfigUnknownProvider tests unknown provider type
func TestNewEngineFromConfigUnknownProvider(t *testing.T) {
	setupRegistries()

	config := engine.Config{
		Provider:   engine.ProviderConfig{Type: "nonexistent"},
		Processors: []engine.ProcessorConfig{},
		Formatter:  engine.FormatterConfig{Type: "json"},
		Output:     engine.OutputConfig{Type: "console"},
	}

	eng, err := NewEngineFromConfig(config)

	if err == nil {
		t.Fatal("Expected error for unknown provider")
	}

	if eng != nil {
		t.Error("Should return nil engine on error")
	}

	if !strings.Contains(err.Error(), "provider") {
		t.Errorf("Error should mention provider, got: %v", err)
	}
}

// TestNewEngineFromConfigUnknownFormatter tests unknown formatter type
func TestNewEngineFromConfigUnknownFormatter(t *testing.T) {
	setupRegistries()

	config := engine.Config{
		Provider:   engine.ProviderConfig{Type: "mock"},
		Processors: []engine.ProcessorConfig{},
		Formatter:  engine.FormatterConfig{Type: "nonexistent"},
		Output:     engine.OutputConfig{Type: "console"},
	}

	eng, err := NewEngineFromConfig(config)

	if err == nil {
		t.Fatal("Expected error for unknown formatter")
	}

	if eng != nil {
		t.Error("Should return nil engine on error")
	}

	if !strings.Contains(err.Error(), "formatter") {
		t.Errorf("Error should mention formatter, got: %v", err)
	}
}

// TestNewEngineFromConfigUnknownOutput tests unknown output type
func TestNewEngineFromConfigUnknownOutput(t *testing.T) {
	setupRegistries()

	config := engine.Config{
		Provider:   engine.ProviderConfig{Type: "mock"},
		Processors: []engine.ProcessorConfig{},
		Formatter:  engine.FormatterConfig{Type: "json"},
		Output:     engine.OutputConfig{Type: "nonexistent"},
	}

	eng, err := NewEngineFromConfig(config)

	if err == nil {
		t.Fatal("Expected error for unknown output")
	}

	if eng != nil {
		t.Error("Should return nil engine on error")
	}

	if !strings.Contains(err.Error(), "output") {
		t.Errorf("Error should mention output, got: %v", err)
	}
}

// testFilter is a simple mock filter for testing
type testFilter struct{}

func (m *testFilter) Keep(row map[string]interface{}) bool {
	return true
}

// TestNewEngineFromConfigWithProcessors tests with processor chain
func TestNewEngineFromConfigWithProcessors(t *testing.T) {
	setupRegistries()

	// Register a mock filter for testing
	registry.RegisterFilter("test_filter", &testFilter{})

	config := engine.Config{
		Provider: engine.ProviderConfig{Type: "mock"},
		Processors: []engine.ProcessorConfig{
			{Type: "test_filter", Params: map[string]string{}},
		},
		Formatter: engine.FormatterConfig{Type: "json"},
		Output:    engine.OutputConfig{Type: "console"},
	}

	eng, err := NewEngineFromConfig(config)

	if err != nil {
		t.Fatalf("NewEngineFromConfig() with processors returned error: %v", err)
	}

	if eng == nil {
		t.Fatal("NewEngineFromConfig() returned nil engine")
	}

	if eng.Processor == nil {
		t.Error("Processor chain is nil")
	}
}

// TestNewEngineFromConfigInvalidProcessor tests with invalid processor
func TestNewEngineFromConfigInvalidProcessor(t *testing.T) {
	setupRegistries()

	config := engine.Config{
		Provider: engine.ProviderConfig{Type: "mock"},
		Processors: []engine.ProcessorConfig{
			{Type: "nonexistent_processor", Params: map[string]string{}},
		},
		Formatter: engine.FormatterConfig{Type: "json"},
		Output:    engine.OutputConfig{Type: "console"},
	}

	eng, err := NewEngineFromConfig(config)

	if err == nil {
		t.Fatal("Expected error for invalid processor")
	}

	if eng != nil {
		t.Error("Should return nil engine on error")
	}
}

// TestNewEngineFromConfigEmptyProcessors tests with empty processor list
func TestNewEngineFromConfigEmptyProcessors(t *testing.T) {
	setupRegistries()

	config := engine.Config{
		Provider:   engine.ProviderConfig{Type: "mock"},
		Processors: []engine.ProcessorConfig{},
		Formatter:  engine.FormatterConfig{Type: "json"},
		Output:     engine.OutputConfig{Type: "console"},
	}

	eng, err := NewEngineFromConfig(config)

	if err != nil {
		t.Fatalf("NewEngineFromConfig() with empty processors returned error: %v", err)
	}

	if eng == nil {
		t.Fatal("NewEngineFromConfig() returned nil engine")
	}

	// Should have base processor
	if eng.Processor == nil {
		t.Error("Processor should not be nil even with empty processor list")
	}
}

// TestNewEngineFromConfigWithParams tests config with parameters
func TestNewEngineFromConfigWithParams(t *testing.T) {
	setupRegistries()

	config := engine.Config{
		Provider: engine.ProviderConfig{
			Type:   "mock",
			Params: map[string]string{"param1": "value1"},
		},
		Processors: []engine.ProcessorConfig{},
		Formatter: engine.FormatterConfig{
			Type:   "json",
			Params: map[string]string{"indent": "2"},
		},
		Output: engine.OutputConfig{
			Type:   "console",
			Params: map[string]string{"verbose": "true"},
		},
	}

	eng, err := NewEngineFromConfig(config)

	if err != nil {
		t.Fatalf("NewEngineFromConfig() with params returned error: %v", err)
	}

	if eng == nil {
		t.Fatal("NewEngineFromConfig() returned nil engine")
	}
}

// TestNewEngineFromConfigMultipleCalls tests creating multiple engines
func TestNewEngineFromConfigMultipleCalls(t *testing.T) {
	setupRegistries()

	config := engine.Config{
		Provider:   engine.ProviderConfig{Type: "mock"},
		Processors: []engine.ProcessorConfig{},
		Formatter:  engine.FormatterConfig{Type: "json"},
		Output:     engine.OutputConfig{Type: "console"},
	}

	eng1, err1 := NewEngineFromConfig(config)
	if err1 != nil {
		t.Fatalf("First NewEngineFromConfig() failed: %v", err1)
	}

	eng2, err2 := NewEngineFromConfig(config)
	if err2 != nil {
		t.Fatalf("Second NewEngineFromConfig() failed: %v", err2)
	}

	// Should create separate instances
	if eng1 == eng2 {
		t.Error("NewEngineFromConfig() should create new instances")
	}
}

// TestNewEngineFromConfigConcurrent tests concurrent engine creation
func TestNewEngineFromConfigConcurrent(t *testing.T) {
	setupRegistries()

	config := engine.Config{
		Provider:   engine.ProviderConfig{Type: "mock"},
		Processors: []engine.ProcessorConfig{},
		Formatter:  engine.FormatterConfig{Type: "json"},
		Output:     engine.OutputConfig{Type: "console"},
	}

	const goroutines = 10
	errors := make(chan error, goroutines)
	engines := make(chan *engine.ReportEngine, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			eng, err := NewEngineFromConfig(config)
			if err != nil {
				errors <- err
				return
			}
			engines <- eng
		}()
	}

	for i := 0; i < goroutines; i++ {
		select {
		case err := <-errors:
			t.Errorf("Concurrent NewEngineFromConfig() failed: %v", err)
		case eng := <-engines:
			if eng == nil {
				t.Error("Concurrent NewEngineFromConfig() returned nil")
			}
		}
	}
}

// TestNewEngineFromConfigRunnable tests that created engine is runnable
func TestNewEngineFromConfigRunnable(t *testing.T) {
	setupRegistries()

	config := engine.Config{
		Provider:   engine.ProviderConfig{Type: "mock"},
		Processors: []engine.ProcessorConfig{},
		Formatter:  engine.FormatterConfig{Type: "json"},
		Output:     engine.OutputConfig{Type: "console"},
	}

	eng, err := NewEngineFromConfig(config)
	if err != nil {
		t.Fatalf("NewEngineFromConfig() failed: %v", err)
	}

	// Should be able to run the engine
	// Note: This will output to console, but that's okay for tests
	ctx := context.Background()
	err = eng.RunWithContext(ctx)
	if err != nil {
		t.Errorf("Created engine Run() failed: %v", err)
	}
}

// BenchmarkNewEngineFromConfig benchmarks engine creation
func BenchmarkNewEngineFromConfig(b *testing.B) {
	setupRegistries()

	config := engine.Config{
		Provider:   engine.ProviderConfig{Type: "mock"},
		Processors: []engine.ProcessorConfig{},
		Formatter:  engine.FormatterConfig{Type: "json"},
		Output:     engine.OutputConfig{Type: "console"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewEngineFromConfig(config)
	}
}

// benchFilter is a simple filter for benchmarking
type benchFilter struct{}

func (m *benchFilter) Keep(row map[string]interface{}) bool {
	return true
}

// BenchmarkNewEngineFromConfigWithProcessors benchmarks with processor chain
func BenchmarkNewEngineFromConfigWithProcessors(b *testing.B) {
	setupRegistries()

	registry.RegisterFilter("bench_filter", &benchFilter{})

	config := engine.Config{
		Provider: engine.ProviderConfig{Type: "mock"},
		Processors: []engine.ProcessorConfig{
			{Type: "bench_filter", Params: map[string]string{}},
		},
		Formatter: engine.FormatterConfig{Type: "json"},
		Output:    engine.OutputConfig{Type: "console"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewEngineFromConfig(config)
	}
}
