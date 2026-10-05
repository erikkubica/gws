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
	Short: "Read and append data to Google Sheets",
}

var sheetsJSON bool

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

func init() {
	sheetsReadCmd.Flags().BoolVar(&sheetsJSON, "json", false, "Output as JSON")
	sheetsCmd.AddCommand(sheetsReadCmd)
	sheetsCmd.AddCommand(sheetsAppendCmd)
}
