package cli

import (
	"context"
	"fmt"

	"github.com/erikkubica/gmcp/internal/services/gmail"
	"github.com/spf13/cobra"
)

var mailCmd = &cobra.Command{
	Use:   "mail",
	Short: "Search, read, send, and draft Gmail messages",
}

var (
	mailQuery string
	mailMax   int64
	mailTo    string
	mailSubj  string
	mailBody  string
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
		fmt.Printf("From: %s\nDate: %s\nSubject: %s\n\n%s\n", msg.From, msg.Date, msg.Subject, msg.Body)
		return nil
	},
}

var mailSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send an email message",
	RunE: func(cmd *cobra.Command, args []string) error {
		if mailTo == "" || mailSubj == "" || mailBody == "" {
			return fmt.Errorf("flags --to, --subject, and --body are all required")
		}
		svc, err := gmail.NewService(context.Background())
		if err != nil {
			return err
		}
		res, err := svc.SendMessage(mailTo, mailSubj, mailBody)
		if err != nil {
			return err
		}
		fmt.Printf("Email sent successfully! (ID: %s)\n", res.Id)
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
		res, err := svc.CreateDraft(mailTo, mailSubj, mailBody)
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

	mailSendCmd.Flags().StringVar(&mailTo, "to", "", "Recipient email")
	mailSendCmd.Flags().StringVar(&mailSubj, "subject", "", "Subject line")
	mailSendCmd.Flags().StringVar(&mailBody, "body", "", "Message body text")

	mailDraftCmd.Flags().StringVar(&mailTo, "to", "", "Recipient email")
	mailDraftCmd.Flags().StringVar(&mailSubj, "subject", "", "Subject line")
	mailDraftCmd.Flags().StringVar(&mailBody, "body", "", "Message body text")

	mailCmd.AddCommand(mailListCmd)
	mailCmd.AddCommand(mailReadCmd)
	mailCmd.AddCommand(mailSendCmd)
	mailCmd.AddCommand(mailDraftCmd)
}
