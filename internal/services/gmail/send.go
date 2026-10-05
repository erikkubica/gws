package gmail

import (
	"encoding/base64"
	"fmt"

	"google.golang.org/api/gmail/v1"
)

// composeRaw formats an RFC 822 email payload and base64url encodes it.
func composeRaw(to, subject, body string) string {
	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s",
		to, subject, body)
	return base64.URLEncoding.EncodeToString([]byte(msg))
}

// SendMessage sends an email directly via the authenticated Gmail account.
func (s *Service) SendMessage(to, subject, body string) (*gmail.Message, error) {
	raw := composeRaw(to, subject, body)
	call := s.client.Users.Messages.Send("me", &gmail.Message{Raw: raw})
	res, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("send email to %s: %w", to, err)
	}
	return res, nil
}

// CreateDraft creates a new draft email without sending it immediately.
func (s *Service) CreateDraft(to, subject, body string) (*gmail.Draft, error) {
	raw := composeRaw(to, subject, body)
	draft := &gmail.Draft{Message: &gmail.Message{Raw: raw}}
	res, err := s.client.Users.Drafts.Create("me", draft).Do()
	if err != nil {
		return nil, fmt.Errorf("create draft to %s: %w", to, err)
	}
	return res, nil
}
