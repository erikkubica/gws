package chat

import (
	"testing"
)

func TestSpaceMatchesQuery(t *testing.T) {
	sp := &SpaceInfo{
		Name:        "spaces/AAAA4_P5byY",
		DisplayName: "GALTON/DEV",
		Type:        "SPACE",
		Members:     []string{"Erik Kubica", "Jan Horecny", "Lucia Závadská"},
	}

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{
			name:     "match display name case insensitive",
			query:    "galton/dev",
			expected: true,
		},
		{
			name:     "match partial display name",
			query:    "dev",
			expected: true,
		},
		{
			name:     "match space ID",
			query:    "AAAA4_P5byY",
			expected: true,
		},
		{
			name:     "match member name nominal",
			query:    "lucia",
			expected: true,
		},
		{
			name:     "non matching query boundary",
			query:    "finance",
			expected: false,
		},
		{
			name:     "empty query boundary",
			query:    "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := spaceMatchesQuery(sp, tt.query)
			if got != tt.expected {
				t.Errorf("spaceMatchesQuery(%+v, %q) = %v, expected %v", sp, tt.query, got, tt.expected)
			}
		})
	}
}
