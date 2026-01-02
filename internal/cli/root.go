package cli

import (
	"os"

	"github.com/amterp/ra"
)

var (
	rootCmd *ra.Cmd

	// Subcommand used flags
	newUsed    *bool
	joinUsed   *bool
	leaveUsed  *bool
	statusUsed *bool
	postUsed   *bool
)

// Run is the main entry point for the CLI
func Run() {
	rootCmd = ra.NewCmd("council")
	rootCmd.SetDescription("Multi-agent collaboration CLI tool")

	// Register subcommands
	newUsed, _ = rootCmd.RegisterCmd(setupNewCmd())
	joinUsed, _ = rootCmd.RegisterCmd(setupJoinCmd())
	leaveUsed, _ = rootCmd.RegisterCmd(setupLeaveCmd())
	statusUsed, _ = rootCmd.RegisterCmd(setupStatusCmd())
	postUsed, _ = rootCmd.RegisterCmd(setupPostCmd())

	rootCmd.ParseOrExit(os.Args[1:])

	// Dispatch to the appropriate handler
	switch {
	case *newUsed:
		handleNew()
	case *joinUsed:
		handleJoin()
	case *leaveUsed:
		handleLeave()
	case *statusUsed:
		handleStatus()
	case *postUsed:
		handlePost()
	}
}
