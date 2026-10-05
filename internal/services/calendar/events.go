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
	MeetURL     string `json:"meet_url,omitempty"`
	SelfRSVP    string `json:"self_rsvp,omitempty"`
}

// EventOptions encapsulates options for creating or updating calendar events.
type EventOptions struct {
	CalendarID  string
	Title       string
	Description string
	Location    string
	Start       string
	End         string
	WithMeet    bool
	Attendees   []string
	SendUpdates string
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

func findSelfRSVP(attendees []*calendar.EventAttendee) string {
	for _, a := range attendees {
		if a.Self {
			return a.ResponseStatus
		}
	}
	return ""
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
			MeetURL:     item.HangoutLink,
			SelfRSVP:    findSelfRSVP(item.Attendees),
		})
	}
	return events
}

// QuickAddEvent creates an event using natural language string.
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

// CreateEventWithOptions creates a calendar event with options (Meet, attendees, etc.).
func (s *Service) CreateEventWithOptions(opts EventOptions) (*calendar.Event, error) {
	if opts.CalendarID == "" {
		opts.CalendarID = "primary"
	}
	event := buildCalendarEvent(opts)
	call := s.client.Events.Insert(opts.CalendarID, event)
	if opts.WithMeet {
		call = call.ConferenceDataVersion(1)
	}
	if opts.SendUpdates != "" {
		call = call.SendUpdates(opts.SendUpdates)
	}
	res, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("insert calendar event: %w", err)
	}
	return res, nil
}

func buildCalendarEvent(opts EventOptions) *calendar.Event {
	event := &calendar.Event{
		Summary:     opts.Title,
		Description: opts.Description,
		Location:    opts.Location,
		Start:       &calendar.EventDateTime{DateTime: opts.Start},
		End:         &calendar.EventDateTime{DateTime: opts.End},
	}
	for _, email := range opts.Attendees {
		if email != "" {
			event.Attendees = append(event.Attendees, &calendar.EventAttendee{Email: email})
		}
	}
	if opts.WithMeet {
		event.ConferenceData = &calendar.ConferenceData{
			CreateRequest: &calendar.CreateConferenceRequest{
				RequestId: fmt.Sprintf("gmcp-meet-%d", time.Now().UnixNano()),
				ConferenceSolutionKey: &calendar.ConferenceSolutionKey{
					Type: "hangoutsMeet",
				},
			},
		}
	}
	return event
}

// CreateEvent creates a structured calendar event with explicit start and end times.
func (s *Service) CreateEvent(calID, title, desc, loc, startISO, endISO string) (*calendar.Event, error) {
	return s.CreateEventWithOptions(EventOptions{
		CalendarID:  calID,
		Title:       title,
		Description: desc,
		Location:    loc,
		Start:       startISO,
		End:         endISO,
	})
}

// DeleteEvent removes an event from the calendar by ID.
func (s *Service) DeleteEvent(calID, eventID string) error {
	if calID == "" {
		calID = "primary"
	}
	if err := s.client.Events.Delete(calID, eventID).Do(); err != nil {
		return fmt.Errorf("delete calendar event %s: %w", eventID, err)
	}
	return nil
}
