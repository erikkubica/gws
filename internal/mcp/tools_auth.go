package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erikkubica/gws/internal/auth"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerAccountTools(s *server.MCPServer) {
	s.AddTool(buildListAccountsTool(), handleListAccounts())
	s.AddTool(buildSwitchAccountTool(), handleSwitchAccount())
}

func buildListAccountsTool() mcp.Tool {
	return mcp.NewTool("workspace_list_accounts",
		mcp.WithDescription("List all authenticated Google accounts in gws and show which one is currently active"),
	)
}

func handleListAccounts() server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		accounts, err := auth.ListAccounts()
		if err != nil {
			return mcp.NewToolResultError("failed to list accounts: " + err.Error()), nil
		}
		b, _ := json.MarshalIndent(accounts, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	}
}

func buildSwitchAccountTool() mcp.Tool {
	return mcp.NewTool("workspace_switch_account",
		mcp.WithDescription("Switch the active default Google account for all workspace operations"),
		mcp.WithString("account", mcp.Required(), mcp.Description("The Google account email to activate")),
	)
}

func handleSwitchAccount() server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		target, err := req.RequireString("account")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if err := auth.SetActiveAccount(target); err != nil {
			return mcp.NewToolResultError("switch failed: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Successfully switched active account to: %s", target)), nil
	}
}
