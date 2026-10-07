package mcp

import (
	"context"

	"github.com/erikkubica/gws/internal/auth"
	"github.com/mark3labs/mcp-go/mcp"
)

func accountOption() mcp.ToolOption {
	return mcp.WithString("account", mcp.Description("Optional Google account email to use (defaults to active account)"))
}

func withAccountContext(reqCtx context.Context, req mcp.CallToolRequest) context.Context {
	acc := req.GetString("account", "")
	if acc != "" {
		return auth.WithAccount(reqCtx, acc)
	}
	return reqCtx
}
