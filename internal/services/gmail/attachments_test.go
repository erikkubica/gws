package gmail

import (
	"encoding/base64"
	"testing"

	"google.golang.org/api/gmail/v1"
)

func TestExtractAttachments(t *testing.T) {
	tests := []struct {
		name     string
		part     *gmail.MessagePart
		expected int
	}{
		{
			name:     "nil part boundary",
			part:     nil,
			expected: 0,
		},
		{
			name: "single attachment nominal",
			part: &gmail.MessagePart{
				Filename: "invoice.pdf",
				MimeType: "application/pdf",
				Body: &gmail.MessagePartBody{
					AttachmentId: "att_123",
					Size:         1024,
				},
			},
			expected: 1,
		},
		{
			name: "multipart with nested attachments",
			part: &gmail.MessagePart{
				MimeType: "multipart/mixed",
				Parts: []*gmail.MessagePart{
					{
						MimeType: "text/plain",
						Body:     &gmail.MessagePartBody{Data: "aGVsbG8="},
					},
					{
						Filename: "photo.png",
						MimeType: "image/png",
						Body: &gmail.MessagePartBody{
							AttachmentId: "att_photo",
							Size:         2048,
						},
					},
					{
						MimeType: "multipart/related",
						Parts: []*gmail.MessagePart{
							{
								Filename: "document.docx",
								MimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
								Body: &gmail.MessagePartBody{
									AttachmentId: "att_doc",
									Size:         4096,
								},
							},
						},
					},
				},
			},
			expected: 2,
		},
		{
			name: "body without filename boundary",
			part: &gmail.MessagePart{
				Filename: "",
				MimeType: "text/html",
				Body: &gmail.MessagePartBody{
					Data: "PGgxPkhlbGxvPC9oMT4=",
					Size: 50,
				},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			atts := extractAttachments(tt.part)
			if len(atts) != tt.expected {
				t.Fatalf("expected %d attachments, got %d", tt.expected, len(atts))
			}
		})
	}
}

func TestFindAttachmentPart(t *testing.T) {
	root := &gmail.MessagePart{
		MimeType: "multipart/mixed",
		Parts: []*gmail.MessagePart{
			{
				Filename: "Specification.PDF",
				MimeType: "application/pdf",
				Body: &gmail.MessagePartBody{
					AttachmentId: "att_spec_99",
					Size:         5000,
				},
			},
			{
				Filename: "notes.txt",
				MimeType: "text/plain",
				Body: &gmail.MessagePartBody{
					AttachmentId: "att_notes_01",
					Size:         100,
				},
			},
		},
	}

	tests := []struct {
		name      string
		query     string
		wantFound bool
		wantName  string
	}{
		{
			name:      "find by attachment ID nominal",
			query:     "att_spec_99",
			wantFound: true,
			wantName:  "Specification.PDF",
		},
		{
			name:      "find by exact filename",
			query:     "notes.txt",
			wantFound: true,
			wantName:  "notes.txt",
		},
		{
			name:      "find by case-insensitive filename",
			query:     "specification.pdf",
			wantFound: true,
			wantName:  "Specification.PDF",
		},
		{
			name:      "not found query boundary",
			query:     "nonexistent.zip",
			wantFound: false,
		},
		{
			name:      "empty query boundary",
			query:     "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			part := findAttachmentPart(root, tt.query)
			if tt.wantFound && (part == nil || part.Filename != tt.wantName) {
				t.Errorf("expected to find %s, got %+v", tt.wantName, part)
			}
			if !tt.wantFound && part != nil {
				t.Errorf("expected nil for query %s, got %+v", tt.query, part)
			}
		})
	}
}

func TestDecodeBase64URL(t *testing.T) {
	nominalRaw := "Hello Google Workspace attachments!"
	padded := base64.URLEncoding.EncodeToString([]byte(nominalRaw))
	rawUnpadded := base64.RawURLEncoding.EncodeToString([]byte(nominalRaw))

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "padded base64url", input: padded, wantErr: false},
		{name: "unpadded base64url", input: rawUnpadded, wantErr: false},
		{name: "empty string boundary", input: "", wantErr: false},
		{name: "malformed base64", input: "!!@@##$$%%", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := decodeBase64URL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("decodeBase64URL(%q) err = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && tt.input != "" && string(data) != nominalRaw {
				t.Errorf("expected %q, got %q", nominalRaw, string(data))
			}
		})
	}
}
