package gmail

import (
	"fmt"
	"strings"

	"google.golang.org/api/gmail/v1"
)

// DraftSummary represents high-level draft metadata.
type DraftSummary struct {
	ID      string `json:"id"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Snippet string `json:"snippet"`
}

// ListDrafts lists existing drafts from the mailbox.
func (s *Service) ListDrafts(max int64) ([]DraftSummary, error) {
	if max <= 0 {
		max = 10
	}
	res, err := s.client.Users.Drafts.List("me").MaxResults(max).Do()
	if err != nil {
		return nil, fmt.Errorf("list drafts: %w", err)
	}

	var drafts []DraftSummary
	for _, d := range res.Drafts {
		detailed, err := s.client.Users.Drafts.Get("me", d.Id).Format("full").Do()
		if err != nil {
			continue
		}
		drafts = append(drafts, extractDraftSummary(detailed))
	}
	return drafts, nil
}

func extractDraftSummary(d *gmail.Draft) DraftSummary {
	var to, subj string
	if d.Message != nil && d.Message.Payload != nil {
		for _, h := range d.Message.Payload.Headers {
			switch strings.ToLower(h.Name) {
			case "to":
				to = h.Value
			case "subject":
				subj = h.Value
			}
		}
	}
	snippet := ""
	if d.Message != nil {
		snippet = d.Message.Snippet
	}
	return DraftSummary{ID: d.Id, To: to, Subject: subj, Snippet: snippet}
}

// SendDraft sends an existing draft by its ID.
func (s *Service) SendDraft(draftID string) (*gmail.Message, error) {
	draft := &gmail.Draft{Id: draftID}
	msg, err := s.client.Users.Drafts.Send("me", draft).Do()
	if err != nil {
		return nil, fmt.Errorf("send draft %s: %w", draftID, err)
	}
	return msg, nil
}

// DeleteDraft deletes a draft by ID.
func (s *Service) DeleteDraft(draftID string) error {
	if err := s.client.Users.Drafts.Delete("me", draftID).Do(); err != nil {
		return fmt.Errorf("delete draft %s: %w", draftID, err)
	}
	return nil
}
