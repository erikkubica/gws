package cli

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/erikkubica/gws/internal/services/gmail"
	"github.com/spf13/cobra"
)

var (
	mailDownloadOut string
	mailDownloadDir string
)

var mailDownloadCmd = &cobra.Command{
	Use:     "download [message_id] [attachment_id_or_filename]",
	Aliases: []string{"att", "attachment"},
	Short:   "Download attachment(s) from a Gmail message",
	Args:    cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		msgID := args[0]
		target := ""
		if len(args) > 1 {
			target = args[1]
		}
		return executeAttachmentDownload(msgID, target)
	},
}

func executeAttachmentDownload(msgID, target string) error {
	svc, err := gmail.NewService(context.Background())
	if err != nil {
		return err
	}
	if target == "" {
		return downloadAllOrSingle(svc, msgID)
	}
	return downloadSpecificAttachment(svc, msgID, target)
}

func downloadAllOrSingle(svc *gmail.Service, msgID string) error {
	atts, err := svc.ListAttachments(msgID)
	if err != nil {
		return err
	}
	if len(atts) == 0 {
		return fmt.Errorf("no attachments found on message %s", msgID)
	}
	if len(atts) == 1 {
		dest := resolveOutputPath(mailDownloadOut, mailDownloadDir, atts[0].Filename)
		return saveAndReport(svc, msgID, atts[0].Filename, dest)
	}
	dir := mailDownloadDir
	if dir == "" {
		dir = "."
	}
	saved, err := svc.DownloadAllAttachments(msgID, dir)
	if err != nil {
		return err
	}
	fmt.Printf("Downloaded %d attachments to %s:\n", len(saved), dir)
	for _, p := range saved {
		fmt.Printf("  📎 %s\n", p)
	}
	return nil
}

func downloadSpecificAttachment(svc *gmail.Service, msgID, target string) error {
	dest := resolveOutputPath(mailDownloadOut, mailDownloadDir, target)
	return saveAndReport(svc, msgID, target, dest)
}

func saveAndReport(svc *gmail.Service, msgID, target, dest string) error {
	file, err := svc.SaveAttachment(msgID, target, dest)
	if err != nil {
		return err
	}
	fmt.Printf("Saved attachment: %s (%s) -> %s\n", file.Filename, formatBytes(file.Size), dest)
	return nil
}

func resolveOutputPath(outFile, outDir, filename string) string {
	if outFile != "" {
		return outFile
	}
	base := filepath.Base(filename)
	if outDir != "" {
		return filepath.Join(outDir, base)
	}
	return base
}

func formatBytes(bytes int64) string {
	if bytes <= 0 {
		return "0 B"
	}
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
}

func initAttachmentCommands() {
	mailDownloadCmd.Flags().StringVarP(&mailDownloadOut, "out", "o", "", "Destination file path (for single attachment)")
	mailDownloadCmd.Flags().StringVarP(&mailDownloadDir, "dir", "d", "", "Destination directory (default: current directory)")

	mailCmd.AddCommand(mailDownloadCmd)
}
