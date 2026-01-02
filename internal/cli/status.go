package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/amterp/council/internal/session"
	"github.com/amterp/ra"
)

var (
	statusCmd         *ra.Cmd
	statusSessionID   *string
	statusAfter       *int
	statusAwait       *bool
	statusParticipant *string
	statusTimeout     *int
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

	statusAwait, _ = ra.NewBool("await").
		SetFlagOnly(true).
		SetOptional(true).
		SetUsage("Block until new events and it's your turn (requires --participant)").
		Register(statusCmd)

	statusParticipant, _ = ra.NewString("participant").
		SetShort("p").
		SetFlagOnly(true).
		SetOptional(true).
		SetUsage("Your participant name (required with --await)").
		Register(statusCmd)

	statusTimeout, _ = ra.NewInt("timeout").
		SetFlagOnly(true).
		SetOptional(true).
		SetUsage("Timeout in seconds for --await (default: 300)").
		Register(statusCmd)

	return statusCmd
}

func handleStatus() {
	afterN := 0
	if statusCmd.Configured("after") {
		afterN = *statusAfter
	}

	// Check if await mode
	awaitMode := statusAwait != nil && *statusAwait
	if awaitMode {
		if statusParticipant == nil || *statusParticipant == "" {
			fmt.Fprintf(os.Stderr, "Error: --await requires --participant\n")
			os.Exit(1)
		}
		handleAwait(*statusSessionID, *statusParticipant, afterN)
		return
	}

	// Normal status mode
	sess, err := session.LoadSession(*statusSessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	output := session.FormatStatus(sess, afterN)
	fmt.Print(output)
}

func handleAwait(sessionID, participant string, afterN int) {
	timeout := 300 // default 5 minutes
	if statusTimeout != nil && *statusTimeout > 0 {
		timeout = *statusTimeout
	}

	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	pollInterval := 2 * time.Second

	currentAfter := afterN

	for {
		if time.Now().After(deadline) {
			fmt.Fprintf(os.Stderr, "Error: Timeout waiting for turn after %d seconds\n", timeout)
			os.Exit(1)
		}

		sess, err := session.LoadSession(sessionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Check if there are new events
		if sess.EventCount() > currentAfter {
			// Check if it's our turn
			nextSpeaker := sess.LatestMessageNext()
			if nextSpeaker == participant {
				// It's our turn! Show the new events
				output := session.FormatStatus(sess, afterN)
				fmt.Print(output)
				return
			}
			// Not our turn, update currentAfter and keep waiting
			currentAfter = sess.EventCount()
		}

		time.Sleep(pollInterval)
	}
}
