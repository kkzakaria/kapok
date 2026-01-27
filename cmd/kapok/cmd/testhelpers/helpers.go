package testhelpers

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// CreateTempDir creates a temporary directory for testing
// The directory is automatically cleaned up when the test finishes
func CreateTempDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// CaptureOutput captures stdout and stderr during function execution
func CaptureOutput(fn func()) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	fn()

	w.Close()
	os.Stdout = old
	out := <-outC

	return out, nil
}

// CaptureOutputWithError captures both output and errors
func CaptureOutputWithError(fn func() error) (string, error) {
	var buf bytes.Buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	outC := make(chan string)
	go func() {
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	err := fn()

	w.Close()
	os.Stdout = old
	out := <-outC

	return out, err
}

// WriteTestFile writes content to a file in a test directory
func WriteTestFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	
	path := dir + "/" + filename
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	return path
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
