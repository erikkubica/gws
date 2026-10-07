package cli

import (
	"strings"
	"testing"

	"github.com/erikkubica/gws/internal/services/chat"
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

func TestBuildWatchEnv(t *testing.T) {
	msgs := []*chat.MessageInfo{
		{
			Name:       "spaces/space123/messages/msg001",
			SenderName: "Alice",
			SenderID:   "users/alice1",
			ThreadName: "spaces/space123/threads/t1",
			Text:       "First message",
		},
		{
			Name:       "spaces/space123/messages/msg002",
			SenderName: "Bob",
			SenderID:   "users/bob2",
			ThreadName: "spaces/space123/threads/t1",
			Text:       "Second message",
		},
	}
	payload := []byte(`{"event":"burst"}`)
	env := buildWatchEnv("all", msgs, payload)

	envMap := make(map[string]string)
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	if envMap["GWS_SPACE_ID"] != "spaces/space123" {
		t.Errorf("expected GWS_SPACE_ID 'spaces/space123', got %q", envMap["GWS_SPACE_ID"])
	}
	if envMap["GWS_MESSAGE_COUNT"] != "2" {
		t.Errorf("expected GWS_MESSAGE_COUNT '2', got %q", envMap["GWS_MESSAGE_COUNT"])
	}
	if envMap["GWS_MESSAGE_ID"] != "spaces/space123/messages/msg002" {
		t.Errorf("expected GWS_MESSAGE_ID of latest message, got %q", envMap["GWS_MESSAGE_ID"])
	}
	if envMap["GWS_SENDER"] != "Bob" {
		t.Errorf("expected GWS_SENDER 'Bob', got %q", envMap["GWS_SENDER"])
	}
	if envMap["GWS_PAYLOAD"] != string(payload) {
		t.Errorf("expected GWS_PAYLOAD %q, got %q", string(payload), envMap["GWS_PAYLOAD"])
	}
}
