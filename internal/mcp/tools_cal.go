package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erikkubica/gmcp/internal/services/calendar"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerCalendarTools(ctx context.Context, s *server.MCPServer) {
	s.AddTool(buildCalListTool(), handleListEvents(ctx))
	s.AddTool(buildCalAddTool(), handleQuickAddEvent(ctx))
	s.AddTool(buildCalCreateTool(), handleCreateEvent(ctx))
	s.AddTool(buildCalDeleteTool(), handleDeleteEvent(ctx))
}

func buildCalListTool() mcp.Tool {
	return mcp.NewTool("calendar_list_events",
		mcp.WithDescription("List upcoming events from Google Calendar"),
		mcp.WithNumber("max", mcp.Description("Max number of events (default 10)")),
		mcp.WithString("calendar_id", mcp.Description("Calendar ID (default 'primary')")),
	)
}

func buildCalAddTool() mcp.Tool {
	return mcp.NewTool("calendar_quick_add",
		mcp.WithDescription("Add a calendar event using natural language (e.g. 'Lunch with John tomorrow at 1pm')"),
		mcp.WithString("text", mcp.Required(), mcp.Description("Natural language event description")),
		mcp.WithString("calendar_id", mcp.Description("Calendar ID (default 'primary')")),
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
		mcp.WithString("calendar_id", mcp.Description("Calendar ID (default 'primary')")),
	)
}

func buildCalDeleteTool() mcp.Tool {
	return mcp.NewTool("calendar_delete_event",
		mcp.WithDescription("Delete a calendar event by ID"),
		mcp.WithString("event_id", mcp.Required(), mcp.Description("The ID of the event to delete")),
		mcp.WithString("calendar_id", mcp.Description("Calendar ID (default 'primary')")),
	)
}

func handleListEvents(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := calendar.NewService(ctx)
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
		svc, err := calendar.NewService(ctx)
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
		svc, err := calendar.NewService(ctx)
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		title, _ := req.RequireString("title")
		start, _ := req.RequireString("start")
		end, _ := req.RequireString("end")
		desc := req.GetString("description", "")
		loc := req.GetString("location", "")
		calID := req.GetString("calendar_id", "primary")

		event, err := svc.CreateEvent(calID, title, desc, loc, start, end)
		if err != nil {
			return mcp.NewToolResultError("create event error: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Event '%s' created (ID: %s, Link: %s)", event.Summary, event.Id, event.HtmlLink)), nil
	}
}

func handleDeleteEvent(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := calendar.NewService(ctx)
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
