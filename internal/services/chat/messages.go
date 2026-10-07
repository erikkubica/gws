package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/api/chat/v1"
)

// MessageOptions contains parameters for delivering a Chat message.
type MessageOptions struct {
	SpaceName  string
	Text       string
	ReplyTo    string
	Attachment string
}

// MessageInfo represents a Google Chat message.
type MessageInfo struct {
	Name       string `json:"name"`
	Text       string `json:"text"`
	SenderName string `json:"sender_name"`
	CreateTime string `json:"create_time"`
	ThreadName string `json:"thread_name,omitempty"`
}

// SendMessage delivers a text message to a designated space.
func (s *Service) SendMessage(spaceName, text string) (*MessageInfo, error) {
	return s.SendMessageWithOptions(MessageOptions{
		SpaceName: spaceName,
		Text:      text,
	})
}

// SendMessageWithOptions delivers a message with optional threading and attachments.
func (s *Service) SendMessageWithOptions(opts MessageOptions) (*MessageInfo, error) {
	resName := NormalizeSpaceName(opts.SpaceName)
	if resName == "" {
		return nil, fmt.Errorf("space name cannot be empty")
	}
	msg := &chat.Message{Text: opts.Text}
	if err := s.attachMedia(resName, opts.Attachment, msg); err != nil {
		return nil, err
	}
	if err := s.configureReplyThread(resName, opts.ReplyTo, msg); err != nil {
		return nil, err
	}

	call := s.client.Spaces.Messages.Create(resName, msg)
	if opts.ReplyTo != "" {
		call = call.MessageReplyOption("REPLY_MESSAGE_FALLBACK_TO_NEW_THREAD")
	}
	res, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("send message to %s: %w", resName, err)
	}
	return formatMessageInfo(res), nil
}

func (s *Service) attachMedia(spaceName, filePath string, msg *chat.Message) error {
	if strings.TrimSpace(filePath) == "" {
		return nil
	}
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open attachment %s: %w", filePath, err)
	}
	defer f.Close()

	req := &chat.UploadAttachmentRequest{Filename: filepath.Base(filePath)}
	res, err := s.client.Media.Upload(spaceName, req).Media(f).Do()
	if err != nil {
		return fmt.Errorf("upload attachment %s: %w", filePath, err)
	}
	msg.Attachment = []*chat.Attachment{{AttachmentDataRef: res.AttachmentDataRef}}
	return nil
}

func (s *Service) configureReplyThread(spaceName, replyTo string, msg *chat.Message) error {
	if strings.TrimSpace(replyTo) == "" {
		return nil
	}
	parentMsgName := NormalizeMessageName(spaceName, replyTo)
	parent, err := s.client.Spaces.Messages.Get(parentMsgName).Do()
	if err != nil {
		return fmt.Errorf("fetch parent message %s: %w", parentMsgName, err)
	}
	if parent.Thread != nil && parent.Thread.Name != "" {
		msg.Thread = &chat.Thread{Name: parent.Thread.Name}
	}
	return nil
}

func formatMessageInfo(res *chat.Message) *MessageInfo {
	sender := ""
	if res.Sender != nil {
		sender = res.Sender.DisplayName
	}
	thread := ""
	if res.Thread != nil {
		thread = res.Thread.Name
	}
	return &MessageInfo{
		Name:       res.Name,
		Text:       res.Text,
		SenderName: sender,
		CreateTime: res.CreateTime,
		ThreadName: thread,
	}
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
		msgs = append(msgs, formatMessageInfo(m))
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
