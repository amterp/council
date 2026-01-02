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

// SessionPath returns the path to a specific session file
func SessionPath(sessionID string) (string, error) {
	sessionsDir, err := SessionsPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(sessionsDir, sessionID+".jsonl"), nil
}

// EnsureSessionsDir creates the sessions directory if it doesn't exist
func EnsureSessionsDir() error {
	path, err := SessionsPath()
	if err != nil {
		return err
	}
	return os.MkdirAll(path, 0755)
}

// SessionExists checks if a session file exists
func SessionExists(sessionID string) (bool, error) {
	path, err := SessionPath(sessionID)
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
