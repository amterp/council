package cli

import (
	"fmt"
	"os"

	"github.com/amterp/council/internal/session"
	"github.com/amterp/ra"
)

var (
	leaveCmd       *ra.Cmd
	leaveSessionID *string
	leaveName      *string
)

func setupLeaveCmd() *ra.Cmd {
	leaveCmd = ra.NewCmd("leave")
	leaveCmd.SetDescription("Leave a session")

	leaveSessionID, _ = ra.NewString("session-id").
		SetUsage("Session ID to leave").
		Register(leaveCmd)

	leaveName, _ = ra.NewString("participant").
		SetShort("p").
		SetFlagOnly(true).
		SetOptional(true).
		SetUsage("Participant name (will prompt if not provided)").
		Register(leaveCmd)

	return leaveCmd
}

func handleLeave() {
	name := *leaveName
	if name == "" {
		name = promptForName("Enter your participant name: ")
	}

	err := session.LeaveSession(*leaveSessionID, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
