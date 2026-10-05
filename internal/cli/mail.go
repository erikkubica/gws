package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/erikkubica/gmcp/internal/services/gmail"
	"github.com/spf13/cobra"
)

var mailCmd = &cobra.Command{
	Use:   "mail",
	Short: "Search, read, send, draft, reply, and schedule Gmail messages",
}

var (
	mailQuery       string
	mailMax         int64
	mailTo          string
	mailSubj        string
	mailBody        string
	mailAttachments []string
	mailDelay       string
	mailAt          string
	mailJSON        bool
)

func waitSchedule(delay, at string) error {
	if at != "" {
		t, err := time.Parse(time.RFC3339, at)
		if err != nil {
			return fmt.Errorf("parse --at (expected RFC3339, e.g. 2026-10-06T09:00:00+07:00): %w", err)
		}
		if d := time.Until(t); d > 0 {
			fmt.Printf("Scheduled send: waiting until %s (%v)...\n", t.Format(time.RFC3339), d.Round(time.Second))
			time.Sleep(d)
		}
	} else if delay != "" {
		d, err := time.ParseDuration(delay)
		if err != nil {
			return fmt.Errorf("parse --delay duration (e.g. 10m, 1h): %w", err)
		}
		if d > 0 {
			fmt.Printf("Scheduled send: waiting for %v...\n", d)
			time.Sleep(d)
		}
	}
	return nil
}

var mailListCmd = &cobra.Command{
	Use:   "list",
	Short: "List or search messages",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := gmail.NewService(context.Background())
		if err != nil {
			return err
		}
		msgs, err := svc.ListMessages(mailQuery, mailMax)
		if err != nil {
			return err
		}
		if mailJSON {
			b, _ := json.MarshalIndent(msgs, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		for _, m := range msgs {
			fmt.Printf("[%s] %s | %s\n  %s\n\n", m.ID, m.Date, m.From, m.Subject)
		}
		return nil
	},
}

var mailReadCmd = &cobra.Command{
	Use:   "read [id]",
	Short: "Read a specific message by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := gmail.NewService(context.Background())
		if err != nil {
			return err
		}
		msg, err := svc.GetMessage(args[0])
		if err != nil {
			return err
		}
		if mailJSON {
			b, _ := json.MarshalIndent(msg, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		fmt.Printf("From: %s\nDate: %s\nSubject: %s\n\n%s\n", msg.From, msg.Date, msg.Subject, msg.Body)
		return nil
	},
}

var mailSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send an email message (supports --attach, --delay, --at)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if mailTo == "" || mailSubj == "" || mailBody == "" {
			return fmt.Errorf("flags --to, --subject, and --body are all required")
		}
		if err := waitSchedule(mailDelay, mailAt); err != nil {
			return err
		}
		svc, err := gmail.NewService(context.Background())
		if err != nil {
			return err
		}
		opts := gmail.EmailOptions{To: mailTo, Subject: mailSubj, Body: mailBody, Attachments: mailAttachments}
		res, err := svc.SendMessage(opts)
		if err != nil {
			return err
		}
		fmt.Printf("Email sent successfully! (ID: %s)\n", res.Id)
		return nil
	},
}

var mailReplyCmd = &cobra.Command{
	Use:   "reply [message_id]",
	Short: "Reply to an existing message thread",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if mailBody == "" {
			return fmt.Errorf("flag --body is required")
		}
		if err := waitSchedule(mailDelay, mailAt); err != nil {
			return err
		}
		svc, err := gmail.NewService(context.Background())
		if err != nil {
			return err
		}
		res, err := svc.ReplyMessage(args[0], mailBody, mailAttachments)
		if err != nil {
			return err
		}
		fmt.Printf("Reply sent successfully! (ID: %s, Thread: %s)\n", res.Id, res.ThreadId)
		return nil
	},
}

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

func init() {
	mailListCmd.Flags().StringVarP(&mailQuery, "query", "q", "", "Search query filter")
	mailListCmd.Flags().Int64VarP(&mailMax, "max", "m", 10, "Max messages")
	mailListCmd.Flags().BoolVar(&mailJSON, "json", false, "Output as JSON")

	mailReadCmd.Flags().BoolVar(&mailJSON, "json", false, "Output as JSON")

	mailSendCmd.Flags().StringVar(&mailTo, "to", "", "Recipient email")
	mailSendCmd.Flags().StringVar(&mailSubj, "subject", "", "Subject line")
	mailSendCmd.Flags().StringVar(&mailBody, "body", "", "Message body text")
	mailSendCmd.Flags().StringSliceVarP(&mailAttachments, "attach", "a", nil, "Files to attach")
	mailSendCmd.Flags().StringVar(&mailDelay, "delay", "", "Delay sending (e.g. 10m, 1h)")
	mailSendCmd.Flags().StringVar(&mailAt, "at", "", "Schedule send at RFC3339 timestamp")

	mailReplyCmd.Flags().StringVar(&mailBody, "body", "", "Reply body text")
	mailReplyCmd.Flags().StringSliceVarP(&mailAttachments, "attach", "a", nil, "Files to attach")
	mailReplyCmd.Flags().StringVar(&mailDelay, "delay", "", "Delay sending (e.g. 10m, 1h)")
	mailReplyCmd.Flags().StringVar(&mailAt, "at", "", "Schedule send at RFC3339 timestamp")

	mailDraftCmd.Flags().StringVar(&mailTo, "to", "", "Recipient email")
	mailDraftCmd.Flags().StringVar(&mailSubj, "subject", "", "Subject line")
	mailDraftCmd.Flags().StringVar(&mailBody, "body", "", "Message body text")
	mailDraftCmd.Flags().StringSliceVarP(&mailAttachments, "attach", "a", nil, "Files to attach")

	mailDraftsListCmd.Flags().Int64VarP(&mailMax, "max", "m", 10, "Max drafts")
	mailDraftsListCmd.Flags().BoolVar(&mailJSON, "json", false, "Output as JSON")

	mailSendDraftCmd.Flags().StringVar(&mailDelay, "delay", "", "Delay sending (e.g. 10m, 1h)")
	mailSendDraftCmd.Flags().StringVar(&mailAt, "at", "", "Schedule send at RFC3339 timestamp")

	mailCmd.AddCommand(mailListCmd)
	mailCmd.AddCommand(mailReadCmd)
	mailCmd.AddCommand(mailSendCmd)
	mailCmd.AddCommand(mailReplyCmd)
	mailCmd.AddCommand(mailDraftCmd)
	mailCmd.AddCommand(mailDraftsListCmd)
	mailCmd.AddCommand(mailSendDraftCmd)
	mailCmd.AddCommand(mailDeleteDraftCmd)
}
