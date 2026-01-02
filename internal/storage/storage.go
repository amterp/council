package storage

import (
	"os"
	"path/filepath"
)

const (
	CouncilDir  = ".council"
	SessionsDir = "sessions"
)

// SessionsPath returns the path to the sessions directory (~/.council/sessions)
func SessionsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, CouncilDir, SessionsDir), nil
}

// SessionDirPath returns the path to a specific session directory
func SessionDirPath(sessionID string) (string, error) {
	sessionsDir, err := SessionsPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(sessionsDir, sessionID), nil
}

// SessionEventsPath returns the path to a session's events.jsonl file
func SessionEventsPath(sessionID string) (string, error) {
	sessionDir, err := SessionDirPath(sessionID)
	if err != nil {
		return "", err
	}
	return filepath.Join(sessionDir, "events.jsonl"), nil
}

// EnsureSessionDir creates a session's directory if it doesn't exist
func EnsureSessionDir(sessionID string) error {
	path, err := SessionDirPath(sessionID)
	if err != nil {
		return err
	}
	return os.MkdirAll(path, 0755)
}

// EnsureSessionsDir creates the sessions directory if it doesn't exist
func EnsureSessionsDir() error {
	path, err := SessionsPath()
	if err != nil {
		return err
	}
	return os.MkdirAll(path, 0755)
}

// SessionExists checks if a session's events file exists
func SessionExists(sessionID string) (bool, error) {
	path, err := SessionEventsPath(sessionID)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
