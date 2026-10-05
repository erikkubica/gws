package mcp

import (
	"context"
	"encoding/json"

	"github.com/erikkubica/gmcp/internal/services/gmail"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerGmailTools(ctx context.Context, s *server.MCPServer) {
	s.AddTool(buildListTool(), handleListMessages(ctx))
	s.AddTool(buildGetTool(), handleGetMessage(ctx))
	s.AddTool(buildSendTool(), handleSendMessage(ctx))
	s.AddTool(buildDraftTool(), handleCreateDraft(ctx))
}

func buildListTool() mcp.Tool {
	return mcp.NewTool("gmail_list_messages",
		mcp.WithDescription("List and search messages in Gmail inbox"),
		mcp.WithString("query", mcp.Description("Gmail search query (e.g. 'is:unread', 'from:someone@domain.com')")),
		mcp.WithNumber("max", mcp.Description("Max number of messages to return (default 10)")),
	)
}

func buildGetTool() mcp.Tool {
	return mcp.NewTool("gmail_get_message",
		mcp.WithDescription("Read full content of a specific Gmail message by ID"),
		mcp.WithString("id", mcp.Required(), mcp.Description("The unique Gmail message ID")),
	)
}

func buildSendTool() mcp.Tool {
	return mcp.NewTool("gmail_send_message",
		mcp.WithDescription("Send an email from the authenticated user's account"),
		mcp.WithString("to", mcp.Required(), mcp.Description("Recipient email address")),
		mcp.WithString("subject", mcp.Required(), mcp.Description("Email subject line")),
		mcp.WithString("body", mcp.Required(), mcp.Description("Email plain text body")),
	)
}

func buildDraftTool() mcp.Tool {
	return mcp.NewTool("gmail_create_draft",
		mcp.WithDescription("Create a new draft in Gmail without sending"),
		mcp.WithString("to", mcp.Required(), mcp.Description("Recipient email address")),
		mcp.WithString("subject", mcp.Required(), mcp.Description("Email subject line")),
		mcp.WithString("body", mcp.Required(), mcp.Description("Email plain text body")),
	)
}

func handleListMessages(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := gmail.NewService(ctx)
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		q := req.GetString("query", "")
		max := int64(req.GetInt("max", 10))
		msgs, err := svc.ListMessages(q, max)
		if err != nil {
			return mcp.NewToolResultError("list failed: " + err.Error()), nil
		}
		b, _ := json.MarshalIndent(msgs, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	}
}

func handleGetMessage(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := gmail.NewService(ctx)
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		msg, err := svc.GetMessage(id)
		if err != nil {
			return mcp.NewToolResultError("fetch failed: " + err.Error()), nil
		}
		b, _ := json.MarshalIndent(msg, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	}
}

func handleSendMessage(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := gmail.NewService(ctx)
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		to, _ := req.RequireString("to")
		subj, _ := req.RequireString("subject")
		body, _ := req.RequireString("body")
		res, err := svc.SendMessage(to, subj, body)
		if err != nil {
			return mcp.NewToolResultError("send failed: " + err.Error()), nil
		}
		return mcp.NewToolResultText("Message sent successfully. ID: " + res.Id), nil
	}
}

func handleCreateDraft(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := gmail.NewService(ctx)
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		to, _ := req.RequireString("to")
		subj, _ := req.RequireString("subject")
		body, _ := req.RequireString("body")
		draft, err := svc.CreateDraft(to, subj, body)
		if err != nil {
			return mcp.NewToolResultError("draft failed: " + err.Error()), nil
		}
		return mcp.NewToolResultText("Draft created successfully. ID: " + draft.Id), nil
	}
}
