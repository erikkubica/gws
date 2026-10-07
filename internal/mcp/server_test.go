package mcp

import (
	"context"
	"testing"

	"github.com/erikkubica/gws/internal/version"
	"github.com/mark3labs/mcp-go/mcp"
)

func makeToolRequest(name string, args map[string]any) mcp.CallToolRequest {
	var req mcp.CallToolRequest
	req.Params.Name = name
	req.Params.Arguments = args
	return req
}

func TestNewServer(t *testing.T) {
	ctx := context.Background()
	s, err := NewServer(ctx)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	if s == nil || s.mcpServer == nil {
		t.Fatal("expected non-nil MCPServer")
	}
	if version.Version != "1.1.1" {
		t.Fatalf("expected version 1.1.1, got %q", version.Version)
	}
}

func TestParseEmailOptions_Validation(t *testing.T) {
	t.Run("nominal_case", func(t *testing.T) {
		req := makeToolRequest("mail_send_message", map[string]any{
			"to":      "test@example.com",
			"subject": "Greetings",
			"body":    "Hello World",
		})
		opts, err := parseEmailOptions(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opts.To != "test@example.com" || opts.Subject != "Greetings" || opts.Body != "Hello World" {
			t.Fatalf("unexpected opts: %+v", opts)
		}
	})

	t.Run("missing_required_arguments", func(t *testing.T) {
		req := makeToolRequest("mail_send_message", map[string]any{
			"to": "test@example.com",
		})
		_, err := parseEmailOptions(req)
		if err == nil {
			t.Fatal("expected error on missing subject and body, got nil")
		}
	})
}

func TestParseChatMessageOptions_Validation(t *testing.T) {
	t.Run("nominal_case", func(t *testing.T) {
		req := makeToolRequest("chat_send_message", map[string]any{
			"space": "spaces/xyz",
			"text":  "hello",
		})
		opts, err := parseChatMessageOptions(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opts.SpaceName != "spaces/xyz" || opts.Text != "hello" {
			t.Fatalf("unexpected opts: %+v", opts)
		}
	})

	t.Run("missing_space_error", func(t *testing.T) {
		req := makeToolRequest("chat_send_message", map[string]any{
			"text": "hello without space",
		})
		_, err := parseChatMessageOptions(req)
		if err == nil {
			t.Fatal("expected error on missing space parameter, got nil")
		}
	})
}

func TestParseEventInput_Validation(t *testing.T) {
	t.Run("nominal_case", func(t *testing.T) {
		req := makeToolRequest("calendar_create_event", map[string]any{
			"title": "Meeting",
			"start": "2026-10-07T10:00:00Z",
			"end":   "2026-10-07T11:00:00Z",
		})
		opts, err := parseEventInput(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opts.Title != "Meeting" || opts.Start != "2026-10-07T10:00:00Z" {
			t.Fatalf("unexpected opts: %+v", opts)
		}
	})

	t.Run("missing_start_error", func(t *testing.T) {
		req := makeToolRequest("calendar_create_event", map[string]any{
			"title": "Meeting without start",
		})
		_, err := parseEventInput(req)
		if err == nil {
			t.Fatal("expected error on missing start, got nil")
		}
	})
}

func TestParseAttendeesList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"nominal multiple attendees", "a@example.com, b@example.com, c@example.com", 3},
		{"empty string boundary", "", 0},
		{"extra commas and whitespace", " , a@example.com , , ", 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := parseAttendeesList(tc.input)
			if len(res) != tc.expected {
				t.Fatalf("expected %d attendees, got %d (%+v)", tc.expected, len(res), res)
			}
		})
	}
}
