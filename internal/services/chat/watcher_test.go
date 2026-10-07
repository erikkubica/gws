package chat

import (
	"testing"
	"time"
)

func TestBurstBuffer_NominalDebounce(t *testing.T) {
	buffer := NewBurstBuffer(30 * time.Second)
	startTime := time.Date(2026, 10, 7, 10, 0, 0, 0, time.UTC)

	msg1 := &MessageInfo{Name: "spaces/A/messages/1", SenderName: "Jan", Text: "could you do this pls"}
	msg2 := &MessageInfo{Name: "spaces/A/messages/2", SenderName: "Jan", Text: "and make it red"}

	if !buffer.Add(msg1, startTime) {
		t.Fatal("expected msg1 to be added")
	}
	if buffer.ShouldFlush(startTime.Add(10 * time.Second)) {
		t.Fatal("buffer should not flush before debounce expires")
	}

	// Add second message at +15s (timer resets)
	if !buffer.Add(msg2, startTime.Add(15*time.Second)) {
		t.Fatal("expected msg2 to be added")
	}
	if buffer.ShouldFlush(startTime.Add(35 * time.Second)) {
		t.Fatal("buffer should not flush because msg2 reset timer to 15s + 30s = 45s")
	}

	// Check boundary threshold at 45s
	if !buffer.ShouldFlush(startTime.Add(45 * time.Second)) {
		t.Fatal("buffer should flush at exact debounce threshold (45s)")
	}

	flushed := buffer.Flush()
	if len(flushed) != 2 {
		t.Fatalf("expected 2 flushed messages, got %d", len(flushed))
	}
	if buffer.ShouldFlush(startTime.Add(50 * time.Second)) {
		t.Fatal("buffer should not flush when empty")
	}
}

func TestBurstBuffer_ZeroDebounceBoundary(t *testing.T) {
	buffer := NewBurstBuffer(0)
	now := time.Now()

	if buffer.ShouldFlush(now) {
		t.Fatal("empty buffer should not flush even with 0 debounce")
	}

	msg := &MessageInfo{Name: "spaces/A/messages/100", SenderName: "Alice", Text: "urgent alert"}
	if !buffer.Add(msg, now) {
		t.Fatal("expected message to be added")
	}
	if !buffer.ShouldFlush(now) {
		t.Fatal("buffer with 0 debounce should flush immediately upon message arrival")
	}
	flushed := buffer.Flush()
	if len(flushed) != 1 || flushed[0].Name != msg.Name {
		t.Fatal("expected single flushed message")
	}
}

func TestBurstBuffer_DeduplicationAndMalformed(t *testing.T) {
	buffer := NewBurstBuffer(10 * time.Second)
	now := time.Now()

	// Malformed inputs
	if buffer.Add(nil, now) {
		t.Fatal("expected nil message to be rejected")
	}
	if buffer.Add(&MessageInfo{Name: ""}, now) {
		t.Fatal("expected empty name message to be rejected")
	}

	// Pre-seed
	existing := []*MessageInfo{
		{Name: "spaces/A/messages/old1"},
		{Name: "spaces/A/messages/old2"},
	}
	buffer.SeedIDs(existing)

	// Attempt to re-add existing
	if buffer.Add(&MessageInfo{Name: "spaces/A/messages/old1", Text: "duplicate"}, now) {
		t.Fatal("expected seeded message to be rejected as duplicate")
	}

	// Add new message twice
	newMsg := &MessageInfo{Name: "spaces/A/messages/new1", Text: "hello"}
	if !buffer.Add(newMsg, now) {
		t.Fatal("expected first add to succeed")
	}
	if buffer.Add(newMsg, now.Add(1*time.Second)) {
		t.Fatal("expected second add to be rejected as duplicate")
	}
}

func TestFormatBurstText(t *testing.T) {
	if FormatBurstText(nil) != "" {
		t.Fatal("expected empty string for nil messages")
	}
	if FormatBurstText([]*MessageInfo{}) != "" {
		t.Fatal("expected empty string for empty slice")
	}

	msgs := []*MessageInfo{
		{SenderName: "Jan", Text: "could you do this pls"},
		{SenderName: "Jan", Text: "and make it red"},
		{SenderName: "", Text: "thanks!"},
	}
	formatted := FormatBurstText(msgs)
	expected := "[Jan]: could you do this pls\n[Jan]: and make it red\n[User]: thanks!"
	if formatted != expected {
		t.Fatalf("unexpected formatted text: got %q, want %q", formatted, expected)
	}
}
