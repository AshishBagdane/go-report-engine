package output

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// FileOutput implements OutputStrategy for writing to the filesystem.
type FileOutput struct {
	Path string
	Mode os.FileMode
	file *os.File
}

// NewFileOutput creates a new instance of FileOutput with defaults.
func NewFileOutput() *FileOutput {
	return &FileOutput{
		Mode: 0644,
	}
}

// Send writes the data to the configured file.
func (f *FileOutput) Send(ctx context.Context, data []byte) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if f.Path == "" {
		return fmt.Errorf("file output: path not configured")
	}

	// Ensure directory exists
	dir := filepath.Dir(f.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("file output: failed to create directory %s: %w", dir, err)
	}

	// Open file (Create = truncate if exists, create if not)
	// We could add an append option later if needed.
	file, err := os.OpenFile(f.Path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode)
	if err != nil {
		return fmt.Errorf("file output: failed to open file %s: %w", f.Path, err)
	}
	defer file.Close()

	// Write data
	// Note: output.go interface says "Send(ctx, data)".
	// Ideally we would write in chunks and check context, but for a single []byte,
	// checking before writing is usually sufficient unless it's huge.
	// For HUGE data we might rely on the OS or file system, which isn't easily interruptible
	// in strict Go "Write" call without specialized handling.
	// Simple write is fine for now.

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("file output: failed to write to file: %w", err)
	}

	return nil
}

// Configure sets up the output from a map of parameters.
// Params:
// - path: Output file path (required)
// - mode: Octal file permission (default: "0644")
func (f *FileOutput) Configure(params map[string]string) error {
	if path, ok := params["path"]; ok {
		f.Path = path
	} else {
		return fmt.Errorf("file output: missing required parameter 'path'")
	}

	if modeStr, ok := params["mode"]; ok {
		// Parse octal string
		mode, err := strconv.ParseUint(modeStr, 8, 32)
		if err != nil {
			return fmt.Errorf("file output: invalid mode %s: %w", modeStr, err)
		}
		f.Mode = os.FileMode(mode)
	}

	return nil
}

// Initialize prepares the output for streaming (opens the file).
func (f *FileOutput) Initialize(ctx context.Context) error {
	if f.Path == "" {
		return fmt.Errorf("file output: path not configured")
	}

	// Ensure directory exists
	dir := filepath.Dir(f.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("file output: failed to create directory %s: %w", dir, err)
	}

	// Open file (truncate)
	file, err := os.OpenFile(f.Path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode)
	if err != nil {
		return fmt.Errorf("file output: failed to open file %s: %w", f.Path, err)
	}
	f.file = file
	return nil
}

// WriteChunk writes a chunk of data to the open file.
func (f *FileOutput) WriteChunk(ctx context.Context, data []byte) error {
	if f.file == nil {
		return fmt.Errorf("file output: file not initialized")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if _, err := f.file.Write(data); err != nil {
		return fmt.Errorf("file output: failed to write chunk: %w", err)
	}
	return nil
}

// Close finalizes the output stream.
func (f *FileOutput) Close(ctx context.Context) error {
	if f.file != nil {
		err := f.file.Close()
		f.file = nil
		if err != nil {
			return fmt.Errorf("file output: failed to close file: %w", err)
		}
	}
	return nil
}
