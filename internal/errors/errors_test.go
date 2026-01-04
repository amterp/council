package errors

import (
	"strings"
	"testing"
)

func TestSessionNotFoundError(t *testing.T) {
	err := &SessionNotFoundError{SessionID: "my-session"}
	msg := err.Error()

	if !strings.Contains(msg, "my-session") {
		t.Errorf("error should contain session ID, got %q", msg)
	}
	if !strings.Contains(msg, "not found") {
		t.Errorf("error should mention 'not found', got %q", msg)
	}
	if !strings.Contains(msg, "council new") {
		t.Errorf("error should suggest 'council new', got %q", msg)
	}
}

func TestNameTakenError(t *testing.T) {
	err := &NameTakenError{Name: "Alice"}
	msg := err.Error()

	if !strings.Contains(msg, "Alice") {
		t.Errorf("error should contain name, got %q", msg)
	}
	if !strings.Contains(msg, "already exists") {
		t.Errorf("error should mention 'already exists', got %q", msg)
	}
}

func TestReservedNameError(t *testing.T) {
	err := &ReservedNameError{Name: "Moderator"}
	msg := err.Error()

	if !strings.Contains(msg, "Moderator") {
		t.Errorf("error should contain name, got %q", msg)
	}
	if !strings.Contains(msg, "reserved") {
		t.Errorf("error should mention 'reserved', got %q", msg)
	}
}

func TestStaleStateError(t *testing.T) {
	err := &StaleStateError{
		ExpectedEventNum: 5,
		ActualEventNum:   7,
		SessionID:        "my-session",
	}
	msg := err.Error()

	if !strings.Contains(msg, "5") {
		t.Errorf("error should contain expected event num, got %q", msg)
	}
	if !strings.Contains(msg, "my-session") {
		t.Errorf("error should contain session ID, got %q", msg)
	}
	if !strings.Contains(msg, "council status") {
		t.Errorf("error should suggest 'council status', got %q", msg)
	}
}

func TestNotAParticipantError(t *testing.T) {
	err := &NotAParticipantError{Name: "Bob", SessionID: "my-session"}
	msg := err.Error()

	if !strings.Contains(msg, "Bob") {
		t.Errorf("error should contain name, got %q", msg)
	}
	if !strings.Contains(msg, "council join") {
		t.Errorf("error should suggest 'council join', got %q", msg)
	}
}

func TestParticipantNotInSessionError(t *testing.T) {
	err := &ParticipantNotInSessionError{Name: "Charlie", SessionID: "my-session"}
	msg := err.Error()

	if !strings.Contains(msg, "Charlie") {
		t.Errorf("error should contain name, got %q", msg)
	}
	if !strings.Contains(msg, "my-session") {
		t.Errorf("error should contain session ID, got %q", msg)
	}
	if !strings.Contains(msg, "not a participant") {
		t.Errorf("error should mention 'not a participant', got %q", msg)
	}
}

func TestInvalidNextParticipantError(t *testing.T) {
	err := &InvalidNextParticipantError{Name: "Dave"}
	msg := err.Error()

	if !strings.Contains(msg, "Dave") {
		t.Errorf("error should contain name, got %q", msg)
	}
	if !strings.Contains(msg, "--next") {
		t.Errorf("error should mention '--next', got %q", msg)
	}
}

func TestErrorInterface(t *testing.T) {
	// Verify all error types implement the error interface
	var _ error = &SessionNotFoundError{}
	var _ error = &NameTakenError{}
	var _ error = &ReservedNameError{}
	var _ error = &StaleStateError{}
	var _ error = &NotAParticipantError{}
	var _ error = &ParticipantNotInSessionError{}
	var _ error = &InvalidNextParticipantError{}
}
