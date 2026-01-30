package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AshishBagdane/go-report-engine/internal/engine"
	"github.com/AshishBagdane/go-report-engine/internal/formatter"
	"github.com/AshishBagdane/go-report-engine/internal/output"
	"github.com/AshishBagdane/go-report-engine/internal/provider"
	"github.com/AshishBagdane/go-report-engine/internal/registry"
)

// setupTestRegistries registers mock components for testing
func setupTestRegistries() {
	registry.ClearProviders()
	registry.ClearProcessors()
	registry.ClearFormatters()
	registry.ClearOutputs()

	registry.RegisterProvider("mock", func() provider.ProviderStrategy {
		return provider.NewMockProvider([]map[string]interface{}{
			{"id": 1, "name": "test"},
		})
	})

	// Register a mock validator for testing
	registry.RegisterValidator("validator", &mockValidator{})

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

// mockValidator is a simple mock validator for testing
type mockValidator struct{}

func (m *mockValidator) Validate(row map[string]interface{}) error {
	return nil
}

// TestLoadAndBuild tests loading and building in one step
func TestLoadAndBuild(t *testing.T) {
	setupTestRegistries()

	yamlContent := `
provider:
  type: mock
processors: []
formatter:
  type: json
output:
  type: console
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	eng, err := LoadAndBuild(configPath)

	if err != nil {
		t.Fatalf("LoadAndBuild() returned error: %v", err)
	}

	if eng == nil {
		t.Fatal("LoadAndBuild() returned nil engine")
	}

	// Verify engine components
	if eng.Provider == nil {
		t.Error("Engine provider is nil")
	}
	if eng.Formatter == nil {
		t.Error("Engine formatter is nil")
	}
	if eng.Output == nil {
		t.Error("Engine output is nil")
	}
}

// TestLoadAndBuildFileNotFound tests error handling for missing file
func TestLoadAndBuildFileNotFound(t *testing.T) {
	setupTestRegistries()

	_, err := LoadAndBuild("/nonexistent/config.yaml")

	if err == nil {
		t.Fatal("LoadAndBuild() should fail for missing file")
	}

	if !strings.Contains(err.Error(), "failed to load config") {
		t.Errorf("Error should mention config loading, got: %v", err)
	}
}

// TestLoadAndBuildInvalidConfig tests error handling for invalid config
func TestLoadAndBuildInvalidConfig(t *testing.T) {
	setupTestRegistries()

	invalidYAML := `
provider:
  type: ""
formatter:
  type: json
output:
  type: console
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := LoadAndBuild(configPath)

	if err == nil {
		t.Fatal("LoadAndBuild() should fail for invalid config")
	}
}

// TestLoadAndBuildWithEnv tests loading with environment overrides
func TestLoadAndBuildWithEnv(t *testing.T) {
	setupTestRegistries()

	os.Setenv("ENGINE_PROVIDER_TYPE", "mock")
	defer os.Unsetenv("ENGINE_PROVIDER_TYPE")

	yamlContent := `
provider:
  type: mock
processors: []
formatter:
  type: json
output:
  type: console
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	eng, err := LoadAndBuildWithEnv(configPath)

	if err != nil {
		t.Fatalf("LoadAndBuildWithEnv() returned error: %v", err)
	}

	if eng == nil {
		t.Fatal("LoadAndBuildWithEnv() returned nil engine")
	}
}

// TestBuildFromBytes tests building from raw bytes
func TestBuildFromBytes(t *testing.T) {
	setupTestRegistries()

	yamlContent := []byte(`
provider:
  type: mock
processors: []
formatter:
  type: json
output:
  type: console
`)

	eng, err := BuildFromBytes(yamlContent, "yaml")

	if err != nil {
		t.Fatalf("BuildFromBytes() returned error: %v", err)
	}

	if eng == nil {
		t.Fatal("BuildFromBytes() returned nil engine")
	}
}

// TestBuildFromBytesJSON tests building from JSON bytes
func TestBuildFromBytesJSON(t *testing.T) {
	setupTestRegistries()

	jsonContent := []byte(`{
  "provider": {"type": "mock"},
  "processors": [],
  "formatter": {"type": "json"},
  "output": {"type": "console"}
}`)

	eng, err := BuildFromBytes(jsonContent, "json")

	if err != nil {
		t.Fatalf("BuildFromBytes() returned error: %v", err)
	}

	if eng == nil {
		t.Fatal("BuildFromBytes() returned nil engine")
	}
}

// TestBuildFromDefault tests building from default config
func TestBuildFromDefault(t *testing.T) {
	setupTestRegistries()

	eng, err := BuildFromDefault()

	if err != nil {
		t.Fatalf("BuildFromDefault() returned error: %v", err)
	}

	if eng == nil {
		t.Fatal("BuildFromDefault() returned nil engine")
	}
}

// TestBuildFromProduction tests building from production config
func TestBuildFromProduction(t *testing.T) {
	setupTestRegistries()

	eng, err := BuildFromProduction()

	if err != nil {
		t.Fatalf("BuildFromProduction() returned error: %v", err)
	}

	if eng == nil {
		t.Fatal("BuildFromProduction() returned nil engine")
	}
}

// TestBuildFromDevelopment tests building from development config
func TestBuildFromDevelopment(t *testing.T) {
	setupTestRegistries()

	eng, err := BuildFromDevelopment()

	if err != nil {
		t.Fatalf("BuildFromDevelopment() returned error: %v", err)
	}

	if eng == nil {
		t.Fatal("BuildFromDevelopment() returned nil engine")
	}
}

// TestBuildFromTesting tests building from testing config
func TestBuildFromTesting(t *testing.T) {
	setupTestRegistries()

	eng, err := BuildFromTesting()

	if err != nil {
		t.Fatalf("BuildFromTesting() returned error: %v", err)
	}

	if eng == nil {
		t.Fatal("BuildFromTesting() returned nil engine")
	}
}

// TestMustLoadAndBuild tests Must variant (success case)
func TestMustLoadAndBuild(t *testing.T) {
	setupTestRegistries()

	yamlContent := `
provider:
  type: mock
processors: []
formatter:
  type: json
output:
  type: console
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	eng := MustLoadAndBuild(configPath)

	if eng == nil {
		t.Fatal("MustLoadAndBuild() returned nil engine")
	}
}

// TestMustLoadAndBuildPanic tests Must variant panics on error
func TestMustLoadAndBuildPanic(t *testing.T) {
	setupTestRegistries()

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustLoadAndBuild() should panic on error")
		}
	}()

	MustLoadAndBuild("/nonexistent/config.yaml")
}

