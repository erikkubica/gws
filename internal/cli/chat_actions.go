package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erikkubica/gws/internal/services/chat"
	"github.com/spf13/cobra"
)

var (
	chatReplyTo    string
	chatAttachment string
)

var chatSendCmd = &cobra.Command{
	Use:   "send [space_id] [message_text]",
	Short: "Send a message to a Google Chat space or direct message",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := chat.NewService(context.Background())
		if err != nil {
			return err
		}
		opts := chat.MessageOptions{
			SpaceName:  args[0],
			Text:       args[1],
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
		opts := chat.MessageOptions{
			SpaceName:  args[0],
			Text:       args[2],
			ReplyTo:    args[1],
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
		if err := svc.AddReaction(args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("Reaction %s added to message!\n", args[1])
		return nil
	},
}

var chatReactionsCmd = &cobra.Command{
	Use:   "reactions [message_name_or_id]",
	Short: "List emoji reactions and users on a specific message",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := chat.NewService(context.Background())
		if err != nil {
			return err
		}
		reactions, err := svc.ListReactions(args[0])
		if err != nil {
			return err
		}
		if len(reactions) == 0 {
			fmt.Println("No reactions on this message.")
			return nil
		}
		for _, r := range reactions {
			user := r.DisplayName
			if user == "" {
				user = r.UserName
			}
			if user == "" {
				user = "Unknown user"
			}
			fmt.Printf("  %s  %s (%s)\n", r.Emoji, user, r.UserName)
		}
		return nil
	},
}

var chatSearchCmd = &cobra.Command{
	Use:     "search [query]",
	Aliases: []string{"find"},
	Short:   "Search messages across all Google Chat conversations",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := chat.NewService(context.Background())
		if err != nil {
			return err
		}
		msgs, err := svc.SearchMessages(args[0], chatMax)
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

func init() {
	chatSendCmd.Flags().StringVar(&chatReplyTo, "reply-to", "", "Message ID to reply to (starts or continues thread)")
	chatSendCmd.Flags().StringVar(&chatAttachment, "attach", "", "Local file path to upload as an attachment")

	chatReplyCmd.Flags().StringVar(&chatAttachment, "attach", "", "Local file path to upload as an attachment")

	chatSearchCmd.Flags().Int64VarP(&chatMax, "max", "m", 20, "Maximum number of messages to return")
	chatSearchCmd.Flags().BoolVar(&chatJSON, "json", false, "Output results in JSON format")

	chatCmd.AddCommand(chatSendCmd)
	chatCmd.AddCommand(chatReplyCmd)
	chatCmd.AddCommand(chatReactCmd)
	chatCmd.AddCommand(chatReactionsCmd)
	chatCmd.AddCommand(chatSearchCmd)
}
