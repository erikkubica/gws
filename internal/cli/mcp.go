package cli

import (
	"context"
	"fmt"

	"github.com/erikkubica/gws/internal/auth"
	"github.com/erikkubica/gws/internal/mcp"
	"github.com/spf13/cobra"
)

var (
	flagMcpPort  int
	flagMcpHost  string
	flagMcpToken string

	flagTokenShow   bool
	flagTokenRevoke bool
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Model Context Protocol (MCP) server and authentication management",
}

var mcpServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Model Context Protocol (MCP) server (stdio by default, or SSE via --port)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		s, err := mcp.NewServer(ctx)
		if err != nil {
			return fmt.Errorf("initialize mcp server: %w", err)
		}
		if flagMcpPort <= 0 {
			return s.ServeStdio()
		}
		return runSSEServer(s)
	},
}

func runSSEServer(s *mcp.Server) error {
	addr := fmt.Sprintf("%s:%d", flagMcpHost, flagMcpPort)
	token := flagMcpToken
	if token == "" {
		token, _ = auth.LoadMCPToken()
	}
	fmt.Printf("Starting gws MCP SSE server on http://%s/sse\n", addr)
	if token != "" {
		fmt.Printf("✔ Authentication enabled: Bearer token required\n")
	} else {
		fmt.Printf("⚠ Warning: No MCP token set. Running without authentication.\n")
	}
	return s.ServeSSE(addr, token)
}

var mcpTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Generate, inspect, or revoke MCP authentication tokens for remote access",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagTokenRevoke {
			return handleTokenRevoke()
		}
		if flagTokenShow {
			return handleTokenShow()
		}
		return handleTokenGenerate()
	},
}

func handleTokenRevoke() error {
	if err := auth.RevokeMCPToken(); err != nil {
		return err
	}
	fmt.Println("✔ Successfully revoked and removed MCP secret token.")
	return nil
}

func handleTokenShow() error {
	token, err := auth.LoadMCPToken()
	if err != nil || token == "" {
		fmt.Println("No MCP token found. Run 'gws mcp token' to generate one.")
		return nil
	}
	path, _ := auth.MCPTokenPath()
	printTokenDetails(token, path)
	return nil
}

func handleTokenGenerate() error {
	token, err := auth.GenerateMCPToken()
	if err != nil {
		return err
	}
	if err := auth.SaveMCPToken(token); err != nil {
		return err
	}
	path, _ := auth.MCPTokenPath()
	fmt.Println("✔ Generated new MCP secret token:")
	printTokenDetails(token, path)
	return nil
}

func printTokenDetails(token, path string) {
	fmt.Printf("\nToken:  %s\n", token)
	fmt.Printf("Stored: %s (mode 0600)\n\n", path)
	fmt.Println("Remote Client Configuration (Claude / Antigravity / Cursor):")
	fmt.Println(`{
  "mcpServers": {
    "gws-remote": {
      "serverUrl": "https://<your-vps-domain>/sse",
      "headers": {
        "Authorization": "Bearer ` + token + `"
      }
    }
  }
}`)
}

func init() {
	mcpServerCmd.Flags().IntVarP(&flagMcpPort, "port", "p", 0, "HTTP port for SSE remote server (default 0 runs stdio mode)")
	mcpServerCmd.Flags().StringVar(&flagMcpHost, "host", "127.0.0.1", "Host address to bind HTTP SSE server")
	mcpServerCmd.Flags().StringVar(&flagMcpToken, "token", "", "Bearer token for authenticating remote MCP requests")

	mcpTokenCmd.Flags().BoolVar(&flagTokenShow, "show", false, "Show existing MCP token without generating a new one")
	mcpTokenCmd.Flags().BoolVar(&flagTokenRevoke, "revoke", false, "Revoke and remove the current MCP token")

	mcpCmd.AddCommand(mcpServerCmd)
	mcpCmd.AddCommand(mcpTokenCmd)
}
