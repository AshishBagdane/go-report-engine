package engine

import (
	"fmt"
	"strings"
)

// Config is the top-level configuration for the Report Engine.
// It defines which provider, processor pipeline, formatter,
// and output module should be used to generate reports.
type Config struct {
	Provider       ProviderConfig        `json:"provider" yaml:"provider"`
	Processors     []ProcessorConfig     `json:"processors" yaml:"processors"`
	Formatter      FormatterConfig       `json:"formatter" yaml:"formatter"`
	Output         OutputConfig          `json:"output" yaml:"output"`
	Retry          *RetryConfig          `json:"retry,omitempty" yaml:"retry,omitempty"`
	CircuitBreaker *CircuitBreakerConfig `json:"circuit_breaker,omitempty" yaml:"circuit_breaker,omitempty"`
}

// RetryConfig defines the retry policy settings.
type RetryConfig struct {
	MaxRetries int     `json:"max_retries" yaml:"max_retries"`
	BaseDelay  string  `json:"base_delay" yaml:"base_delay"` // Parsed to time.Duration
	MaxDelay   string  `json:"max_delay" yaml:"max_delay"`   // Parsed to time.Duration
	Factor     float64 `json:"factor" yaml:"factor"`
	Jitter     bool    `json:"jitter" yaml:"jitter"`
}

// CircuitBreakerConfig defines circuit breaker settings.
type CircuitBreakerConfig struct {
	FailureThreshold uint   `json:"failure_threshold" yaml:"failure_threshold"`
	ResetTimeout     string `json:"reset_timeout" yaml:"reset_timeout"` // Parsed to time.Duration
}

// ProviderConfig represents the selected provider and its parameters.
type ProviderConfig struct {
	Type   string            `json:"type" yaml:"type"` // e.g., "mock", "sql", "file"
	Params map[string]string `json:"params" yaml:"params"`
}

// ProcessorConfig represents a single processor in the processing pipeline.
type ProcessorConfig struct {
	Type   string            `json:"type" yaml:"type"` // e.g., "sanitize", "aggregate"
	Params map[string]string `json:"params" yaml:"params"`
}

// FormatterConfig represents output formatting settings.
type FormatterConfig struct {
	Type   string            `json:"type" yaml:"type"` // e.g., "json", "csv", "html"
	Params map[string]string `json:"params" yaml:"params"`
}

// OutputConfig represents how the formatted data is delivered.
type OutputConfig struct {
	Type   string            `json:"type" yaml:"type"` // e.g., "console", "file", "s3"
	Params map[string]string `json:"params" yaml:"params"`
}

