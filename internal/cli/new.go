package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/amterp/council/internal/session"
	"github.com/amterp/ra"
	petname "github.com/dustinkirkland/golang-petname"
)

var (
	newCmd  *ra.Cmd
	newCopy *bool
)

func setupNewCmd() *ra.Cmd {
	newCmd = ra.NewCmd("new")
	newCmd.SetDescription("Create a new collaboration session")

	newCopy, _ = ra.NewBool("copy").
		SetShort("c").
		SetFlagOnly(true).
		SetOptional(true).
		SetUsage("Copy session ID to clipboard").
		Register(newCmd)

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

	if *newCopy {
		if err := copyToClipboard(sessionID); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to copy to clipboard: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "Copied to clipboard.")
	}
}

func copyToClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		cmd = exec.Command("xclip", "-selection", "clipboard")
	case "windows":
		cmd = exec.Command("clip.exe")
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}
