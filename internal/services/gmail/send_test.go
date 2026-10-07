package gmail

import (
	"encoding/base64"
	"mime"
	"strings"
	"testing"

	gmailapi "google.golang.org/api/gmail/v1"
)

func TestComposeRaw(t *testing.T) {
	tests := []struct {
		name         string
		to           string
		subject      string
		body         string
		expectedSubj string
	}{
		{
			name:         "nominal case",
			to:           "recruiter@example.com",
			subject:      "Senior Developer Application",
			body:         "Hello, I am interested in this position.",
			expectedSubj: "Senior Developer Application",
		},
		{
			name:         "empty strings boundary",
			to:           "",
			subject:      "",
			body:         "",
			expectedSubj: "",
		},
		{
			name:         "special characters and diacritics",
			to:           "user@example.com",
			subject:      "B2B Spolupráca: Senior Full-Stack & AI Inžinier",
			body:         "Dobrý deň, môj pracovný deň začína o 6:00 ráno.",
			expectedSubj: "B2B Spolupráca: Senior Full-Stack & AI Inžinier",
		},
		{
			name:         "crlf header injection attempt",
			to:           "user@example.com",
			subject:      "Legit Subject\r\nBcc: evil@example.com",
			body:         "Testing CRLF prevention",
			expectedSubj: "Legit Subject Bcc: evil@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := composeRaw(tt.to, tt.subject, tt.body)
			decoded, err := base64.URLEncoding.DecodeString(raw)
			if err != nil {
				t.Fatalf("failed to decode base64url string: %v", err)
			}
			s := string(decoded)
			if !strings.Contains(s, "To: "+tt.to) {
				t.Errorf("expected To: header missing: %s", s)
			}

			// Ensure raw \r\n injection did not create an unescaped Bcc header
			if strings.Contains(s, "\r\nBcc:") {
				t.Errorf("found unescaped CRLF header injection in: %s", s)
			}

			if tt.expectedSubj != "" {
				encoded := encodeSubject(tt.subject)
				if !strings.Contains(s, "Subject: "+encoded) {
					t.Errorf("expected Subject: %s, got: %s", encoded, s)
				}
				decHeader, err := (&mime.WordDecoder{}).DecodeHeader(encoded)
				if err != nil {
					t.Fatalf("failed to decode Q-encoded header: %v", err)
				}
				if decHeader != tt.expectedSubj {
					t.Errorf("expected decoded subject %q, got %q", tt.expectedSubj, decHeader)
				}
			}
			if !strings.Contains(s, tt.body) {
				t.Errorf("expected body missing: %s", s)
			}
		})
	}
}

func TestDynamicBoundaryGeneration(t *testing.T) {
	raw1, err := composeMIME(EmailOptions{To: "a@a.com", Subject: "test", Body: "1"})
	if err != nil {
		t.Fatalf("composeMIME 1 failed: %v", err)
	}
	raw2, err := composeMIME(EmailOptions{To: "a@a.com", Subject: "test", Body: "2"})
	if err != nil {
		t.Fatalf("composeMIME 2 failed: %v", err)
	}

	dec1, _ := base64.URLEncoding.DecodeString(raw1)
	dec2, _ := base64.URLEncoding.DecodeString(raw2)

	// Ensure static "gws_boundary_part" is not used
	if strings.Contains(string(dec1), "gws_boundary_part") {
		t.Fatal("found static boundary 'gws_boundary_part'")
	}

	// Verify boundaries are different across calls
	b1 := extractBoundary(string(dec1))
	b2 := extractBoundary(string(dec2))
	if b1 == "" || b2 == "" {
		t.Fatalf("failed to extract boundary: b1=%q, b2=%q", b1, b2)
	}
	if b1 == b2 {
		t.Errorf("expected distinct boundaries, got identical: %s", b1)
	}
}

func extractBoundary(raw string) string {
	const prefix = "boundary=\""
	idx := strings.Index(raw, prefix)
	if idx == -1 {
		return ""
	}
	rest := raw[idx+len(prefix):]
	endIdx := strings.Index(rest, "\"")
	if endIdx == -1 {
		return ""
	}
	return rest[:endIdx]
}

func TestBuildReplyOptions(t *testing.T) {
	orig := &gmailapi.Message{
		ThreadId: "thread-xyz",
		Payload: &gmailapi.MessagePart{
			Headers: []*gmailapi.MessagePartHeader{
				{Name: "From", Value: "boss@example.com"},
				{Name: "Subject", Value: "Quarterly Planning"},
				{Name: "Message-ID", Value: "<msg1@example.com>"},
				{Name: "References", Value: "<root@example.com>"},
			},
		},
	}
	opts := buildReplyOptions(orig, "Acknowledged.", []string{"report.pdf"})
	if opts.To != "boss@example.com" {
		t.Errorf("expected To 'boss@example.com', got %q", opts.To)
	}
	if opts.Subject != "Re: Quarterly Planning" {
		t.Errorf("expected Subject 'Re: Quarterly Planning', got %q", opts.Subject)
	}
	if opts.InReplyTo != "<msg1@example.com>" {
		t.Errorf("expected InReplyTo '<msg1@example.com>', got %q", opts.InReplyTo)
	}
	if opts.References != "<root@example.com> <msg1@example.com>" {
		t.Errorf("expected References '<root@example.com> <msg1@example.com>', got %q", opts.References)
	}
	if opts.ThreadID != "thread-xyz" {
		t.Errorf("expected ThreadID 'thread-xyz', got %q", opts.ThreadID)
	}

	// Boundary: existing "re:" prefix should not be duplicated
	orig2 := &gmailapi.Message{
		Payload: &gmailapi.MessagePart{
			Headers: []*gmailapi.MessagePartHeader{
				{Name: "Subject", Value: "re: status update"},
				{Name: "Message-ID", Value: "<msg2@example.com>"},
			},
		},
	}
	opts2 := buildReplyOptions(orig2, "ok", nil)
	if opts2.Subject != "re: status update" {
		t.Errorf("expected subject preserved without double 'Re: ', got %q", opts2.Subject)
	}
	if opts2.References != "<msg2@example.com>" {
		t.Errorf("expected references fallback to msgID, got %q", opts2.References)
	}
}
