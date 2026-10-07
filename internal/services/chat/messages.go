package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/api/chat/v1"
)

// MessageInfo represents a Google Chat message.
type MessageInfo struct {
	Name       string `json:"name"`
	Text       string `json:"text"`
	SenderName string `json:"sender_name"`
	CreateTime string `json:"create_time"`
}

// SendMessage delivers a text message to a designated space or direct message room.
func (s *Service) SendMessage(spaceName, text string) (*MessageInfo, error) {
	resName := NormalizeSpaceName(spaceName)
	if resName == "" {
		return nil, fmt.Errorf("space name cannot be empty")
	}
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("message text cannot be empty")
	}

	msg := &chat.Message{Text: text}
	res, err := s.client.Spaces.Messages.Create(resName, msg).Do()
	if err != nil {
		return nil, fmt.Errorf("send message to %s: %w", resName, err)
	}

	sender := ""
	if res.Sender != nil {
		sender = res.Sender.DisplayName
	}
	return &MessageInfo{
		Name:       res.Name,
		Text:       res.Text,
		SenderName: sender,
		CreateTime: res.CreateTime,
	}, nil
}

// ListMessages retrieves recent messages from a space.
func (s *Service) ListMessages(spaceName string, pageSize int64) ([]*MessageInfo, error) {
	resName := NormalizeSpaceName(spaceName)
	if resName == "" {
		return nil, fmt.Errorf("space name cannot be empty")
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	call := s.client.Spaces.Messages.List(resName).PageSize(pageSize)
	res, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("list messages for %s: %w", resName, err)
	}

	var msgs []*MessageInfo
	for _, m := range res.Messages {
		sender := ""
		if m.Sender != nil {
			sender = m.Sender.DisplayName
		}
		msgs = append(msgs, &MessageInfo{
			Name:       m.Name,
			Text:       m.Text,
			SenderName: sender,
			CreateTime: m.CreateTime,
		})
	}
	return msgs, nil
}

// SendWebhook posts a message to an incoming Google Chat webhook URL.
func (s *Service) SendWebhook(webhookURL, text string) error {
	if strings.TrimSpace(webhookURL) == "" {
		return fmt.Errorf("webhook URL cannot be empty")
	}
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("message text cannot be empty")
	}

	payload, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return fmt.Errorf("encode webhook payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("create webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-success HTTP status %d", resp.StatusCode)
	}
	return nil
}
