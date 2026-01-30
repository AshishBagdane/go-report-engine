package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewLoader tests loader creation
func TestNewLoader(t *testing.T) {
	loader := NewLoader()

	if loader == nil {
		t.Fatal("NewLoader() returned nil")
	}

	if loader.applyEnvOverrides {
		t.Error("New loader should have env overrides disabled by default")
	}
}

// TestLoaderWithEnvOverrides tests enabling env overrides
func TestLoaderWithEnvOverrides(t *testing.T) {
	loader := NewLoader().WithEnvOverrides()

	// Should return loader for chaining
	if loader == nil {
		t.Error("WithEnvOverrides() should return loader")
		return
	}

	if !loader.applyEnvOverrides {
		t.Error("WithEnvOverrides() should enable env overrides")
	}
}

// TestLoadFromFileYAML tests loading YAML config
func TestLoadFromFileYAML(t *testing.T) {
	// Create temporary YAML file
	yamlContent := `
provider:
  type: mock
  params:
    data: "test"
processors: []
formatter:
  type: json
  params:
    indent: "2"
output:
  type: console
  params: {}
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewLoader()
	config, err := loader.LoadFromFile(configPath)

	if err != nil {
		t.Fatalf("LoadFromFile() returned error: %v", err)
	}

	if config == nil {
		t.Fatal("LoadFromFile() returned nil config")
	}

	// Verify loaded values
	if config.Provider.Type != "mock" {
		t.Errorf("Provider type = %s, expected mock", config.Provider.Type)
	}

	if config.Formatter.Type != "json" {
		t.Errorf("Formatter type = %s, expected json", config.Formatter.Type)
	}

	if config.Output.Type != "console" {
		t.Errorf("Output type = %s, expected console", config.Output.Type)
	}

	// Verify params
	if config.Provider.Params["data"] != "test" {
		t.Error("Provider params not loaded correctly")
	}

	if config.Formatter.Params["indent"] != "2" {
		t.Error("Formatter params not loaded correctly")
	}
}

// TestLoadFromFileJSON tests loading JSON config
func TestLoadFromFileJSON(t *testing.T) {
	jsonContent := `{
  "provider": {
    "type": "mock",
    "params": {
      "data": "test"
    }
  },
  "processors": [],
  "formatter": {
    "type": "json",
    "params": {
      "indent": "2"
    }
  },
  "output": {
    "type": "console",
    "params": {}
  }
}`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(configPath, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewLoader()
	config, err := loader.LoadFromFile(configPath)

	if err != nil {
		t.Fatalf("LoadFromFile() returned error: %v", err)
	}

	if config == nil {
		t.Fatal("LoadFromFile() returned nil config")
	}

	// Verify loaded values
	if config.Provider.Type != "mock" {
		t.Errorf("Provider type = %s, expected mock", config.Provider.Type)
	}
}

// TestLoadFromFileYMLExtension tests .yml extension
func TestLoadFromFileYMLExtension(t *testing.T) {
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
	configPath := filepath.Join(tmpDir, "config.yml")

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewLoader()
	config, err := loader.LoadFromFile(configPath)

	if err != nil {
		t.Fatalf("LoadFromFile() returned error: %v", err)
	}

	if config == nil {
		t.Fatal("LoadFromFile() returned nil config")
	}
}

// TestLoadFromFileNotFound tests missing file error
func TestLoadFromFileNotFound(t *testing.T) {
	loader := NewLoader()
	_, err := loader.LoadFromFile("/nonexistent/config.yaml")

	if err == nil {
		t.Fatal("LoadFromFile() should fail for missing file")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error should mention file not found, got: %v", err)
	}
}

// TestLoadFromFileUnsupportedFormat tests unsupported file format
func TestLoadFromFileUnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.txt")

	if err := os.WriteFile(configPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewLoader()
	_, err := loader.LoadFromFile(configPath)

	if err == nil {
		t.Fatal("LoadFromFile() should fail for unsupported format")
	}

	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("Error should mention unsupported format, got: %v", err)
	}
}

// TestLoadFromFileInvalidYAML tests invalid YAML content
func TestLoadFromFileInvalidYAML(t *testing.T) {
	invalidYAML := `
provider: this is not valid yaml
  type: [missing indent
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewLoader()
	_, err := loader.LoadFromFile(configPath)

	if err == nil {
		t.Fatal("LoadFromFile() should fail for invalid YAML")
	}

	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("Error should mention parse failure, got: %v", err)
	}
}

