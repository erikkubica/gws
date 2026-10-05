package mcp

import (
	"context"
	"encoding/json"

	"github.com/erikkubica/gmcp/internal/services/drive"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerDriveTools(ctx context.Context, s *server.MCPServer) {
	s.AddTool(buildDriveListTool(), handleListDriveFiles(ctx))
	s.AddTool(buildDriveReadTool(), handleReadDriveFile(ctx))
}

func buildDriveListTool() mcp.Tool {
	return mcp.NewTool("drive_list_files",
		mcp.WithDescription("List and search files in Google Drive"),
		mcp.WithString("query", mcp.Description("Optional filename search term")),
		mcp.WithNumber("max", mcp.Description("Max number of files (default 10)")),
	)
}

func buildDriveReadTool() mcp.Tool {
	return mcp.NewTool("drive_read_file",
		mcp.WithDescription("Read or export text/csv content of a Google Drive file or Doc"),
		mcp.WithString("file_id", mcp.Required(), mcp.Description("The ID of the file to read")),
	)
}

func handleListDriveFiles(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := drive.NewService(ctx)
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		q := req.GetString("query", "")
		max := int64(req.GetInt("max", 10))
		files, err := svc.ListFiles(q, max)
		if err != nil {
			return mcp.NewToolResultError("drive error: " + err.Error()), nil
		}
		b, _ := json.MarshalIndent(files, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	}
}

func handleReadDriveFile(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := drive.NewService(ctx)
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		id, err := req.RequireString("file_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		content, err := svc.ReadFile(id)
		if err != nil {
			return mcp.NewToolResultError("read error: " + err.Error()), nil
		}
		return mcp.NewToolResultText(content), nil
	}
}
