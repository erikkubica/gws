package cli

import (
	"fmt"

	"github.com/erikkubica/gws/internal/services/docs"
	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Create, read, and append text to Google Docs",
}

var docsCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new Google Doc",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := docs.NewService(cmd.Context())
		if err != nil {
			return err
		}
		doc, err := svc.CreateDocument(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Google Doc created: %s\nID: %s\nURL: https://docs.google.com/document/d/%s/edit\n",
			doc.Title, doc.DocumentId, doc.DocumentId)
		return nil
	},
}

var docsAppendCmd = &cobra.Command{
	Use:   "append [doc_id] [text]",
	Short: "Append text to an existing Google Doc",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := docs.NewService(cmd.Context())
		if err != nil {
			return err
		}
		if err := svc.AppendText(args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("Text appended successfully to doc %s\n", args[0])
		return nil
	},
}

var docsReadCmd = &cobra.Command{
	Use:   "read [doc_id]",
	Short: "Read full text content of a Google Doc",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := docs.NewService(cmd.Context())
		if err != nil {
			return err
		}
		text, err := svc.GetDocumentText(args[0])
		if err != nil {
			return err
		}
		fmt.Println(text)
		return nil
	},
}

func init() {
	docsCmd.AddCommand(docsCreateCmd)
	docsCmd.AddCommand(docsAppendCmd)
	docsCmd.AddCommand(docsReadCmd)
}
