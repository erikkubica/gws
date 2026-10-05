package cli

import (
	"context"
	"fmt"

	"github.com/erikkubica/gmcp/internal/services/calendar"
	"github.com/spf13/cobra"
)

var calCmd = &cobra.Command{
	Use:   "cal",
	Short: "View, create, and manage Google Calendar events",
}

var (
	calMax      int64
	calTitle    string
	calStart    string
	calEnd      string
	calDesc     string
	calLoc      string
	calCalendar string
)

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
		if len(events) == 0 {
			fmt.Println("No upcoming events found.")
			return nil
		}
		for _, e := range events {
			fmt.Printf("[%s] %s (ID: %s)\n  Time: %s -> %s\n  Location: %s\n\n",
				e.Summary, e.ID, e.ID, e.Start, e.End, e.Location)
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
	Short: "Create an event with explicit start and end times (RFC3339)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if calTitle == "" || calStart == "" || calEnd == "" {
			return fmt.Errorf("flags --title, --start, and --end are all required")
		}
		svc, err := calendar.NewService(context.Background())
		if err != nil {
			return err
		}
		ev, err := svc.CreateEvent(calCalendar, calTitle, calDesc, calLoc, calStart, calEnd)
		if err != nil {
			return err
		}
		fmt.Printf("Event created: %s (ID: %s)\nLink: %s\n", ev.Summary, ev.Id, ev.HtmlLink)
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

func init() {
	calListCmd.Flags().Int64VarP(&calMax, "max", "m", 10, "Max events")
	calListCmd.Flags().StringVarP(&calCalendar, "calendar", "c", "primary", "Calendar ID")

	calAddCmd.Flags().StringVarP(&calCalendar, "calendar", "c", "primary", "Calendar ID")

	calCreateCmd.Flags().StringVarP(&calTitle, "title", "t", "", "Event title/summary")
	calCreateCmd.Flags().StringVarP(&calStart, "start", "s", "", "Start time in RFC3339 (e.g. 2026-10-06T10:00:00+07:00)")
	calCreateCmd.Flags().StringVarP(&calEnd, "end", "e", "", "End time in RFC3339 (e.g. 2026-10-06T11:00:00+07:00)")
	calCreateCmd.Flags().StringVarP(&calDesc, "desc", "d", "", "Description")
	calCreateCmd.Flags().StringVarP(&calLoc, "loc", "l", "", "Location")
	calCreateCmd.Flags().StringVarP(&calCalendar, "calendar", "c", "primary", "Calendar ID")

	calDeleteCmd.Flags().StringVarP(&calCalendar, "calendar", "c", "primary", "Calendar ID")

	calCmd.AddCommand(calListCmd)
	calCmd.AddCommand(calAddCmd)
	calCmd.AddCommand(calCreateCmd)
	calCmd.AddCommand(calDeleteCmd)
}
