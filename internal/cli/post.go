package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/amterp/council/internal/session"
	"github.com/amterp/ra"
)

var (
	postCmd         *ra.Cmd
	postSessionID   *string
	postParticipant *string
	postAfter       *int
	postFile        *string
	postNext        *string
)

func setupPostCmd() *ra.Cmd {
	postCmd = ra.NewCmd("post")
	postCmd.SetDescription("Post a message to the session")

	postSessionID, _ = ra.NewString("session-id").
		SetUsage("Session ID to post to").
		Register(postCmd)

	postParticipant, _ = ra.NewString("participant").
		SetShort("p").
		SetFlagOnly(true).
		SetUsage("Participant name posting the message").
		Register(postCmd)

	postAfter, _ = ra.NewInt("after").
		SetFlagOnly(true).
		SetUsage("Only post if latest event is exactly N").
		Register(postCmd)

	postFile, _ = ra.NewString("file").
		SetShort("f").
		SetFlagOnly(true).
		SetOptional(true).
		SetUsage("Read content from file instead of stdin").
		Register(postCmd)

	postNext, _ = ra.NewString("next").
		SetShort("n").
		SetFlagOnly(true).
		SetOptional(true).
		SetUsage("Designate the next speaker (defaults to previous speaker)").
		Register(postCmd)

	return postCmd
}

func handlePost() {
	content, err := readContent(*postFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// postNext may be nil if optional and not provided
	next := ""
	if postNext != nil {
		next = *postNext
	}

	eventNum, err := session.PostMessage(*postSessionID, *postParticipant, content, next, *postAfter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Posted as event #%d.\n", eventNum)
}

// readContent reads message content from file or stdin
func readContent(filePath string) (string, error) {
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	// Read from stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
