package calendar

import (
	"fmt"
	"time"

	"google.golang.org/api/calendar/v3"
)

// EventSummary contains high-level calendar event information.
type EventSummary struct {
	ID          string `json:"id"`
	Summary     string `json:"summary"`
	Start       string `json:"start"`
	End         string `json:"end"`
	Location    string `json:"location,omitempty"`
	Description string `json:"description,omitempty"`
}

// ListUpcomingEvents retrieves upcoming calendar events starting from now.
func (s *Service) ListUpcomingEvents(calID string, max int64) ([]EventSummary, error) {
	if calID == "" {
		calID = "primary"
	}
	if max <= 0 {
		max = 10
	}
	tMin := time.Now().Format(time.RFC3339)
	res, err := s.client.Events.List(calID).ShowDeleted(false).
		SingleEvents(true).TimeMin(tMin).MaxResults(max).OrderBy("startTime").Do()
	if err != nil {
		return nil, fmt.Errorf("list calendar events: %w", err)
	}

	return mapEvents(res.Items), nil
}

// mapEvents maps Google Calendar event structs to EventSummary.
func mapEvents(items []*calendar.Event) []EventSummary {
	var events []EventSummary
	for _, item := range items {
		start := item.Start.DateTime
		if start == "" {
			start = item.Start.Date
		}
		end := item.End.DateTime
		if end == "" {
			end = item.End.Date
		}
		events = append(events, EventSummary{
			ID:          item.Id,
			Summary:     item.Summary,
			Start:       start,
			End:         end,
			Location:    item.Location,
			Description: item.Description,
		})
	}
	return events
}

// QuickAddEvent creates an event using natural language string (e.g. "Meeting with John tomorrow at 3pm").
func (s *Service) QuickAddEvent(calID, text string) (*calendar.Event, error) {
	if calID == "" {
		calID = "primary"
	}
	event, err := s.client.Events.QuickAdd(calID, text).Do()
	if err != nil {
		return nil, fmt.Errorf("quickadd event: %w", err)
	}
	return event, nil
}
