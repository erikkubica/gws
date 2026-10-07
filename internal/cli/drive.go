package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erikkubica/gws/internal/services/drive"
	"github.com/spf13/cobra"
)

var driveCmd = &cobra.Command{
	Use:   "drive",
	Short: "Search, list, read, upload, download, and create files on Google Drive",
}

var (
	driveQuery      string
	driveMax        int64
	driveUploadName string
	driveCreateMime string
	driveJSON       bool
)

var driveListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List or search files (aliases: ls)",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := drive.NewService(context.Background())
		if err != nil {
			return err
		}
		files, err := svc.ListFiles(driveQuery, driveMax)
		if err != nil {
			return err
		}
		if driveJSON {
			b, _ := json.MarshalIndent(files, "", "  ")
			fmt.Println(string(b))
			return nil
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

var driveDownloadCmd = &cobra.Command{
	Use:   "download [file_id] [destination_path]",
	Short: "Download a file or exported doc from Google Drive",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := drive.NewService(context.Background())
		if err != nil {
			return err
		}
		if err := svc.DownloadFile(args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("Downloaded %s to %s successfully.\n", args[0], args[1])
		return nil
	},
}

var driveCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new file or doc in Google Drive",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := drive.NewService(context.Background())
		if err != nil {
			return err
		}
		f, err := svc.CreateEmptyFile(args[0], driveCreateMime)
		if err != nil {
			return err
		}
		fmt.Printf("File created: %s (ID: %s, MIME: %s)\n", f.Name, f.ID, f.MimeType)
		return nil
	},
}

var driveDeleteCmd = &cobra.Command{
	Use:     "delete [file_id]",
	Aliases: []string{"rm"},
	Short:   "Permanently delete a file from Google Drive (aliases: rm)",
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
	driveListCmd.Flags().BoolVar(&driveJSON, "json", false, "Output as JSON")

	driveUploadCmd.Flags().StringVarP(&driveUploadName, "name", "n", "", "Custom name on Drive")
	driveCreateCmd.Flags().StringVarP(&driveCreateMime, "mime", "m", "text/plain", "MIME type (e.g. 'application/vnd.google-apps.document')")

	driveCmd.AddCommand(driveListCmd)
	driveCmd.AddCommand(driveReadCmd)
	driveCmd.AddCommand(driveUploadCmd)
	driveCmd.AddCommand(driveDownloadCmd)
	driveCmd.AddCommand(driveCreateCmd)
	driveCmd.AddCommand(driveDeleteCmd)
}
