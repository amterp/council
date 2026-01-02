package errors

import "fmt"

// SessionNotFoundError indicates the session file does not exist
type SessionNotFoundError struct {
	SessionID string
}

func (e *SessionNotFoundError) Error() string {
	return fmt.Sprintf("Session '%s' not found. Run 'council new' to create a session.", e.SessionID)
}

// NameTakenError indicates the participant name is already in use
type NameTakenError struct {
	Name string
}

func (e *NameTakenError) Error() string {
	return fmt.Sprintf("Participant '%s' already exists in this session. Choose a different name.", e.Name)
}

// ReservedNameError indicates the name is reserved
type ReservedNameError struct {
	Name string
}

func (e *ReservedNameError) Error() string {
	return fmt.Sprintf("'%s' is a reserved name. Choose a different name.", e.Name)
}

// StaleStateError indicates optimistic lock failure
type StaleStateError struct {
	ExpectedEventNum int
	ActualEventNum   int
	SessionID        string
}

func (e *StaleStateError) Error() string {
	return fmt.Sprintf("New activity since event #%d. Re-read with 'council status %s --after %d' before posting.",
		e.ExpectedEventNum, e.SessionID, e.ExpectedEventNum)
}

// NotAParticipantError indicates the user hasn't joined
type NotAParticipantError struct {
	Name      string
	SessionID string
}

func (e *NotAParticipantError) Error() string {
	return fmt.Sprintf("'%s' must join the session before posting. Run 'council join %s'.", e.Name, e.SessionID)
}

// ParticipantNotInSessionError indicates participant not found for leave
type ParticipantNotInSessionError struct {
	Name      string
	SessionID string
}

func (e *ParticipantNotInSessionError) Error() string {
	return fmt.Sprintf("'%s' is not a participant in session '%s'.", e.Name, e.SessionID)
}
