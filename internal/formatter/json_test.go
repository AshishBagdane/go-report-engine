package formatter

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestNewJSONFormatter tests the factory function
func TestNewJSONFormatter(t *testing.T) {
	formatter := NewJSONFormatter()

	if formatter == nil {
		t.Fatal("NewJSONFormatter() returned nil")
	}

	// Verify it implements FormatStrategy
	var _ FormatStrategy = formatter
}

// TestJSONFormatterFormat tests basic formatting
func TestJSONFormatterFormat(t *testing.T) {
	formatter := NewJSONFormatter()

	testData := []map[string]interface{}{
		{"id": 1, "name": "Alice", "score": 95},
		{"id": 2, "name": "Bob", "score": 88},
	}

	result, err := formatter.Format(testData)

	if err != nil {
		t.Fatalf("Format() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Format() returned nil")
	}

	if len(result) == 0 {
		t.Fatal("Format() returned empty data")
	}

	// Verify it's valid JSON
	var parsed []map[string]interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("Format() did not produce valid JSON: %v", err)
	}

	// Verify data integrity
	if len(parsed) != len(testData) {
		t.Errorf("Parsed JSON has %d records, expected %d", len(parsed), len(testData))
	}
}

// TestJSONFormatterFormatEmpty tests formatting empty data
func TestJSONFormatterFormatEmpty(t *testing.T) {
	formatter := NewJSONFormatter()

	testData := []map[string]interface{}{}

	result, err := formatter.Format(testData)

	if err != nil {
		t.Fatalf("Format() returned error: %v", err)
	}

	// Verify it's valid JSON (empty array)
	var parsed []map[string]interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("Format() did not produce valid JSON: %v", err)
	}

	if len(parsed) != 0 {
		t.Errorf("Parsed JSON has %d records, expected 0", len(parsed))
	}
}

// TestJSONFormatterFormatSingleRecord tests single record
func TestJSONFormatterFormatSingleRecord(t *testing.T) {
	formatter := NewJSONFormatter()

	testData := []map[string]interface{}{
		{"id": 1, "name": "Alice"},
	}

	result, err := formatter.Format(testData)

	if err != nil {
		t.Fatalf("Format() returned error: %v", err)
	}

	var parsed []map[string]interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("Format() did not produce valid JSON: %v", err)
	}

	if len(parsed) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(parsed))
	}

	if parsed[0]["name"] != "Alice" {
		t.Errorf("Name field = %v, expected 'Alice'", parsed[0]["name"])
	}
}

// TestJSONFormatterFormatDataTypes tests various data types
func TestJSONFormatterFormatDataTypes(t *testing.T) {
	formatter := NewJSONFormatter()

	testData := []map[string]interface{}{
		{
			"string": "hello",
			"int":    42,
			"float":  3.14,
			"bool":   true,
			"null":   nil,
			"array":  []int{1, 2, 3},
			"nested": map[string]interface{}{"key": "value"},
		},
	}

	result, err := formatter.Format(testData)

	if err != nil {
		t.Fatalf("Format() returned error: %v", err)
	}

	var parsed []map[string]interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("Format() did not produce valid JSON: %v", err)
	}

	if len(parsed) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(parsed))
	}

	record := parsed[0]

	// Verify string
	if record["string"] != "hello" {
		t.Errorf("String field = %v, expected 'hello'", record["string"])
	}

	// Verify int (JSON numbers are float64)
	if record["int"].(float64) != 42 {
		t.Errorf("Int field = %v, expected 42", record["int"])
	}

	// Verify float
	if record["float"].(float64) != 3.14 {
		t.Errorf("Float field = %v, expected 3.14", record["float"])
	}

	// Verify bool
	if record["bool"] != true {
		t.Errorf("Bool field = %v, expected true", record["bool"])
	}

	// Verify null
	if record["null"] != nil {
		t.Errorf("Null field = %v, expected nil", record["null"])
	}

	// Verify array exists
	if _, ok := record["array"]; !ok {
		t.Error("Array field is missing")
	}

	// Verify nested object exists
	if _, ok := record["nested"]; !ok {
		t.Error("Nested field is missing")
	}
}