// TestMustLoadAndBuildWithEnv tests Must variant with env (success case)
func TestMustLoadAndBuildWithEnv(t *testing.T) {
	setupTestRegistries()

	yamlContent := `
provider:
  type: mock
processors: []
formatter:
  type: json
output:
  type: console
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	eng := MustLoadAndBuildWithEnv(configPath)

	if eng == nil {
		t.Fatal("MustLoadAndBuildWithEnv() returned nil engine")
	}
}

// TestMustLoadAndBuildWithEnvPanic tests Must variant with env panics on error
func TestMustLoadAndBuildWithEnvPanic(t *testing.T) {
	setupTestRegistries()

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustLoadAndBuildWithEnv() should panic on error")
		}
	}()

	MustLoadAndBuildWithEnv("/nonexistent/config.yaml")
}

// TestMustBuildFromDefault tests Must variant for default config
func TestMustBuildFromDefault(t *testing.T) {
	setupTestRegistries()

	eng := MustBuildFromDefault()

	if eng == nil {
		t.Fatal("MustBuildFromDefault() returned nil engine")
	}
}

// TestMustBuildFromProduction tests Must variant for production config
func TestMustBuildFromProduction(t *testing.T) {
	setupTestRegistries()

	eng := MustBuildFromProduction()

	if eng == nil {
		t.Fatal("MustBuildFromProduction() returned nil engine")
	}
}

// TestMustBuildFromDevelopment tests Must variant for development config
func TestMustBuildFromDevelopment(t *testing.T) {
	setupTestRegistries()

	eng := MustBuildFromDevelopment()

	if eng == nil {
		t.Fatal("MustBuildFromDevelopment() returned nil engine")
	}
}

// TestMustBuildFromTesting tests Must variant for testing config
func TestMustBuildFromTesting(t *testing.T) {
	setupTestRegistries()

	eng := MustBuildFromTesting()

	if eng == nil {
		t.Fatal("MustBuildFromTesting() returned nil engine")
	}
}

// TestValidateAndBuild tests explicit validation before build
func TestValidateAndBuild(t *testing.T) {
	setupTestRegistries()

	cfg := DefaultConfig()
	eng, err := ValidateAndBuild(cfg)

	if err != nil {
		t.Fatalf("ValidateAndBuild() returned error: %v", err)
	}

	if eng == nil {
		t.Fatal("ValidateAndBuild() returned nil engine")
	}
}

// TestValidateAndBuildInvalidConfig tests validation failure
func TestValidateAndBuildInvalidConfig(t *testing.T) {
	setupTestRegistries()

	cfg := engine.Config{
		Provider:   engine.ProviderConfig{Type: ""},
		Processors: []engine.ProcessorConfig{},
		Formatter:  engine.FormatterConfig{Type: "json"},
		Output:     engine.OutputConfig{Type: "console"},
	}

	_, err := ValidateAndBuild(cfg)

	if err == nil {
		t.Fatal("ValidateAndBuild() should fail for invalid config")
	}

	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("Error should mention validation, got: %v", err)
	}
}

// TestLoadOrDefault tests fallback to default
func TestLoadOrDefault(t *testing.T) {
	setupTestRegistries()

	// Try to load nonexistent file
	cfg, err := LoadOrDefault("/nonexistent/config.yaml")

	if err == nil {
		t.Error("LoadOrDefault() should return error when using default")
	}

	if cfg == nil {
		t.Fatal("LoadOrDefault() should return default config on error")
	}

	// Should be valid default config
	if validateErr := cfg.Validate(); validateErr != nil {
		t.Errorf("Default config should be valid, got: %v", validateErr)
	}
}

// TestLoadOrDefaultSuccess tests successful load
func TestLoadOrDefaultSuccess(t *testing.T) {
	setupTestRegistries()

	yamlContent := `
provider:
  type: mock
processors: []
formatter:
  type: json
output:
  type: console
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg, err := LoadOrDefault(configPath)

	if err != nil {
		t.Errorf("LoadOrDefault() should not return error on success, got: %v", err)
	}

	if cfg == nil {
		t.Fatal("LoadOrDefault() returned nil config")
	}
}

