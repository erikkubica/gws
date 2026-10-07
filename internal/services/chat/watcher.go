package chat

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// FlushedBurst contains the messages belonging to an expired debounce window for a specific space.
type FlushedBurst struct {
	SpaceID  string         `json:"space_id"`
	Messages []*MessageInfo `json:"messages"`
}

type spaceBuffer struct {
	lastMessageTime time.Time
	messages        []*MessageInfo
}

// BurstBuffer manages isolated debounce buffers per Google Chat space.
type BurstBuffer struct {
	mu               sync.Mutex
	debounceDuration time.Duration
	buffers          map[string]*spaceBuffer
	seenIDs          map[string]bool
}

// NewBurstBuffer creates a new buffer for per-space message debouncing.
func NewBurstBuffer(debounce time.Duration) *BurstBuffer {
	return &BurstBuffer{
		debounceDuration: debounce,
		buffers:          make(map[string]*spaceBuffer),
		seenIDs:          make(map[string]bool),
	}
}

// SeedIDs marks pre-existing message IDs as seen to prevent re-processing.
func (b *BurstBuffer) SeedIDs(msgs []*MessageInfo) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, m := range msgs {
		if m != nil && m.Name != "" {
			b.seenIDs[m.Name] = true
		}
	}
}

// Add appends a message to its space's isolated buffer and resets that space's debounce timer.
func (b *BurstBuffer) Add(spaceID string, msg *MessageInfo, arrivalTime time.Time) bool {
	if msg == nil || msg.Name == "" {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.seenIDs[msg.Name] {
		return false
	}
	b.seenIDs[msg.Name] = true

	targetSpace := spaceID
	if targetSpace == "" || targetSpace == "all" {
		targetSpace = ExtractSpaceID(msg.Name)
	}

	sb, exists := b.buffers[targetSpace]
	if !exists {
		sb = &spaceBuffer{messages: make([]*MessageInfo, 0)}
		b.buffers[targetSpace] = sb
	}
	sb.messages = append(sb.messages, msg)
	sb.lastMessageTime = arrivalTime
	return true
}

// FlushedBursts checks all space buffers and returns bursts whose debounce window has elapsed.
func (b *BurstBuffer) FlushedBursts(now time.Time) []FlushedBurst {
	b.mu.Lock()
	defer b.mu.Unlock()
	var flushed []FlushedBurst

	for spaceID, sb := range b.buffers {
		if len(sb.messages) == 0 {
			continue
		}
		if b.debounceDuration <= 0 || !now.Before(sb.lastMessageTime.Add(b.debounceDuration)) {
			flushed = append(flushed, FlushedBurst{
				SpaceID:  spaceID,
				Messages: sb.messages,
			})
			delete(b.buffers, spaceID)
		}
	}
	return flushed
}

// ExtractSpaceID parses the parent space name from a full message resource path.
func ExtractSpaceID(messageResourceName string) string {
	parts := strings.Split(messageResourceName, "/")
	if len(parts) >= 2 {
		return parts[0] + "/" + parts[1]
	}
	return messageResourceName
}

// FormatBurstText formats a burst of messages into a synthesized conversational text.
func FormatBurstText(msgs []*MessageInfo) string {
	if len(msgs) == 0 {
		return ""
	}
	var lines []string
	for _, m := range msgs {
		if m == nil {
			continue
		}
		sender := m.SenderName
		if sender == "" {
			sender = "User"
		}
		cleanText := strings.TrimSpace(m.Text)
		if cleanText != "" {
			lines = append(lines, fmt.Sprintf("[%s]: %s", sender, cleanText))
		}
	}
	return strings.Join(lines, "\n")
}
