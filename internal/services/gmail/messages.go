package gmail

import (
	"encoding/base64"
	"fmt"
	"strings"

	"google.golang.org/api/gmail/v1"
)

// MessageSummary contains minimal headers for list displays.
type MessageSummary struct {
	ID      string `json:"id"`
	From    string `json:"from"`
	Subject string `json:"subject"`
	Snippet string `json:"snippet"`
	Date    string `json:"date"`
}

// MessageDetail contains full body and headers.
type MessageDetail struct {
	MessageSummary
	Body string `json:"body"`
}

// ListMessages queries the Gmail API for messages matching a standard search query.
func (s *Service) ListMessages(query string, max int64) ([]MessageSummary, error) {
	if max <= 0 {
		max = 10
	}
	req := s.client.Users.Messages.List("me").MaxResults(max)
	if query != "" {
		req = req.Q(query)
	}
	res, err := req.Do()
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}

	return s.fetchSummaries(res.Messages)
}

// fetchSummaries converts message stubs into summarized messages.
func (s *Service) fetchSummaries(stubs []*gmail.Message) ([]MessageSummary, error) {
	var summaries []MessageSummary
	for _, stub := range stubs {
		msg, err := s.client.Users.Messages.Get("me", stub.Id).Format("metadata").MetadataHeaders("From", "Subject", "Date").Do()
		if err != nil {
			continue
		}
		summaries = append(summaries, extractSummary(msg))
	}
	return summaries, nil
}

// extractSummary parses headers and snippet from a Gmail API message.
func extractSummary(m *gmail.Message) MessageSummary {
	var from, subject, date string
	for _, h := range m.Payload.Headers {
		switch strings.ToLower(h.Name) {
		case "from":
			from = h.Value
		case "subject":
			subject = h.Value
		case "date":
			date = h.Value
		}
	}
	return MessageSummary{
		ID:      m.Id,
		From:    from,
		Subject: subject,
		Snippet: m.Snippet,
		Date:    date,
	}
}

// GetMessage retrieves full text of a specific message.
func (s *Service) GetMessage(id string) (*MessageDetail, error) {
	msg, err := s.client.Users.Messages.Get("me", id).Format("full").Do()
	if err != nil {
		return nil, fmt.Errorf("get message %s: %w", id, err)
	}

	summary := extractSummary(msg)
	body := extractBody(msg.Payload)
	return &MessageDetail{MessageSummary: summary, Body: body}, nil
}

// extractBody traverses payload parts to decode plain or html text.
func extractBody(p *gmail.MessagePart) string {
	if p.Body != nil && p.Body.Data != "" {
		decoded, err := base64.URLEncoding.DecodeString(p.Body.Data)
		if err == nil {
			return string(decoded)
		}
	}
	for _, sub := range p.Parts {
		if content := extractBody(sub); content != "" {
			return content
		}
	}
	return ""
}
