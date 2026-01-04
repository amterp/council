package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSessionsPath(t *testing.T) {
	path, err := SessionsPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should end with .council/sessions
	if !strings.HasSuffix(path, filepath.Join(CouncilDir, SessionsDir)) {
		t.Errorf("path should end with %s, got %q", filepath.Join(CouncilDir, SessionsDir), path)
	}

	// Should be an absolute path
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got %q", path)
	}
}

func TestSessionDirPath(t *testing.T) {
	path, err := SessionDirPath("my-session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should end with the session ID
	if !strings.HasSuffix(path, "my-session") {
		t.Errorf("path should end with session ID, got %q", path)
	}

	// Should contain .council/sessions
	if !strings.Contains(path, filepath.Join(CouncilDir, SessionsDir)) {
		t.Errorf("path should contain .council/sessions, got %q", path)
	}
}

func TestSessionEventsPath(t *testing.T) {
	path, err := SessionEventsPath("test-session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should end with events.jsonl
	if !strings.HasSuffix(path, "events.jsonl") {
		t.Errorf("path should end with events.jsonl, got %q", path)
	}

	// Should contain the session ID
	if !strings.Contains(path, "test-session") {
		t.Errorf("path should contain session ID, got %q", path)
	}
}

func TestPathConsistency(t *testing.T) {
	sessionID := "consistent-test"

	sessionsPath, _ := SessionsPath()
	dirPath, _ := SessionDirPath(sessionID)
	eventsPath, _ := SessionEventsPath(sessionID)

	// SessionDirPath should be under SessionsPath
	if !strings.HasPrefix(dirPath, sessionsPath) {
		t.Errorf("session dir %q should be under sessions path %q", dirPath, sessionsPath)
	}

	// SessionEventsPath should be under SessionDirPath
	if !strings.HasPrefix(eventsPath, dirPath) {
		t.Errorf("events path %q should be under session dir %q", eventsPath, dirPath)
	}

	// Verify the full structure
	expectedEventsPath := filepath.Join(sessionsPath, sessionID, "events.jsonl")
	if eventsPath != expectedEventsPath {
		t.Errorf("expected %q, got %q", expectedEventsPath, eventsPath)
	}
}

func TestEnsureSessionDir(t *testing.T) {
	// Use test name for unique session ID (safe for parallel runs)
	sessionID := "test-ensure-dir-" + t.Name()

	// Clean up after test
	t.Cleanup(func() {
		dirPath, _ := SessionDirPath(sessionID)
		os.RemoveAll(dirPath)
	})

	// Ensure the directory
	err := EnsureSessionDir(sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it exists
	dirPath, _ := SessionDirPath(sessionID)
	info, err := os.Stat(dirPath)
	if err != nil {
		t.Fatalf("directory should exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("should be a directory")
	}

	// Calling again should not error (idempotent)
	err = EnsureSessionDir(sessionID)
	if err != nil {
		t.Fatalf("second call should not error: %v", err)
	}
}

func TestEnsureSessionsDir(t *testing.T) {
	err := EnsureSessionsDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it exists
	path, _ := SessionsPath()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("directory should exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("should be a directory")
	}
}

func TestSessionExists(t *testing.T) {
	// Non-existent session (use test name for uniqueness)
	exists, err := SessionExists("nonexistent-session-" + t.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("session should not exist")
	}

	// Create a session
	sessionID := "test-exists-" + t.Name()
	t.Cleanup(func() {
		dirPath, _ := SessionDirPath(sessionID)
		os.RemoveAll(dirPath)
	})

	err = EnsureSessionDir(sessionID)
	if err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	eventsPath, _ := SessionEventsPath(sessionID)
	err = os.WriteFile(eventsPath, []byte(`{"type":"session_created"}`), 0644)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Now it should exist
	exists, err = SessionExists(sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("session should exist after creating file")
	}
}

func TestConstants(t *testing.T) {
	if CouncilDir != ".council" {
		t.Errorf("CouncilDir should be '.council', got %q", CouncilDir)
	}
	if SessionsDir != "sessions" {
		t.Errorf("SessionsDir should be 'sessions', got %q", SessionsDir)
	}
}
