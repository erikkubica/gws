package mcp

import (
	"context"
	"fmt"

	"github.com/erikkubica/gws/internal/services/docs"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerDocsTools(ctx context.Context, s *server.MCPServer) {
	s.AddTool(buildDocsCreateTool(), handleCreateDoc(ctx))
	s.AddTool(buildDocsAppendTool(), handleAppendDoc(ctx))
	s.AddTool(buildDocsReadTool(), handleReadDoc(ctx))
}

func buildDocsCreateTool() mcp.Tool {
	return mcp.NewTool("docs_create_document",
		mcp.WithDescription("Create a new Google Document"),
		mcp.WithString("title", mcp.Required(), mcp.Description("Title of the new document")),
		accountOption(),
	)
}

func buildDocsAppendTool() mcp.Tool {
	return mcp.NewTool("docs_append_text",
		mcp.WithDescription("Append text to an existing Google Document"),
		mcp.WithString("doc_id", mcp.Required(), mcp.Description("The ID of the document")),
		mcp.WithString("text", mcp.Required(), mcp.Description("Text to append to the document")),
		accountOption(),
	)
}

func buildDocsReadTool() mcp.Tool {
	return mcp.NewTool("docs_read_document",
		mcp.WithDescription("Read full text content of a Google Document"),
		mcp.WithString("doc_id", mcp.Required(), mcp.Description("The ID of the document")),
		accountOption(),
	)
}

func handleCreateDoc(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := docs.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		title, err := req.RequireString("title")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		doc, err := svc.CreateDocument(title)
		if err != nil {
			return mcp.NewToolResultError("create doc error: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Google Doc '%s' created (ID: %s, URL: https://docs.google.com/document/d/%s/edit)",
			doc.Title, doc.DocumentId, doc.DocumentId)), nil
	}
}

func handleAppendDoc(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := docs.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		docID, err := req.RequireString("doc_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		text, err := req.RequireString("text")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if err := svc.AppendText(docID, text); err != nil {
			return mcp.NewToolResultError("append text error: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Text appended successfully to document %s", docID)), nil
	}
}

func handleReadDoc(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := docs.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		docID, err := req.RequireString("doc_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		text, err := svc.GetDocumentText(docID)
		if err != nil {
			return mcp.NewToolResultError("read doc error: " + err.Error()), nil
		}
		return mcp.NewToolResultText(text), nil
	}
}
