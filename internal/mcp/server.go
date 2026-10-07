package mcp

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"
	"time"

	"github.com/erikkubica/gws/internal/version"
	"github.com/mark3labs/mcp-go/server"
)

// Server encapsulates the MCP server instance and registered tools.
type Server struct {
	mcpServer *server.MCPServer
}

// NewServer builds and registers all tool suites for Google Workspace services.
func NewServer(ctx context.Context) (*Server, error) {
	s := server.NewMCPServer("gws", version.Version, server.WithResourceCapabilities(true, true))

	registerGmailTools(ctx, s)
	registerCalendarTools(ctx, s)
	registerDriveTools(ctx, s)
	registerYouTubeTools(ctx, s)
	registerTasksTools(ctx, s)
	registerSheetsTools(ctx, s)
	registerDocsTools(ctx, s)
	registerMeetTools(ctx, s)
	registerAccountTools(s)
	registerChatTools(s)

	return &Server{mcpServer: s}, nil
}

// ServeStdio executes the server on standard input/output.
func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.mcpServer)
}

// ServeSSE starts an SSE-based HTTP server with Bearer token authentication.
func (s *Server) ServeSSE(addr, token string) error {
	sseServer := server.NewSSEServer(s.mcpServer)
	handler := authMiddleware(sseServer, token)
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return httpServer.ListenAndServe()
}

func authMiddleware(next http.Handler, validToken string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if validToken != "" {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Unauthorized: missing Bearer token", http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if subtle.ConstantTimeCompare([]byte(token), []byte(validToken)) != 1 {
				http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
