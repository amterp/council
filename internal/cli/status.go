package cli

import (
	"fmt"
	"os"

	"github.com/amterp/council/internal/session"
	"github.com/amterp/ra"
)

var (
	statusCmd       *ra.Cmd
	statusSessionID *string
	statusAfter     *int
)

func setupStatusCmd() *ra.Cmd {
	statusCmd = ra.NewCmd("status")
	statusCmd.SetDescription("Display session state")

	statusSessionID, _ = ra.NewString("session-id").
		SetUsage("Session ID to check").
		Register(statusCmd)

	statusAfter, _ = ra.NewInt("after").
		SetFlagOnly(true).
		SetOptional(true).
		SetUsage("Only show events after event number N").
		Register(statusCmd)

	return statusCmd
}

func handleStatus() {
	sess, err := session.LoadSession(*statusSessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	afterN := 0
	if statusCmd.Configured("after") {
		afterN = *statusAfter
	}

	output := session.FormatStatus(sess, afterN)
	fmt.Print(output)
}
