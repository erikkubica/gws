package cli

import (
	"context"
	"fmt"

	"github.com/erikkubica/gmcp/internal/mcp"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Model Context Protocol (MCP) server over stdio",
	Long:  "Launches the stdio-based MCP server for Claude Desktop, Cursor, Antigravity, and Zed.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		s, err := mcp.NewServer(ctx)
		if err != nil {
			return fmt.Errorf("initialize mcp server: %w", err)
		}
		return s.ServeStdio()
	},
}
