package engine_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AshishBagdane/go-report-engine/internal/engine"
	"github.com/AshishBagdane/go-report-engine/internal/formatter"
	"github.com/AshishBagdane/go-report-engine/internal/output"
	"github.com/AshishBagdane/go-report-engine/internal/processor"
	"github.com/AshishBagdane/go-report-engine/internal/provider"
)

func TestStreamingPipeline_Integration(t *testing.T) {
	// 1. Setup: Create a large CSV file
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "large_input.csv")
	jsonPath := filepath.Join(tmpDir, "large_output.json")

	createLargeCSV(t, csvPath, 5000) // 5000 records

	// 2. Setup Components
	// Provider
	csvProv := provider.NewCSVProvider()
	err := csvProv.Configure(map[string]string{
		"file_path": csvPath,
	})
	if err != nil {
		t.Fatalf("Failed to configure CSV provider: %v", err)
	}

	// Processor (Pass-through)
	baseProc := &processor.BaseProcessor{}

	// Formatter (Buffered/Streaming JSON)
	jsonFormatter := formatter.NewJSONFormatter("") // Compact

	// Output (Streaming File)
	fileOut := output.NewFileOutput()
	err = fileOut.Configure(map[string]string{
		"path": jsonPath,
	})
	if err != nil {
		t.Fatalf("Failed to configure File output: %v", err)
	}

	// 3. Build Engine
	eng := &engine.ReportEngine{
		Provider:  csvProv,
		Processor: baseProc,
		Formatter: jsonFormatter,
		Output:    fileOut,
	}

	// Set valid chunk size
	eng.WithChunkSize(100)

	// 4. Run Pipeline
	ctx := context.Background()
	startTime := time.Now()
	if err := eng.RunWithContext(ctx); err != nil {
		t.Fatalf("Engine run failed: %v", err)
	}
	duration := time.Since(startTime)
	t.Logf("Streaming pipeline took %v", duration)

	// 5. Verify Output
	// Read back the JSON file
	bytes, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Verify it is valid JSON
	var results []map[string]interface{}
	if err := json.Unmarshal(bytes, &results); err != nil {
		t.Fatalf("Output is not valid JSON: %v. \nFirst 100 chars: %s", err, string(bytes[:min(100, len(bytes))]))
	}

	// Verify count
	if len(results) != 5000 {
		t.Errorf("Expected 5000 records, got %d", len(results))
	}

	// Verify content of first and last
	if results[0]["id"] != "1" {
		t.Errorf("First record ID mismatch. Got %v, want 1", results[0]["id"])
	}
	if results[4999]["id"] != "5000" {
		t.Errorf("Last record ID mismatch. Got %v, want 5000", results[4999]["id"])
	}
}

func createLargeCSV(t *testing.T, path string, records int) {
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create CSV: %v", err)
	}
	defer f.Close()

	// Header
	f.WriteString("id,name,value\n")

	// Records
	for i := 1; i <= records; i++ {
		line := fmt.Sprintf("%d,User%d,%d\n", i, i, i*10)
		f.WriteString(line)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
