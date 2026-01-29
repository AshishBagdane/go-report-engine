package provider

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

// CSVProvider implements ProviderStrategy for reading CSV files.
type CSVProvider struct {
	FilePath  string
	Delimiter rune
	HasHeader bool
}

// NewCSVProvider creates a new instance of CSVProvider with defaults.
func NewCSVProvider() *CSVProvider {
	return &CSVProvider{
		Delimiter: ',',
		HasHeader: true,
	}
}

// Fetch reads data from the CSV file and returns it as a slice of maps.
func (p *CSVProvider) Fetch(ctx context.Context) ([]map[string]interface{}, error) {
	if p.FilePath == "" {
		return nil, fmt.Errorf("csv provider: file path not configured")
	}

	// check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	file, err := os.Open(p.FilePath)
	if err != nil {
		return nil, fmt.Errorf("csv provider: failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = p.Delimiter

	// Handle LazyQuotes if needed, but for now stick to standard compliant CSV
	// reader.LazyQuotes = true

	var headers []string
	var records []map[string]interface{}

	// Read first record
	firstRecord, err := reader.Read()
	if err == io.EOF {
		return []map[string]interface{}{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("csv provider: failed to read header/first row: %w", err)
	}

	if p.HasHeader {
		headers = firstRecord
	} else {
		// Generate default headers col_1, col_2, etc.
		headers = make([]string, len(firstRecord))
		for i := range headers {
			headers[i] = fmt.Sprintf("col_%d", i+1)
		}
		// If no header, the first record is actual data
		recordMap := make(map[string]interface{})
		for i, val := range firstRecord {
			if i < len(headers) {
				recordMap[headers[i]] = val
			}
		}
		records = append(records, recordMap)
	}

	// Read remaining records
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("csv provider: error reading row: %w", err)
		}

		// Skip empty rows or rows with mismatching length if strict,
		// but Go's csv reader handles field count validation by default unless FieldsPerRecord is -1

		recordMap := make(map[string]interface{})
		for i, val := range row {
			if i < len(headers) {
				recordMap[headers[i]] = val
			}
		}
		records = append(records, recordMap)
	}

	return records, nil
}

// Configure sets up the provider from a map of parameters.
// Params:
// - file_path: Path to the CSV file (required)
// - delimiter: Character separator (default: ",")
// - has_header: "true" or "false" (default: "true")
func (p *CSVProvider) Configure(params map[string]string) error {
	if path, ok := params["file_path"]; ok {
		p.FilePath = path
	} else {
		return fmt.Errorf("csv provider: missing required parameter 'file_path'")
	}

	if delim, ok := params["delimiter"]; ok {
		if len(delim) != 1 {
			return fmt.Errorf("csv provider: delimiter must be a single character")
		}
		p.Delimiter = rune(delim[0])
	}

	if hasHeader, ok := params["has_header"]; ok {
		if strings.ToLower(hasHeader) == "false" {
			p.HasHeader = false
		} else {
			p.HasHeader = true
		}
	}

	return nil
}
