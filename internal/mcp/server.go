package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/server"
)

// Server encapsulates the MCP server instance and registered tools.
type Server struct {
	mcpServer *server.MCPServer
}

// NewServer builds and registers all tool suites for Google Workspace services.
func NewServer(ctx context.Context) (*Server, error) {
	s := server.NewMCPServer("gmcp", "1.0.0", server.WithResourceCapabilities(true, true))

	registerGmailTools(ctx, s)
	registerCalendarTools(ctx, s)
	registerDriveTools(ctx, s)
	registerYouTubeTools(ctx, s)
	registerTasksTools(ctx, s)
	registerSheetsTools(ctx, s)

	return &Server{mcpServer: s}, nil
}

// ServeStdio executes the server on standard input/output.
func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.mcpServer)
}
