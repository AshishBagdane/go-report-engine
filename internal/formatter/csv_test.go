package formatter

import (
	"context"
	"strings"
	"testing"
)

func TestCSVFormatter_Format(t *testing.T) {
	data := []map[string]interface{}{
		{"name": "Alice", "age": 30, "city": "NYC"},
		{"name": "Bob", "age": 25, "city": "SF"},
	}

	tests := []struct {
		name          string
		delimiter     rune
		includeHeader bool
		wantContains  []string
	}{
		{
			name:          "standard csv",
			delimiter:     ',',
			includeHeader: true,
			wantContains:  []string{"age,city,name", "30,NYC,Alice", "25,SF,Bob"},
		},
		{
			name:          "pipe delimiter",
			delimiter:     '|',
			includeHeader: true,
			wantContains:  []string{"age|city|name", "30|NYC|Alice"},
		},
		{
			name:          "no header",
			delimiter:     ',',
			includeHeader: false,
			wantContains:  []string{"30,NYC,Alice"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewCSVFormatter()
			f.Delimiter = tt.delimiter
			f.IncludeHeader = tt.includeHeader

			got, err := f.Format(context.Background(), data)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}

			str := string(got)
			for _, want := range tt.wantContains {
				if !strings.Contains(str, want) {
					t.Errorf("Format() expected to contain %q, but got:\n%s", want, str)
				}
			}

			if !tt.includeHeader && strings.Contains(str, "age") {
				// Check strictly that header is not present if disabled
				// Note: this assumes data values don't contain key names
				t.Errorf("Format() contained header when it should not")
			}
		})
	}
}

func TestCSVFormatter_Configure(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]string
		wantErr bool
	}{
		{
			name:    "valid config",
			params:  map[string]string{"delimiter": ";", "include_header": "false"},
			wantErr: false,
		},
		{
			name:    "invalid delimiter",
			params:  map[string]string{"delimiter": "xx"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewCSVFormatter()
			if err := f.Configure(tt.params); (err != nil) != tt.wantErr {
				t.Errorf("CSVFormatter.Configure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCSVFormatter_EmptyData(t *testing.T) {
	f := NewCSVFormatter()
	got, err := f.Format(context.Background(), []map[string]interface{}{})
	if err != nil {
		t.Errorf("Format() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Expected empty output for empty data, got: %s", string(got))
	}
}
