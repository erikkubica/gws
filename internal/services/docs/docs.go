package docs

import (
	"fmt"
	"strings"

	"google.golang.org/api/docs/v1"
)

// CreateDocument creates a new Google Doc with the specified title.
func (s *Service) CreateDocument(title string) (*docs.Document, error) {
	doc := &docs.Document{Title: title}
	res, err := s.client.Documents.Create(doc).Do()
	if err != nil {
		return nil, fmt.Errorf("create document: %w", err)
	}
	return res, nil
}

// AppendText inserts text at the end of the Google Doc.
func (s *Service) AppendText(docID, text string) error {
	doc, err := s.client.Documents.Get(docID).Do()
	if err != nil {
		return fmt.Errorf("fetch document: %w", err)
	}

	endIndex := int64(1)
	if len(doc.Body.Content) > 0 {
		endIndex = doc.Body.Content[len(doc.Body.Content)-1].EndIndex - 1
		if endIndex < 1 {
			endIndex = 1
		}
	}

	req := &docs.Request{
		InsertText: &docs.InsertTextRequest{
			Location: &docs.Location{Index: endIndex},
			Text:     text,
		},
	}

	batch := &docs.BatchUpdateDocumentRequest{Requests: []*docs.Request{req}}
	if _, err := s.client.Documents.BatchUpdate(docID, batch).Do(); err != nil {
		return fmt.Errorf("append text to doc: %w", err)
	}
	return nil
}

// GetDocumentText extracts and returns plain text from the Google Doc.
func (s *Service) GetDocumentText(docID string) (string, error) {
	doc, err := s.client.Documents.Get(docID).Do()
	if err != nil {
		return "", fmt.Errorf("fetch document: %w", err)
	}

	var sb strings.Builder
	for _, elem := range doc.Body.Content {
		if elem.Paragraph == nil {
			continue
		}
		for _, pElem := range elem.Paragraph.Elements {
			if pElem.TextRun != nil {
				sb.WriteString(pElem.TextRun.Content)
			}
		}
	}
	return sb.String(), nil
}