// TestLoadFromFileInvalidJSON tests invalid JSON content
func TestLoadFromFileInvalidJSON(t *testing.T) {
	invalidJSON := `{
  "provider": {
    "type": "mock"
  }
  missing comma
}`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(configPath, []byte(invalidJSON), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewLoader()
	_, err := loader.LoadFromFile(configPath)

	if err == nil {
		t.Fatal("LoadFromFile() should fail for invalid JSON")
	}

	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("Error should mention parse failure, got: %v", err)
	}
}

// TestLoadFromFileValidationFailure tests config validation
func TestLoadFromFileValidationFailure(t *testing.T) {
	invalidConfig := `
provider:
  type: ""
formatter:
  type: json
output:
  type: console
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewLoader()
	_, err := loader.LoadFromFile(configPath)

	if err == nil {
		t.Fatal("LoadFromFile() should fail for invalid config")
	}

	if !strings.Contains(err.Error(), "invalid configuration") {
		t.Errorf("Error should mention invalid configuration, got: %v", err)
	}
}

// TestLoadFromBytesYAML tests loading from YAML bytes
func TestLoadFromBytesYAML(t *testing.T) {
	yamlContent := []byte(`
provider:
  type: mock
processors: []
formatter:
  type: json
output:
  type: console
`)

	loader := NewLoader()
	config, err := loader.LoadFromBytes(yamlContent, "yaml")

	if err != nil {
		t.Fatalf("LoadFromBytes() returned error: %v", err)
	}

	if config.Provider.Type != "mock" {
		t.Errorf("Provider type = %s, expected mock", config.Provider.Type)
	}
}

// TestLoadFromBytesJSON tests loading from JSON bytes
func TestLoadFromBytesJSON(t *testing.T) {
	jsonContent := []byte(`{
  "provider": {"type": "mock"},
  "processors": [],
  "formatter": {"type": "json"},
  "output": {"type": "console"}
}`)

	loader := NewLoader()
	config, err := loader.LoadFromBytes(jsonContent, "json")

	if err != nil {
		t.Fatalf("LoadFromBytes() returned error: %v", err)
	}

	if config.Provider.Type != "mock" {
		t.Errorf("Provider type = %s, expected mock", config.Provider.Type)
	}
}

// TestLoadFromBytesUnsupportedFormat tests unsupported format
func TestLoadFromBytesUnsupportedFormat(t *testing.T) {
	loader := NewLoader()
	_, err := loader.LoadFromBytes([]byte("test"), "xml")

	if err == nil {
		t.Fatal("LoadFromBytes() should fail for unsupported format")
	}

	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Error should mention unsupported format, got: %v", err)
	}
}

// TestEnvironmentOverrides tests environment variable overrides
func TestEnvironmentOverrides(t *testing.T) {
	// Set environment variables
	_ = os.Setenv("ENGINE_PROVIDER_TYPE", "postgres")
	_ = os.Setenv("ENGINE_FORMATTER_TYPE", "csv")
	_ = os.Setenv("ENGINE_OUTPUT_TYPE", "file")
	defer func() {
		_ = os.Unsetenv("ENGINE_PROVIDER_TYPE")
		_ = os.Unsetenv("ENGINE_FORMATTER_TYPE")
		_ = os.Unsetenv("ENGINE_OUTPUT_TYPE")
	}()

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

	loader := NewLoader().WithEnvOverrides()
	config, err := loader.LoadFromFile(configPath)

	if err != nil {
		t.Fatalf("LoadFromFile() returned error: %v", err)
	}

	// Environment variables should override file values
	if config.Provider.Type != "postgres" {
		t.Errorf("Provider type = %s, expected postgres (from env)", config.Provider.Type)
	}

	if config.Formatter.Type != "csv" {
		t.Errorf("Formatter type = %s, expected csv (from env)", config.Formatter.Type)
	}

	if config.Output.Type != "file" {
		t.Errorf("Output type = %s, expected file (from env)", config.Output.Type)
	}
}

// TestEnvironmentParamOverrides tests param overrides
func TestEnvironmentParamOverrides(t *testing.T) {
	// Set parameter environment variables
	_ = os.Setenv("ENGINE_PROVIDER_PARAM_HOST", "localhost")
	_ = os.Setenv("ENGINE_PROVIDER_PARAM_PORT", "5432")
	_ = os.Setenv("ENGINE_FORMATTER_PARAM_INDENT", "4")
	_ = os.Setenv("ENGINE_OUTPUT_PARAM_PATH", "/tmp/output.json")
	defer func() {
		_ = os.Unsetenv("ENGINE_PROVIDER_PARAM_HOST")
		_ = os.Unsetenv("ENGINE_PROVIDER_PARAM_PORT")
		_ = os.Unsetenv("ENGINE_FORMATTER_PARAM_INDENT")
		_ = os.Unsetenv("ENGINE_OUTPUT_PARAM_PATH")
	}()

	yamlContent := `
provider:
  type: mock
  params:
    existing: "value"
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

	loader := NewLoader().WithEnvOverrides()
	config, err := loader.LoadFromFile(configPath)

	if err != nil {
		t.Fatalf("LoadFromFile() returned error: %v", err)
	}

	// Verify env params were added
	if config.Provider.Params["host"] != "localhost" {
		t.Error("Provider param 'host' not set from env")
	}

	if config.Provider.Params["port"] != "5432" {
		t.Error("Provider param 'port' not set from env")
	}

	// Verify existing param preserved
	if config.Provider.Params["existing"] != "value" {
		t.Error("Existing provider param should be preserved")
	}

	if config.Formatter.Params["indent"] != "4" {
		t.Error("Formatter param 'indent' not set from env")
	}

	if config.Output.Params["path"] != "/tmp/output.json" {
		t.Error("Output param 'path' not set from env")
	}
}

