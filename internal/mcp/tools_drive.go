package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erikkubica/gmcp/internal/services/drive"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerDriveTools(ctx context.Context, s *server.MCPServer) {
	s.AddTool(buildDriveListTool(), handleListDriveFiles(ctx))
	s.AddTool(buildDriveReadTool(), handleReadDriveFile(ctx))
	s.AddTool(buildDriveUploadTool(), handleUploadDriveFile(ctx))
	s.AddTool(buildDriveDeleteTool(), handleDeleteDriveFile(ctx))
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

func buildDriveUploadTool() mcp.Tool {
	return mcp.NewTool("drive_upload_file",
		mcp.WithDescription("Upload a local file to Google Drive"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Local file path to upload")),
		mcp.WithString("name", mcp.Description("Custom filename on Google Drive (optional)")),
	)
}

func buildDriveDeleteTool() mcp.Tool {
	return mcp.NewTool("drive_delete_file",
		mcp.WithDescription("Permanently delete a file from Google Drive by ID"),
		mcp.WithString("file_id", mcp.Required(), mcp.Description("The ID of the file to delete")),
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

func handleUploadDriveFile(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := drive.NewService(ctx)
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		path, err := req.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		name := req.GetString("name", "")
		uploaded, err := svc.UploadFile(path, name)
		if err != nil {
			return mcp.NewToolResultError("upload error: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Uploaded '%s' (ID: %s, %d bytes)", uploaded.Name, uploaded.ID, uploaded.Size)), nil
	}
}

func handleDeleteDriveFile(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := drive.NewService(ctx)
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		id, err := req.RequireString("file_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if err := svc.DeleteFile(id); err != nil {
			return mcp.NewToolResultError("delete error: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("File %s deleted successfully", id)), nil
	}
}
