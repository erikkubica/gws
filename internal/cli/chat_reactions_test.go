package cli

import (
	"strings"
	"testing"

	"github.com/erikkubica/gws/internal/services/chat"
)

func TestFormatReactionsString(t *testing.T) {
	tests := []struct {
		name     string
		input    []chat.ReactionSummary
		expected string
	}{
		{
			name:     "empty reactions boundary",
			input:    nil,
			expected: "",
		},
		{
			name: "nominal case with users and display names",
			input: []chat.ReactionSummary{
				{
					Emoji: "👍",
					Count: 2,
					Users: []chat.ReactingUser{
						{DisplayName: "Jan Horecny", Name: "users/109815"},
						{DisplayName: "Lucia Závadská", Name: "users/104231"},
					},
				},
				{
					Emoji: "❤️",
					Count: 1,
					Users: []chat.ReactingUser{
						{DisplayName: "Erik Kubica", Name: "users/100001"},
					},
				},
			},
			expected: "  Reactions: 👍 2 (Jan Horecny, Lucia Závadská)  ❤️ 1 (Erik Kubica)\n",
		},
		{
			name: "fallback to user resource name when display name is empty",
			input: []chat.ReactionSummary{
				{
					Emoji: "🎉",
					Count: 1,
					Users: []chat.ReactingUser{
						{DisplayName: "", Name: "users/999888"},
					},
				},
			},
			expected: "  Reactions: 🎉 1 (users/999888)\n",
		},
		{
			name: "count only when users list is empty boundary",
			input: []chat.ReactionSummary{
				{
					Emoji: "🚀",
					Count: 3,
					Users: nil,
				},
			},
			expected: "  Reactions: 🚀 3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatReactionsString(tt.input)
			if got != tt.expected {
				t.Errorf("formatReactionsString() = %q, expected %q", got, tt.expected)
			}
		})
	}
}

func TestFormatAttachmentsString(t *testing.T) {
	tests := []struct {
		name     string
		input    []chat.AttachmentInfo
		expected string
	}{
		{
			name:     "nil attachments boundary",
			input:    nil,
			expected: "",
		},
		{
			name: "nominal attachments with download url",
			input: []chat.AttachmentInfo{
				{
					ContentName: "spec.pdf",
					DownloadURL: "https://chat.googleapis.com/download/spec.pdf",
				},
			},
			expected: "  📎 spec.pdf (https://chat.googleapis.com/download/spec.pdf)\n",
		},
		{
			name: "fallback when content name is empty",
			input: []chat.AttachmentInfo{
				{
					ContentName: "",
					DownloadURL: "",
				},
			},
			expected: "  📎 attachment\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatAttachmentsString(tt.input)
			if !strings.Contains(got, tt.expected) {
				t.Errorf("formatAttachmentsString() = %q, expected to contain %q", got, tt.expected)
			}
		})
	}
}
