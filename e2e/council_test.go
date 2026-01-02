package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// councilBinary returns the path to the council binary.
// Tests assume the binary has been built before running.
func councilBinary() string {
	// Look for binary in project root (relative to this test file)
	// Tests run from e2e/ directory, so go up one level
	return "../council"
}

// runCouncil executes council with the given args and optional stdin
func runCouncil(t *testing.T, stdin string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(councilBinary(), args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	// CombinedOutput captures both stdout and stderr
	output, err := cmd.CombinedOutput()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	// For simplicity, we return combined output as both stdout and stderr
	// The tests check for content presence which works with combined output
	return string(output), string(output), exitCode
}

// createSession creates a new session and returns the session ID
func createSession(t *testing.T) string {
	t.Helper()
	stdout, stderr, exitCode := runCouncil(t, "", "new")
	if exitCode != 0 {
		t.Fatalf("council new failed: %s", stderr)
	}
	return strings.TrimSpace(stdout)
}

// joinSession joins a session with the given name
func joinSession(t *testing.T, sessionID, name string) {
	t.Helper()
	_, stderr, exitCode := runCouncil(t, "", "join", sessionID, "--participant", name)
	if exitCode != 0 {
		t.Fatalf("council join failed: %s", stderr)
	}
}

func TestNewSession(t *testing.T) {
	sessionID := createSession(t)

	// Verify session ID format (3 words separated by hyphens)
	parts := strings.Split(sessionID, "-")
	if len(parts) != 3 {
		t.Errorf("expected 3-word session ID, got %d parts: %s", len(parts), sessionID)
	}

	// Verify session file was created in the new directory structure
	home, _ := os.UserHomeDir()
	sessionPath := filepath.Join(home, ".council", "sessions", sessionID, "events.jsonl")
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		t.Errorf("session file not created at %s", sessionPath)
	}
}

func TestJoinSession(t *testing.T) {
	sessionID := createSession(t)

	// Join the session and verify output
	stdout, stderr, exitCode := runCouncil(t, "", "join", sessionID, "--participant", "Engineer")
	if exitCode != 0 {
		t.Fatalf("council join failed: %s", stderr)
	}

	// Verify join output shows event number
	if !strings.Contains(stdout, "Joined session as event #2") {
		t.Errorf("expected 'Joined session as event #2', got: %s", stdout)
	}
	if !strings.Contains(stdout, "Use --after 2 for your first post") {
		t.Errorf("expected '--after 2' guidance, got: %s", stdout)
	}

	// Verify participant appears in status
	stdout, _, exitCode = runCouncil(t, "", "status", sessionID)
	if exitCode != 0 {
		t.Fatalf("council status failed")
	}

	if !strings.Contains(stdout, "Participants: Engineer") {
		t.Errorf("Engineer not found in participants, got: %s", stdout)
	}

	if !strings.Contains(stdout, "Engineer Joined") {
		t.Errorf("Engineer Joined event not found, got: %s", stdout)
	}
}

func TestDuplicateNameRejection(t *testing.T) {
	sessionID := createSession(t)
	joinSession(t, sessionID, "Engineer")

	// Try to join with same name
	_, stderr, exitCode := runCouncil(t, "", "join", sessionID, "--participant", "Engineer")
	if exitCode == 0 {
		t.Error("expected error for duplicate name, but command succeeded")
	}

	if !strings.Contains(stderr, "already exists") {
		t.Errorf("expected 'already exists' error, got: %s", stderr)
	}
}

func TestReservedNameRejection(t *testing.T) {
	sessionID := createSession(t)

	// Try to join as Moderator
	_, stderr, exitCode := runCouncil(t, "", "join", sessionID, "--participant", "Moderator")
	if exitCode == 0 {
		t.Error("expected error for reserved name, but command succeeded")
	}

	if !strings.Contains(stderr, "reserved name") {
		t.Errorf("expected 'reserved name' error, got: %s", stderr)
	}
}

