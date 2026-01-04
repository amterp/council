package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amterp/council/internal/storage"
	"github.com/amterp/council/internal/web"
	"github.com/amterp/ra"
)

var (
	watchCmd       *ra.Cmd
	watchSessionID *string
	watchPort      *int
	watchNoOpen    *bool
)

func setupWatchCmd() *ra.Cmd {
	watchCmd = ra.NewCmd("watch")
	watchCmd.SetDescription("Watch a session via web interface")

	watchSessionID, _ = ra.NewString("session").
		SetShort("s").
		SetFlagOnly(true).
		SetUsage("Session ID to watch").
		Register(watchCmd)

	watchPort, _ = ra.NewInt("port").
		SetShort("p").
		SetFlagOnly(true).
		SetOptional(true).
		SetUsage("Port to serve on (auto-finds available port if not specified)").
		Register(watchCmd)

	watchNoOpen, _ = ra.NewBool("no-open").
		SetFlagOnly(true).
		SetOptional(true).
		SetUsage("Don't auto-open browser").
		Register(watchCmd)

	return watchCmd
}

// runWatchServer starts the watch web server for the given session.
// If port is 0, it finds an available port starting from 3000.
// If openBrowser is true, it auto-opens the URL in the default browser.
// Blocks until Ctrl+C or server error.
func runWatchServer(sessionID string, port int, openBrowser bool) {
	if port == 0 {
		port = web.FindAvailablePort(3000)
	}

	server := web.NewServer(sessionID, port)
	serverErr := make(chan error, 1)

	go func() {
		if err := server.Start(); err != nil {
			serverErr <- err
		}
	}()

	url := fmt.Sprintf("http://localhost:%d?session=%s", port, sessionID)
	fmt.Printf("Watching session at %s\n", url)

	if openBrowser {
		if err := web.OpenBrowser(url); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not open browser: %v\n", err)
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigChan:
		fmt.Println("\nShutting down...")
	case err := <-serverErr:
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func handleWatch() {
	// Validate session ID provided
	if watchSessionID == nil || *watchSessionID == "" {
		fmt.Fprintf(os.Stderr, "Error: --session is required\n")
		os.Exit(1)
	}

	// Validate session exists
	exists, err := storage.SessionExists(*watchSessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if !exists {
		fmt.Fprintf(os.Stderr, "Error: Session '%s' not found\n", *watchSessionID)
		os.Exit(1)
	}

	// Determine port (0 means auto-find)
	port := 0
	if watchPort != nil && *watchPort > 0 {
		port = *watchPort
	}

	// Determine if we should open browser
	openBrowser := watchNoOpen == nil || !*watchNoOpen

	runWatchServer(*watchSessionID, port, openBrowser)
}
