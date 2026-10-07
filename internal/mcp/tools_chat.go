package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/erikkubica/gws/internal/services/chat"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerChatTools(s *server.MCPServer) {
	s.AddTool(buildChatListSpacesTool(), handleChatListSpaces())
	s.AddTool(buildChatSendMessageTool(), handleChatSendMessage())
	s.AddTool(buildChatListMessagesTool(), handleChatListMessages())
	s.AddTool(buildChatSearchMessagesTool(), handleChatSearchMessages())
	s.AddTool(buildChatReactTool(), handleChatReact())
	s.AddTool(buildChatListReactionsTool(), handleChatListReactions())
}

func buildChatListSpacesTool() mcp.Tool {
	return mcp.NewTool("chat_list_spaces",
		mcp.WithDescription("List or search joined Google Chat spaces, rooms, and direct message conversations"),
		mcp.WithString("query", mcp.Description("Optional filter keyword matching space name or participant name")),
		mcp.WithNumber("max", mcp.Description("Max spaces to return (default 20)")),
		accountOption(),
	)
}

func buildChatSendMessageTool() mcp.Tool {
	return mcp.NewTool("chat_send_message",
		mcp.WithDescription("Send a message, reply to a thread, or upload an attachment to a Google Chat space"),
		mcp.WithString("space", mcp.Required(), mcp.Description("Space name or ID (e.g. 'spaces/AAAA...')")),
		mcp.WithString("text", mcp.Required(), mcp.Description("Message plain text")),
		mcp.WithString("reply_to", mcp.Description("Optional parent message ID to reply to in a thread")),
		mcp.WithString("attachment", mcp.Description("Optional absolute local file path to attach")),
		accountOption(),
	)
}

func buildChatReactTool() mcp.Tool {
	return mcp.NewTool("chat_react_message",
		mcp.WithDescription("Add an emoji reaction to a Google Chat message (e.g. '👍', '❤️', '🔥')"),
		mcp.WithString("message_name", mcp.Required(), mcp.Description("Full message resource name or ID")),
		mcp.WithString("emoji", mcp.Required(), mcp.Description("Emoji character")),
		accountOption(),
	)
}

func buildChatListReactionsTool() mcp.Tool {
	return mcp.NewTool("chat_list_reactions",
		mcp.WithDescription("List emoji reactions on a Google Chat message"),
		mcp.WithString("message_name", mcp.Required(), mcp.Description("Full message resource name or ID")),
		accountOption(),
	)
}

func buildChatListMessagesTool() mcp.Tool {
	return mcp.NewTool("chat_list_messages",
		mcp.WithDescription("List messages from a Google Chat space (newest first by default)"),
		mcp.WithString("space", mcp.Required(), mcp.Description("Space resource name or ID (e.g. 'spaces/AAAA...')")),
		mcp.WithNumber("max", mcp.Description("Max messages to return (default 20)")),
		mcp.WithString("order", mcp.Description("Order by create time: 'DESC' (newest first, default) or 'ASC' (oldest first)")),
		accountOption(),
	)
}

func buildChatSearchMessagesTool() mcp.Tool {
	return mcp.NewTool("chat_search_messages",
		mcp.WithDescription("Search messages across all Google Chat spaces and conversations"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search filter or keyword across messages (e.g. 'godot', 'has_link()')")),
		mcp.WithNumber("max", mcp.Description("Max messages to return (default 20)")),
		accountOption(),
	)
}

func handleChatListSpaces() server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := chat.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		max := int64(req.GetInt("max", 20))
		q := req.GetString("query", "")
		spaces, err := svc.SearchSpaces(q, max)
		if err != nil {
			return mcp.NewToolResultError("chat error: " + err.Error()), nil
		}
		b, _ := json.MarshalIndent(spaces, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	}
}

func handleChatSearchMessages() server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := chat.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		q, err := req.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		max := int64(req.GetInt("max", 20))
		msgs, err := svc.SearchMessages(q, max)
		if err != nil {
			return mcp.NewToolResultError("search error: " + err.Error()), nil
		}
		b, _ := json.MarshalIndent(msgs, "", "  ")
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
		opts := chat.MessageOptions{
			SpaceName:  space,
			Text:       text,
			ReplyTo:    req.GetString("reply_to", ""),
			Attachment: req.GetString("attachment", ""),
		}
		msg, err := svc.SendMessageWithOptions(opts)
		if err != nil {
			return mcp.NewToolResultError("send error: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Message delivered! ID: %s", msg.Name)), nil
	}
}

func handleChatReact() server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := chat.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		msgName, _ := req.RequireString("message_name")
		emoji, _ := req.RequireString("emoji")
		if err := svc.AddReaction(msgName, emoji); err != nil {
			return mcp.NewToolResultError("react error: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Reaction %s added successfully!", emoji)), nil
	}
}

func handleChatListReactions() server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := chat.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		msgName, _ := req.RequireString("message_name")
		reactions, err := svc.ListReactions(msgName)
		if err != nil {
			return mcp.NewToolResultError("list reactions error: " + err.Error()), nil
		}
		b, _ := json.MarshalIndent(reactions, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
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
		order := strings.ToUpper(req.GetString("order", "DESC"))
		msgs, err := svc.ListMessagesWithOrder(space, max, order)
		if err != nil {
			return mcp.NewToolResultError("list messages error: " + err.Error()), nil
		}
		b, _ := json.MarshalIndent(msgs, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	}
}
