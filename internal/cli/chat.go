package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/erikkubica/gws/internal/services/chat"
	"github.com/spf13/cobra"
)

var (
	chatMax         int64
	chatJSON        bool
	chatAsc         bool
	chatSpacesQuery string
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Send messages, reply in threads, react, and interact with Google Chat",
}

var chatSpacesCmd = &cobra.Command{
	Use:     "spaces [query]",
	Aliases: []string{"rooms", "dms"},
	Short:   "List or search joined Google Chat spaces and direct messages (aliases: rooms, dms)",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := chat.NewService(context.Background())
		if err != nil {
			return err
		}
		query := chatSpacesQuery
		if len(args) > 0 {
			query = args[0]
		}
		spaces, err := svc.SearchSpaces(query, chatMax)
		if err != nil {
			return err
		}
		if chatJSON {
			b, _ := json.MarshalIndent(spaces, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		printSpacesTable(spaces)
		return nil
	},
}

var chatMessagesCmd = &cobra.Command{
	Use:     "messages [space_id]",
	Aliases: []string{"msgs"},
	Short:   "List messages from a Google Chat space or direct message (aliases: msgs)",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("space_id is required (e.g. 'gws chat messages spaces/AAAA...'). Run 'gws chat spaces' to list available spaces")
		}
		return runListMessages(args[0])
	},
}

func printSpacesTable(spaces []*chat.SpaceInfo) {
	if len(spaces) == 0 {
		fmt.Println("No Google Chat spaces found.")
		return
	}
	fmt.Printf("%-28s  %-16s  %-20s  %s\n", "NAME / ID", "TYPE", "LAST ACTIVE", "CONVERSATION / SPACE NAME")
	for _, sp := range spaces {
		title := sp.DisplayName
		if title == "" {
			title = "(Direct Message)"
		}
		lastActive := sp.LastActiveTime
		if len(lastActive) > 19 {
			lastActive = lastActive[:10] + " " + lastActive[11:19]
		}
		if lastActive == "" {
			lastActive = "-"
		}
		fmt.Printf("%-28s  %-16s  %-20s  %s\n", sp.Name, sp.SpaceType, lastActive, title)
	}
}

var chatListCmd = &cobra.Command{
	Use:     "list [space_id|spaces]",
	Aliases: []string{"ls"},
	Short:   "List spaces, or messages from a space (aliases: ls)",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 || args[0] == "spaces" {
			return chatSpacesCmd.RunE(cmd, args)
		}
		return runListMessages(args[0])
	},
}

func runListMessages(space string) error {
	svc, err := chat.NewService(context.Background())
	if err != nil {
		return err
	}
	order := "DESC"
	if chatAsc {
		order = "ASC"
	}
	msgs, err := svc.ListMessagesWithOrder(space, chatMax, order)
	if err != nil {
		return err
	}
	if chatJSON {
		b, _ := json.MarshalIndent(msgs, "", "  ")
		fmt.Println(string(b))
		return nil
	}
	printMessagesTable(msgs)
	return nil
}

func printMessagesTable(msgs []*chat.MessageInfo) {
	if len(msgs) == 0 {
		fmt.Println("No messages found in this space.")
		return
	}
	for _, m := range msgs {
		printMessageItem(m)
	}
}

func printMessageItem(m *chat.MessageInfo) {
	sender := m.SenderName
	if sender == "" {
		sender = "Unknown"
	}
	if m.SenderID != "" {
		sender = fmt.Sprintf("%s (%s)", sender, m.SenderID)
	}
	fmt.Printf("[%s] %s (ID: %s):\n  %s\n", m.CreateTime, sender, m.Name, m.Text)
	if attachStr := formatAttachmentsString(m.Attachments); attachStr != "" {
		fmt.Print(attachStr)
	}
	if reactionsStr := formatReactionsString(m.Reactions); reactionsStr != "" {
		fmt.Print(reactionsStr)
	}
	fmt.Println()
}

func formatAttachmentsString(attachments []chat.AttachmentInfo) string {
	if len(attachments) == 0 {
		return ""
	}
	var parts []string
	for _, a := range attachments {
		name := a.ContentName
		if name == "" {
			name = "attachment"
		}
		if a.DownloadURL != "" {
			parts = append(parts, fmt.Sprintf("📎 %s (%s)", name, a.DownloadURL))
		} else {
			parts = append(parts, fmt.Sprintf("📎 %s", name))
		}
	}
	return "  " + strings.Join(parts, "\n  ") + "\n"
}

func formatReactionsString(reactions []chat.ReactionSummary) string {
	if len(reactions) == 0 {
		return ""
	}
	var parts []string
	for _, r := range reactions {
		parts = append(parts, formatSingleReaction(r))
	}
	return "  Reactions: " + strings.Join(parts, "  ") + "\n"
}

func formatSingleReaction(r chat.ReactionSummary) string {
	if len(r.Users) == 0 {
		return fmt.Sprintf("%s %d", r.Emoji, r.Count)
	}
	var names []string
	for _, u := range r.Users {
		name := u.DisplayName
		if name == "" {
			name = u.Name
		}
		names = append(names, name)
	}
	return fmt.Sprintf("%s %d (%s)", r.Emoji, r.Count, strings.Join(names, ", "))
}

func init() {
	chatSpacesCmd.Flags().Int64VarP(&chatMax, "max", "m", 20, "Maximum number of spaces to return")
	chatSpacesCmd.Flags().StringVarP(&chatSpacesQuery, "query", "q", "", "Filter spaces by name or member")
	chatSpacesCmd.Flags().BoolVar(&chatJSON, "json", false, "Output results in JSON format")

	chatMessagesCmd.Flags().Int64VarP(&chatMax, "max", "m", 20, "Maximum number of messages to return")
	chatMessagesCmd.Flags().BoolVar(&chatJSON, "json", false, "Output results in JSON format")
	chatMessagesCmd.Flags().BoolVar(&chatAsc, "asc", false, "List in ascending order (oldest first; default is newest first)")

	chatListCmd.Flags().Int64VarP(&chatMax, "max", "m", 20, "Maximum number of messages to return")
	chatListCmd.Flags().BoolVar(&chatJSON, "json", false, "Output results in JSON format")
	chatListCmd.Flags().BoolVar(&chatAsc, "asc", false, "List in ascending order (oldest first; default is newest first)")

	chatCmd.AddCommand(chatSpacesCmd)
	chatCmd.AddCommand(chatMessagesCmd)
	chatCmd.AddCommand(chatListCmd)
}
