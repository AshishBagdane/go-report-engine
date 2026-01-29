package processor

import (
	"context"
	"testing"
)

func TestDeduplicateProcessor_Process(t *testing.T) {
	tests := []struct {
		name      string
		fields    []string
		input     []map[string]interface{}
		wantCount int
	}{
		{
			name:   "dedupe all fields",
			fields: nil,
			input: []map[string]interface{}{
				{"id": 1, "val": "a"},
				{"id": 1, "val": "a"},
				{"id": 2, "val": "b"},
			},
			wantCount: 2,
		},
		{
			name:   "dedupe specific field",
			fields: []string{"id"},
			input: []map[string]interface{}{
				{"id": 1, "val": "a"},
				{"id": 1, "val": "b"}, // Duplicate id
				{"id": 2, "val": "c"},
			},
			wantCount: 2,
		},
		{
			name:   "no duplicates",
			fields: nil,
			input: []map[string]interface{}{
				{"id": 1},
				{"id": 2},
			},
			wantCount: 2,
		},
		{
			name:   "missing field treated as null",
			fields: []string{"key"},
			input: []map[string]interface{}{
				{"other": 1}, // key missing
				{"other": 2}, // key missing -> same signature key:<missing>
			},
			wantCount: 1, // should dedupe on missing key collision
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewDeduplicateProcessor(tt.fields)
			got, err := p.Process(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("Process() error = %v", err)
			}
			if len(got) != tt.wantCount {
				t.Errorf("Process() count = %d, want %d", len(got), tt.wantCount)
			}
		})
	}
}
