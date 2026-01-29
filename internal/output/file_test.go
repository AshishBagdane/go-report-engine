package output

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFileOutput_Send(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		path    string
		data    []byte
		mode    string
		wantErr bool
	}{
		{
			name:    "simple write",
			path:    filepath.Join(tmpDir, "output.txt"),
			data:    []byte("hello world"),
			wantErr: false,
		},
		{
			name:    "missing path",
			path:    "",
			data:    []byte("data"),
			wantErr: true,
		},
		{
			name:    "nested directory",
			path:    filepath.Join(tmpDir, "subdir", "output.txt"),
			data:    []byte("hello subdir"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFileOutput()
			params := map[string]string{}
			if tt.path != "" {
				params["path"] = tt.path
			}
			if tt.mode != "" {
				params["mode"] = tt.mode
			}

			// Configure manually if path is missing to triggers Configure error or Send error?
			// Send checks Path field. Configure sets it.
			// Let's set fields directly or use Configure.
			// For validation of Send, we can set fields directly.
			if tt.path != "" {
				f.Path = tt.path
			}

			err := f.Send(context.Background(), tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileOutput.Send() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify file content
				content, err := os.ReadFile(tt.path)
				if err != nil {
					t.Fatalf("Failed to read output file: %v", err)
				}
				if string(content) != string(tt.data) {
					t.Errorf("File content = %q, want %q", content, tt.data)
				}
			}
		})
	}
}

func TestFileOutput_Configure(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]string
		wantErr bool
	}{
		{
			name:    "valid config",
			params:  map[string]string{"path": "/tmp/out.txt", "mode": "0600"},
			wantErr: false,
		},
		{
			name:    "missing path",
			params:  map[string]string{"mode": "0600"},
			wantErr: true,
		},
		{
			name:    "invalid mode",
			params:  map[string]string{"path": "out.txt", "mode": "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFileOutput()
			if err := f.Configure(tt.params); (err != nil) != tt.wantErr {
				t.Errorf("FileOutput.Configure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
