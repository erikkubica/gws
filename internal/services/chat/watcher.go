package chat

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// BurstBuffer aggregates rapid message bursts and flushes after a debounce duration.
type BurstBuffer struct {
	mu               sync.Mutex
	debounceDuration time.Duration
	lastMessageTime  time.Time
	messages         []*MessageInfo
	seenIDs          map[string]bool
}

// NewBurstBuffer creates a new buffer for debouncing chat messages.
func NewBurstBuffer(debounce time.Duration) *BurstBuffer {
	return &BurstBuffer{
		debounceDuration: debounce,
		seenIDs:          make(map[string]bool),
		messages:         make([]*MessageInfo, 0),
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

// Add appends a new message to the buffer and updates the burst timer.
func (b *BurstBuffer) Add(msg *MessageInfo, arrivalTime time.Time) bool {
	if msg == nil || msg.Name == "" {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.seenIDs[msg.Name] {
		return false
	}
	b.seenIDs[msg.Name] = true
	b.messages = append(b.messages, msg)
	b.lastMessageTime = arrivalTime
	return true
}

// ShouldFlush returns true if the buffer has messages and debounce has expired.
func (b *BurstBuffer) ShouldFlush(now time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.messages) == 0 {
		return false
	}
	if b.debounceDuration <= 0 {
		return true
	}
	return !now.Before(b.lastMessageTime.Add(b.debounceDuration))
}

// Flush returns all accumulated messages and resets the burst buffer.
func (b *BurstBuffer) Flush() []*MessageInfo {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.messages) == 0 {
		return nil
	}
	flushed := b.messages
	b.messages = make([]*MessageInfo, 0)
	return flushed
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
