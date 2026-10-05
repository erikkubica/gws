package calendar

import (
	"fmt"
	"time"

	"google.golang.org/api/calendar/v3"
)

// CreateQuickMeet provisions a Google Meet session with optional attendees.
func (s *Service) CreateQuickMeet(title, startISO, endISO string, attendees []string) (*calendar.Event, string, error) {
	if title == "" {
		title = "Google Meet Video Call"
	}
	startISO, endISO = defaultMeetTimes(startISO, endISO)
	opts := EventOptions{
		CalendarID:  "primary",
		Title:       title,
		Description: "Created via gmcp Google Meet",
		Start:       startISO,
		End:         endISO,
		WithMeet:    true,
		Attendees:   attendees,
		SendUpdates: "all",
	}
	ev, err := s.CreateEventWithOptions(opts)
	if err != nil {
		return nil, "", fmt.Errorf("create meet: %w", err)
	}
	return ev, extractMeetURL(ev), nil
}

func defaultMeetTimes(startISO, endISO string) (string, string) {
	now := time.Now()
	if startISO == "" {
		startISO = now.Format(time.RFC3339)
	}
	if endISO == "" {
		endISO = now.Add(30 * time.Minute).Format(time.RFC3339)
	}
	return startISO, endISO
}

func extractMeetURL(ev *calendar.Event) string {
	if ev.HangoutLink != "" {
		return ev.HangoutLink
	}
	if ev.ConferenceData != nil {
		for _, ep := range ev.ConferenceData.EntryPoints {
			if ep.EntryPointType == "video" {
				return ep.Uri
			}
		}
	}
	return ""
}
