package formatter

import (
	"context"
	"encoding/json"
)

// JSONFormatter formats data as JSON with optional indentation.
// It provides clean, readable JSON output with context-aware cancellation.
//
// Context handling:
//   - Checks context before formatting
//   - Returns ctx.Err() if canceled
//   - Uses standard library json.Marshal (atomic operation)
//
// Note: json.Marshal is a single operation that doesn't check context
// during execution. For extremely large datasets, consider implementing
// a streaming JSON formatter that checks context periodically.
//
// Thread-safe: Yes. JSONFormatter is immutable after creation.
type JSONFormatter struct {
	// Indent controls JSON indentation for readability
	// Empty string produces compact JSON
	// "  " (two spaces) produces indented JSON
	Indent string
}

// NewJSONFormatter creates a new JSONFormatter with indentation.
// This is the recommended way to create a JSONFormatter.
//
// Example with indentation:
//
//	formatter := formatter.NewJSONFormatter("  ")
//	result, err := formatter.Format(ctx, data)
//	// Produces pretty-printed JSON
//
// Example without indentation (compact):
//
//	formatter := formatter.NewJSONFormatter("")
//	result, err := formatter.Format(ctx, data)
//	// Produces compact JSON
//
// Parameters:
//   - indent: Indentation string (e.g., "", "  ", "\t")
//
// Returns:
//   - *JSONFormatter: A new formatter instance
func NewJSONFormatter(indent string) *JSONFormatter {
	return &JSONFormatter{
		Indent: indent,
	}
}

// Format converts data to JSON format.
// The output is either compact or indented based on the Indent setting.
//
// Context handling:
//   - Checks context before formatting
//   - Returns ctx.Err() if already canceled
//   - json.Marshal itself doesn't check context (atomic operation)
//
// For large datasets (>10MB), the json.Marshal operation may take
// considerable time. In production, consider implementing a streaming
// JSON formatter that checks context periodically.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - data: Data to format as JSON
//
// Returns:
//   - []byte: JSON-formatted data
//   - error: ctx.Err() if context canceled, or json.Marshal error
func (j *JSONFormatter) Format(ctx context.Context, data []map[string]interface{}) ([]byte, error) {
	// Check if context is already canceled
	// This provides a fast path before expensive JSON serialization
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Format as JSON
	// For indented output, use MarshalIndent
	// For compact output, use Marshal
	var result []byte
	var err error

	if j.Indent != "" {
		// Pretty-printed JSON with indentation
		result, err = json.MarshalIndent(data, "", j.Indent)
	} else {
		// Compact JSON without indentation
		result, err = json.Marshal(data)
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}
