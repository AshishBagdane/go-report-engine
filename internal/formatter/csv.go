package formatter

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"sort"
	"strings"
)

// CSVFormatter implements FormatterStrategy for CSV output.
type CSVFormatter struct {
	Delimiter     rune
	IncludeHeader bool
}

// NewCSVFormatter creates a new instance of CSVFormatter with defaults.
func NewCSVFormatter() *CSVFormatter {
	return &CSVFormatter{
		Delimiter:     ',',
		IncludeHeader: true,
	}
}

// Format converts the data into CSV format.
func (f *CSVFormatter) Format(ctx context.Context, data []map[string]interface{}) ([]byte, error) {
	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if len(data) == 0 {
		return []byte(""), nil
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.Comma = f.Delimiter

	// Extract headers from the first record
	headers := make([]string, 0, len(data[0]))
	for k := range data[0] {
		headers = append(headers, k)
	}
	sort.Strings(headers) // Sort headers for deterministic output

	// Write header
	if f.IncludeHeader {
		if err := writer.Write(headers); err != nil {
			return nil, fmt.Errorf("csv formatter: failed to write header: %w", err)
		}
	}

	// Write records
	for _, record := range data {
		// Check context periodically
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		row := make([]string, len(headers))
		for i, h := range headers {
			if val, ok := record[h]; ok {
				row[i] = fmt.Sprintf("%v", val)
			} else {
				row[i] = ""
			}
		}
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("csv formatter: failed to write record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("csv formatter: flush error: %w", err)
	}

	return buf.Bytes(), nil
}

// Configure sets up the formatter from a map of parameters.
// Params:
// - delimiter: Character separator (default: ",")
// - include_header: "true" or "false" (default: "true")
func (f *CSVFormatter) Configure(params map[string]string) error {
	if delim, ok := params["delimiter"]; ok {
		if len(delim) != 1 {
			return fmt.Errorf("csv formatter: delimiter must be a single character")
		}
		f.Delimiter = rune(delim[0])
	}

	if includeHeader, ok := params["include_header"]; ok {
		if strings.ToLower(includeHeader) == "false" {
			f.IncludeHeader = false
		} else {
			f.IncludeHeader = true
		}
	}

	return nil
}
