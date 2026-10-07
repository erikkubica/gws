package chat

import (
	"testing"
)

func TestNormalizeSpaceName(t *testing.T) {
	t.Run("nominal_cases", func(t *testing.T) {
		cases := []struct {
			input    string
			expected string
		}{
			{"spaces/AAAABBBB", "spaces/AAAABBBB"},
			{"AAAABBBB", "spaces/AAAABBBB"},
			{"  spaces/XYZ123  ", "spaces/XYZ123"},
			{"  ROOM_456  ", "spaces/ROOM_456"},
		}
		for _, tc := range cases {
			got := NormalizeSpaceName(tc.input)
			if got != tc.expected {
				t.Errorf("NormalizeSpaceName(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		}
	})

	t.Run("boundary_thresholds", func(t *testing.T) {
		got := NormalizeSpaceName("x")
		if got != "spaces/x" {
			t.Errorf("expected 'spaces/x', got %q", got)
		}
	})

	t.Run("malformed_inputs", func(t *testing.T) {
		malformed := []string{
			"",
			"   ",
			"\t\n",
		}
		for _, input := range malformed {
			got := NormalizeSpaceName(input)
			if got != "" {
				t.Errorf("expected empty string for %q, got %q", input, got)
			}
		}
	})
}
