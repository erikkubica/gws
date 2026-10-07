package cli

import (
	"testing"
)

func TestMatchesExcludeList(t *testing.T) {
	excludes := []string{"spaces/AAQADfwHDeE", "WP Plugin", "Dev-Spam"}

	tests := []struct {
		name        string
		spaceID     string
		displayName string
		expected    bool
	}{
		// Nominal cases
		{"match by exact space ID", "spaces/AAQADfwHDeE", "Support", true},
		{"match by partial name", "spaces/XYZ", "GALTON SUPPORT - WP Plugin", true},
		{"match case insensitive", "spaces/XYZ", "wp plugin alerts", true},
		{"non-matching space", "spaces/AAAA4_P5byY", "GALTON/DEV", false},

		// Boundary thresholds
		{"empty space and name", "", "", false},
		{"empty exclude list", "spaces/XYZ", "Any Name", false},

		// Malformed inputs
		{"whitespace in spaceID", "  spaces/AAQADfwHDeE  ", "Support", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			list := excludes
			if tc.name == "empty exclude list" {
				list = []string{}
			}
			result := matchesExcludeList(tc.spaceID, tc.displayName, list)
			if result != tc.expected {
				t.Errorf("matchesExcludeList(%q, %q) = %v, want %v", tc.spaceID, tc.displayName, result, tc.expected)
			}
		})
	}
}
