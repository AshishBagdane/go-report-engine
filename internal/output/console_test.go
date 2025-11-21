package output

import (
	"bytes"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
)

// captureStdout captures stdout output during test execution
// Returns the captured output as a string
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a channel to signal completion
	done := make(chan struct{})
	var buf bytes.Buffer

	// Read in a goroutine to prevent blocking
	go func() {
		io.Copy(&buf, r)
		close(done)
	}()

	// Execute the function
	f()

	// Close writer and wait for reader
	w.Close()
	<-done
	os.Stdout = old

	return buf.String()
}

// TestNewConsoleOutput tests the factory function
func TestNewConsoleOutput(t *testing.T) {
	output := NewConsoleOutput()

	if output == nil {
		t.Fatal("NewConsoleOutput() returned nil")
	}

	// Verify it implements OutputStrategy
	var _ OutputStrategy = output
}

// TestConsoleOutputSend tests basic send functionality
func TestConsoleOutputSend(t *testing.T) {
	output := NewConsoleOutput()
	testData := []byte("test output")

	captured := captureStdout(func() {
		output.Send(testData)
	})

	expectedOutput := "test output\n"
	if captured != expectedOutput {
		t.Errorf("Send() output = %q, expected %q", captured, expectedOutput)
	}
}

// TestConsoleOutputSendEmpty tests sending empty data
func TestConsoleOutputSendEmpty(t *testing.T) {
	output := NewConsoleOutput()
	testData := []byte("")

	captured := captureStdout(func() {
		output.Send(testData)
	})

	expectedOutput := "\n"
	if captured != expectedOutput {
		t.Errorf("Send() output = %q, expected %q", captured, expectedOutput)
	}
}

// TestConsoleOutputSendNil tests sending nil data
func TestConsoleOutputSendNil(t *testing.T) {
	output := NewConsoleOutput()

	captured := captureStdout(func() {
		output.Send(nil)
	})

	expectedOutput := "\n"
	if captured != expectedOutput {
		t.Errorf("Send(nil) output = %q, expected %q", captured, expectedOutput)
	}
}

// TestConsoleOutputSendJSON tests sending JSON data
func TestConsoleOutputSendJSON(t *testing.T) {
	output := NewConsoleOutput()
	jsonData := []byte(`{"id":1,"name":"Alice","score":95}`)

	captured := captureStdout(func() {
		output.Send(jsonData)
	})

	if !strings.Contains(captured, `"id":1`) {
		t.Error("JSON data not found in output")
	}
}

// TestConsoleOutputSendMultiline tests sending multiline data
func TestConsoleOutputSendMultiline(t *testing.T) {
	output := NewConsoleOutput()
	multilineData := []byte("line1\nline2\nline3")

	captured := captureStdout(func() {
		output.Send(multilineData)
	})

	if !strings.Contains(captured, "line1") || !strings.Contains(captured, "line2") {
		t.Error("Multiline data not preserved in output")
	}
}

// TestConsoleOutputSendLargeData tests sending large data
func TestConsoleOutputSendLargeData(t *testing.T) {
	output := NewConsoleOutput()

	// Create large data (100KB instead of 1MB for faster tests)
	largeData := bytes.Repeat([]byte("x"), 100*1024)

	captured := captureStdout(func() {
		output.Send(largeData)
	})

	// Verify size (should be original + newline)
	expectedSize := len(largeData) + 1
	if len(captured) != expectedSize {
		t.Errorf("Output size = %d, expected %d", len(captured), expectedSize)
	}
}

// TestConsoleOutputSendSpecialCharacters tests special characters
func TestConsoleOutputSendSpecialCharacters(t *testing.T) {
	output := NewConsoleOutput()
	specialData := []byte("Hello\tWorld\r\n")

	captured := captureStdout(func() {
		output.Send(specialData)
	})

	if !strings.Contains(captured, "Hello") || !strings.Contains(captured, "World") {
		t.Error("Special characters affected output")
	}
}

// TestConsoleOutputSendUnicode tests Unicode handling
func TestConsoleOutputSendUnicode(t *testing.T) {
	output := NewConsoleOutput()
	unicodeData := []byte("Hello, 世界! 🎉")

	captured := captureStdout(func() {
		output.Send(unicodeData)
	})

	if !strings.Contains(captured, "世界") || !strings.Contains(captured, "🎉") {
		t.Error("Unicode not preserved in output")
	}
}

// TestConsoleOutputMultipleSends tests multiple send operations
func TestConsoleOutputMultipleSends(t *testing.T) {
	output := NewConsoleOutput()

	sends := [][]byte{
		[]byte("first"),
		[]byte("second"),
		[]byte("third"),
	}

	captured := captureStdout(func() {
		for _, data := range sends {
			output.Send(data)
		}
	})

	// Verify all sends are present
	for _, data := range sends {
		if !strings.Contains(captured, string(data)) {
			t.Errorf("Output missing %q", string(data))
		}
	}
}

// TestConsoleOutputImplementsInterface verifies interface implementation
func TestConsoleOutputImplementsInterface(t *testing.T) {
	var _ OutputStrategy = (*ConsoleOutput)(nil)
}

// TestConsoleOutputZeroValue tests zero value behavior
func TestConsoleOutputZeroValue(t *testing.T) {
	var output ConsoleOutput
	testData := []byte("test")

	captured := captureStdout(func() {
		output.Send(testData)
	})

	if !strings.Contains(captured, "test") {
		t.Error("Zero value Send() did not output data")
	}
}

// TestConsoleOutputConcurrentSends tests concurrent send operations
func TestConsoleOutputConcurrentSends(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	output := NewConsoleOutput()

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Suppress output for this test to avoid clutter
	old := os.Stdout
	os.Stdout = nil

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			data := []byte(strings.Repeat("x", 100))
			err := output.Send(data)
			if err != nil {
				t.Errorf("Concurrent Send() failed: %v", err)
			}
		}(i)
	}

	wg.Wait()
	os.Stdout = old
}

// BenchmarkConsoleOutputSend benchmarks basic send
func BenchmarkConsoleOutputSend(b *testing.B) {
	output := NewConsoleOutput()
	testData := []byte("benchmark test data")

	// Suppress output for benchmark
	old := os.Stdout
	os.Stdout = nil

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output.Send(testData)
	}

	os.Stdout = old
}

// BenchmarkConsoleOutputSendLarge benchmarks large data
func BenchmarkConsoleOutputSendLarge(b *testing.B) {
	output := NewConsoleOutput()
	testData := bytes.Repeat([]byte("x"), 10000)

	// Suppress output for benchmark
	old := os.Stdout
	os.Stdout = nil

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output.Send(testData)
	}

	os.Stdout = old
}

// BenchmarkNewConsoleOutput benchmarks output creation
func BenchmarkNewConsoleOutput(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewConsoleOutput()
	}
}
