package drive

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read failure")
}

func TestReadLimitedStream(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		maxBytes  int64
		expectErr bool
	}{
		{
			name:      "nominal under limit",
			content:   "hello drive",
			maxBytes:  1024,
			expectErr: false,
		},
		{
			name:      "exact boundary",
			content:   "12345",
			maxBytes:  5,
			expectErr: false,
		},
		{
			name:      "exceeds limit",
			content:   "123456",
			maxBytes:  5,
			expectErr: true,
		},
		{
			name:      "empty stream boundary",
			content:   "",
			maxBytes:  10,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := readLimitedStream(strings.NewReader(tt.content), tt.maxBytes)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error for content length %d with max %d, got nil", len(tt.content), tt.maxBytes)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res != tt.content {
				t.Fatalf("expected content %q, got %q", tt.content, res)
			}
		})
	}
}

func TestReadLimitedStream_ErrorReader(t *testing.T) {
	_, err := readLimitedStream(&errorReader{}, 1024)
	if err == nil {
		t.Fatal("expected error on failed stream read, got nil")
	}
}

func TestWriteStreamToFile(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "testfile.txt")
	content := "atomic drive download content"

	err := writeStreamToFile(targetPath, strings.NewReader(content))
	if err != nil {
		t.Fatalf("writeStreamToFile failed: %v", err)
	}

	data, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read destination file: %v", err)
	}
	if string(data) != content {
		t.Fatalf("expected file content %q, got %q", content, string(data))
	}

	// Verify temp file does not remain
	if _, err := os.Stat(targetPath + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("temp file %s.tmp was not cleaned up", targetPath)
	}
}

func TestWriteStreamToFile_FailureCleansUpTemp(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "failed_file.txt")

	err := writeStreamToFile(targetPath, &errorReader{})
	if err == nil {
		t.Fatal("expected writeStreamToFile to fail, got nil")
	}

	// Verify target file was not created
	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		t.Errorf("target file should not exist on failure: %v", err)
	}

	// Verify temp file was cleaned up
	if _, err := os.Stat(targetPath + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("temp file should be deleted on failure: %v", err)
	}
}

func TestBuildDriveQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "nominal query",
			input:    "annual_report",
			expected: "name contains 'annual_report' and trashed = false",
		},
		{
			name:     "empty query boundary",
			input:    "",
			expected: "trashed = false",
		},
		{
			name:     "whitespace query boundary",
			input:    "   ",
			expected: "trashed = false",
		},
		{
			name:     "single quote escaping injection prevention",
			input:    "John's doc's",
			expected: "name contains 'John\\'s doc\\'s' and trashed = false",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildDriveQuery(tc.input)
			if got != tc.expected {
				t.Fatalf("expected query %q, got %q", tc.expected, got)
			}
		})
	}
}
