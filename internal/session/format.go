package session

import (
	"fmt"
	"sort"
	"strings"
)

// FormatStatus generates the human-readable status output
func FormatStatus(sess *Session, afterN int) string {
	var b strings.Builder

	// Header
	fmt.Fprintf(&b, "=== Session: %s ===\n", sess.ID)

	// Participants (excluding Moderator, sorted for consistency)
	participants := sess.ActiveParticipants()
	sort.Strings(participants)
	if len(participants) > 0 {
		fmt.Fprintf(&b, "Participants: %s\n", strings.Join(participants, ", "))
	} else {
		fmt.Fprintf(&b, "Participants: (none)\n")
	}
	b.WriteString("\n")

	// Events (starting from afterN, 1-indexed for display)
	for i, event := range sess.Events {
		eventNum := i + 1 // 1-indexed for display
		if eventNum <= afterN {
			continue
		}

		switch e := event.(type) {
		case *SessionCreatedEvent:
			// Don't show session_created in output
			continue
		case *JoinedEvent:
			// Don't show Moderator join events
			if e.Participant != "Moderator" {
				fmt.Fprintf(&b, "--- #%d | %s Joined ---\n\n", eventNum, e.Participant)
			}
		case *LeftEvent:
			fmt.Fprintf(&b, "--- #%d | %s Left ---\n\n", eventNum, e.Participant)
		case *MessageEvent:
			fmt.Fprintf(&b, "--- #%d | %s ---\n", eventNum, e.Participant)
			b.WriteString(e.Content)
			if !strings.HasSuffix(e.Content, "\n") {
				b.WriteString("\n")
			}
			if e.Next != "" {
				fmt.Fprintf(&b, "--- End #%d | %s | Next: %s ---\n\n", eventNum, e.Participant, e.Next)
			} else {
				fmt.Fprintf(&b, "--- End #%d | %s ---\n\n", eventNum, e.Participant)
			}
		}
	}

	return b.String()
}
