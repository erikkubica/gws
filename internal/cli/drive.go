package cli

import (
	"context"
	"fmt"

	"github.com/erikkubica/gmcp/internal/services/drive"
	"github.com/spf13/cobra"
)

var driveCmd = &cobra.Command{
	Use:   "drive",
	Short: "Search, list, read, upload, and delete files on Google Drive",
}

var (
	driveQuery      string
	driveMax        int64
	driveUploadName string
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

var driveUploadCmd = &cobra.Command{
	Use:   "upload [file_path]",
	Short: "Upload a local file to Google Drive",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := drive.NewService(context.Background())
		if err != nil {
			return err
		}
		uploaded, err := svc.UploadFile(args[0], driveUploadName)
		if err != nil {
			return err
		}
		fmt.Printf("File uploaded successfully!\nID: %s\nName: %s\nSize: %d bytes\n",
			uploaded.ID, uploaded.Name, uploaded.Size)
		return nil
	},
}

var driveDeleteCmd = &cobra.Command{
	Use:   "delete [file_id]",
	Short: "Permanently delete a file from Google Drive",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := drive.NewService(context.Background())
		if err != nil {
			return err
		}
		if err := svc.DeleteFile(args[0]); err != nil {
			return err
		}
		fmt.Printf("File %s deleted successfully.\n", args[0])
		return nil
	},
}

func init() {
	driveListCmd.Flags().StringVarP(&driveQuery, "query", "q", "", "Filename query")
	driveListCmd.Flags().Int64VarP(&driveMax, "max", "m", 10, "Max files")
	driveUploadCmd.Flags().StringVarP(&driveUploadName, "name", "n", "", "Custom name on Drive (defaults to local filename)")

	driveCmd.AddCommand(driveListCmd)
	driveCmd.AddCommand(driveReadCmd)
	driveCmd.AddCommand(driveUploadCmd)
	driveCmd.AddCommand(driveDeleteCmd)
}