// TestJSONFormatterFormatIndentation tests pretty-printing
func TestJSONFormatterFormatIndentation(t *testing.T) {
	formatter := NewJSONFormatter()

	testData := []map[string]interface{}{
		{"id": 1, "name": "Alice"},
	}

	result, err := formatter.Format(testData)

	if err != nil {
		t.Fatalf("Format() returned error: %v", err)
	}

	// Check for indentation (MarshalIndent uses 2 spaces)
	resultStr := string(result)
	if !strings.Contains(resultStr, "  ") {
		t.Error("JSON output does not appear to be indented")
	}

	// Check for newlines
	if !strings.Contains(resultStr, "\n") {
		t.Error("JSON output does not contain newlines")
	}
}

// TestJSONFormatterFormatSpecialCharacters tests special character handling
func TestJSONFormatterFormatSpecialCharacters(t *testing.T) {
	formatter := NewJSONFormatter()

	testData := []map[string]interface{}{
		{
			"quotes":    `"quoted"`,
			"newline":   "line1\nline2",
			"tab":       "col1\tcol2",
			"backslash": `path\to\file`,
			"unicode":   "Hello, 世界",
		},
	}

	result, err := formatter.Format(testData)

	if err != nil {
		t.Fatalf("Format() returned error: %v", err)
	}

	// Verify it's valid JSON
	var parsed []map[string]interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("Format() did not produce valid JSON: %v", err)
	}

	// Verify special characters are preserved
	record := parsed[0]
	if record["quotes"] != `"quoted"` {
		t.Error("Quotes not preserved correctly")
	}

	if record["unicode"] != "Hello, 世界" {
		t.Error("Unicode not preserved correctly")
	}
}

// TestJSONFormatterFormatNil tests nil data handling
func TestJSONFormatterFormatNil(t *testing.T) {
	formatter := NewJSONFormatter()

	result, err := formatter.Format(nil)

	if err != nil {
		t.Fatalf("Format(nil) returned error: %v", err)
	}

	// Should produce valid JSON for null
	var parsed interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("Format(nil) did not produce valid JSON: %v", err)
	}
}

// TestJSONFormatterFormatLargeDataset tests with large dataset
func TestJSONFormatterFormatLargeDataset(t *testing.T) {
	formatter := NewJSONFormatter()

	// Create large dataset
	const recordCount = 1000
	testData := make([]map[string]interface{}, recordCount)
	for i := 0; i < recordCount; i++ {
		testData[i] = map[string]interface{}{
			"id":    i,
			"name":  "User",
			"value": i * 2,
		}
	}

	result, err := formatter.Format(testData)

	if err != nil {
		t.Fatalf("Format() failed on large dataset: %v", err)
	}

	var parsed []map[string]interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("Format() did not produce valid JSON: %v", err)
	}

	if len(parsed) != recordCount {
		t.Errorf("Parsed JSON has %d records, expected %d", len(parsed), recordCount)
	}
}

// TestJSONFormatterFormatNestedStructures tests deeply nested data
func TestJSONFormatterFormatNestedStructures(t *testing.T) {
	formatter := NewJSONFormatter()

	testData := []map[string]interface{}{
		{
			"id": 1,
			"user": map[string]interface{}{
				"name": "Alice",
				"address": map[string]interface{}{
					"street": "123 Main St",
					"city":   "Springfield",
				},
				"tags": []string{"admin", "user"},
			},
		},
	}

	result, err := formatter.Format(testData)

	if err != nil {
		t.Fatalf("Format() returned error: %v", err)
	}

	var parsed []map[string]interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("Format() did not produce valid JSON: %v", err)
	}

	// Verify nested structure is preserved
	user := parsed[0]["user"].(map[string]interface{})
	if user["name"] != "Alice" {
		t.Error("Nested name not preserved")
	}

	address := user["address"].(map[string]interface{})
	if address["city"] != "Springfield" {
		t.Error("Deeply nested city not preserved")
	}
}

