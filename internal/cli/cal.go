package cli

import (
	"context"
	"fmt"

	"github.com/erikkubica/gmcp/internal/services/calendar"
	"github.com/spf13/cobra"
)

var calCmd = &cobra.Command{
	Use:   "cal",
	Short: "View and add Google Calendar events",
}

var calMax int64

var calListCmd = &cobra.Command{
	Use:   "list",
	Short: "List upcoming events",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := calendar.NewService(context.Background())
		if err != nil {
			return err
		}
		events, err := svc.ListUpcomingEvents("primary", calMax)
		if err != nil {
			return err
		}
		if len(events) == 0 {
			fmt.Println("No upcoming events found.")
			return nil
		}
		for _, e := range events {
			fmt.Printf("[%s] %s -> %s\n  Location: %s\n\n", e.Start, e.Summary, e.End, e.Location)
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
		ev, err := svc.QuickAddEvent("primary", args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Event created: %s\nLink: %s\n", ev.Summary, ev.HtmlLink)
		return nil
	},
}

func init() {
	calListCmd.Flags().Int64VarP(&calMax, "max", "m", 10, "Max events")
	calCmd.AddCommand(calListCmd)
	calCmd.AddCommand(calAddCmd)
}
