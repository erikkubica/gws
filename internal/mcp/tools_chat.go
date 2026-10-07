package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erikkubica/gws/internal/services/chat"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerChatTools(s *server.MCPServer) {
	s.AddTool(buildChatListSpacesTool(), handleChatListSpaces())
	s.AddTool(buildChatSendMessageTool(), handleChatSendMessage())
	s.AddTool(buildChatListMessagesTool(), handleChatListMessages())
	s.AddTool(buildChatSendWebhookTool(), handleChatSendWebhook())
}

func buildChatListSpacesTool() mcp.Tool {
	return mcp.NewTool("chat_list_spaces",
		mcp.WithDescription("List joined Google Chat spaces, rooms, and direct message conversations"),
		mcp.WithNumber("max", mcp.Description("Max spaces to return (default 20)")),
		accountOption(),
	)
}

func buildChatSendMessageTool() mcp.Tool {
	return mcp.NewTool("chat_send_message",
		mcp.WithDescription("Send a text message to a Google Chat space or direct message room"),
		mcp.WithString("space", mcp.Required(), mcp.Description("Space resource name or ID (e.g. 'spaces/AAAA...' or 'AAAA...')")),
		mcp.WithString("text", mcp.Required(), mcp.Description("Message plain text")),
		accountOption(),
	)
}

func buildChatListMessagesTool() mcp.Tool {
	return mcp.NewTool("chat_list_messages",
		mcp.WithDescription("List recent messages from a Google Chat space"),
		mcp.WithString("space", mcp.Required(), mcp.Description("Space resource name or ID (e.g. 'spaces/AAAA...')")),
		mcp.WithNumber("max", mcp.Description("Max messages to return (default 20)")),
		accountOption(),
	)
}

func buildChatSendWebhookTool() mcp.Tool {
	return mcp.NewTool("chat_send_webhook",
		mcp.WithDescription("Post a message to a Google Chat incoming webhook URL"),
		mcp.WithString("webhook_url", mcp.Required(), mcp.Description("Incoming webhook URL")),
		mcp.WithString("text", mcp.Required(), mcp.Description("Message plain text")),
	)
}

func handleChatListSpaces() server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := chat.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		max := int64(req.GetInt("max", 20))
		spaces, err := svc.ListSpaces(max)
		if err != nil {
			return mcp.NewToolResultError("chat error: " + err.Error()), nil
		}
		b, _ := json.MarshalIndent(spaces, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	}
}

func handleChatSendMessage() server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := chat.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		space, _ := req.RequireString("space")
		text, _ := req.RequireString("text")
		msg, err := svc.SendMessage(space, text)
		if err != nil {
			return mcp.NewToolResultError("send error: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Message delivered to %s! ID: %s", space, msg.Name)), nil
	}
}

func handleChatListMessages() server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := chat.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		space, _ := req.RequireString("space")
		max := int64(req.GetInt("max", 20))
		msgs, err := svc.ListMessages(space, max)
		if err != nil {
			return mcp.NewToolResultError("list messages error: " + err.Error()), nil
		}
		b, _ := json.MarshalIndent(msgs, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	}
}

func handleChatSendWebhook() server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := chat.NewService(c)
		if err != nil {
			return mcp.NewToolResultError("chat service error: " + err.Error()), nil
		}
		url, _ := req.RequireString("webhook_url")
		text, _ := req.RequireString("text")
		if err := svc.SendWebhook(url, text); err != nil {
			return mcp.NewToolResultError("webhook error: " + err.Error()), nil
		}
		return mcp.NewToolResultText("Webhook message delivered successfully!"), nil
	}
}
