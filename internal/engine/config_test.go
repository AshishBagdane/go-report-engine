package engine

import (
	"strings"
	"testing"
)

// TestConfigValidateSuccess tests successful validation
func TestConfigValidateSuccess(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "minimal valid config",
			config: Config{
				Provider:   ProviderConfig{Type: "mock"},
				Processors: []ProcessorConfig{},
				Formatter:  FormatterConfig{Type: "json"},
				Output:     OutputConfig{Type: "console"},
			},
		},
		{
			name: "config with parameters",
			config: Config{
				Provider: ProviderConfig{
					Type:   "postgres",
					Params: map[string]string{"host": "localhost", "port": "5432"},
				},
				Processors: []ProcessorConfig{
					{Type: "filter", Params: map[string]string{"threshold": "50"}},
				},
				Formatter: FormatterConfig{
					Type:   "json",
					Params: map[string]string{"indent": "2"},
				},
				Output: OutputConfig{
					Type:   "file",
					Params: map[string]string{"path": "/tmp/output.json"},
				},
			},
		},
		{
			name: "config with multiple processors",
			config: Config{
				Provider: ProviderConfig{Type: "mock"},
				Processors: []ProcessorConfig{
					{Type: "filter", Params: map[string]string{}},
					{Type: "validator", Params: map[string]string{}},
					{Type: "transformer", Params: map[string]string{}},
				},
				Formatter: FormatterConfig{Type: "json"},
				Output:    OutputConfig{Type: "console"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if err != nil {
				t.Errorf("Validate() should succeed, got error: %v", err)
			}
		})
	}
}

// TestConfigValidateFailures tests validation failures
func TestConfigValidateFailures(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError string
	}{
		{
			name: "missing provider type",
			config: Config{
				Provider:   ProviderConfig{Type: ""},
				Processors: []ProcessorConfig{},
				Formatter:  FormatterConfig{Type: "json"},
				Output:     OutputConfig{Type: "console"},
			},
			expectError: "provider.type is required",
		},
		{
			name: "whitespace provider type",
			config: Config{
				Provider:   ProviderConfig{Type: "   "},
				Processors: []ProcessorConfig{},
				Formatter:  FormatterConfig{Type: "json"},
				Output:     OutputConfig{Type: "console"},
			},
			expectError: "whitespace",
		},
		{
			name: "missing formatter type",
			config: Config{
				Provider:   ProviderConfig{Type: "mock"},
				Processors: []ProcessorConfig{},
				Formatter:  FormatterConfig{Type: ""},
				Output:     OutputConfig{Type: "console"},
			},
			expectError: "formatter.type is required",
		},
		{
			name: "whitespace formatter type",
			config: Config{
				Provider:   ProviderConfig{Type: "mock"},
				Processors: []ProcessorConfig{},
				Formatter:  FormatterConfig{Type: "  \t  "},
				Output:     OutputConfig{Type: "console"},
			},
			expectError: "whitespace",
		},
		{
			name: "missing output type",
			config: Config{
				Provider:   ProviderConfig{Type: "mock"},
				Processors: []ProcessorConfig{},
				Formatter:  FormatterConfig{Type: "json"},
				Output:     OutputConfig{Type: ""},
			},
			expectError: "output.type is required",
		},
		{
			name: "whitespace output type",
			config: Config{
				Provider:   ProviderConfig{Type: "mock"},
				Processors: []ProcessorConfig{},
				Formatter:  FormatterConfig{Type: "json"},
				Output:     OutputConfig{Type: "\n\t"},
			},
			expectError: "whitespace",
		},
		{
			name: "empty processor type",
			config: Config{
				Provider: ProviderConfig{Type: "mock"},
				Processors: []ProcessorConfig{
					{Type: "", Params: map[string]string{}},
				},
				Formatter: FormatterConfig{Type: "json"},
				Output:    OutputConfig{Type: "console"},
			},
			expectError: "processor[0].type is required",
		},
		{
			name: "whitespace processor type",
			config: Config{
				Provider: ProviderConfig{Type: "mock"},
				Processors: []ProcessorConfig{
					{Type: "  ", Params: map[string]string{}},
				},
				Formatter: FormatterConfig{Type: "json"},
				Output:    OutputConfig{Type: "console"},
			},
			expectError: "processor[0].type cannot be empty or whitespace",
		},
		{
			name: "empty param key in provider",
			config: Config{
				Provider: ProviderConfig{
					Type:   "postgres",
					Params: map[string]string{"": "value"},
				},
				Processors: []ProcessorConfig{},
				Formatter:  FormatterConfig{Type: "json"},
				Output:     OutputConfig{Type: "console"},
			},
			expectError: "provider.params contains empty key",
		},
		{
			name: "whitespace param key in formatter",
			config: Config{
				Provider:   ProviderConfig{Type: "mock"},
				Processors: []ProcessorConfig{},
				Formatter: FormatterConfig{
					Type:   "json",
					Params: map[string]string{"  ": "value"},
				},
				Output: OutputConfig{Type: "console"},
			},
			expectError: "formatter.params contains whitespace-only key",
		},
		{
			name: "multiple validation errors",
			config: Config{
				Provider:   ProviderConfig{Type: ""},
				Processors: []ProcessorConfig{},
				Formatter:  FormatterConfig{Type: ""},
				Output:     OutputConfig{Type: ""},
			},
			expectError: "config validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if err == nil {
				t.Fatal("Validate() should return error")
			}

			if !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("Error should contain %q, got: %v", tt.expectError, err)
			}
		})
	}
}

