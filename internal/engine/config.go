package engine

import "fmt"

// Config is the top-level configuration for the Report Engine.
// It defines which provider, processor pipeline, formatter,
// and output module should be used to generate reports.
type Config struct {
	Provider   ProviderConfig    `json:"provider" yaml:"provider"`
	Processors []ProcessorConfig `json:"processors" yaml:"processors"`
	Formatter  FormatterConfig   `json:"formatter" yaml:"formatter"`
	Output     OutputConfig      `json:"output" yaml:"output"`
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

// Validate ensures required fields are present.
func (c Config) Validate() error {
	if c.Provider.Type == "" {
		return ErrMissingProvider
	}
	if c.Formatter.Type == "" {
		return ErrMissingFormatter
	}
	if c.Output.Type == "" {
		return ErrMissingOutput
	}
	return nil
}

var (
	ErrMissingProvider  = fmt.Errorf("provider.type is required")
	ErrMissingFormatter = fmt.Errorf("formatter.type is required")
	ErrMissingOutput    = fmt.Errorf("output.type is required")
)
