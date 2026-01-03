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

	// Determine port
	port := 3000
	if watchPort != nil && *watchPort > 0 {
		port = *watchPort
	} else {
		port = web.FindAvailablePort(3000)
	}

	// Start server in goroutine
	server := web.NewServer(*watchSessionID, port)
	serverErr := make(chan error, 1)

	go func() {
		if err := server.Start(); err != nil {
			serverErr <- err
		}
	}()

	url := fmt.Sprintf("http://localhost:%d?session=%s", port, *watchSessionID)
	fmt.Printf("Watching session at %s\n", url)

	// Open browser (unless --no-open)
	if watchNoOpen == nil || !*watchNoOpen {
		if err := web.OpenBrowser(url); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not open browser: %v\n", err)
		}
	}

	// Block until interrupt or server error
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
