package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/erikkubica/gws/internal/services/calendar"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerCalendarTools(ctx context.Context, s *server.MCPServer) {
	s.AddTool(buildCalListTool(), handleListEvents(ctx))
	s.AddTool(buildCalAddTool(), handleQuickAddEvent(ctx))
	s.AddTool(buildCalCreateTool(), handleCreateEvent(ctx))
	s.AddTool(buildCalDeleteTool(), handleDeleteEvent(ctx))
	s.AddTool(buildCalRespondTool(), handleRespondEvent(ctx))
}

func buildCalListTool() mcp.Tool {
	return mcp.NewTool("calendar_list_events",
		mcp.WithDescription("List upcoming events from Google Calendar"),
		mcp.WithNumber("max", mcp.Description("Max number of events (default 10)")),
		mcp.WithString("calendar_id", mcp.Description("Calendar ID (default 'primary')")),
		accountOption(),
	)
}

func buildCalAddTool() mcp.Tool {
	return mcp.NewTool("calendar_quick_add",
		mcp.WithDescription("Add a calendar event using natural language (e.g. 'Lunch with John tomorrow at 1pm')"),
		mcp.WithString("text", mcp.Required(), mcp.Description("Natural language event description")),
		mcp.WithString("calendar_id", mcp.Description("Calendar ID (default 'primary')")),
		accountOption(),
	)
}

func buildCalCreateTool() mcp.Tool {
	return mcp.NewTool("calendar_create_event",
		mcp.WithDescription("Create a structured calendar event with explicit start and end times"),
		mcp.WithString("title", mcp.Required(), mcp.Description("Event title")),
		mcp.WithString("start", mcp.Required(), mcp.Description("Start time in RFC3339 (e.g. '2026-10-06T10:00:00+07:00')")),
		mcp.WithString("end", mcp.Required(), mcp.Description("End time in RFC3339")),
		mcp.WithString("description", mcp.Description("Event description")),
		mcp.WithString("location", mcp.Description("Event location")),
		mcp.WithBoolean("with_meet", mcp.Description("Generate Google Meet video conference link")),
		mcp.WithString("attendees", mcp.Description("Comma-separated attendee email addresses")),
		mcp.WithString("calendar_id", mcp.Description("Calendar ID (default 'primary')")),
		accountOption(),
	)
}

func buildCalRespondTool() mcp.Tool {
	return mcp.NewTool("calendar_respond_event",
		mcp.WithDescription("Respond to a calendar event invitation (RSVP)"),
		mcp.WithString("event_id", mcp.Required(), mcp.Description("The ID of the event to respond to")),
		mcp.WithString("response", mcp.Required(), mcp.Description("RSVP response: 'accepted', 'declined', or 'tentative'")),
		mcp.WithString("calendar_id", mcp.Description("Calendar ID (default 'primary')")),
		accountOption(),
	)
}

func buildCalDeleteTool() mcp.Tool {
	return mcp.NewTool("calendar_delete_event",
		mcp.WithDescription("Delete a calendar event by ID"),
		mcp.WithString("event_id", mcp.Required(), mcp.Description("The ID of the event to delete")),
		mcp.WithString("calendar_id", mcp.Description("Calendar ID (default 'primary')")),
		accountOption(),
	)
}

func handleListEvents(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := calendar.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		max := int64(req.GetInt("max", 10))
		calID := req.GetString("calendar_id", "primary")
		events, err := svc.ListUpcomingEvents(calID, max)
		if err != nil {
			return mcp.NewToolResultError("calendar error: " + err.Error()), nil
		}
		b, _ := json.MarshalIndent(events, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	}
}

func handleQuickAddEvent(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := calendar.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		text, err := req.RequireString("text")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		calID := req.GetString("calendar_id", "primary")
		event, err := svc.QuickAddEvent(calID, text)
		if err != nil {
			return mcp.NewToolResultError("add event error: " + err.Error()), nil
		}
		return mcp.NewToolResultText("Event created: " + event.Summary + " (Link: " + event.HtmlLink + ")"), nil
	}
}

func handleCreateEvent(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := calendar.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		title, _ := req.RequireString("title")
		start, _ := req.RequireString("start")
		end, _ := req.RequireString("end")
		opts := calendar.EventOptions{
			CalendarID:  req.GetString("calendar_id", "primary"),
			Title:       title,
			Description: req.GetString("description", ""),
			Location:    req.GetString("location", ""),
			Start:       start,
			End:         end,
			WithMeet:    req.GetBool("with_meet", false),
			Attendees:   parseAttendeesList(req.GetString("attendees", "")),
			SendUpdates: "all",
		}
		ev, err := svc.CreateEventWithOptions(opts)
		if err != nil {
			return mcp.NewToolResultError("create event error: " + err.Error()), nil
		}
		msg := fmt.Sprintf("Event '%s' created (ID: %s, Link: %s)", ev.Summary, ev.Id, ev.HtmlLink)
		if ev.HangoutLink != "" {
			msg += "\nGoogle Meet: " + ev.HangoutLink
		}
		return mcp.NewToolResultText(msg), nil
	}
}

func parseAttendeesList(raw string) []string {
	var attendees []string
	if raw == "" {
		return attendees
	}
	for _, a := range strings.Split(raw, ",") {
		if s := strings.TrimSpace(a); s != "" {
			attendees = append(attendees, s)
		}
	}
	return attendees
}

func handleRespondEvent(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := calendar.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		id, _ := req.RequireString("event_id")
		status, _ := req.RequireString("response")
		calID := req.GetString("calendar_id", "primary")
		ev, err := svc.RespondToEvent(calID, id, status, "all")
		if err != nil {
			return mcp.NewToolResultError("respond error: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("RSVP for '%s' set to: %s", ev.Summary, status)), nil
	}
}

func handleDeleteEvent(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := calendar.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		id, err := req.RequireString("event_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		calID := req.GetString("calendar_id", "primary")
		if err := svc.DeleteEvent(calID, id); err != nil {
			return mcp.NewToolResultError("delete event error: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Calendar event %s deleted successfully", id)), nil
	}
}
