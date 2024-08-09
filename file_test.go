package doraemon

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestLazyFileWriter(t *testing.T) {
	// Setup a temporary directory for testing
	tempDir := os.TempDir()
	if tempDir == "" || !DirIsExist(tempDir).IsTrue() {
		t.Fatalf("Failed to get temp dir")
	}

	// Define the file path
	filePath := filepath.Join(tempDir, "testfile.txt")
	_ = os.Remove(filePath)

	// Create a new LazyFileWriter
	writer := NewLazyFileWriter(filePath)

	// Test writing to the file
	data := []byte("Hello, World!")
	n, err := writer.Write(data)
	if err != nil {
		t.Fatalf("Failed to write to file: %v", err)
	}
	if n != len(data) {
		t.Fatalf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	// Test if file is created
	if !writer.IsCreated() {
		t.Fatal("File should be created after writing")
	}

	// Test syncing the file
	if err := writer.Sync(); err != nil {
		t.Fatalf("Failed to sync file: %v", err)
	}

	// Test closing the file
	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close file: %v", err)
	}

	// Test writing to a closed file
	_, err = writer.Write(data)
	if err != os.ErrClosed {
		t.Fatalf("Expected error %v, got %v", os.ErrClosed, err)
	}

	// Test closing an already closed file
	err = writer.Close()
	if err != os.ErrClosed {
		t.Fatalf("Expected error %v, got %v", os.ErrClosed, err)
	}

	// Test file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != string(data) {
		t.Fatalf("Expected file content %s, got %s", data, content)
	}

	// Test file name and path
	if writer.Name() != "testfile.txt" {
		t.Fatalf("Expected file name testfile.txt, got %s", writer.Name())
	}
	if writer.Path() != filePath {
		t.Fatalf("Expected file path %s, got %s", filePath, writer.Path())
	}
}

func TestLazyFileWriterConcurrency(t *testing.T) {
	// Setup a temporary directory for testing
	tempDir := os.TempDir()
	if tempDir == "" || !DirIsExist(tempDir).IsTrue() {
		t.Fatalf("Failed to get temp dir")
	}

	// Define the file path
	filePath := filepath.Join(tempDir, "concurrentfile.txt")
	_ = os.Remove(filePath)

	// Create a new LazyFileWriter
	writer := NewLazyFileWriter(filePath)

	// Test concurrent writes
	var wg sync.WaitGroup
	data := []byte("Concurrent Write\n")
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := writer.Write(data)
			if err != nil {
				t.Errorf("Failed to write to file: %v", err)
			}
		}()
	}
	wg.Wait()
	err := writer.Sync()
	if err != nil {
		t.Fatalf("Failed to sync file: %v", err)
	}

	// Test file content length
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	expectedLength := len(data) * 10
	if len(content) != expectedLength {
		t.Fatalf("Expected file content length %d, got %d", expectedLength, len(content))
	}

	// Close the file
	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close file: %v", err)
	}
}
