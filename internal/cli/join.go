package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/amterp/council/internal/session"
	"github.com/amterp/ra"
)

var (
	joinCmd       *ra.Cmd
	joinSessionID *string
	joinName      *string
)

func setupJoinCmd() *ra.Cmd {
	joinCmd = ra.NewCmd("join")
	joinCmd.SetDescription("Join a session as a participant")

	joinSessionID, _ = ra.NewString("session-id").
		SetUsage("Session ID to join").
		Register(joinCmd)

	joinName, _ = ra.NewString("participant").
		SetShort("p").
		SetFlagOnly(true).
		SetOptional(true).
		SetUsage("Participant name (will prompt if not provided)").
		Register(joinCmd)

	return joinCmd
}

func handleJoin() {
	name := *joinName
	if name == "" {
		name = promptForName("Enter your participant name: ")
	}

	eventNum, err := session.JoinSession(*joinSessionID, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Joined session as event #%d. Use --after %d for your first post.\n", eventNum, eventNum)
}

// promptForName prompts the user for a name via stdin
func promptForName(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	name, _ := reader.ReadString('\n')
	return strings.TrimSpace(name)
}
