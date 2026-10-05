package mcp

import (
	"context"
	"encoding/json"

	"github.com/erikkubica/gmcp/internal/services/calendar"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerCalendarTools(ctx context.Context, s *server.MCPServer) {
	s.AddTool(buildCalListTool(), handleListEvents(ctx))
	s.AddTool(buildCalAddTool(), handleQuickAddEvent(ctx))
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
