package chat

import (
	"testing"
	"time"
)

func TestBurstBuffer_PerChannelDebounceIsolation(t *testing.T) {
	buffer := NewBurstBuffer(30 * time.Second)
	startTime := time.Date(2026, 10, 7, 10, 0, 0, 0, time.UTC)

	// Space A messages (at t=0s and t=10s)
	msgA1 := &MessageInfo{Name: "spaces/dev/messages/1", SenderName: "Jan", Text: "fix login"}
	msgA2 := &MessageInfo{Name: "spaces/dev/messages/2", SenderName: "Jan", Text: "make it red"}

	// Space B message (at t=15s)
	msgB1 := &MessageInfo{Name: "spaces/general/messages/1", SenderName: "Peter", Text: "lunch?"}

	buffer.Add("spaces/dev", msgA1, startTime)
	buffer.Add("spaces/dev", msgA2, startTime.Add(10*time.Second))
	buffer.Add("spaces/general", msgB1, startTime.Add(15*time.Second))

	// At t=35s: Space A timer expires at 10s + 30s = 40s. Neither should flush yet.
	if flushed := buffer.FlushedBursts(startTime.Add(35 * time.Second)); len(flushed) != 0 {
		t.Fatalf("expected 0 flushed bursts at 35s, got %d", len(flushed))
	}

	// At t=40s: Space A should flush alone (its 30s of silence passed from t=10s)
	// Space B should NOT flush because its message arrived at t=15s (needs until t=45s)
	flushed40 := buffer.FlushedBursts(startTime.Add(40 * time.Second))
	if len(flushed40) != 1 {
		t.Fatalf("expected 1 flushed burst at 40s, got %d", len(flushed40))
	}
	if flushed40[0].SpaceID != "spaces/dev" || len(flushed40[0].Messages) != 2 {
		t.Fatalf("unexpected burst flushed at 40s: %+v", flushed40[0])
	}

	// At t=45s: Space B should now flush
	flushed45 := buffer.FlushedBursts(startTime.Add(45 * time.Second))
	if len(flushed45) != 1 {
		t.Fatalf("expected 1 flushed burst at 45s, got %d", len(flushed45))
	}
	if flushed45[0].SpaceID != "spaces/general" || len(flushed45[0].Messages) != 1 {
		t.Fatalf("unexpected burst flushed at 45s: %+v", flushed45[0])
	}

	// Afterwards, buffer is empty
	if flushedEmpty := buffer.FlushedBursts(startTime.Add(50 * time.Second)); len(flushedEmpty) != 0 {
		t.Fatalf("expected 0 flushed bursts after drain, got %d", len(flushedEmpty))
	}
}

func TestBurstBuffer_ZeroDebounceBoundary(t *testing.T) {
	buffer := NewBurstBuffer(0)
	now := time.Now()

	msg := &MessageInfo{Name: "spaces/A/messages/100", SenderName: "Alice", Text: "urgent alert"}
	buffer.Add("spaces/A", msg, now)

	flushed := buffer.FlushedBursts(now)
	if len(flushed) != 1 || len(flushed[0].Messages) != 1 {
		t.Fatal("expected 1 flushed burst with 0 debounce")
	}
}

func TestBurstBuffer_DeduplicationAndMalformed(t *testing.T) {
	buffer := NewBurstBuffer(10 * time.Second)
	now := time.Now()

	if buffer.Add("spaces/A", nil, now) {
		t.Fatal("expected nil message to be rejected")
	}
	if buffer.Add("spaces/A", &MessageInfo{Name: ""}, now) {
		t.Fatal("expected empty name message to be rejected")
	}

	existing := []*MessageInfo{
		{Name: "spaces/A/messages/old1"},
	}
	buffer.SeedIDs(existing)

	if buffer.Add("spaces/A", &MessageInfo{Name: "spaces/A/messages/old1"}, now) {
		t.Fatal("expected seeded message to be rejected as duplicate")
	}
}

func TestExtractSpaceID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"spaces/AAAA4_P5byY/messages/msg123", "spaces/AAAA4_P5byY"},
		{"spaces/xRjOZgAAAAE/messages/xyz", "spaces/xRjOZgAAAAE"},
		{"invalidpath", "invalidpath"},
		{"", ""},
	}
	for _, tc := range tests {
		result := ExtractSpaceID(tc.input)
		if result != tc.expected {
			t.Errorf("ExtractSpaceID(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestFormatBurstText(t *testing.T) {
	msgs := []*MessageInfo{
		{SenderName: "Jan", Text: "could you do this pls"},
		{SenderName: "Jan", Text: "and make it red"},
	}
	formatted := FormatBurstText(msgs)
	expected := "[Jan]: could you do this pls\n[Jan]: and make it red"
	if formatted != expected {
		t.Fatalf("unexpected formatted text: got %q, want %q", formatted, expected)
	}
}

func TestBurstBuffer_BoundedSeenIDs(t *testing.T) {
	buffer := NewBurstBuffer(10 * time.Second)
	now := time.Now()

	// Insert maxSeenIDs + 500 distinct messages
	total := maxSeenIDs + 500
	for i := 0; i < total; i++ {
		msg := &MessageInfo{
			Name: "spaces/A/messages/" + time.Now().Format("20060102150405.000000") + "_" + string(rune('a'+(i%26))) + "_" + time.Duration(i).String(),
		}
		buffer.Add("spaces/A", msg, now)
	}

	buffer.mu.Lock()
	defer buffer.mu.Unlock()

	if len(buffer.seenIDs) > maxSeenIDs {
		t.Fatalf("seenIDs exceeded max capacity: got %d, max %d", len(buffer.seenIDs), maxSeenIDs)
	}
	if len(buffer.seenOrder) > maxSeenIDs {
		t.Fatalf("seenOrder exceeded max capacity: got %d, max %d", len(buffer.seenOrder), maxSeenIDs)
	}
}
