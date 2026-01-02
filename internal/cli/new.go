package cli

import (
	"fmt"
	"os"

	"github.com/amterp/council/internal/session"
	"github.com/amterp/ra"
	petname "github.com/dustinkirkland/golang-petname"
)

var newCmd *ra.Cmd

func setupNewCmd() *ra.Cmd {
	newCmd = ra.NewCmd("new")
	newCmd.SetDescription("Create a new collaboration session")
	return newCmd
}

func handleNew() {
	// Generate session ID (3 words, hyphen-separated)
	sessionID := petname.Generate(3, "-")

	// Create session file with session_created event
	err := session.CreateSession(sessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Output session ID
	fmt.Println(sessionID)
}
