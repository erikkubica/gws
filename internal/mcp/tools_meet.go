package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/erikkubica/gws/internal/services/calendar"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerMeetTools(ctx context.Context, s *server.MCPServer) {
	s.AddTool(buildMeetCreateTool(), handleCreateMeet(ctx))
}

func buildMeetCreateTool() mcp.Tool {
	return mcp.NewTool("meet_create_session",
		mcp.WithDescription("Provision a Google Meet video conference with link and calendar invitation"),
		mcp.WithString("title", mcp.Required(), mcp.Description("Meeting title")),
		mcp.WithString("start", mcp.Description("Start time in RFC3339 (defaults to now)")),
		mcp.WithString("end", mcp.Description("End time in RFC3339 (defaults to start + 30m)")),
		mcp.WithString("attendees", mcp.Description("Comma-separated attendee emails to invite")),
	)
}

func handleCreateMeet(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := calendar.NewService(ctx)
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		title, _ := req.RequireString("title")
		start := req.GetString("start", "")
		end := req.GetString("end", "")
		rawAtt := req.GetString("attendees", "")
		var att []string
		if rawAtt != "" {
			for _, a := range strings.Split(rawAtt, ",") {
				if s := strings.TrimSpace(a); s != "" {
					att = append(att, s)
				}
			}
		}
		ev, meetURL, err := svc.CreateQuickMeet(title, start, end, att)
		if err != nil {
			return mcp.NewToolResultError("create meet error: " + err.Error()), nil
		}
		msg := fmt.Sprintf("Google Meet Ready!\nURL: %s\nEvent ID: %s\nCalendar: %s", meetURL, ev.Id, ev.HtmlLink)
		return mcp.NewToolResultText(msg), nil
	}
}
