package session

import (
	"strings"
	"testing"
)

func TestNewSession(t *testing.T) {
	s := NewSession("test-session")
	if s.ID != "test-session" {
		t.Errorf("expected ID 'test-session', got %q", s.ID)
	}
	if len(s.Events) != 0 {
		t.Errorf("expected 0 events, got %d", len(s.Events))
	}
	if len(s.Participants) != 0 {
		t.Errorf("expected 0 participants, got %d", len(s.Participants))
	}
}

func TestEventCount(t *testing.T) {
	s := NewSession("test")
	if s.EventCount() != 0 {
		t.Errorf("expected 0 events, got %d", s.EventCount())
	}

	s.addEvent(NewJoinedEvent("Alice"))
	if s.EventCount() != 1 {
		t.Errorf("expected 1 event, got %d", s.EventCount())
	}

	s.addEvent(NewJoinedEvent("Bob"))
	if s.EventCount() != 2 {
		t.Errorf("expected 2 events, got %d", s.EventCount())
	}
}

func TestActiveParticipants(t *testing.T) {
	s := NewSession("test")

	// Initially empty
	active := s.ActiveParticipants()
	if len(active) != 0 {
		t.Errorf("expected 0 active participants, got %d", len(active))
	}

	// Add participants
	s.addEvent(NewJoinedEvent("Alice"))
	s.addEvent(NewJoinedEvent("Bob"))

	active = s.ActiveParticipants()
	if len(active) != 2 {
		t.Errorf("expected 2 active participants, got %d", len(active))
	}

	// One leaves
	s.addEvent(NewLeftEvent("Bob"))
	active = s.ActiveParticipants()
	if len(active) != 1 {
		t.Errorf("expected 1 active participant, got %d", len(active))
	}
	if active[0] != "Alice" {
		t.Errorf("expected Alice to be active, got %q", active[0])
	}
}

func TestActiveParticipantsExcludesModerator(t *testing.T) {
	s := NewSession("test")
	s.Participants["Moderator"] = true
	s.Participants["Alice"] = true

	active := s.ActiveParticipants()
	if len(active) != 1 {
		t.Errorf("expected 1 active participant (excluding Moderator), got %d", len(active))
	}
	if active[0] != "Alice" {
		t.Errorf("expected Alice, got %q", active[0])
	}
}

func TestIsActiveParticipant(t *testing.T) {
	s := NewSession("test")

	if s.IsActiveParticipant("Alice") {
		t.Error("Alice should not be active initially")
	}

	s.addEvent(NewJoinedEvent("Alice"))
	if !s.IsActiveParticipant("Alice") {
		t.Error("Alice should be active after joining")
	}

	s.addEvent(NewLeftEvent("Alice"))
	if s.IsActiveParticipant("Alice") {
		t.Error("Alice should not be active after leaving")
	}
}

func TestPreviousSpeaker(t *testing.T) {
	s := NewSession("test")

	// No messages yet
	prev := s.PreviousSpeaker("Alice")
	if prev != "" {
		t.Errorf("expected empty string, got %q", prev)
	}

	// Add messages
	s.addEvent(NewMessageEvent("Alice", "hello", "Bob"))
	s.addEvent(NewMessageEvent("Bob", "hi", "Alice"))
	s.addEvent(NewMessageEvent("Alice", "how are you", "Bob"))

	// Previous speaker before Alice is Bob
	prev = s.PreviousSpeaker("Alice")
	if prev != "Bob" {
		t.Errorf("expected Bob, got %q", prev)
	}

	// Previous speaker before Bob is Alice
	prev = s.PreviousSpeaker("Bob")
	if prev != "Alice" {
		t.Errorf("expected Alice, got %q", prev)
	}
}

func TestPreviousSpeakerSkipsSelf(t *testing.T) {
	s := NewSession("test")
	s.addEvent(NewMessageEvent("Alice", "msg1", "Bob"))
	s.addEvent(NewMessageEvent("Alice", "msg2", "Bob"))
	s.addEvent(NewMessageEvent("Alice", "msg3", "Bob"))

	// Should return empty since only Alice has posted
	prev := s.PreviousSpeaker("Alice")
	if prev != "" {
		t.Errorf("expected empty string when only self has posted, got %q", prev)
	}
}