// TestEnvironmentOverridesWithoutFlag tests that env vars are ignored without flag
func TestEnvironmentOverridesWithoutFlag(t *testing.T) {
	_ = os.Setenv("ENGINE_PROVIDER_TYPE", "postgres")
	defer func() { _ = os.Unsetenv("ENGINE_PROVIDER_TYPE") }()

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

	// Load without env overrides enabled
	loader := NewLoader()
	config, err := loader.LoadFromFile(configPath)

	if err != nil {
		t.Fatalf("LoadFromFile() returned error: %v", err)
	}

	// Should use file value, not env
	if config.Provider.Type != "mock" {
		t.Errorf("Provider type = %s, expected mock (env should be ignored)", config.Provider.Type)
	}
}

// TestLoadFromFileConvenience tests convenience function
func TestLoadFromFileConvenience(t *testing.T) {
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

	config, err := LoadFromFile(configPath)

	if err != nil {
		t.Fatalf("LoadFromFile() returned error: %v", err)
	}

	if config == nil {
		t.Fatal("LoadFromFile() returned nil config")
	}
}

// TestLoadFromFileWithEnvConvenience tests convenience function with env
func TestLoadFromFileWithEnvConvenience(t *testing.T) {
	_ = os.Setenv("ENGINE_PROVIDER_TYPE", "postgres")
	defer func() { _ = os.Unsetenv("ENGINE_PROVIDER_TYPE") }()

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

	config, err := LoadFromFileWithEnv(configPath)

	if err != nil {
		t.Fatalf("LoadFromFileWithEnv() returned error: %v", err)
	}

	// Should apply env override
	if config.Provider.Type != "postgres" {
		t.Errorf("Provider type = %s, expected postgres", config.Provider.Type)
	}
}

