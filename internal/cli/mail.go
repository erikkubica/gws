package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/erikkubica/gws/internal/services/calendar"
	"github.com/erikkubica/gws/internal/services/gmail"
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
	mailWithMeet    bool
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
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List or search messages (aliases: ls)",
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
			fmt.Printf("[%s] %s | %s\n  %s\n", m.ID, m.Date, m.From, m.Subject)
			for _, att := range m.Attachments {
				fmt.Printf("  📎 %s (%s)\n", att.Filename, formatBytes(att.Size))
			}
			fmt.Println()
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
		fmt.Printf("From: %s\nDate: %s\nSubject: %s\n", msg.From, msg.Date, msg.Subject)
		if len(msg.Attachments) > 0 {
			fmt.Println("\nAttachments:")
			for _, att := range msg.Attachments {
				attID := att.AttachmentID
				if attID == "" {
					attID = "inline"
				}
				fmt.Printf("  📎 %s (%s) [ID: %s]\n", att.Filename, formatBytes(att.Size), attID)
			}
		}
		fmt.Printf("\n%s\n", msg.Body)
		return nil
	},
}

func appendMeetIfNeeded(subj, to, body string) (string, error) {
	if !mailWithMeet {
		return body, nil
	}
	calSvc, err := calendar.NewService(context.Background())
	if err != nil {
		return "", err
	}
	_, meetURL, err := calSvc.CreateQuickMeet(subj, "", "", []string{to})
	if err != nil {
		return "", err
	}
	return body + fmt.Sprintf("\n\n---\nGoogle Meet Link: %s\n", meetURL), nil
}

var mailSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send an email message (supports --attach, --meet, --delay, --at)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if mailTo == "" || mailSubj == "" || mailBody == "" {
			return fmt.Errorf("flags --to, --subject, and --body are all required")
		}
		if err := waitSchedule(mailDelay, mailAt); err != nil {
			return err
		}
		body, err := appendMeetIfNeeded(mailSubj, mailTo, mailBody)
		if err != nil {
			return err
		}
		svc, err := gmail.NewService(context.Background())
		if err != nil {
			return err
		}
		opts := gmail.EmailOptions{To: mailTo, Subject: mailSubj, Body: body, Attachments: mailAttachments}
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
	Short: "Reply to an existing message thread (supports --meet)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if mailBody == "" {
			return fmt.Errorf("flag --body is required")
		}
		if err := waitSchedule(mailDelay, mailAt); err != nil {
			return err
		}
		body, err := appendMeetIfNeeded("Meeting Followup", mailTo, mailBody)
		if err != nil {
			return err
		}
		svc, err := gmail.NewService(context.Background())
		if err != nil {
			return err
		}
		res, err := svc.ReplyMessage(args[0], body, mailAttachments)
		if err != nil {
			return err
		}
		fmt.Printf("Reply sent successfully! (ID: %s, Thread: %s)\n", res.Id, res.ThreadId)
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
	mailSendCmd.Flags().BoolVar(&mailWithMeet, "meet", false, "Generate and append Google Meet link")
	mailSendCmd.Flags().StringVar(&mailDelay, "delay", "", "Delay sending (e.g. 10m, 1h)")
	mailSendCmd.Flags().StringVar(&mailAt, "at", "", "Schedule send at RFC3339 timestamp")

	mailReplyCmd.Flags().StringVar(&mailBody, "body", "", "Reply body text")
	mailReplyCmd.Flags().StringSliceVarP(&mailAttachments, "attach", "a", nil, "Files to attach")
	mailReplyCmd.Flags().BoolVar(&mailWithMeet, "meet", false, "Generate and append Google Meet link")
	mailReplyCmd.Flags().StringVar(&mailDelay, "delay", "", "Delay sending (e.g. 10m, 1h)")
	mailReplyCmd.Flags().StringVar(&mailAt, "at", "", "Schedule send at RFC3339 timestamp")

	mailCmd.AddCommand(mailListCmd)
	mailCmd.AddCommand(mailReadCmd)
	mailCmd.AddCommand(mailSendCmd)
	mailCmd.AddCommand(mailReplyCmd)

	initDraftCommands()
	initAttachmentCommands()
}
