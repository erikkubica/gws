package gmail

import (
	"encoding/base64"
	"fmt"
	"strings"
	"sync"

	"google.golang.org/api/gmail/v1"
)

// MessageSummary contains headers and attachments for list displays.
type MessageSummary struct {
	ID          string           `json:"id"`
	From        string           `json:"from"`
	Subject     string           `json:"subject"`
	Snippet     string           `json:"snippet"`
	Date        string           `json:"date"`
	Attachments []AttachmentInfo `json:"attachments,omitempty"`
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

// fetchSummaries converts message stubs into summarized messages concurrently.
func (s *Service) fetchSummaries(stubs []*gmail.Message) ([]MessageSummary, error) {
	summaries := make([]MessageSummary, len(stubs))
	var wg sync.WaitGroup
	for i, stub := range stubs {
		wg.Add(1)
		go func(idx int, id string) {
			defer wg.Done()
			msg, err := s.client.Users.Messages.Get("me", id).Format("full").Do()
			if err == nil {
				summaries[idx] = extractSummary(msg)
			}
		}(i, stub.Id)
	}
	wg.Wait()

	var result []MessageSummary
	for _, sum := range summaries {
		if sum.ID != "" {
			result = append(result, sum)
		}
	}
	return result, nil
}

// extractSummary parses headers, snippet, and attachments from a Gmail API message.
func extractSummary(m *gmail.Message) MessageSummary {
	var from, subject, date string
	if m.Payload != nil {
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
	}
	return MessageSummary{
		ID:          m.Id,
		From:        from,
		Subject:     subject,
		Snippet:     m.Snippet,
		Date:        date,
		Attachments: extractAttachments(m.Payload),
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
