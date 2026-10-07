package cli

import (
	"context"
	"fmt"
	"strings"

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
	Short: "List emoji reactions on a specific message",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := chat.NewService(context.Background())
		if err != nil {
			return err
		}
		emojis, err := svc.ListReactions(args[0])
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

var chatWebhookCmd = &cobra.Command{
	Use:   "webhook [webhook_url] [message_text]",
	Short: "Post a message to an incoming Google Chat webhook URL",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := chat.NewService(context.Background())
		if err != nil {
			return err
		}
		if err := svc.SendWebhook(args[0], args[1]); err != nil {
			return err
		}
		fmt.Println("Webhook message delivered successfully!")
		return nil
	},
}

func init() {
	chatSendCmd.Flags().StringVar(&chatReplyTo, "reply-to", "", "Message ID to reply to (starts or continues thread)")
	chatSendCmd.Flags().StringVar(&chatAttachment, "attach", "", "Local file path to upload as an attachment")

	chatReplyCmd.Flags().StringVar(&chatAttachment, "attach", "", "Local file path to upload as an attachment")

	chatCmd.AddCommand(chatSendCmd)
	chatCmd.AddCommand(chatReplyCmd)
	chatCmd.AddCommand(chatReactCmd)
	chatCmd.AddCommand(chatReactionsCmd)
	chatCmd.AddCommand(chatWebhookCmd)
}
