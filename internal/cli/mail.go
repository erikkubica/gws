package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erikkubica/gmcp/internal/services/gmail"
	"github.com/spf13/cobra"
)

var mailCmd = &cobra.Command{
	Use:   "mail",
	Short: "Search, read, send, draft, and reply to Gmail messages",
}

var (
	mailQuery       string
	mailMax         int64
	mailTo          string
	mailSubj        string
	mailBody        string
	mailAttachments []string
	mailJSON        bool
)

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
	Short: "Send an email message (supports --attach)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if mailTo == "" || mailSubj == "" || mailBody == "" {
			return fmt.Errorf("flags --to, --subject, and --body are all required")
		}
		svc, err := gmail.NewService(context.Background())
		if err != nil {
			return err
		}
		opts := gmail.EmailOptions{
			To: mailTo, Subject: mailSubj, Body: mailBody, Attachments: mailAttachments,
		}
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
		opts := gmail.EmailOptions{
			To: mailTo, Subject: mailSubj, Body: mailBody, Attachments: mailAttachments,
		}
		res, err := svc.CreateDraft(opts)
		if err != nil {
			return err
		}
		fmt.Printf("Draft created successfully! (ID: %s)\n", res.Id)
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

	mailReplyCmd.Flags().StringVar(&mailBody, "body", "", "Reply body text")
	mailReplyCmd.Flags().StringSliceVarP(&mailAttachments, "attach", "a", nil, "Files to attach")

	mailDraftCmd.Flags().StringVar(&mailTo, "to", "", "Recipient email")
	mailDraftCmd.Flags().StringVar(&mailSubj, "subject", "", "Subject line")
	mailDraftCmd.Flags().StringVar(&mailBody, "body", "", "Message body text")
	mailDraftCmd.Flags().StringSliceVarP(&mailAttachments, "attach", "a", nil, "Files to attach")

	mailCmd.AddCommand(mailListCmd)
	mailCmd.AddCommand(mailReadCmd)
	mailCmd.AddCommand(mailSendCmd)
	mailCmd.AddCommand(mailReplyCmd)
	mailCmd.AddCommand(mailDraftCmd)
}