// TestLoadFromBytesConvenience tests convenience function
func TestLoadFromBytesConvenience(t *testing.T) {
	yamlContent := []byte(`
provider:
  type: mock
processors: []
formatter:
  type: json
output:
  type: console
`)

	config, err := LoadFromBytes(yamlContent, "yaml")

	if err != nil {
		t.Fatalf("LoadFromBytes() returned error: %v", err)
	}

	if config == nil {
		t.Fatal("LoadFromBytes() returned nil config")
	}
}

// TestLoadFromFileWithProcessors tests loading config with processors
func TestLoadFromFileWithProcessors(t *testing.T) {
	yamlContent := `
provider:
  type: mock
processors:
  - type: filter
    params:
      min_score: "90"
  - type: validator
    params:
      required_fields: "id,name"
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

	loader := NewLoader()
	config, err := loader.LoadFromFile(configPath)

	if err != nil {
		t.Fatalf("LoadFromFile() returned error: %v", err)
	}

	if len(config.Processors) != 2 {
		t.Fatalf("Expected 2 processors, got %d", len(config.Processors))
	}

	if config.Processors[0].Type != "filter" {
		t.Errorf("First processor type = %s, expected filter", config.Processors[0].Type)
	}

	if config.Processors[0].Params["min_score"] != "90" {
		t.Error("First processor params not loaded correctly")
	}

	if config.Processors[1].Type != "validator" {
		t.Errorf("Second processor type = %s, expected validator", config.Processors[1].Type)
	}
}

// TestApplyParamOverridesNilMap tests param overrides with nil map
func TestApplyParamOverridesNilMap(t *testing.T) {
	_ = os.Setenv("ENGINE_TEST_PARAM_KEY", "value")
	defer func() { _ = os.Unsetenv("ENGINE_TEST_PARAM_KEY") }()

	var params map[string]string
	applyParamOverrides(&params, "ENGINE_TEST_PARAM_")

	if params == nil {
		t.Fatal("applyParamOverrides should initialize nil map")
	}

	if params["key"] != "value" {
		t.Error("Param override not applied to nil map")
	}
}

// TestLoadFromFileCompleteConfig tests loading a complete config with all fields
func TestLoadFromFileCompleteConfig(t *testing.T) {
	yamlContent := `
provider:
  type: postgres
  params:
    host: localhost
    port: "5432"
    database: mydb
processors:
  - type: filter
    params:
      condition: "score > 80"
  - type: transformer
    params:
      format: "uppercase"
formatter:
  type: json
  params:
    indent: "2"
    pretty: "true"
output:
  type: file
  params:
    path: /tmp/report.json
    mode: "0644"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewLoader()
	config, err := loader.LoadFromFile(configPath)

	if err != nil {
		t.Fatalf("LoadFromFile() returned error: %v", err)
	}

	// Verify all components loaded
	if config.Provider.Type != "postgres" {
		t.Error("Provider type not loaded correctly")
	}
	if len(config.Provider.Params) != 3 {
		t.Error("Provider params count incorrect")
	}

	if len(config.Processors) != 2 {
		t.Error("Processors count incorrect")
	}

	if config.Formatter.Type != "json" {
		t.Error("Formatter type not loaded correctly")
	}
	if len(config.Formatter.Params) != 2 {
		t.Error("Formatter params count incorrect")
	}

	if config.Output.Type != "file" {
		t.Error("Output type not loaded correctly")
	}
	if len(config.Output.Params) != 2 {
		t.Error("Output params count incorrect")
	}
}

// TestLoadFromFileEmptyParams tests loading config with empty params
func TestLoadFromFileEmptyParams(t *testing.T) {
	yamlContent := `
provider:
  type: mock
  params: {}
processors: []
formatter:
  type: json
  params: {}
output:
  type: console
  params: {}
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewLoader()
	config, err := loader.LoadFromFile(configPath)

	if err != nil {
		t.Fatalf("LoadFromFile() returned error: %v", err)
	}

	// Empty params should be valid
	if config.Provider.Params == nil {
		t.Error("Provider params should not be nil")
	}
	if config.Formatter.Params == nil {
		t.Error("Formatter params should not be nil")
	}
	if config.Output.Params == nil {
		t.Error("Output params should not be nil")
	}
}
