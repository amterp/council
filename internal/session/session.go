package session

import (
	"bufio"
	"io"
	"os"
	"syscall"

	"github.com/amterp/council/internal/errors"
	"github.com/amterp/council/internal/storage"
)

// Session represents the in-memory state of a session
type Session struct {
	ID           string
	Events       []Event
	Participants map[string]bool // currently active participants (true = joined, false = left)
}

// NewSession creates a new empty session with the given ID
func NewSession(id string) *Session {
	return &Session{
		ID:           id,
		Events:       make([]Event, 0),
		Participants: make(map[string]bool),
	}
}

// EventCount returns the number of events in the session
func (s *Session) EventCount() int {
	return len(s.Events)
}

// ActiveParticipants returns a list of currently active participants (excluding Moderator)
func (s *Session) ActiveParticipants() []string {
	result := []string{}
	for name, active := range s.Participants {
		if active && name != "Moderator" {
			result = append(result, name)
		}
	}
	return result
}

// IsActiveParticipant checks if a name is currently an active participant
func (s *Session) IsActiveParticipant(name string) bool {
	return s.Participants[name]
}

// PreviousSpeaker returns the participant who posted the message before the given one
// Returns empty string if there's no previous message
func (s *Session) PreviousSpeaker(excludeParticipant string) string {
	// Walk backwards through events to find the last message not by excludeParticipant
	for i := len(s.Events) - 1; i >= 0; i-- {
		if msg, ok := s.Events[i].(*MessageEvent); ok {
			if msg.Participant != excludeParticipant {
				return msg.Participant
			}
		}
	}
	return ""
}

// RandomActiveParticipant returns a random active participant excluding the given name
func (s *Session) RandomActiveParticipant(exclude string) string {
	active := s.ActiveParticipants()
	var candidates []string
	for _, name := range active {
		if name != exclude {
			candidates = append(candidates, name)
		}
	}
	if len(candidates) == 0 {
		return ""
	}
	// For determinism in tests, just pick the first one (alphabetically sorted would be better)
	// In practice, we could use rand but for now this is fine
	return candidates[0]
}

// LatestMessageNext returns the Next field from the most recent message event
// Returns empty string if no messages exist
func (s *Session) LatestMessageNext() string {
	for i := len(s.Events) - 1; i >= 0; i-- {
		if msg, ok := s.Events[i].(*MessageEvent); ok {
			return msg.Next
		}
	}
	return ""
}

// addEvent adds an event and updates participant state
func (s *Session) addEvent(event Event) {
	s.Events = append(s.Events, event)

	switch e := event.(type) {
	case *JoinedEvent:
		s.Participants[e.Participant] = true
	case *LeftEvent:
		s.Participants[e.Participant] = false
	}
}

// LoadSession reads and parses a session file
func LoadSession(sessionID string) (*Session, error) {
	path, err := storage.SessionEventsPath(sessionID)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, &errors.SessionNotFoundError{SessionID: sessionID}
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return readSessionFromReader(sessionID, file)
}

// readSessionFromReader parses session events from a reader
func readSessionFromReader(sessionID string, r io.Reader) (*Session, error) {
	session := NewSession(sessionID)
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		event, err := ParseEvent(line)
		if err != nil {
			return nil, err
		}
		session.addEvent(event)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return session, nil
}

// FileLocker provides exclusive file locking
type FileLocker struct {
	file *os.File
}

// AcquireLock opens the file and acquires an exclusive lock
func AcquireLock(path string) (*FileLocker, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	// Acquire exclusive lock (blocking)
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, err
	}

	return &FileLocker{file: f}, nil
}

// Release unlocks and closes the file
func (l *FileLocker) Release() error {
	syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	return l.file.Close()
}

// File returns the underlying file for reading/writing
func (l *FileLocker) File() *os.File {
	return l.file
}

// CreateSession creates a new session file with a session_created event
func CreateSession(sessionID string) error {
	if err := storage.EnsureSessionDir(sessionID); err != nil {
		return err
	}

	path, err := storage.SessionEventsPath(sessionID)
	if err != nil {
		return err
	}

	lock, err := AcquireLock(path)
	if err != nil {
		return err
	}
	defer lock.Release()

	event := NewSessionCreatedEvent(sessionID)
	eventBytes, err := MarshalEvent(event)
	if err != nil {
		return err
	}

	_, err = lock.File().Write(append(eventBytes, '\n'))
	return err
}

