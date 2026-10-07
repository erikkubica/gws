package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erikkubica/gws/internal/services/sheets"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerSheetsTools(ctx context.Context, s *server.MCPServer) {
	s.AddTool(buildSheetsReadTool(), handleReadSheet(ctx))
	s.AddTool(buildSheetsAppendTool(), handleAppendSheet(ctx))
}

func buildSheetsReadTool() mcp.Tool {
	return mcp.NewTool("sheets_read_range",
		mcp.WithDescription("Read cell values from a Google Spreadsheet range (e.g. 'Sheet1!A1:E10')"),
		mcp.WithString("spreadsheet_id", mcp.Required(), mcp.Description("The ID of the Google Spreadsheet")),
		mcp.WithString("range", mcp.Required(), mcp.Description("A1 notation range to read")),
		accountOption(),
	)
}

func buildSheetsAppendTool() mcp.Tool {
	return mcp.NewTool("sheets_append_row",
		mcp.WithDescription("Append a row of values to a Google Spreadsheet table or sheet"),
		mcp.WithString("spreadsheet_id", mcp.Required(), mcp.Description("The ID of the Google Spreadsheet")),
		mcp.WithString("range", mcp.Required(), mcp.Description("A1 notation range or sheet name (e.g. 'Sheet1!A1')")),
		mcp.WithString("values", mcp.Required(), mcp.Description("Comma-separated or JSON array of cell values for the row")),
		accountOption(),
	)
}

func handleReadSheet(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := sheets.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		sheetID, err := req.RequireString("spreadsheet_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		r, err := req.RequireString("range")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		rows, err := svc.ReadRange(sheetID, r)
		if err != nil {
			return mcp.NewToolResultError("read sheet error: " + err.Error()), nil
		}
		b, _ := json.MarshalIndent(rows, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	}
}

func handleAppendSheet(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := sheets.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		sheetID, err := req.RequireString("spreadsheet_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		r, err := req.RequireString("range")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		valStr, err := req.RequireString("values")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		row := parseSheetRowValues(valStr)
		if err := svc.AppendRow(sheetID, r, row); err != nil {
			return mcp.NewToolResultError("append sheet error: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Row appended successfully to spreadsheet %s", sheetID)), nil
	}
}

func parseSheetRowValues(valStr string) []interface{} {
	var row []interface{}
	if err := json.Unmarshal([]byte(valStr), &row); err != nil {
		row = append(row, valStr)
	}
	return row
}
