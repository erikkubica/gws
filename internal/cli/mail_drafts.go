package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erikkubica/gmcp/internal/services/gmail"
	"github.com/spf13/cobra"
)

var mailDraftCmd = &cobra.Command{
	Use:   "draft",
	Short: "Create an email draft",
	RunE: func(cmd *cobra.Command, args []string) error {
		if mailTo == "" || mailSubj == "" || mailBody == "" {
			return fmt.Errorf("flags --to, --subject, and --body are all required")
		}
		svc, err := gmail.NewService(context.Background())
		if err != nil {
			return err
		}
		opts := gmail.EmailOptions{To: mailTo, Subject: mailSubj, Body: mailBody, Attachments: mailAttachments}
		res, err := svc.CreateDraft(opts)
		if err != nil {
			return err
		}
		fmt.Printf("Draft created successfully! (ID: %s)\n", res.Id)
		return nil
	},
}

var mailDraftsListCmd = &cobra.Command{
	Use:   "drafts",
	Short: "List drafts in the mailbox",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := gmail.NewService(context.Background())
		if err != nil {
			return err
		}
		drafts, err := svc.ListDrafts(mailMax)
		if err != nil {
			return err
		}
		if mailJSON {
			b, _ := json.MarshalIndent(drafts, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		for _, d := range drafts {
			fmt.Printf("[%s] To: %s | %s\n  %s\n\n", d.ID, d.To, d.Subject, d.Snippet)
		}
		return nil
	},
}

var mailSendDraftCmd = &cobra.Command{
	Use:   "send-draft [draft_id]",
	Short: "Send an existing draft (supports --delay, --at)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := waitSchedule(mailDelay, mailAt); err != nil {
			return err
		}
		svc, err := gmail.NewService(context.Background())
		if err != nil {
			return err
		}
		msg, err := svc.SendDraft(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Draft sent successfully! (ID: %s)\n", msg.Id)
		return nil
	},
}

var mailDeleteDraftCmd = &cobra.Command{
	Use:   "delete-draft [draft_id]",
	Short: "Delete an existing draft",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := gmail.NewService(context.Background())
		if err != nil {
			return err
		}
		if err := svc.DeleteDraft(args[0]); err != nil {
			return err
		}
		fmt.Printf("Draft %s deleted successfully.\n", args[0])
		return nil
	},
}

func initDraftCommands() {
	mailDraftCmd.Flags().StringVar(&mailTo, "to", "", "Recipient email")
	mailDraftCmd.Flags().StringVar(&mailSubj, "subject", "", "Subject line")
	mailDraftCmd.Flags().StringVar(&mailBody, "body", "", "Message body text")
	mailDraftCmd.Flags().StringSliceVarP(&mailAttachments, "attach", "a", nil, "Files to attach")

	mailDraftsListCmd.Flags().Int64VarP(&mailMax, "max", "m", 10, "Max drafts")
	mailDraftsListCmd.Flags().BoolVar(&mailJSON, "json", false, "Output as JSON")

	mailSendDraftCmd.Flags().StringVar(&mailDelay, "delay", "", "Delay sending (e.g. 10m, 1h)")
	mailSendDraftCmd.Flags().StringVar(&mailAt, "at", "", "Schedule send at RFC3339 timestamp")

	mailCmd.AddCommand(mailDraftCmd)
	mailCmd.AddCommand(mailDraftsListCmd)
	mailCmd.AddCommand(mailSendDraftCmd)
	mailCmd.AddCommand(mailDeleteDraftCmd)
}
