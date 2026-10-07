package gmail

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/api/gmail/v1"
)

// AttachmentInfo describes a file attachment on an email message.
type AttachmentInfo struct {
	AttachmentID string `json:"attachment_id,omitempty"`
	Filename     string `json:"filename"`
	MimeType     string `json:"mime_type,omitempty"`
	Size         int64  `json:"size,omitempty"`
}

// AttachmentFile contains the decoded attachment file content.
type AttachmentFile struct {
	AttachmentInfo
	Data []byte `json:"-"`
}

// extractAttachments recursively extracts attachment metadata from MIME message parts.
func extractAttachments(p *gmail.MessagePart) []AttachmentInfo {
	if p == nil {
		return nil
	}
	var atts []AttachmentInfo
	if p.Filename != "" {
		var attID string
		var size int64
		if p.Body != nil {
			attID = p.Body.AttachmentId
			size = p.Body.Size
		}
		atts = append(atts, AttachmentInfo{
			AttachmentID: attID,
			Filename:     p.Filename,
			MimeType:     p.MimeType,
			Size:         size,
		})
	}
	for _, sub := range p.Parts {
		atts = append(atts, extractAttachments(sub)...)
	}
	return atts
}

// decodeBase64URL decodes URL-safe base64 data with or without padding.
func decodeBase64URL(data string) ([]byte, error) {
	b, err := base64.URLEncoding.DecodeString(data)
	if err == nil {
		return b, nil
	}
	return base64.RawURLEncoding.DecodeString(data)
}

// GetAttachmentContent fetches raw binary content of an attachment by ID.
func (s *Service) GetAttachmentContent(messageID, attachmentID string) ([]byte, error) {
	if attachmentID == "" {
		return nil, fmt.Errorf("attachment ID is required")
	}
	res, err := s.client.Users.Messages.Attachments.Get("me", messageID, attachmentID).Do()
	if err != nil {
		return nil, fmt.Errorf("fetch attachment: %w", err)
	}
	data, err := decodeBase64URL(res.Data)
	if err != nil {
		return nil, fmt.Errorf("decode attachment data: %w", err)
	}
	return data, nil
}

// findAttachmentPart locates the MessagePart matching attachment ID or filename.
func findAttachmentPart(p *gmail.MessagePart, idOrName string) *gmail.MessagePart {
	if p == nil || idOrName == "" {
		return nil
	}
	if p.Filename != "" {
		isID := p.Body != nil && p.Body.AttachmentId != "" && p.Body.AttachmentId == idOrName
		isName := strings.EqualFold(p.Filename, idOrName) || strings.EqualFold(filepath.Base(p.Filename), idOrName)
		if isID || isName {
			return p
		}
	}
	for _, sub := range p.Parts {
		if match := findAttachmentPart(sub, idOrName); match != nil {
			return match
		}
	}
	return nil
}

// extractPartData resolves binary bytes from inline data or via Attachments API.
func (s *Service) extractPartData(messageID string, part *gmail.MessagePart) ([]byte, error) {
	if part.Body == nil {
		return nil, fmt.Errorf("attachment part has no body")
	}
	if part.Body.Data != "" {
		return decodeBase64URL(part.Body.Data)
	}
	if part.Body.AttachmentId != "" {
		return s.GetAttachmentContent(messageID, part.Body.AttachmentId)
	}
	return nil, fmt.Errorf("attachment has neither inline data nor attachment ID")
}

func (s *Service) resolveDirectAttachment(messageID, idOrName string) (*AttachmentFile, error) {
	data, err := s.GetAttachmentContent(messageID, idOrName)
	if err != nil {
		return nil, fmt.Errorf("attachment %q not found in message %s", idOrName, messageID)
	}
	return &AttachmentFile{
		AttachmentInfo: AttachmentInfo{
			AttachmentID: idOrName,
			Filename:     "attachment",
			Size:         int64(len(data)),
		},
		Data: data,
	}, nil
}

// FetchAttachment retrieves an attachment file by message ID and attachment ID or filename.
func (s *Service) FetchAttachment(messageID, idOrName string) (*AttachmentFile, error) {
	msg, err := s.client.Users.Messages.Get("me", messageID).Format("full").Do()
	if err != nil {
		return nil, fmt.Errorf("get message %s: %w", messageID, err)
	}
	part := findAttachmentPart(msg.Payload, idOrName)
	if part == nil && idOrName == "" {
		atts := extractAttachments(msg.Payload)
		if len(atts) == 1 {
			part = findAttachmentPart(msg.Payload, atts[0].Filename)
		}
	}
	if part == nil {
		if idOrName != "" {
			return s.resolveDirectAttachment(messageID, idOrName)
		}
		return nil, fmt.Errorf("no attachments found in message %s", messageID)
	}
	data, err := s.extractPartData(messageID, part)
	if err != nil {
		return nil, err
	}
	var attID string
	if part.Body != nil {
		attID = part.Body.AttachmentId
	}
	return &AttachmentFile{
		AttachmentInfo: AttachmentInfo{
			AttachmentID: attID,
			Filename:     filepath.Base(part.Filename),
			MimeType:     part.MimeType,
			Size:         int64(len(data)),
		},
		Data: data,
	}, nil
}

// ListAttachments returns all attachments metadata for a specific message.
func (s *Service) ListAttachments(messageID string) ([]AttachmentInfo, error) {
	msg, err := s.client.Users.Messages.Get("me", messageID).Format("full").Do()
	if err != nil {
		return nil, fmt.Errorf("get message %s: %w", messageID, err)
	}
	return extractAttachments(msg.Payload), nil
}

// SaveAttachment downloads and safely writes an attachment to a local file.
func (s *Service) SaveAttachment(messageID, idOrName, destPath string) (*AttachmentFile, error) {
	file, err := s.FetchAttachment(messageID, idOrName)
	if err != nil {
		return nil, err
	}
	targetPath := destPath
	fi, err := os.Stat(destPath)
	if err == nil && fi.IsDir() {
		targetPath = filepath.Join(destPath, file.Filename)
	}
	if err := os.WriteFile(targetPath, file.Data, 0644); err != nil {
		return nil, fmt.Errorf("write attachment to %s: %w", targetPath, err)
	}
	return file, nil
}

// DownloadAllAttachments fetches and saves all attachments from a message to the target directory.
func (s *Service) DownloadAllAttachments(messageID, targetDir string) ([]string, error) {
	atts, err := s.ListAttachments(messageID)
	if err != nil {
		return nil, err
	}
	if len(atts) == 0 {
		return nil, fmt.Errorf("no attachments found on message %s", messageID)
	}
	if targetDir == "" {
		targetDir = "."
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}
	var saved []string
	for _, a := range atts {
		dest := filepath.Join(targetDir, filepath.Base(a.Filename))
		if _, err := s.SaveAttachment(messageID, a.Filename, dest); err != nil {
			return nil, err
		}
		saved = append(saved, dest)
	}
	return saved, nil
}
