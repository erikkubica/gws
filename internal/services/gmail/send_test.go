package gmail

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestComposeRaw(t *testing.T) {
	tests := []struct {
		name    string
		to      string
		subject string
		body    string
	}{
		{
			name:    "nominal case",
			to:      "recruiter@example.com",
			subject: "Senior Developer Application",
			body:    "Hello, I am interested in this position.",
		},
		{
			name:    "empty strings boundary",
			to:      "",
			subject: "",
			body:    "",
		},
		{
			name:    "special characters and diacritics",
			to:      "user@example.com",
			subject: "B2B Spolupráca: Senior Full-Stack & AI Inžinier",
			body:    "Dobrý deň, môj pracovný deň začína o 6:00 ráno.",
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
			if !strings.Contains(s, "Subject: "+tt.subject) {
				t.Errorf("expected Subject: header missing: %s", s)
			}
			if !strings.Contains(s, tt.body) {
				t.Errorf("expected body missing: %s", s)
			}
		})
	}
}
