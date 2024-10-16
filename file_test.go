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

func TestGenerateUniqueFilepath(t *testing.T) {
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testCases := []struct {
		name     string
		setup    func(string) string
		input    string
		expected string
	}{
		{
			name:     "New file",
			setup:    func(dir string) string { return filepath.Join(dir, "test.txt") },
			input:    "test.txt",
			expected: "test.txt",
		},
		{
			name: "Existing file",
			setup: func(dir string) string {
				path := filepath.Join(dir, "test.txt")
				os.Create(path)
				return path
			},
			input:    "test.txt",
			expected: "test(1).txt",
		},
		{
			name: "Multiple existing files",
			setup: func(dir string) string {
				os.Create(filepath.Join(dir, "test.txt"))
				os.Create(filepath.Join(dir, "test(1).txt"))
				return filepath.Join(dir, "test.txt")
			},
			input:    "test.txt",
			expected: "test(2).txt",
		},
		{
			name:     "File without extension",
			setup:    func(dir string) string { return filepath.Join(dir, "test") },
			input:    "test",
			expected: "test",
		},
		{
			name: "Existing file without extension",
			setup: func(dir string) string {
				path := filepath.Join(dir, "test")
				os.Create(path)
				return path
			},
			input:    "test",
			expected: "test(1)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filePath := tc.setup(tempDir)
			result := GenerateUniqueFilepath(filePath)
			expected := filepath.Join(tempDir, tc.expected)
			if result != expected {
				t.Errorf("Expected %s, got %s", expected, result)
			}
		})
	}
}
