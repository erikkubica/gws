package chat

import (
	"fmt"
	"strings"

	"google.golang.org/api/chat/v1"
)

// NormalizeMessageName ensures the resource name follows 'spaces/{space}/messages/{message}'.
func NormalizeMessageName(spaceName, messageID string) string {
	cleanMsg := strings.TrimSpace(messageID)
	if strings.HasPrefix(cleanMsg, "spaces/") && strings.Contains(cleanMsg, "/messages/") {
		return cleanMsg
	}
	cleanSpace := NormalizeSpaceName(spaceName)
	return fmt.Sprintf("%s/messages/%s", cleanSpace, cleanMsg)
}

// AddReaction adds an emoji reaction to a Google Chat message.
func (s *Service) AddReaction(messageName, emoji string) error {
	cleanMsg := strings.TrimSpace(messageName)
	if cleanMsg == "" {
		return fmt.Errorf("message name cannot be empty")
	}
	cleanEmoji := strings.TrimSpace(emoji)
	if cleanEmoji == "" {
		return fmt.Errorf("emoji cannot be empty")
	}

	reaction := &chat.Reaction{
		Emoji: &chat.Emoji{
			Unicode: cleanEmoji,
		},
	}
	_, err := s.client.Spaces.Messages.Reactions.Create(cleanMsg, reaction).Do()
	if err != nil {
		return fmt.Errorf("add reaction to %s: %w", cleanMsg, err)
	}
	return nil
}

// UserReaction represents an emoji reaction with the identity of the user who reacted.
type UserReaction struct {
	Emoji       string `json:"emoji"`
	UserName    string `json:"user_name,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

// ListReactions returns the emoji reactions and the users who reacted on a message.
func (s *Service) ListReactions(messageName string) ([]UserReaction, error) {
	cleanMsg := strings.TrimSpace(messageName)
	if cleanMsg == "" {
		return nil, fmt.Errorf("message name cannot be empty")
	}
	res, err := s.client.Spaces.Messages.Reactions.List(cleanMsg).Do()
	if err != nil {
		return nil, fmt.Errorf("list reactions for %s: %w", cleanMsg, err)
	}
	var reactions []UserReaction
	for _, r := range res.Reactions {
		reactions = append(reactions, formatUserReaction(r))
	}
	return reactions, nil
}

func formatUserReaction(r *chat.Reaction) UserReaction {
	emoji := ""
	if r.Emoji != nil {
		emoji = r.Emoji.Unicode
	}
	userName, displayName := "", ""
	if r.User != nil {
		userName = r.User.Name
		displayName = r.User.DisplayName
	}
	return UserReaction{
		Emoji:       emoji,
		UserName:    userName,
		DisplayName: displayName,
	}
}
