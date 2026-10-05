package cli

import (
	"context"
	"fmt"

	"github.com/erikkubica/gmcp/internal/services/drive"
	"github.com/spf13/cobra"
)

var driveCmd = &cobra.Command{
	Use:   "drive",
	Short: "Search, list, and read files from Google Drive",
}

var (
	driveQuery string
	driveMax   int64
)

var driveListCmd = &cobra.Command{
	Use:   "list",
	Short: "List or search files",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := drive.NewService(context.Background())
		if err != nil {
			return err
		}
		files, err := svc.ListFiles(driveQuery, driveMax)
		if err != nil {
			return err
		}
		for _, f := range files {
			fmt.Printf("[%s] %-35s (%s)\n", f.ID, f.Name, f.MimeType)
		}
		return nil
	},
}

var driveReadCmd = &cobra.Command{
	Use:   "read [file_id]",
	Short: "Read or export file content as text",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := drive.NewService(context.Background())
		if err != nil {
			return err
		}
		content, err := svc.ReadFile(args[0])
		if err != nil {
			return err
		}
		fmt.Println(content)
		return nil
	},
}

func init() {
	driveListCmd.Flags().StringVarP(&driveQuery, "query", "q", "", "Filename query")
	driveListCmd.Flags().Int64VarP(&driveMax, "max", "m", 10, "Max files")
	driveCmd.AddCommand(driveListCmd)
	driveCmd.AddCommand(driveReadCmd)
}