// TestJSONFormatterImplementsInterface verifies interface implementation
func TestJSONFormatterImplementsInterface(t *testing.T) {
	var _ FormatStrategy = (*JSONFormatter)(nil)
}

// TestJSONFormatterZeroValue tests zero value behavior
func TestJSONFormatterZeroValue(t *testing.T) {
	var formatter JSONFormatter

	testData := []map[string]interface{}{
		{"id": 1, "name": "test"},
	}

	result, err := formatter.Format(testData)

	if err != nil {
		t.Errorf("Zero value Format() returned error: %v", err)
	}

	if result == nil {
		t.Error("Zero value Format() returned nil")
	}

	// Verify it's valid JSON
	var parsed []map[string]interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("Zero value Format() did not produce valid JSON: %v", err)
	}
}

// TestJSONFormatterConcurrentAccess tests concurrent formatting
func TestJSONFormatterConcurrentAccess(t *testing.T) {
	formatter := NewJSONFormatter()

	testData := []map[string]interface{}{
		{"id": 1, "name": "test"},
	}

	const goroutines = 10
	errors := make(chan error, goroutines)
	results := make(chan []byte, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			result, err := formatter.Format(testData)
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}()
	}

	for i := 0; i < goroutines; i++ {
		select {
		case err := <-errors:
			t.Errorf("Concurrent Format() failed: %v", err)
		case result := <-results:
			if result == nil {
				t.Error("Concurrent Format() returned nil")
			}
		}
	}
}

// TestJSONFormatterOutputConsistency tests that multiple calls produce same output
func TestJSONFormatterOutputConsistency(t *testing.T) {
	formatter := NewJSONFormatter()

	testData := []map[string]interface{}{
		{"id": 1, "name": "Alice", "score": 95},
	}

	result1, err1 := formatter.Format(testData)
	if err1 != nil {
		t.Fatalf("First Format() failed: %v", err1)
	}

	result2, err2 := formatter.Format(testData)
	if err2 != nil {
		t.Fatalf("Second Format() failed: %v", err2)
	}

	// Results should be identical
	if string(result1) != string(result2) {
		t.Error("Multiple Format() calls produced different output")
	}
}

// BenchmarkJSONFormatterFormat benchmarks basic formatting
func BenchmarkJSONFormatterFormat(b *testing.B) {
	formatter := NewJSONFormatter()

	testData := []map[string]interface{}{
		{"id": 1, "name": "Alice", "score": 95},
		{"id": 2, "name": "Bob", "score": 88},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatter.Format(testData)
	}
}

// BenchmarkJSONFormatterFormatLarge benchmarks large dataset
func BenchmarkJSONFormatterFormatLarge(b *testing.B) {
	formatter := NewJSONFormatter()

	testData := make([]map[string]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		testData[i] = map[string]interface{}{
			"id":    i,
			"name":  "User",
			"value": i * 2,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatter.Format(testData)
	}
}

// BenchmarkJSONFormatterFormatNested benchmarks nested structures
func BenchmarkJSONFormatterFormatNested(b *testing.B) {
	formatter := NewJSONFormatter()

	testData := []map[string]interface{}{
		{
			"id": 1,
			"user": map[string]interface{}{
				"name": "Alice",
				"address": map[string]interface{}{
					"street": "123 Main St",
					"city":   "Springfield",
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatter.Format(testData)
	}
}

// BenchmarkNewJSONFormatter benchmarks formatter creation
func BenchmarkNewJSONFormatter(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewJSONFormatter()
	}
}