// TestLoadOrDefaultWithEnv tests fallback to default with env
func TestLoadOrDefaultWithEnv(t *testing.T) {
	setupTestRegistries()

	cfg, err := LoadOrDefaultWithEnv("/nonexistent/config.yaml")

	if err == nil {
		t.Error("LoadOrDefaultWithEnv() should return error when using default")
	}

	if cfg == nil {
		t.Fatal("LoadOrDefaultWithEnv() should return default config on error")
	}
}

// TestBuildFromConfigOrFile tests with provided config
func TestBuildFromConfigOrFile(t *testing.T) {
	setupTestRegistries()

	cfg := DefaultConfig()
	eng, err := BuildFromConfigOrFile(&cfg, "")

	if err != nil {
		t.Fatalf("BuildFromConfigOrFile() with config returned error: %v", err)
	}

	if eng == nil {
		t.Fatal("BuildFromConfigOrFile() returned nil engine")
	}
}

// TestBuildFromConfigOrFileFromFile tests fallback to file
func TestBuildFromConfigOrFileFromFile(t *testing.T) {
	setupTestRegistries()

	yamlContent := `
provider:
  type: mock
processors: []
formatter:
  type: json
output:
  type: console
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	eng, err := BuildFromConfigOrFile(nil, configPath)

	if err != nil {
		t.Fatalf("BuildFromConfigOrFile() from file returned error: %v", err)
	}

	if eng == nil {
		t.Fatal("BuildFromConfigOrFile() returned nil engine")
	}
}

// TestBuildFromConfigOrFileFileNotFound tests error when both fail
func TestBuildFromConfigOrFileFileNotFound(t *testing.T) {
	setupTestRegistries()

	_, err := BuildFromConfigOrFile(nil, "/nonexistent/config.yaml")

	if err == nil {
		t.Fatal("BuildFromConfigOrFile() should fail when file not found")
	}
}

// TestAllIntegrationFunctions tests that all integration functions work
func TestAllIntegrationFunctions(t *testing.T) {
	setupTestRegistries()

	// Create a valid config file
	yamlContent := `
provider:
  type: mock
processors: []
formatter:
  type: json
output:
  type: console
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name string
		fn   func() (*engine.ReportEngine, error)
	}{
		{
			name: "LoadAndBuild",
			fn:   func() (*engine.ReportEngine, error) { return LoadAndBuild(configPath) },
		},
		{
			name: "LoadAndBuildWithEnv",
			fn:   func() (*engine.ReportEngine, error) { return LoadAndBuildWithEnv(configPath) },
		},
		{
			name: "BuildFromDefault",
			fn:   BuildFromDefault,
		},
		{
			name: "BuildFromProduction",
			fn:   BuildFromProduction,
		},
		{
			name: "BuildFromDevelopment",
			fn:   BuildFromDevelopment,
		},
		{
			name: "BuildFromTesting",
			fn:   BuildFromTesting,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eng, err := tt.fn()
			if err != nil {
				t.Errorf("%s returned error: %v", tt.name, err)
			}
			if eng == nil {
				t.Errorf("%s returned nil engine", tt.name)
			}
		})
	}
}

// BenchmarkLoadAndBuild benchmarks loading and building
func BenchmarkLoadAndBuild(b *testing.B) {
	setupTestRegistries()

	yamlContent := `
provider:
  type: mock
processors: []
formatter:
  type: json
output:
  type: console
`
	tmpDir := b.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LoadAndBuild(configPath)
	}
}

// BenchmarkBuildFromDefault benchmarks default config build
func BenchmarkBuildFromDefault(b *testing.B) {
	setupTestRegistries()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildFromDefault()
	}
}

// BenchmarkBuildFromBytes benchmarks building from bytes
func BenchmarkBuildFromBytes(b *testing.B) {
	setupTestRegistries()

	yamlContent := []byte(`
provider:
  type: mock
processors: []
formatter:
  type: json
output:
  type: console
`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildFromBytes(yamlContent, "yaml")
	}
}
