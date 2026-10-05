package mcp

import (
	"context"
	"encoding/json"

	"github.com/erikkubica/gmcp/internal/services/youtube"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerYouTubeTools(ctx context.Context, s *server.MCPServer) {
	s.AddTool(buildYTSearchTool(), handleYTSearch(ctx))
	s.AddTool(buildYTDetailsTool(), handleYTDetails(ctx))
}

func buildYTSearchTool() mcp.Tool {
	return mcp.NewTool("youtube_search",
		mcp.WithDescription("Search videos on YouTube"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search terms")),
		mcp.WithNumber("max", mcp.Description("Max number of videos (default 10)")),
	)
}

func buildYTDetailsTool() mcp.Tool {
	return mcp.NewTool("youtube_video_details",
		mcp.WithDescription("Retrieve stats, views, and description for a specific YouTube video"),
		mcp.WithString("video_id", mcp.Required(), mcp.Description("The YouTube video ID")),
	)
}

func handleYTSearch(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := youtube.NewService(ctx)
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		q, err := req.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		max := int64(req.GetInt("max", 10))
		videos, err := svc.SearchVideos(q, max)
		if err != nil {
			return mcp.NewToolResultError("youtube search error: " + err.Error()), nil
		}
		b, _ := json.MarshalIndent(videos, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	}
}

func handleYTDetails(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := youtube.NewService(ctx)
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		id, err := req.RequireString("video_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		details, err := svc.GetVideoDetails(id)
		if err != nil {
			return mcp.NewToolResultError("youtube details error: " + err.Error()), nil
		}
		b, _ := json.MarshalIndent(details, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	}
}
