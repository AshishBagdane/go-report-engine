package formatter

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

// YAMLFormatter implements FormatterStrategy for YAML output.
type YAMLFormatter struct {
	Indent int
}

// NewYAMLFormatter creates a new instance of YAMLFormatter with default indentation (2 spaces).
func NewYAMLFormatter() *YAMLFormatter {
	return &YAMLFormatter{
		Indent: 2,
	}
}

// Format converts the data into YAML format.
func (f *YAMLFormatter) Format(ctx context.Context, data []map[string]interface{}) ([]byte, error) {
	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if len(data) == 0 {
		return []byte("[]\n"), nil
	}

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(f.Indent)
	defer encoder.Close()

	if err := encoder.Encode(data); err != nil {
		return nil, fmt.Errorf("yaml formatter: failed to encode data: %w", err)
	}

	return buf.Bytes(), nil
}

// Configure sets up the formatter from a map of parameters.
// Params:
// - indent: Number of spaces for indentation (default: "2")
func (f *YAMLFormatter) Configure(params map[string]string) error {
	if indentStr, ok := params["indent"]; ok {
		indent, err := strconv.Atoi(indentStr)
		if err != nil {
			return fmt.Errorf("yaml formatter: invalid indent %s: %w", indentStr, err)
		}
		if indent < 1 || indent > 8 {
			// yaml.v3 might not have strict limits, but 2 or 4 is standard.
			// Letting reasonable range.
		}
		f.Indent = indent
	}

	return nil
}