func TestPostMessage(t *testing.T) {
	sessionID := createSession(t)
	joinSession(t, sessionID, "Engineer")

	// Post a message (after event #2 = session_created + joined)
	stdout, stderr, exitCode := runCouncil(t, "Hello world!", "post", sessionID, "--participant", "Engineer", "--after", "2")
	if exitCode != 0 {
		t.Fatalf("council post failed: %s", stderr)
	}

	// Verify post output shows event number
	if !strings.Contains(stdout, "Posted as event #3") {
		t.Errorf("expected 'Posted as event #3', got: %s", stdout)
	}

	// Verify message appears in status
	stdout, _, _ = runCouncil(t, "", "status", sessionID)
	if !strings.Contains(stdout, "Hello world!") {
		t.Errorf("message not found in status, got: %s", stdout)
	}
	if !strings.Contains(stdout, "--- #3 | Engineer ---") {
		t.Errorf("message header not found, got: %s", stdout)
	}
	// Verify end marker includes author (Next defaults to Moderator when no other participants)
	if !strings.Contains(stdout, "--- End #3 | Engineer | Next: Moderator ---") {
		t.Errorf("message end marker not found with correct format, got: %s", stdout)
	}
}

func TestOptimisticLocking(t *testing.T) {
	sessionID := createSession(t)
	joinSession(t, sessionID, "Engineer")

	// Post first message (after event #2)
	_, stderr, exitCode := runCouncil(t, "First message", "post", sessionID, "--participant", "Engineer", "--after", "2")
	if exitCode != 0 {
		t.Fatalf("first post failed: %s", stderr)
	}

	// Try to post with stale --after value (should fail)
	_, stderr, exitCode = runCouncil(t, "Stale message", "post", sessionID, "--participant", "Engineer", "--after", "2")
	if exitCode == 0 {
		t.Error("expected error for stale post, but command succeeded")
	}

	if !strings.Contains(stderr, "New activity since event #2") {
		t.Errorf("expected stale state error, got: %s", stderr)
	}
}

func TestNotAParticipant(t *testing.T) {
	sessionID := createSession(t)

	// Try to post without joining
	_, stderr, exitCode := runCouncil(t, "Unauthorized message", "post", sessionID, "--participant", "Outsider", "--after", "1")
	if exitCode == 0 {
		t.Error("expected error for non-participant, but command succeeded")
	}

	if !strings.Contains(stderr, "must join the session") {
		t.Errorf("expected 'must join' error, got: %s", stderr)
	}
}

func TestLeaveSession(t *testing.T) {
	sessionID := createSession(t)
	joinSession(t, sessionID, "Engineer")

	// Leave the session
	_, stderr, exitCode := runCouncil(t, "", "leave", sessionID, "--participant", "Engineer")
	if exitCode != 0 {
		t.Fatalf("council leave failed: %s", stderr)
	}

	// Verify participant is gone from active list
	stdout, _, _ := runCouncil(t, "", "status", sessionID)
	if !strings.Contains(stdout, "Participants: (none)") {
		t.Errorf("expected no participants after leave, got: %s", stdout)
	}

	// Verify left event appears
	if !strings.Contains(stdout, "Engineer Left") {
		t.Errorf("left event not found, got: %s", stdout)
	}
}

func TestSessionNotFound(t *testing.T) {
	_, stderr, exitCode := runCouncil(t, "", "status", "nonexistent-session-xyz")
	if exitCode == 0 {
		t.Error("expected error for nonexistent session, but command succeeded")
	}

	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got: %s", stderr)
	}
}

func TestStatusAfterFilter(t *testing.T) {
	sessionID := createSession(t)
	joinSession(t, sessionID, "Engineer")

	// Post a message
	runCouncil(t, "Hello!", "post", sessionID, "--participant", "Engineer", "--after", "2")

	// Get status with --after 2 (should skip joined event)
	stdout, _, _ := runCouncil(t, "", "status", sessionID, "--after", "2")

	// Should NOT contain the joined event
	if strings.Contains(stdout, "Engineer Joined") {
		t.Errorf("--after filter should have excluded joined event, got: %s", stdout)
	}

	// Should contain the message
	if !strings.Contains(stdout, "Hello!") {
		t.Errorf("message should still appear, got: %s", stdout)
	}
}

