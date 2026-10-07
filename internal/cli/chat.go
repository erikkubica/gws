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
	chatMax        int64
	chatJSON       bool
	chatReplyTo    string
	chatAttachment string
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Send messages, reply in threads, react, and interact with Google Chat",
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
	fmt.Printf("%-28s  %-16s  %s\n", "NAME / ID", "TYPE", "CONVERSATION / SPACE NAME")
	for _, sp := range spaces {
		title := sp.DisplayName
		if title == "" {
			title = "(Direct Message)"
		}
		fmt.Printf("%-28s  %-16s  %s\n", sp.Name, sp.SpaceType, title)
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
		opts := chat.MessageOptions{
			SpaceName:  space,
			Text:       text,
			ReplyTo:    chatReplyTo,
			Attachment: chatAttachment,
		}
		msg, err := svc.SendMessageWithOptions(opts)
		if err != nil {
			return err
		}
		fmt.Printf("Message delivered! ID: %s (Time: %s)\n", msg.Name, msg.CreateTime)
		return nil
	},
}

var chatReplyCmd = &cobra.Command{
	Use:   "reply [space_id] [message_id] [message_text]",
	Short: "Reply to a specific message in a Google Chat thread",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := chat.NewService(context.Background())
		if err != nil {
			return err
		}
		space := args[0]
		msgID := args[1]
		text := args[2]
		opts := chat.MessageOptions{
			SpaceName:  space,
			Text:       text,
			ReplyTo:    msgID,
			Attachment: chatAttachment,
		}
		msg, err := svc.SendMessageWithOptions(opts)
		if err != nil {
			return err
		}
		fmt.Printf("Reply posted! ID: %s (Thread: %s)\n", msg.Name, msg.ThreadName)
		return nil
	},
}

var chatReactCmd = &cobra.Command{
	Use:   "react [message_name_or_id] [emoji]",
	Short: "Add an emoji reaction to a message (e.g. '👍', '❤️', '🔥')",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := chat.NewService(context.Background())
		if err != nil {
			return err
		}
		msgName := args[0]
		emoji := args[1]
		if err := svc.AddReaction(msgName, emoji); err != nil {
			return err
		}
		fmt.Printf("Reaction %s added to message!\n", emoji)
		return nil
	},
}

var chatReactionsCmd = &cobra.Command{
	Use:   "reactions [message_name_or_id]",
	Short: "List emoji reactions on a specific message",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := chat.NewService(context.Background())
		if err != nil {
			return err
		}
		msgName := args[0]
		emojis, err := svc.ListReactions(msgName)
		if err != nil {
			return err
		}
		if len(emojis) == 0 {
			fmt.Println("No reactions on this message.")
			return nil
		}
		fmt.Printf("Reactions: %s\n", strings.Join(emojis, " "))
		return nil
	},
}

var chatListCmd = &cobra.Command{
	Use:     "list [space_id|spaces]",
	Aliases: []string{"messages", "msgs"},
	Short:   "List Google Chat spaces, or messages from a space",
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
		fmt.Printf("[%s] %s (ID: %s):\n  %s\n\n", m.CreateTime, sender, m.Name, m.Text)
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

	chatSendCmd.Flags().StringVar(&chatReplyTo, "reply-to", "", "Message ID to reply to (starts or continues thread)")
	chatSendCmd.Flags().StringVar(&chatAttachment, "attach", "", "Local file path to upload as an attachment")

	chatReplyCmd.Flags().StringVar(&chatAttachment, "attach", "", "Local file path to upload as an attachment")

	chatCmd.AddCommand(chatSpacesCmd)
	chatCmd.AddCommand(chatSendCmd)
	chatCmd.AddCommand(chatReplyCmd)
	chatCmd.AddCommand(chatReactCmd)
	chatCmd.AddCommand(chatReactionsCmd)
	chatCmd.AddCommand(chatListCmd)
	chatCmd.AddCommand(chatWebhookCmd)
}