// Validate performs comprehensive validation of the configuration.
// It checks all required fields and validates parameter structures.
//
// Returns nil if valid, or a detailed error describing what is invalid.
func (c Config) Validate() error {
	var errors []string

	// Validate Provider
	if err := c.validateProvider(); err != nil {
		errors = append(errors, err.Error())
	}

	// Validate Processors
	if err := c.validateProcessors(); err != nil {
		errors = append(errors, err.Error())
	}

	// Validate Formatter
	if err := c.validateFormatter(); err != nil {
		errors = append(errors, err.Error())
	}

	// Validate Output
	if err := c.validateOutput(); err != nil {
		errors = append(errors, err.Error())
	}

	// Validate Retry (Optional, but if present must be valid)
	// We don't have a strict validator for it yet as it's optional,
	// but we could ensure Factor >= 1.0 if specified.

	if len(errors) > 0 {
		return fmt.Errorf("config validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// validateProvider validates the provider configuration
func (c Config) validateProvider() error {
	if c.Provider.Type == "" {
		return ErrMissingProvider
	}

	// Validate type is not just whitespace
	if strings.TrimSpace(c.Provider.Type) == "" {
		return fmt.Errorf("provider.type cannot be empty or whitespace")
	}

	// Validate parameters if present
	if err := validateParams(c.Provider.Params, "provider"); err != nil {
		return err
	}

	return nil
}

// validateProcessors validates the processor configurations
func (c Config) validateProcessors() error {
	// Empty processor list is valid (will use base processor)
	if len(c.Processors) == 0 {
		return nil
	}

	for i, proc := range c.Processors {
		if proc.Type == "" {
			return fmt.Errorf("processor[%d].type is required", i)
		}

		// Validate type is not just whitespace
		if strings.TrimSpace(proc.Type) == "" {
			return fmt.Errorf("processor[%d].type cannot be empty or whitespace", i)
		}

		// Validate parameters if present
		if err := validateParams(proc.Params, fmt.Sprintf("processor[%d]", i)); err != nil {
			return err
		}
	}

	return nil
}

// validateFormatter validates the formatter configuration
func (c Config) validateFormatter() error {
	if c.Formatter.Type == "" {
		return ErrMissingFormatter
	}

	// Validate type is not just whitespace
	if strings.TrimSpace(c.Formatter.Type) == "" {
		return fmt.Errorf("formatter.type cannot be empty or whitespace")
	}

	// Validate parameters if present
	if err := validateParams(c.Formatter.Params, "formatter"); err != nil {
		return err
	}

	return nil
}

// validateOutput validates the output configuration
func (c Config) validateOutput() error {
	if c.Output.Type == "" {
		return ErrMissingOutput
	}

	// Validate type is not just whitespace
	if strings.TrimSpace(c.Output.Type) == "" {
		return fmt.Errorf("output.type cannot be empty or whitespace")
	}

	// Validate parameters if present
	if err := validateParams(c.Output.Params, "output"); err != nil {
		return err
	}

	return nil
}

// validateParams validates parameter map for empty keys or values
func validateParams(params map[string]string, context string) error {
	if params == nil {
		return nil
	}

	for key, value := range params {
		// Check for empty keys
		if key == "" {
			return fmt.Errorf("%s.params contains empty key", context)
		}

		// Check for whitespace-only keys
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("%s.params contains whitespace-only key", context)
		}

		// Check for empty values (warn but don't fail - might be intentional)
		if value == "" {
			// Empty values are allowed but we could log a warning in production
			continue
		}
	}

	return nil
}

// Predefined errors for common validation failures
var (
	// ErrMissingProvider indicates provider.type is not set
	ErrMissingProvider = fmt.Errorf("provider.type is required")

	// ErrMissingFormatter indicates formatter.type is not set
	ErrMissingFormatter = fmt.Errorf("formatter.type is required")

	// ErrMissingOutput indicates output.type is not set
	ErrMissingOutput = fmt.Errorf("output.type is required")

	// ErrInvalidConfig indicates the overall config is invalid
	ErrInvalidConfig = fmt.Errorf("invalid configuration")
)

// ValidateProviderConfig validates a provider config independently
func ValidateProviderConfig(cfg ProviderConfig) error {
	if cfg.Type == "" || strings.TrimSpace(cfg.Type) == "" {
		return fmt.Errorf("provider type is required and cannot be empty")
	}
	return validateParams(cfg.Params, "provider")
}

// ValidateProcessorConfig validates a processor config independently
func ValidateProcessorConfig(cfg ProcessorConfig) error {
	if cfg.Type == "" || strings.TrimSpace(cfg.Type) == "" {
		return fmt.Errorf("processor type is required and cannot be empty")
	}
	return validateParams(cfg.Params, "processor")
}

// ValidateFormatterConfig validates a formatter config independently
func ValidateFormatterConfig(cfg FormatterConfig) error {
	if cfg.Type == "" || strings.TrimSpace(cfg.Type) == "" {
		return fmt.Errorf("formatter type is required and cannot be empty")
	}
	return validateParams(cfg.Params, "formatter")
}

// ValidateOutputConfig validates an output config independently
func ValidateOutputConfig(cfg OutputConfig) error {
	if cfg.Type == "" || strings.TrimSpace(cfg.Type) == "" {
		return fmt.Errorf("output type is required and cannot be empty")
	}
	return validateParams(cfg.Params, "output")
}