// JoinSession adds a participant to a session
// Returns the new event number (1-indexed for display)
func JoinSession(sessionID, name string) (int, error) {
	// Validate reserved name
	if IsReservedName(name) {
		return 0, &errors.ReservedNameError{Name: name}
	}

	path, err := storage.SessionEventsPath(sessionID)
	if err != nil {
		return 0, err
	}

	// Check session exists
	exists, err := storage.SessionExists(sessionID)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, &errors.SessionNotFoundError{SessionID: sessionID}
	}

	lock, err := AcquireLock(path)
	if err != nil {
		return 0, err
	}
	defer lock.Release()

	// Read current state
	lock.File().Seek(0, io.SeekStart)
	session, err := readSessionFromReader(sessionID, lock.File())
	if err != nil {
		return 0, err
	}

	// Check for duplicate name
	if session.IsActiveParticipant(name) {
		return 0, &errors.NameTakenError{Name: name}
	}

	// Append joined event
	event := NewJoinedEvent(name)
	eventBytes, err := MarshalEvent(event)
	if err != nil {
		return 0, err
	}

	lock.File().Seek(0, io.SeekEnd)
	_, err = lock.File().Write(append(eventBytes, '\n'))
	if err != nil {
		return 0, err
	}

	// Return 1-indexed event number
	return session.EventCount() + 1, nil
}

// LeaveSession removes a participant from a session
func LeaveSession(sessionID, name string) error {
	path, err := storage.SessionEventsPath(sessionID)
	if err != nil {
		return err
	}

	// Check session exists
	exists, err := storage.SessionExists(sessionID)
	if err != nil {
		return err
	}
	if !exists {
		return &errors.SessionNotFoundError{SessionID: sessionID}
	}

	lock, err := AcquireLock(path)
	if err != nil {
		return err
	}
	defer lock.Release()

	// Read current state
	lock.File().Seek(0, io.SeekStart)
	session, err := readSessionFromReader(sessionID, lock.File())
	if err != nil {
		return err
	}

	// Check participant is active
	if !session.IsActiveParticipant(name) {
		return &errors.ParticipantNotInSessionError{Name: name, SessionID: sessionID}
	}

	// Append left event
	event := NewLeftEvent(name)
	eventBytes, err := MarshalEvent(event)
	if err != nil {
		return err
	}

	lock.File().Seek(0, io.SeekEnd)
	_, err = lock.File().Write(append(eventBytes, '\n'))
	return err
}

// PostMessage posts a message to a session with optimistic locking
// Returns the new event number (1-indexed for display)
func PostMessage(sessionID, participant, content, next string, afterEventNum int) (int, error) {
	path, err := storage.SessionEventsPath(sessionID)
	if err != nil {
		return 0, err
	}

	// Check session exists
	exists, err := storage.SessionExists(sessionID)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, &errors.SessionNotFoundError{SessionID: sessionID}
	}

	lock, err := AcquireLock(path)
	if err != nil {
		return 0, err
	}
	defer lock.Release()

	// Read current state
	lock.File().Seek(0, io.SeekStart)
	session, err := readSessionFromReader(sessionID, lock.File())
	if err != nil {
		return 0, err
	}

	// Optimistic lock check
	if session.EventCount() != afterEventNum {
		return 0, &errors.StaleStateError{
			ExpectedEventNum: afterEventNum,
			ActualEventNum:   session.EventCount(),
			SessionID:        sessionID,
		}
	}

	// Check participant is active (Moderator is always allowed to post)
	if participant != "Moderator" && !session.IsActiveParticipant(participant) {
		return 0, &errors.NotAParticipantError{Name: participant, SessionID: sessionID}
	}

	// Determine next speaker if not provided
	if next == "" {
		// Fallback chain: previous speaker (if still active) -> random active -> Moderator
		prevSpeaker := session.PreviousSpeaker(participant)
		if prevSpeaker != "" && session.IsActiveParticipant(prevSpeaker) {
			next = prevSpeaker
		}
		if next == "" {
			next = session.RandomActiveParticipant(participant)
		}
		if next == "" {
			next = "Moderator"
		}
	}

	// Validate next is an active participant or "Moderator"
	if next != "Moderator" && !session.IsActiveParticipant(next) {
		return 0, &errors.InvalidNextParticipantError{Name: next}
	}

	// Append message event
	event := NewMessageEvent(participant, content, next)
	eventBytes, err := MarshalEvent(event)
	if err != nil {
		return 0, err
	}

	lock.File().Seek(0, io.SeekEnd)
	_, err = lock.File().Write(append(eventBytes, '\n'))
	if err != nil {
		return 0, err
	}

	// Return 1-indexed event number
	return session.EventCount() + 1, nil
}
