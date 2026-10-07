package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erikkubica/gws/internal/services/gmail"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerGmailTools(ctx context.Context, s *server.MCPServer) {
	s.AddTool(buildListTool(), handleListMessages(ctx))
	s.AddTool(buildGetTool(), handleGetMessage(ctx))
	s.AddTool(buildSendTool(), handleSendMessage(ctx))
	s.AddTool(buildReplyTool(), handleReplyMessage(ctx))
	s.AddTool(buildDraftTool(), handleCreateDraft(ctx))
	s.AddTool(buildDraftsListTool(), handleListDrafts(ctx))
	s.AddTool(buildSendDraftTool(), handleSendDraft(ctx))
	s.AddTool(buildDownloadAttachmentTool(), handleDownloadAttachment(ctx))
}

func buildListTool() mcp.Tool {
	return mcp.NewTool("gmail_list_messages",
		mcp.WithDescription("List and search messages in Gmail inbox"),
		mcp.WithString("query", mcp.Description("Gmail search query (e.g. 'is:unread', 'from:someone@domain.com')")),
		mcp.WithNumber("max", mcp.Description("Max number of messages to return (default 10)")),
		accountOption(),
	)
}

func buildGetTool() mcp.Tool {
	return mcp.NewTool("gmail_get_message",
		mcp.WithDescription("Read full content of a specific Gmail message by ID"),
		mcp.WithString("id", mcp.Required(), mcp.Description("The unique Gmail message ID")),
		accountOption(),
	)
}

func buildSendTool() mcp.Tool {
	return mcp.NewTool("gmail_send_message",
		mcp.WithDescription("Send an email from the authenticated user's account"),
		mcp.WithString("to", mcp.Required(), mcp.Description("Recipient email address")),
		mcp.WithString("subject", mcp.Required(), mcp.Description("Email subject line")),
		mcp.WithString("body", mcp.Required(), mcp.Description("Email plain text body")),
		mcp.WithString("attachment", mcp.Description("Optional absolute local file path to attach")),
		accountOption(),
	)
}

func buildReplyTool() mcp.Tool {
	return mcp.NewTool("gmail_reply_message",
		mcp.WithDescription("Reply to an existing email message thread"),
		mcp.WithString("message_id", mcp.Required(), mcp.Description("The original message ID to reply to")),
		mcp.WithString("body", mcp.Required(), mcp.Description("Reply body text")),
		mcp.WithString("attachment", mcp.Description("Optional absolute local file path to attach")),
		accountOption(),
	)
}

func buildDraftTool() mcp.Tool {
	return mcp.NewTool("gmail_create_draft",
		mcp.WithDescription("Create a new draft in Gmail without sending"),
		mcp.WithString("to", mcp.Required(), mcp.Description("Recipient email address")),
		mcp.WithString("subject", mcp.Required(), mcp.Description("Email subject line")),
		mcp.WithString("body", mcp.Required(), mcp.Description("Email plain text body")),
		mcp.WithString("attachment", mcp.Description("Optional absolute local file path to attach")),
		accountOption(),
	)
}

func buildDraftsListTool() mcp.Tool {
	return mcp.NewTool("gmail_list_drafts",
		mcp.WithDescription("List existing draft messages in Gmail"),
		mcp.WithNumber("max", mcp.Description("Max number of drafts to return (default 10)")),
		accountOption(),
	)
}

func buildSendDraftTool() mcp.Tool {
	return mcp.NewTool("gmail_send_draft",
		mcp.WithDescription("Send an existing draft email by its draft ID"),
		mcp.WithString("draft_id", mcp.Required(), mcp.Description("The ID of the draft to send")),
		accountOption(),
	)
}

func handleListMessages(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := gmail.NewService(withAccountContext(c, req))
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
		svc, err := gmail.NewService(withAccountContext(c, req))
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
		svc, err := gmail.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		opts, err := parseEmailOptions(req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		res, err := svc.SendMessage(opts)
		if err != nil {
			return mcp.NewToolResultError("send failed: " + err.Error()), nil
		}
		return mcp.NewToolResultText("Message sent successfully. ID: " + res.Id), nil
	}
}

func parseEmailOptions(req mcp.CallToolRequest) (gmail.EmailOptions, error) {
	to, err := req.RequireString("to")
	if err != nil {
		return gmail.EmailOptions{}, err
	}
	subj, err := req.RequireString("subject")
	if err != nil {
		return gmail.EmailOptions{}, err
	}
	body, err := req.RequireString("body")
	if err != nil {
		return gmail.EmailOptions{}, err
	}
	return gmail.EmailOptions{
		To:          to,
		Subject:     subj,
		Body:        body,
		Attachments: parseAttachmentOption(req),
	}, nil
}

func parseAttachmentOption(req mcp.CallToolRequest) []string {
	if a := req.GetString("attachment", ""); a != "" {
		return []string{a}
	}
	return nil
}

func handleReplyMessage(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := gmail.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		msgID, err := req.RequireString("message_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		body, err := req.RequireString("body")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		res, err := svc.ReplyMessage(msgID, body, parseAttachmentOption(req))
		if err != nil {
			return mcp.NewToolResultError("reply failed: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Reply sent successfully. ID: %s (Thread: %s)", res.Id, res.ThreadId)), nil
	}
}

func handleCreateDraft(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := gmail.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		opts, err := parseEmailOptions(req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		draft, err := svc.CreateDraft(opts)
		if err != nil {
			return mcp.NewToolResultError("draft failed: " + err.Error()), nil
		}
		return mcp.NewToolResultText("Draft created successfully. ID: " + draft.Id), nil
	}
}

func handleListDrafts(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := gmail.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		max := int64(req.GetInt("max", 10))
		drafts, err := svc.ListDrafts(max)
		if err != nil {
			return mcp.NewToolResultError("list drafts failed: " + err.Error()), nil
		}
		b, _ := json.MarshalIndent(drafts, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	}
}

func handleSendDraft(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := gmail.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		id, err := req.RequireString("draft_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		msg, err := svc.SendDraft(id)
		if err != nil {
			return mcp.NewToolResultError("send draft failed: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Draft sent successfully. ID: %s", msg.Id)), nil
	}
}

func buildDownloadAttachmentTool() mcp.Tool {
	return mcp.NewTool("gmail_download_attachment",
		mcp.WithDescription("Download an attachment from a Gmail message to a local file"),
		mcp.WithString("message_id", mcp.Required(), mcp.Description("The message ID containing the attachment")),
		mcp.WithString("attachment_id", mcp.Required(), mcp.Description("The attachment ID or filename to download")),
		mcp.WithString("destination_path", mcp.Description("Local destination file or directory path (defaults to current directory)")),
		accountOption(),
	)
}

func handleDownloadAttachment(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := gmail.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		msgID, err := req.RequireString("message_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		attID, err := req.RequireString("attachment_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		dest := req.GetString("destination_path", ".")
		file, err := svc.SaveAttachment(msgID, attID, dest)
		if err != nil {
			return mcp.NewToolResultError("download failed: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved attachment '%s' (%d bytes) to %s", file.Filename, file.Size, dest)), nil
	}
}
