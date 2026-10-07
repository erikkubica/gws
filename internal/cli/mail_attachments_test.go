package cli

import (
	"path/filepath"
	"testing"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{name: "zero boundary", input: 0, expected: "0 B"},
		{name: "negative boundary", input: -10, expected: "0 B"},
		{name: "bytes nominal", input: 512, expected: "512 B"},
		{name: "1023 bytes threshold", input: 1023, expected: "1023 B"},
		{name: "exact 1 KB threshold", input: 1024, expected: "1.0 KB"},
		{name: "kilobytes nominal", input: 1536, expected: "1.5 KB"},
		{name: "exact 1 MB threshold", input: 1048576, expected: "1.0 MB"},
		{name: "megabytes nominal", input: 2621440, expected: "2.5 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBytes(tt.input)
			if got != tt.expected {
				t.Errorf("formatBytes(%d) = %s, expected %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestResolveOutputPath(t *testing.T) {
	tests := []struct {
		name     string
		outFile  string
		outDir   string
		filename string
		expected string
	}{
		{
			name:     "explicit output file overrides everything",
			outFile:  "/custom/path/doc.pdf",
			outDir:   "/tmp",
			filename: "original.pdf",
			expected: "/custom/path/doc.pdf",
		},
		{
			name:     "custom directory with simple filename",
			outFile:  "",
			outDir:   "/downloads",
			filename: "report.pdf",
			expected: filepath.Join("/downloads", "report.pdf"),
		},
		{
			name:     "directory defense against path traversal attempt",
			outFile:  "",
			outDir:   "/safe/dir",
			filename: "../../../../etc/passwd",
			expected: filepath.Join("/safe/dir", "passwd"),
		},
		{
			name:     "no out and no dir defaults to base filename",
			outFile:  "",
			outDir:   "",
			filename: "notes.txt",
			expected: "notes.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveOutputPath(tt.outFile, tt.outDir, tt.filename)
			if got != tt.expected {
				t.Errorf("resolveOutputPath(%q, %q, %q) = %q, expected %q",
					tt.outFile, tt.outDir, tt.filename, got, tt.expected)
			}
		})
	}
}
