package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erikkubica/gmcp/internal/services/calendar"
	"github.com/erikkubica/gmcp/internal/services/gmail"
	"github.com/spf13/cobra"
	googlecal "google.golang.org/api/calendar/v3"
)

var meetCmd = &cobra.Command{
	Use:   "meet",
	Short: "Create and share Google Meet video conferences",
}

var (
	meetStart     string
	meetEnd       string
	meetAttendees []string
	meetTo        string
	meetJSON      bool
)

var meetCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Generate a Google Meet link and calendar event",
	RunE: func(cmd *cobra.Command, args []string) error {
		title := "Google Meet Meeting"
		if len(args) > 0 {
			title = args[0]
		}
		svc, err := calendar.NewService(context.Background())
		if err != nil {
			return err
		}
		ev, meetURL, err := svc.CreateQuickMeet(title, meetStart, meetEnd, meetAttendees)
		if err != nil {
			return err
		}
		return outputMeetResult(ev, meetURL)
	},
}

func outputMeetResult(ev *googlecal.Event, meetURL string) error {
	if meetJSON {
		b, _ := json.MarshalIndent(map[string]string{
			"event_id": ev.Id,
			"title":    ev.Summary,
			"meet_url": meetURL,
			"link":     ev.HtmlLink,
		}, "", "  ")
		fmt.Println(string(b))
		return nil
	}
	fmt.Printf("Google Meet Ready!\n")
	fmt.Printf("Meeting URL: %s\n", meetURL)
	fmt.Printf("Event ID:    %s\n", ev.Id)
	fmt.Printf("Calendar:    %s\n", ev.HtmlLink)
	return nil
}

var meetSendCmd = &cobra.Command{
	Use:   "send [title]",
	Short: "Create a Google Meet link and email the invitation to a recipient",
	RunE: func(cmd *cobra.Command, args []string) error {
		if meetTo == "" {
			return fmt.Errorf("flag --to is required")
		}
		title := "Google Meet Meeting"
		if len(args) > 0 {
			title = args[0]
		}
		calSvc, err := calendar.NewService(context.Background())
		if err != nil {
			return err
		}
		ev, meetURL, err := calSvc.CreateQuickMeet(title, meetStart, meetEnd, []string{meetTo})
		if err != nil {
			return err
		}
		return sendMeetEmail(meetTo, title, meetURL, ev.HtmlLink)
	},
}

func sendMeetEmail(to, title, meetURL, calLink string) error {
	mailSvc, err := gmail.NewService(context.Background())
	if err != nil {
		return err
	}
	body := fmt.Sprintf("Hi,\n\nYou have been invited to a Google Meet session: %s\n\nJoin with Google Meet:\n%s\n\nCalendar Event: %s\n",
		title, meetURL, calLink)
	opts := gmail.EmailOptions{To: to, Subject: "Invitation: " + title, Body: body}
	res, err := mailSvc.SendMessage(opts)
	if err != nil {
		return err
	}
	fmt.Printf("Meet invitation emailed to %s! (Email ID: %s)\n", to, res.Id)
	fmt.Printf("Google Meet URL: %s\n", meetURL)
	return nil
}

func init() {
	meetCreateCmd.Flags().StringVarP(&meetStart, "start", "s", "", "Start time RFC3339 (default: now)")
	meetCreateCmd.Flags().StringVarP(&meetEnd, "end", "e", "", "End time RFC3339 (default: now + 30m)")
	meetCreateCmd.Flags().StringSliceVarP(&meetAttendees, "attendees", "a", nil, "Attendee emails to invite")
	meetCreateCmd.Flags().BoolVar(&meetJSON, "json", false, "Output as JSON")

	meetSendCmd.Flags().StringVar(&meetTo, "to", "", "Recipient email address")
	meetSendCmd.Flags().StringVarP(&meetStart, "start", "s", "", "Start time RFC3339 (default: now)")
	meetSendCmd.Flags().StringVarP(&meetEnd, "end", "e", "", "End time RFC3339 (default: now + 30m)")

	meetCmd.AddCommand(meetCreateCmd)
	meetCmd.AddCommand(meetSendCmd)
}
