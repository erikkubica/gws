package gmail

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"
)

// EmailOptions encapsulates all parameters for sending or drafting messages.
type EmailOptions struct {
	To          string
	Subject     string
	Body        string
	InReplyTo   string
	References  string
	ThreadID    string
	Attachments []string
}

// composeRaw formats an RFC 822 email payload and base64url encodes it.
func composeRaw(to, subject, body string) string {
	raw, _ := composeMIME(EmailOptions{To: to, Subject: subject, Body: body})
	return raw
}

// composeMIME formats a MIME message (with multipart if attachments exist).
func composeMIME(opts EmailOptions) (string, error) {
	buf := new(bytes.Buffer)
	boundary := generateBoundary()
	writeHeaders(buf, opts, boundary)

	fmt.Fprintf(buf, "--%s\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s\r\n", boundary, opts.Body)

	if err := appendAttachments(buf, boundary, opts.Attachments); err != nil {
		return "", err
	}
	fmt.Fprintf(buf, "--%s--\r\n", boundary)
	return base64.URLEncoding.EncodeToString(buf.Bytes()), nil
}

func sanitizeHeader(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}

func encodeSubject(subject string) string {
	clean := sanitizeHeader(subject)
	if clean == "" {
		return ""
	}
	return mime.QEncoding.Encode("utf-8", clean)
}

func generateBoundary() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("gws_boundary_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("gws_boundary_%x", b)
}

func writeHeaders(buf *bytes.Buffer, opts EmailOptions, boundary string) {
	cleanTo := sanitizeHeader(opts.To)
	fmt.Fprintf(buf, "To: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\n", cleanTo, encodeSubject(opts.Subject))
	if opts.InReplyTo != "" {
		fmt.Fprintf(buf, "In-Reply-To: %s\r\n", sanitizeHeader(opts.InReplyTo))
	}
	if opts.References != "" {
		fmt.Fprintf(buf, "References: %s\r\n", sanitizeHeader(opts.References))
	}
	fmt.Fprintf(buf, "Content-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", boundary)
}

func appendAttachments(buf *bytes.Buffer, boundary string, files []string) error {
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read attachment %s: %w", f, err)
		}
		filename := filepath.Base(f)
		mType := mime.TypeByExtension(filepath.Ext(f))
		if mType == "" {
			mType = "application/octet-stream"
		}
		encoded := base64.StdEncoding.EncodeToString(data)
		fmt.Fprintf(buf, "--%s\r\nContent-Type: %s; name=\"%s\"\r\nContent-Disposition: attachment; filename=\"%s\"\r\nContent-Transfer-Encoding: base64\r\n\r\n%s\r\n",
			boundary, mType, filename, filename, encoded)
	}
	return nil
}

// SendMessage sends an email with full options.
func (s *Service) SendMessage(opts EmailOptions) (*gmail.Message, error) {
	raw, err := composeMIME(opts)
	if err != nil {
		return nil, err
	}
	msg := &gmail.Message{Raw: raw}
	if opts.ThreadID != "" {
		msg.ThreadId = opts.ThreadID
	}
	res, err := s.client.Users.Messages.Send("me", msg).Do()
	if err != nil {
		return nil, fmt.Errorf("send email to %s: %w", opts.To, err)
	}
	return res, nil
}

// CreateDraft creates a new draft email with full options.
func (s *Service) CreateDraft(opts EmailOptions) (*gmail.Draft, error) {
	raw, err := composeMIME(opts)
	if err != nil {
		return nil, err
	}
	draft := &gmail.Draft{Message: &gmail.Message{Raw: raw}}
	if opts.ThreadID != "" {
		draft.Message.ThreadId = opts.ThreadID
	}
	res, err := s.client.Users.Drafts.Create("me", draft).Do()
	if err != nil {
		return nil, fmt.Errorf("create draft to %s: %w", opts.To, err)
	}
	return res, nil
}

// ReplyMessage prepares and sends a reply to an existing message thread.
func (s *Service) ReplyMessage(msgID, body string, attachments []string) (*gmail.Message, error) {
	orig, err := s.client.Users.Messages.Get("me", msgID).Format("metadata").
		MetadataHeaders("From", "Subject", "Message-ID", "References").Do()
	if err != nil {
		return nil, fmt.Errorf("retrieve original message %s: %w", msgID, err)
	}
	opts := buildReplyOptions(orig, body, attachments)
	return s.SendMessage(opts)
}

func buildReplyOptions(m *gmail.Message, body string, attachments []string) EmailOptions {
	to, subj, msgID, refs := extractReplyHeaders(m.Payload.Headers)
	if !strings.HasPrefix(strings.ToLower(subj), "re:") {
		subj = "Re: " + subj
	}
	if refs == "" {
		refs = msgID
	} else {
		refs = refs + " " + msgID
	}
	return EmailOptions{
		To: to, Subject: subj, Body: body, InReplyTo: msgID,
		References: refs, ThreadID: m.ThreadId, Attachments: attachments,
	}
}

func extractReplyHeaders(headers []*gmail.MessagePartHeader) (to, subj, msgID, refs string) {
	for _, h := range headers {
		switch strings.ToLower(h.Name) {
		case "from":
			to = h.Value
		case "subject":
			subj = h.Value
		case "message-id":
			msgID = h.Value
		case "references":
			refs = h.Value
		}
	}
	return
}
