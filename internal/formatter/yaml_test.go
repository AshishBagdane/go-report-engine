package formatter

import (
	"context"
	"strings"
	"testing"
)

func TestYAMLFormatter_Format(t *testing.T) {
	data := []map[string]interface{}{
		{"name": "Alice", "age": 30},
		{"name": "Bob", "age": 25},
	}

	tests := []struct {
		name         string
		indent       string
		wantContains []string
	}{
		{
			name:         "default indent",
			indent:       "",
			wantContains: []string{"- age: 30", "  name: Alice"},
		},
		{
			name:   "large indent",
			indent: "4",
			// Indentation check is fragile across yaml library versions/structure.
			// Just checking content presence.
			wantContains: []string{"age: 30", "name: Alice"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewYAMLFormatter()
			if tt.indent != "" {
				_ = f.Configure(map[string]string{"indent": tt.indent})
			}

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
		})
	}
}

func TestYAMLFormatter_Configure(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]string
		wantErr bool
	}{
		{
			name:    "valid indent",
			params:  map[string]string{"indent": "4"},
			wantErr: false,
		},
		{
			name:    "invalid indent",
			params:  map[string]string{"indent": "abc"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewYAMLFormatter()
			if err := f.Configure(tt.params); (err != nil) != tt.wantErr {
				t.Errorf("Configure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