func TestRandomActiveParticipant(t *testing.T) {
	s := NewSession("test")

	// No participants
	result := s.RandomActiveParticipant("Alice")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}

	// Add participants
	s.addEvent(NewJoinedEvent("Alice"))
	s.addEvent(NewJoinedEvent("Bob"))
	s.addEvent(NewJoinedEvent("Charlie"))

	// Should return someone other than Alice
	result = s.RandomActiveParticipant("Alice")
	if result == "" {
		t.Error("expected a participant, got empty string")
	}
	if result == "Alice" {
		t.Error("should not return excluded participant")
	}
}

func TestRandomActiveParticipantAllExcluded(t *testing.T) {
	s := NewSession("test")
	s.addEvent(NewJoinedEvent("Alice"))

	// Alice is the only participant, exclude her
	result := s.RandomActiveParticipant("Alice")
	if result != "" {
		t.Errorf("expected empty string when all excluded, got %q", result)
	}
}

func TestLatestMessageNext(t *testing.T) {
	s := NewSession("test")

	// No messages
	next := s.LatestMessageNext()
	if next != "" {
		t.Errorf("expected empty string, got %q", next)
	}

	// Add messages
	s.addEvent(NewMessageEvent("Alice", "hello", "Bob"))
	next = s.LatestMessageNext()
	if next != "Bob" {
		t.Errorf("expected Bob, got %q", next)
	}

	s.addEvent(NewMessageEvent("Bob", "hi", "Charlie"))
	next = s.LatestMessageNext()
	if next != "Charlie" {
		t.Errorf("expected Charlie, got %q", next)
	}
}

func TestLatestMessageNextIgnoresNonMessageEvents(t *testing.T) {
	s := NewSession("test")
	s.addEvent(NewMessageEvent("Alice", "hello", "Bob"))
	s.addEvent(NewJoinedEvent("Charlie"))
	s.addEvent(NewLeftEvent("Dave"))

	// Should still return Bob from the message event
	next := s.LatestMessageNext()
	if next != "Bob" {
		t.Errorf("expected Bob, got %q", next)
	}
}

func TestAddEventUpdatesParticipantState(t *testing.T) {
	s := NewSession("test")

	s.addEvent(NewJoinedEvent("Alice"))
	if !s.Participants["Alice"] {
		t.Error("Alice should be active after joined event")
	}

	s.addEvent(NewLeftEvent("Alice"))
	if s.Participants["Alice"] {
		t.Error("Alice should be inactive after left event")
	}
}

func TestReadSessionFromReader(t *testing.T) {
	input := `{"type":"session_created","timestamp_millis":1234567890,"id":"test-session"}
{"type":"joined","timestamp_millis":1234567891,"participant":"Alice"}
{"type":"joined","timestamp_millis":1234567892,"participant":"Bob"}
{"type":"message","timestamp_millis":1234567893,"participant":"Alice","content":"Hello","next":"Bob"}
{"type":"left","timestamp_millis":1234567894,"participant":"Bob"}
`

	session, err := readSessionFromReader("test-session", strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if session.ID != "test-session" {
		t.Errorf("expected ID 'test-session', got %q", session.ID)
	}

	if len(session.Events) != 5 {
		t.Errorf("expected 5 events, got %d", len(session.Events))
	}

	if !session.IsActiveParticipant("Alice") {
		t.Error("Alice should be active")
	}

	if session.IsActiveParticipant("Bob") {
		t.Error("Bob should not be active (left)")
	}
}

func TestReadSessionFromReaderEmptyLines(t *testing.T) {
	input := `{"type":"session_created","timestamp_millis":1234567890,"id":"test"}

{"type":"joined","timestamp_millis":1234567891,"participant":"Alice"}

`

	session, err := readSessionFromReader("test", strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(session.Events) != 2 {
		t.Errorf("expected 2 events (empty lines ignored), got %d", len(session.Events))
	}
}

func TestReadSessionFromReaderInvalidJSON(t *testing.T) {
	input := `{"type":"session_created","timestamp_millis":1234567890,"id":"test"}
not valid json
`

	_, err := readSessionFromReader("test", strings.NewReader(input))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestReadSessionFromReaderUnknownEventType(t *testing.T) {
	input := `{"type":"unknown_type","timestamp_millis":1234567890}
`

	_, err := readSessionFromReader("test", strings.NewReader(input))
	if err == nil {
		t.Error("expected error for unknown event type")
	}
}
