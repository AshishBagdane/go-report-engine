package provider

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/AshishBagdane/go-report-engine/internal/memory"
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

// Stream returns an Iterator for streaming data access.
func (p *CSVProvider) Stream(ctx context.Context) (Iterator, error) {
	if p.FilePath == "" {
		return nil, fmt.Errorf("csv provider: file path not configured")
	}

	file, err := os.Open(p.FilePath)
	if err != nil {
		return nil, fmt.Errorf("csv provider: failed to open file: %w", err)
	}

	reader := csv.NewReader(file)
	reader.Comma = p.Delimiter

	var headers []string

	// Read first record to determine headers
	firstRecord, err := reader.Read()
	if err == io.EOF {
		// Empty file
		_ = file.Close()
		return &CSVIterator{
			file:   nil, // Closed
			err:    nil, // Empty stream
			reader: nil,
		}, nil
	}
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("csv provider: failed to read header/first row: %w", err)
	}

	if p.HasHeader {
		headers = firstRecord
	} else {
		// Generate default headers
		headers = make([]string, len(firstRecord))
		for i := range headers {
			headers[i] = fmt.Sprintf("col_%d", i+1)
		}

		// If NO header, we need to make sure the first record (which is data)
		// is returned by the first call to Next().
		// However, standard csv.Reader tracks position.
		// If we consumed it, we can't un-read it easily without rewinding (seek).
		// Since we cannot rely on Seek (might be a stream), we have a "pushback" problem.
		//
		// Solution: CSVIterator can have an optional `firstRecord` buffer.
	}

	it := &CSVIterator{
		file:    file,
		reader:  reader,
		headers: headers,
	}

	if !p.HasHeader {
		// Pre-populate the buffer with the already-read first record
		it.nextRec = firstRecord
		it.hasBuf = true
	}

	return it, nil
}

type CSVIterator struct {
	file    *os.File
	reader  *csv.Reader
	headers []string
	current map[string]interface{}
	err     error

	// For handling no-header case where we consumed the first row
	nextRec []string
	hasBuf  bool
}

func (it *CSVIterator) Next() bool {
	if it.err != nil {
		return false
	}
	if it.reader == nil { // Handle empty file case
		return false
	}

	var record []string
	var err error

	if it.hasBuf {
		record = it.nextRec
		it.hasBuf = false
	} else {
		record, err = it.reader.Read()
		if err == io.EOF {
			return false
		}
		if err != nil {
			it.err = err
			return false
		}
	}

	it.current = memory.GetMap()
	for i, val := range record {
		if i < len(it.headers) {
			it.current[it.headers[i]] = val
		}
	}
	return true
}

func (it *CSVIterator) Value() map[string]interface{} {
	return it.current
}

func (it *CSVIterator) Err() error {
	return it.err
}

func (it *CSVIterator) Close() error {
	if it.file != nil {
		return it.file.Close()
	}
	return nil
}
