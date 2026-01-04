package session

import (
	"encoding/json"
	"testing"
)

func TestParseEventSessionCreated(t *testing.T) {
	input := `{"type":"session_created","timestamp_millis":1234567890,"id":"my-session"}`

	event, err := ParseEvent([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	created, ok := event.(*SessionCreatedEvent)
	if !ok {
		t.Fatalf("expected *SessionCreatedEvent, got %T", event)
	}

	if created.GetType() != EventTypeSessionCreated {
		t.Errorf("expected type %q, got %q", EventTypeSessionCreated, created.GetType())
	}
	if created.GetTimestamp() != 1234567890 {
		t.Errorf("expected timestamp 1234567890, got %d", created.GetTimestamp())
	}
	if created.ID != "my-session" {
		t.Errorf("expected ID 'my-session', got %q", created.ID)
	}
}

func TestParseEventJoined(t *testing.T) {
	input := `{"type":"joined","timestamp_millis":1234567890,"participant":"Alice"}`

	event, err := ParseEvent([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	joined, ok := event.(*JoinedEvent)
	if !ok {
		t.Fatalf("expected *JoinedEvent, got %T", event)
	}

	if joined.GetType() != EventTypeJoined {
		t.Errorf("expected type %q, got %q", EventTypeJoined, joined.GetType())
	}
	if joined.Participant != "Alice" {
		t.Errorf("expected participant 'Alice', got %q", joined.Participant)
	}
}

func TestParseEventLeft(t *testing.T) {
	input := `{"type":"left","timestamp_millis":1234567890,"participant":"Bob"}`

	event, err := ParseEvent([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	left, ok := event.(*LeftEvent)
	if !ok {
		t.Fatalf("expected *LeftEvent, got %T", event)
	}

	if left.GetType() != EventTypeLeft {
		t.Errorf("expected type %q, got %q", EventTypeLeft, left.GetType())
	}
	if left.Participant != "Bob" {
		t.Errorf("expected participant 'Bob', got %q", left.Participant)
	}
}

func TestParseEventMessage(t *testing.T) {
	input := `{"type":"message","timestamp_millis":1234567890,"participant":"Alice","content":"Hello world","next":"Bob"}`

	event, err := ParseEvent([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg, ok := event.(*MessageEvent)
	if !ok {
		t.Fatalf("expected *MessageEvent, got %T", event)
	}

	if msg.GetType() != EventTypeMessage {
		t.Errorf("expected type %q, got %q", EventTypeMessage, msg.GetType())
	}
	if msg.Participant != "Alice" {
		t.Errorf("expected participant 'Alice', got %q", msg.Participant)
	}
	if msg.Content != "Hello world" {
		t.Errorf("expected content 'Hello world', got %q", msg.Content)
	}
	if msg.Next != "Bob" {
		t.Errorf("expected next 'Bob', got %q", msg.Next)
	}
}

func TestParseEventInvalidJSON(t *testing.T) {
	input := `not valid json`

	_, err := ParseEvent([]byte(input))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseEventUnknownType(t *testing.T) {
	input := `{"type":"unknown","timestamp_millis":1234567890}`

	_, err := ParseEvent([]byte(input))
	if err == nil {
		t.Error("expected error for unknown event type")
	}
}

func TestParseEventMalformedSessionCreated(t *testing.T) {
	// Valid type but malformed payload
	input := `{"type":"session_created","timestamp_millis":"not_a_number"}`

	_, err := ParseEvent([]byte(input))
	if err == nil {
		t.Error("expected error for malformed event")
	}
}

func TestMarshalEvent(t *testing.T) {
	tests := []struct {
		name  string
		event Event
	}{
		{
			name:  "session_created",
			event: &SessionCreatedEvent{BaseEvent: BaseEvent{Type: EventTypeSessionCreated, TimestampMillis: 1234567890}, ID: "test"},
		},
		{
			name:  "joined",
			event: &JoinedEvent{BaseEvent: BaseEvent{Type: EventTypeJoined, TimestampMillis: 1234567890}, Participant: "Alice"},
		},
		{
			name:  "left",
			event: &LeftEvent{BaseEvent: BaseEvent{Type: EventTypeLeft, TimestampMillis: 1234567890}, Participant: "Bob"},
		},
		{
			name:  "message",
			event: &MessageEvent{BaseEvent: BaseEvent{Type: EventTypeMessage, TimestampMillis: 1234567890}, Participant: "Alice", Content: "Hello", Next: "Bob"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := MarshalEvent(tt.event)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify it's valid JSON
			var raw map[string]any
			if err := json.Unmarshal(data, &raw); err != nil {
				t.Fatalf("marshaled event is not valid JSON: %v", err)
			}

			// Verify type field
			if raw["type"] != string(tt.event.GetType()) {
				t.Errorf("expected type %q, got %q", tt.event.GetType(), raw["type"])
			}
		})
	}
}

func TestMarshalAndParseRoundTrip(t *testing.T) {
	original := NewMessageEvent("Alice", "Hello, world!", "Bob")

	data, err := MarshalEvent(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	parsed, err := ParseEvent(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	msg, ok := parsed.(*MessageEvent)
	if !ok {
		t.Fatalf("expected *MessageEvent, got %T", parsed)
	}

	if msg.Participant != original.Participant {
		t.Errorf("participant mismatch: %q vs %q", msg.Participant, original.Participant)
	}
	if msg.Content != original.Content {
		t.Errorf("content mismatch: %q vs %q", msg.Content, original.Content)
	}
	if msg.Next != original.Next {
		t.Errorf("next mismatch: %q vs %q", msg.Next, original.Next)
	}
}

func TestNewEventConstructors(t *testing.T) {
	t.Run("NewSessionCreatedEvent", func(t *testing.T) {
		e := NewSessionCreatedEvent("my-id")
		if e.Type != EventTypeSessionCreated {
			t.Errorf("wrong type: %q", e.Type)
		}
		if e.ID != "my-id" {
			t.Errorf("wrong ID: %q", e.ID)
		}
		if e.TimestampMillis == 0 {
			t.Error("timestamp should be set")
		}
	})

	t.Run("NewJoinedEvent", func(t *testing.T) {
		e := NewJoinedEvent("Alice")
		if e.Type != EventTypeJoined {
			t.Errorf("wrong type: %q", e.Type)
		}
		if e.Participant != "Alice" {
			t.Errorf("wrong participant: %q", e.Participant)
		}
	})

	t.Run("NewLeftEvent", func(t *testing.T) {
		e := NewLeftEvent("Bob")
		if e.Type != EventTypeLeft {
			t.Errorf("wrong type: %q", e.Type)
		}
		if e.Participant != "Bob" {
			t.Errorf("wrong participant: %q", e.Participant)
		}
	})

	t.Run("NewMessageEvent", func(t *testing.T) {
		e := NewMessageEvent("Alice", "content", "Bob")
		if e.Type != EventTypeMessage {
			t.Errorf("wrong type: %q", e.Type)
		}
		if e.Participant != "Alice" {
			t.Errorf("wrong participant: %q", e.Participant)
		}
		if e.Content != "content" {
			t.Errorf("wrong content: %q", e.Content)
		}
		if e.Next != "Bob" {
			t.Errorf("wrong next: %q", e.Next)
		}
	})
}
