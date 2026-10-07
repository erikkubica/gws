package tasks

import (
	"testing"
)

func TestNormalizeDueDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Nominal cases
		{"nominal RFC3339", "2026-10-08T15:30:00Z", "2026-10-08T15:30:00Z"},
		{"nominal YYYY-MM-DD", "2026-10-08", "2026-10-08T00:00:00Z"},
		{"nominal YYYY-MM-DD with whitespace", "  2026-10-08  ", "2026-10-08T00:00:00Z"},

		// Boundary thresholds
		{"empty string", "", ""},
		{"only whitespace", "   ", ""},

		// Malformed inputs
		{"invalid date format", "tomorrow morning", "tomorrow morning"},
		{"invalid day 99", "2026-10-99", "2026-10-99"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := NormalizeDueDate(tc.input)
			if result != tc.expected {
				t.Errorf("NormalizeDueDate(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}
