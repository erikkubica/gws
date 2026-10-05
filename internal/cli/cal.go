package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erikkubica/gmcp/internal/services/calendar"
	"github.com/spf13/cobra"
)

var calCmd = &cobra.Command{
	Use:   "cal",
	Short: "View, create, and manage Google Calendar events",
}

var (
	calMax         int64
	calTitle       string
	calStart       string
	calEnd         string
	calDesc        string
	calLoc         string
	calCalendar    string
	calWithMeet    bool
	calAttendees   []string
	calSendUpdates string
	calJSON        bool
)

func printCalendarItem(e calendar.EventSummary) {
	fmt.Printf("[%s] (ID: %s)\n  Time: %s -> %s\n", e.Summary, e.ID, e.Start, e.End)
	if e.Location != "" {
		fmt.Printf("  Location: %s\n", e.Location)
	}
	if e.MeetURL != "" {
		fmt.Printf("  Google Meet: %s\n", e.MeetURL)
	}
	if e.SelfRSVP != "" {
		fmt.Printf("  RSVP: %s\n", e.SelfRSVP)
	}
	fmt.Println()
}

var calListCmd = &cobra.Command{
	Use:   "list",
	Short: "List upcoming events",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := calendar.NewService(context.Background())
		if err != nil {
			return err
		}
		events, err := svc.ListUpcomingEvents(calCalendar, calMax)
		if err != nil {
			return err
		}
		if calJSON {
			b, _ := json.MarshalIndent(events, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		if len(events) == 0 {
			fmt.Println("No upcoming events found.")
			return nil
		}
		for _, e := range events {
			printCalendarItem(e)
		}
		return nil
	},
}

var calAddCmd = &cobra.Command{
	Use:   "add [text]",
	Short: "Quick-add an event with natural language",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := calendar.NewService(context.Background())
		if err != nil {
			return err
		}
		ev, err := svc.QuickAddEvent(calCalendar, args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Event created: %s (ID: %s)\nLink: %s\n", ev.Summary, ev.Id, ev.HtmlLink)
		return nil
	},
}

var calCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an event with explicit start and end times (supports --meet, --attendees)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if calTitle == "" || calStart == "" || calEnd == "" {
			return fmt.Errorf("flags --title, --start, and --end are all required")
		}
		svc, err := calendar.NewService(context.Background())
		if err != nil {
			return err
		}
		opts := calendar.EventOptions{
			CalendarID:  calCalendar,
			Title:       calTitle,
			Description: calDesc,
			Location:    calLoc,
			Start:       calStart,
			End:         calEnd,
			WithMeet:    calWithMeet,
			Attendees:   calAttendees,
			SendUpdates: calSendUpdates,
		}
		ev, err := svc.CreateEventWithOptions(opts)
		if err != nil {
			return err
		}
		fmt.Printf("Event created: %s (ID: %s)\nLink: %s\n", ev.Summary, ev.Id, ev.HtmlLink)
		if ev.HangoutLink != "" {
			fmt.Printf("Google Meet: %s\n", ev.HangoutLink)
		}
		return nil
	},
}

var calDeleteCmd = &cobra.Command{
	Use:   "delete [event_id]",
	Short: "Delete a calendar event by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := calendar.NewService(context.Background())
		if err != nil {
			return err
		}
		if err := svc.DeleteEvent(calCalendar, args[0]); err != nil {
			return err
		}
		fmt.Printf("Calendar event %s deleted successfully.\n", args[0])
		return nil
	},
}

func runCalRSVP(eventID, status string) error {
	svc, err := calendar.NewService(context.Background())
	if err != nil {
		return err
	}
	ev, err := svc.RespondToEvent(calCalendar, eventID, status, calSendUpdates)
	if err != nil {
		return err
	}
	fmt.Printf("RSVP for '%s' updated to: %s\n", ev.Summary, status)
	return nil
}

var calAcceptCmd = &cobra.Command{
	Use:   "accept [event_id]",
	Short: "Accept a calendar invitation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCalRSVP(args[0], "accepted")
	},
}

var calDeclineCmd = &cobra.Command{
	Use:   "decline [event_id]",
	Short: "Decline a calendar invitation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCalRSVP(args[0], "declined")
	},
}

var calMaybeCmd = &cobra.Command{
	Use:   "maybe [event_id]",
	Short: "Tentatively accept a calendar invitation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCalRSVP(args[0], "tentative")
	},
}

var calRespondCmd = &cobra.Command{
	Use:   "respond [event_id] [status]",
	Short: "Respond to an event invitation (accepted, declined, tentative)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCalRSVP(args[0], args[1])
	},
}

func init() {
	calListCmd.Flags().Int64VarP(&calMax, "max", "m", 10, "Max events")
	calListCmd.Flags().StringVarP(&calCalendar, "calendar", "c", "primary", "Calendar ID")
	calListCmd.Flags().BoolVar(&calJSON, "json", false, "Output as JSON")

	calAddCmd.Flags().StringVarP(&calCalendar, "calendar", "c", "primary", "Calendar ID")

	calCreateCmd.Flags().StringVarP(&calTitle, "title", "t", "", "Event title/summary")
	calCreateCmd.Flags().StringVarP(&calStart, "start", "s", "", "Start time in RFC3339 (e.g. 2026-10-06T10:00:00+07:00)")
	calCreateCmd.Flags().StringVarP(&calEnd, "end", "e", "", "End time in RFC3339 (e.g. 2026-10-06T11:00:00+07:00)")
	calCreateCmd.Flags().StringVarP(&calDesc, "desc", "d", "", "Description")
	calCreateCmd.Flags().StringVarP(&calLoc, "loc", "l", "", "Location")
	calCreateCmd.Flags().StringVarP(&calCalendar, "calendar", "c", "primary", "Calendar ID")
	calCreateCmd.Flags().BoolVar(&calWithMeet, "meet", false, "Generate Google Meet conference link")
	calCreateCmd.Flags().StringSliceVarP(&calAttendees, "attendees", "a", nil, "Attendee emails to invite")
	calCreateCmd.Flags().StringVar(&calSendUpdates, "send-updates", "all", "Send updates (all, none, externalOnly)")

	calDeleteCmd.Flags().StringVarP(&calCalendar, "calendar", "c", "primary", "Calendar ID")

	calAcceptCmd.Flags().StringVarP(&calCalendar, "calendar", "c", "primary", "Calendar ID")
	calAcceptCmd.Flags().StringVar(&calSendUpdates, "send-updates", "all", "Send updates notification")

	calDeclineCmd.Flags().StringVarP(&calCalendar, "calendar", "c", "primary", "Calendar ID")
	calDeclineCmd.Flags().StringVar(&calSendUpdates, "send-updates", "all", "Send updates notification")

	calMaybeCmd.Flags().StringVarP(&calCalendar, "calendar", "c", "primary", "Calendar ID")
	calMaybeCmd.Flags().StringVar(&calSendUpdates, "send-updates", "all", "Send updates notification")

	calRespondCmd.Flags().StringVarP(&calCalendar, "calendar", "c", "primary", "Calendar ID")
	calRespondCmd.Flags().StringVar(&calSendUpdates, "send-updates", "all", "Send updates notification")

	calCmd.AddCommand(calListCmd)
	calCmd.AddCommand(calAddCmd)
	calCmd.AddCommand(calCreateCmd)
	calCmd.AddCommand(calDeleteCmd)
	calCmd.AddCommand(calAcceptCmd)
	calCmd.AddCommand(calDeclineCmd)
	calCmd.AddCommand(calMaybeCmd)
	calCmd.AddCommand(calRespondCmd)
}