func TestMultipleParticipants(t *testing.T) {
	sessionID := createSession(t)
	joinSession(t, sessionID, "Engineer")
	joinSession(t, sessionID, "Designer")

	// Check both appear in participants
	stdout, _, _ := runCouncil(t, "", "status", sessionID)

	if !strings.Contains(stdout, "Designer") || !strings.Contains(stdout, "Engineer") {
		t.Errorf("expected both participants, got: %s", stdout)
	}
}

func TestRejoinAfterLeave(t *testing.T) {
	sessionID := createSession(t)
	joinSession(t, sessionID, "Engineer")

	// Leave
	runCouncil(t, "", "leave", sessionID, "--participant", "Engineer")

	// Should be able to rejoin
	_, stderr, exitCode := runCouncil(t, "", "join", sessionID, "--participant", "Engineer")
	if exitCode != 0 {
		t.Fatalf("rejoin after leave should succeed: %s", stderr)
	}

	// Verify back in participants
	stdout, _, _ := runCouncil(t, "", "status", sessionID)
	if !strings.Contains(stdout, "Participants: Engineer") {
		t.Errorf("Engineer should be back in participants after rejoin, got: %s", stdout)
	}
}

func TestCannotPostAfterLeave(t *testing.T) {
	sessionID := createSession(t)
	joinSession(t, sessionID, "Engineer")
	runCouncil(t, "", "leave", sessionID, "--participant", "Engineer")

	// Try to post after leaving
	_, stderr, exitCode := runCouncil(t, "Ghost message", "post", sessionID, "--participant", "Engineer", "--after", "3")
	if exitCode == 0 {
		t.Error("should not be able to post after leaving")
	}

	if !strings.Contains(stderr, "must join") {
		t.Errorf("expected 'must join' error, got: %s", stderr)
	}
}

func TestNextFlag(t *testing.T) {
	sessionID := createSession(t)
	joinSession(t, sessionID, "Engineer")
	joinSession(t, sessionID, "Designer")

	// Post with explicit --next
	_, stderr, exitCode := runCouncil(t, "Hello Designer!", "post", sessionID, "--participant", "Engineer", "--after", "3", "--next", "Designer")
	if exitCode != 0 {
		t.Fatalf("council post with --next failed: %s", stderr)
	}

	// Verify Next appears in status output
	stdout, _, _ := runCouncil(t, "", "status", sessionID)
	if !strings.Contains(stdout, "Next: Designer") {
		t.Errorf("expected 'Next: Designer' in output, got: %s", stdout)
	}
}

func TestNextFlagDefaulting(t *testing.T) {
	sessionID := createSession(t)
	joinSession(t, sessionID, "Engineer")
	joinSession(t, sessionID, "Designer")

	// First post from Engineer (no previous speaker, should pick random other = Designer)
	runCouncil(t, "First message", "post", sessionID, "--participant", "Engineer", "--after", "3")

	// Second post from Designer (previous speaker = Engineer)
	runCouncil(t, "Second message", "post", sessionID, "--participant", "Designer", "--after", "4")

	// Verify Next defaults to previous speaker (Engineer)
	stdout, _, _ := runCouncil(t, "", "status", sessionID)
	if !strings.Contains(stdout, "--- End #5 | Designer | Next: Engineer ---") {
		t.Errorf("expected Next to default to Engineer, got: %s", stdout)
	}
}

func TestNextFlagInvalidParticipant(t *testing.T) {
	sessionID := createSession(t)
	joinSession(t, sessionID, "Engineer")

	// Try to post with --next pointing to non-existent participant
	_, stderr, exitCode := runCouncil(t, "Hello!", "post", sessionID, "--participant", "Engineer", "--after", "2", "--next", "Ghost")
	if exitCode == 0 {
		t.Error("expected error for invalid --next, but command succeeded")
	}

	if !strings.Contains(stderr, "not an active participant") {
		t.Errorf("expected 'not an active participant' error, got: %s", stderr)
	}
}
