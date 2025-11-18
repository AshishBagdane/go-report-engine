package api

import "fmt"

// Configurable allows a strategy to read parameters from the Engine's config file.
// Any component (Processor, Provider, Formatter, or Output) can optionally implement this.
type Configurable interface {
	Configure(params map[string]string) error
}

// FilterStrategy determines if a record should be kept.
// Users implement this to remove unwanted records.
type FilterStrategy interface {
	Keep(row map[string]interface{}) bool
}

// ValidatorStrategy checks if a record is valid.
// Users implement this to assert data correctness.
type ValidatorStrategy interface {
	Validate(row map[string]interface{}) error
}

// TransformerStrategy modifies a record.
// Users implement this to perform data mapping or enrichment.
type TransformerStrategy interface {
	Transform(row map[string]interface{}) map[string]interface{}
}

// ErrMissingParam is a utility error for configuration issues.
var ErrMissingParam = func(p string) error { return fmt.Errorf("configuration error: missing required parameter '%s'", p) }
