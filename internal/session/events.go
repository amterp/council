package session

import (
	"encoding/json"
	"fmt"
	"time"
)

// EventType represents the type of session event
type EventType string

const (
	EventTypeSessionCreated EventType = "session_created"
	EventTypeJoined         EventType = "joined"
	EventTypeLeft           EventType = "left"
	EventTypeMessage        EventType = "message"
)

// Event is the interface for all event types
type Event interface {
	GetType() EventType
	GetTimestamp() int64
}

// BaseEvent contains common fields for all events
type BaseEvent struct {
	Type            EventType `json:"type"`
	TimestampMillis int64     `json:"timestamp_millis"`
}

func (e *BaseEvent) GetType() EventType {
	return e.Type
}

func (e *BaseEvent) GetTimestamp() int64 {
	return e.TimestampMillis
}

// SessionCreatedEvent represents session creation
type SessionCreatedEvent struct {
	BaseEvent
	ID string `json:"id"`
}

// JoinedEvent represents a participant joining
type JoinedEvent struct {
	BaseEvent
	Participant string `json:"participant"`
}

// LeftEvent represents a participant leaving
type LeftEvent struct {
	BaseEvent
	Participant string `json:"participant"`
}

// MessageEvent represents a message posted
type MessageEvent struct {
	BaseEvent
	Participant string `json:"participant"`
	Content     string `json:"content"`
}

// Now returns the current timestamp in milliseconds
func Now() int64 {
	return time.Now().UnixMilli()
}

// NewSessionCreatedEvent creates a new session_created event
func NewSessionCreatedEvent(id string) *SessionCreatedEvent {
	return &SessionCreatedEvent{
		BaseEvent: BaseEvent{
			Type:            EventTypeSessionCreated,
			TimestampMillis: Now(),
		},
		ID: id,
	}
}

// NewJoinedEvent creates a new joined event
func NewJoinedEvent(participant string) *JoinedEvent {
	return &JoinedEvent{
		BaseEvent: BaseEvent{
			Type:            EventTypeJoined,
			TimestampMillis: Now(),
		},
		Participant: participant,
	}
}

// NewLeftEvent creates a new left event
func NewLeftEvent(participant string) *LeftEvent {
	return &LeftEvent{
		BaseEvent: BaseEvent{
			Type:            EventTypeLeft,
			TimestampMillis: Now(),
		},
		Participant: participant,
	}
}

// NewMessageEvent creates a new message event
func NewMessageEvent(participant, content string) *MessageEvent {
	return &MessageEvent{
		BaseEvent: BaseEvent{
			Type:            EventTypeMessage,
			TimestampMillis: Now(),
		},
		Participant: participant,
		Content:     content,
	}
}

// rawEvent is used for initial JSON parsing to determine event type
type rawEvent struct {
	Type EventType `json:"type"`
}

// ParseEvent parses a JSON line into the appropriate event type
func ParseEvent(line []byte) (Event, error) {
	var raw rawEvent
	if err := json.Unmarshal(line, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse event type: %w", err)
	}

	var event Event
	switch raw.Type {
	case EventTypeSessionCreated:
		var e SessionCreatedEvent
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, fmt.Errorf("failed to parse session_created event: %w", err)
		}
		event = &e
	case EventTypeJoined:
		var e JoinedEvent
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, fmt.Errorf("failed to parse joined event: %w", err)
		}
		event = &e
	case EventTypeLeft:
		var e LeftEvent
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, fmt.Errorf("failed to parse left event: %w", err)
		}
		event = &e
	case EventTypeMessage:
		var e MessageEvent
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, fmt.Errorf("failed to parse message event: %w", err)
		}
		event = &e
	default:
		return nil, fmt.Errorf("unknown event type: %s", raw.Type)
	}

	return event, nil
}

// MarshalEvent serializes an event to JSON
func MarshalEvent(e Event) ([]byte, error) {
	return json.Marshal(e)
}