// TestConfigValidateEmptyProcessors tests empty processor list is valid
func TestConfigValidateEmptyProcessors(t *testing.T) {
	config := Config{
		Provider:   ProviderConfig{Type: "mock"},
		Processors: []ProcessorConfig{},
		Formatter:  FormatterConfig{Type: "json"},
		Output:     OutputConfig{Type: "console"},
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Empty processor list should be valid, got: %v", err)
	}
}

// TestConfigValidateNilProcessors tests nil processor list is valid
func TestConfigValidateNilProcessors(t *testing.T) {
	config := Config{
		Provider:   ProviderConfig{Type: "mock"},
		Processors: nil,
		Formatter:  FormatterConfig{Type: "json"},
		Output:     OutputConfig{Type: "console"},
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Nil processor list should be valid, got: %v", err)
	}
}

// TestConfigValidateEmptyParams tests empty params are allowed
func TestConfigValidateEmptyParams(t *testing.T) {
	config := Config{
		Provider: ProviderConfig{
			Type:   "mock",
			Params: map[string]string{"key": ""}, // Empty value is allowed
		},
		Processors: []ProcessorConfig{},
		Formatter:  FormatterConfig{Type: "json"},
		Output:     OutputConfig{Type: "console"},
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Empty param values should be allowed, got: %v", err)
	}
}

// TestValidateProviderConfig tests standalone provider validation
func TestValidateProviderConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      ProviderConfig
		shouldError bool
	}{
		{
			name:        "valid config",
			config:      ProviderConfig{Type: "mock"},
			shouldError: false,
		},
		{
			name:        "empty type",
			config:      ProviderConfig{Type: ""},
			shouldError: true,
		},
		{
			name:        "whitespace type",
			config:      ProviderConfig{Type: "  "},
			shouldError: true,
		},
		{
			name: "valid with params",
			config: ProviderConfig{
				Type:   "postgres",
				Params: map[string]string{"host": "localhost"},
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProviderConfig(tt.config)
			if (err != nil) != tt.shouldError {
				t.Errorf("ValidateProviderConfig() error = %v, shouldError = %v", err, tt.shouldError)
			}
		})
	}
}

// TestValidateProcessorConfig tests standalone processor validation
func TestValidateProcessorConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      ProcessorConfig
		shouldError bool
	}{
		{
			name:        "valid config",
			config:      ProcessorConfig{Type: "filter"},
			shouldError: false,
		},
		{
			name:        "empty type",
			config:      ProcessorConfig{Type: ""},
			shouldError: true,
		},
		{
			name:        "whitespace type",
			config:      ProcessorConfig{Type: "\t\n"},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProcessorConfig(tt.config)
			if (err != nil) != tt.shouldError {
				t.Errorf("ValidateProcessorConfig() error = %v, shouldError = %v", err, tt.shouldError)
			}
		})
	}
}

// TestValidateFormatterConfig tests standalone formatter validation
func TestValidateFormatterConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      FormatterConfig
		shouldError bool
	}{
		{
			name:        "valid config",
			config:      FormatterConfig{Type: "json"},
			shouldError: false,
		},
		{
			name:        "empty type",
			config:      FormatterConfig{Type: ""},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFormatterConfig(tt.config)
			if (err != nil) != tt.shouldError {
				t.Errorf("ValidateFormatterConfig() error = %v, shouldError = %v", err, tt.shouldError)
			}
		})
	}
}

// TestValidateOutputConfig tests standalone output validation
func TestValidateOutputConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      OutputConfig
		shouldError bool
	}{
		{
			name:        "valid config",
			config:      OutputConfig{Type: "console"},
			shouldError: false,
		},
		{
			name:        "empty type",
			config:      OutputConfig{Type: ""},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOutputConfig(tt.config)
			if (err != nil) != tt.shouldError {
				t.Errorf("ValidateOutputConfig() error = %v, shouldError = %v", err, tt.shouldError)
			}
		})
	}
}

// TestPredefinedErrors tests predefined error constants
func TestPredefinedErrors(t *testing.T) {
	if ErrMissingProvider == nil {
		t.Error("ErrMissingProvider should not be nil")
	}
	if ErrMissingFormatter == nil {
		t.Error("ErrMissingFormatter should not be nil")
	}
	if ErrMissingOutput == nil {
		t.Error("ErrMissingOutput should not be nil")
	}
	if ErrInvalidConfig == nil {
		t.Error("ErrInvalidConfig should not be nil")
	}
}

// BenchmarkConfigValidate benchmarks config validation
func BenchmarkConfigValidate(b *testing.B) {
	config := Config{
		Provider: ProviderConfig{
			Type:   "postgres",
			Params: map[string]string{"host": "localhost", "port": "5432"},
		},
		Processors: []ProcessorConfig{
			{Type: "filter", Params: map[string]string{"threshold": "50"}},
			{Type: "validator", Params: map[string]string{}},
		},
		Formatter: FormatterConfig{Type: "json"},
		Output:    OutputConfig{Type: "console"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}
