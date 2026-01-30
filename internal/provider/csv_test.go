package provider

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCSVProvider_Configure(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]string
		wantErr bool
	}{
		{
			name:    "valid configuration",
			params:  map[string]string{"file_path": "test.csv", "delimiter": ",", "has_header": "true"},
			wantErr: false,
		},
		{
			name:    "missing file path",
			params:  map[string]string{"delimiter": ","},
			wantErr: true,
		},
		{
			name:    "invalid delimiter",
			params:  map[string]string{"file_path": "test.csv", "delimiter": "xx"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewCSVProvider()
			if err := p.Configure(tt.params); (err != nil) != tt.wantErr {
				t.Errorf("CSVProvider.Configure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCSVProvider_Fetch(t *testing.T) {
	// Create a temporary CSV file
	content := `name,age,city
Alice,30,New York
Bob,25,San Francisco
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.csv")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	tests := []struct {
		name      string
		params    map[string]string
		wantRows  int
		checkRow0 func(map[string]interface{}) bool
		wantErr   bool
	}{
		{
			name:     "standard csv with header",
			params:   map[string]string{"file_path": tmpFile},
			wantRows: 2,
			checkRow0: func(row map[string]interface{}) bool {
				return row["name"] == "Alice" && row["age"] == "30"
			},
			wantErr: false,
		},
		{
			name:     "non-existent file",
			params:   map[string]string{"file_path": "non-existent.csv"},
			wantRows: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewCSVProvider()
			if err := p.Configure(tt.params); err != nil && !tt.wantErr {
				t.Fatalf("Configure failed: %v", err)
			} else if err != nil && tt.wantErr {
				return // Expected error in configure if any (though here we test fetch)
			}

			// Some tests expect Configure to pass but Fetch to fail
			result, err := p.Fetch(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Fetch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(result) != tt.wantRows {
					t.Errorf("Fetch() got %d rows, want %d", len(result), tt.wantRows)
				}
				if len(result) > 0 && tt.checkRow0 != nil {
					if !tt.checkRow0(result[0]) {
						t.Errorf("Row 0 content mismatch: %v", result[0])
					}
				}
			}
		})
	}
}

func TestCSVProvider_Fetch_NoHeader(t *testing.T) {
	content := `Alice,30
Bob,25`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "noheader.csv")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	p := NewCSVProvider()
	_ = p.Configure(map[string]string{
		"file_path":  tmpFile,
		"has_header": "false",
	})

	result, err := p.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(result))
	}

	// Default headers should be col_1, col_2
	firstRow := result[0]
	if firstRow["col_1"] != "Alice" || firstRow["col_2"] != "30" {
		t.Errorf("Unexpected row content for no-header CSV: %v", firstRow)
	}
}
