package calendar

import (
	"fmt"

	"google.golang.org/api/calendar/v3"
)

// RespondToEvent updates the current user's RSVP status for an event.
// Valid statuses: "accepted", "declined", "tentative", "needsAction".
func (s *Service) RespondToEvent(calID, eventID, status, sendUpdates string) (*calendar.Event, error) {
	if calID == "" {
		calID = "primary"
	}
	if sendUpdates == "" {
		sendUpdates = "all"
	}
	ev, err := s.client.Events.Get(calID, eventID).Do()
	if err != nil {
		return nil, fmt.Errorf("fetch event %s: %w", eventID, err)
	}
	patch := &calendar.Event{Attendees: updateSelfResponse(ev, status)}
	res, err := s.client.Events.Patch(calID, eventID, patch).SendUpdates(sendUpdates).Do()
	if err != nil {
		return nil, fmt.Errorf("update rsvp status for %s: %w", eventID, err)
	}
	return res, nil
}

func updateSelfResponse(ev *calendar.Event, status string) []*calendar.EventAttendee {
	selfEmail := getSelfEmail(ev)
	found := false
	for _, a := range ev.Attendees {
		if a.Self || (selfEmail != "" && a.Email == selfEmail) {
			a.ResponseStatus = status
			found = true
			break
		}
	}
	if !found && selfEmail != "" {
		ev.Attendees = append(ev.Attendees, &calendar.EventAttendee{
			Email:          selfEmail,
			Self:           true,
			ResponseStatus: status,
		})
	}
	return ev.Attendees
}

func getSelfEmail(ev *calendar.Event) string {
	if ev.Organizer != nil && ev.Organizer.Self {
		return ev.Organizer.Email
	}
	if ev.Creator != nil && ev.Creator.Self {
		return ev.Creator.Email
	}
	if ev.Organizer != nil && ev.Organizer.Email != "" {
		return ev.Organizer.Email
	}
	if ev.Creator != nil && ev.Creator.Email != "" {
		return ev.Creator.Email
	}
	return ""
}
