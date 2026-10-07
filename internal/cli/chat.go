package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erikkubica/gws/internal/services/chat"
	"github.com/spf13/cobra"
)

var (
	chatMax  int64
	chatJSON bool
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Send messages, list spaces, and interact with Google Chat",
}

var chatSpacesCmd = &cobra.Command{
	Use:   "spaces",
	Short: "List joined Google Chat spaces and direct messages",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := chat.NewService(context.Background())
		if err != nil {
			return err
		}
		spaces, err := svc.ListSpaces(chatMax)
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

func printSpacesTable(spaces []*chat.SpaceInfo) {
	if len(spaces) == 0 {
		fmt.Println("No Google Chat spaces found.")
		return
	}
	fmt.Printf("%-32s  %-24s  %s\n", "NAME / ID", "TYPE", "DISPLAY NAME")
	for _, sp := range spaces {
		name := sp.Name
		title := sp.DisplayName
		if title == "" {
			title = "(Direct Message)"
		}
		fmt.Printf("%-32s  %-24s  %s\n", name, sp.SpaceType, title)
	}
}

var chatSendCmd = &cobra.Command{
	Use:   "send [space_id] [message_text]",
	Short: "Send a message to a Google Chat space or direct message",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := chat.NewService(context.Background())
		if err != nil {
			return err
		}
		space := args[0]
		text := args[1]
		msg, err := svc.SendMessage(space, text)
		if err != nil {
			return err
		}
		fmt.Printf("Message delivered! ID: %s (Time: %s)\n", msg.Name, msg.CreateTime)
		return nil
	},
}

var chatListCmd = &cobra.Command{
	Use:   "list [space_id]",
	Short: "List recent messages from a Google Chat space",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := chat.NewService(context.Background())
		if err != nil {
			return err
		}
		space := args[0]
		msgs, err := svc.ListMessages(space, chatMax)
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
	},
}

func printMessagesTable(msgs []*chat.MessageInfo) {
	if len(msgs) == 0 {
		fmt.Println("No messages found in this space.")
		return
	}
	for _, m := range msgs {
		sender := m.SenderName
		if sender == "" {
			sender = "Unknown"
		}
		fmt.Printf("[%s] %s:\n  %s\n\n", m.CreateTime, sender, m.Text)
	}
}

var chatWebhookCmd = &cobra.Command{
	Use:   "webhook [webhook_url] [message_text]",
	Short: "Post a message to an incoming Google Chat webhook URL",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := chat.NewService(context.Background())
		if err != nil {
			return err
		}
		url := args[0]
		text := args[1]
		if err := svc.SendWebhook(url, text); err != nil {
			return err
		}
		fmt.Println("Webhook message delivered successfully!")
		return nil
	},
}

func init() {
	chatSpacesCmd.Flags().Int64VarP(&chatMax, "max", "m", 20, "Maximum number of spaces to return")
	chatSpacesCmd.Flags().BoolVar(&chatJSON, "json", false, "Output results in JSON format")

	chatListCmd.Flags().Int64VarP(&chatMax, "max", "m", 20, "Maximum number of messages to return")
	chatListCmd.Flags().BoolVar(&chatJSON, "json", false, "Output results in JSON format")

	chatCmd.AddCommand(chatSpacesCmd)
	chatCmd.AddCommand(chatSendCmd)
	chatCmd.AddCommand(chatListCmd)
	chatCmd.AddCommand(chatWebhookCmd)
}
