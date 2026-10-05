package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erikkubica/gmcp/internal/services/sheets"
	"github.com/spf13/cobra"
)

var sheetsCmd = &cobra.Command{
	Use:   "sheets",
	Short: "Create, read, append, update, and manage Google Sheets",
}

var sheetsJSON bool

var sheetsCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new Google Spreadsheet",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := sheets.NewService(context.Background())
		if err != nil {
			return err
		}
		ss, err := svc.CreateSpreadsheet(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Spreadsheet created: %s\nID: %s\nURL: %s\n", ss.Properties.Title, ss.SpreadsheetId, ss.SpreadsheetUrl)
		return nil
	},
}

var sheetsAddSheetCmd = &cobra.Command{
	Use:   "add-sheet [spreadsheet_id] [sheet_title]",
	Short: "Add a new sheet/tab to a spreadsheet",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := sheets.NewService(context.Background())
		if err != nil {
			return err
		}
		if err := svc.AddSheet(args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("Sheet '%s' added successfully to %s\n", args[1], args[0])
		return nil
	},
}

var sheetsReadCmd = &cobra.Command{
	Use:   "read [spreadsheet_id] [range]",
	Short: "Read cell range from a spreadsheet (e.g. 'Sheet1!A1:D10')",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := sheets.NewService(context.Background())
		if err != nil {
			return err
		}
		rows, err := svc.ReadRange(args[0], args[1])
		if err != nil {
			return err
		}
		if sheetsJSON {
			b, _ := json.MarshalIndent(rows, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		for _, row := range rows {
			for _, cell := range row {
				fmt.Printf("%-20v ", cell)
			}
			fmt.Println()
		}
		return nil
	},
}

var sheetsAppendCmd = &cobra.Command{
	Use:   "append [spreadsheet_id] [range] [col1] [col2] ...",
	Short: "Append a row of values to a spreadsheet",
	Args:  cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := sheets.NewService(context.Background())
		if err != nil {
			return err
		}
		sheetID := args[0]
		targetRange := args[1]
		var row []interface{}
		for _, val := range args[2:] {
			row = append(row, val)
		}
		if err := svc.AppendRow(sheetID, targetRange, row); err != nil {
			return err
		}
		fmt.Printf("Row appended successfully to %s (%s)\n", sheetID, targetRange)
		return nil
	},
}

var sheetsUpdateCmd = &cobra.Command{
	Use:   "update [spreadsheet_id] [range] [col1] [col2] ...",
	Short: "Update cells at range with values",
	Args:  cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := sheets.NewService(context.Background())
		if err != nil {
			return err
		}
		sheetID := args[0]
		targetRange := args[1]
		var row []interface{}
		for _, val := range args[2:] {
			row = append(row, val)
		}
		if err := svc.UpdateRange(sheetID, targetRange, [][]interface{}{row}); err != nil {
			return err
		}
		fmt.Printf("Updated %s at %s successfully\n", sheetID, targetRange)
		return nil
	},
}

func init() {
	sheetsReadCmd.Flags().BoolVar(&sheetsJSON, "json", false, "Output as JSON")
	sheetsCmd.AddCommand(sheetsCreateCmd)
	sheetsCmd.AddCommand(sheetsAddSheetCmd)
	sheetsCmd.AddCommand(sheetsReadCmd)
	sheetsCmd.AddCommand(sheetsAppendCmd)
	sheetsCmd.AddCommand(sheetsUpdateCmd)
}
