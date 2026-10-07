package cli

import (
	"fmt"

	"github.com/erikkubica/gws/internal/mcp"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Model Context Protocol (MCP) server over stdio",
	Long:  "Launches the stdio-based MCP server for Claude Desktop, Cursor, Antigravity, and Zed.",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := mcp.NewServer(cmd.Context())
		if err != nil {
			return fmt.Errorf("initialize mcp server: %w", err)
		}
		return s.ServeStdio()
	},
}
