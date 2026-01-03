package cli

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/amterp/ra"
)

//go:embed skill.md
var skillMd string

var (
	rootCmd *ra.Cmd

	// Subcommand used flags
	newUsed     *bool
	joinUsed    *bool
	leaveUsed   *bool
	statusUsed  *bool
	postUsed    *bool
	installUsed *bool
	watchUsed   *bool
)

// Run is the main entry point for the CLI
func Run() {
	rootCmd = ra.NewCmd("council")
	rootCmd.SetDescription("Multi-agent collaboration CLI tool")
	rootCmd.SetCustomUsage(printUsage)

	// Register subcommands
	newUsed, _ = rootCmd.RegisterCmd(setupNewCmd())
	joinUsed, _ = rootCmd.RegisterCmd(setupJoinCmd())
	leaveUsed, _ = rootCmd.RegisterCmd(setupLeaveCmd())
	statusUsed, _ = rootCmd.RegisterCmd(setupStatusCmd())
	postUsed, _ = rootCmd.RegisterCmd(setupPostCmd())
	installUsed, _ = rootCmd.RegisterCmd(setupInstallCmd())
	watchUsed, _ = rootCmd.RegisterCmd(setupWatchCmd())

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
	case *installUsed:
		handleInstall()
	case *watchUsed:
		handleWatch()
	}
}

func printUsage(isLongHelp bool) {
	fmt.Print(rootCmd.GenerateShortUsage())

	if isLongHelp {
		fmt.Println()
		fmt.Print(stripFrontmatter(skillMd))
	} else {
		fmt.Println("\nRun 'council --help' for full agent participation instructions.")
	}
}

// stripFrontmatter removes YAML frontmatter (--- delimited) from markdown content
func stripFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---") {
		return content
	}

	// Find the closing ---
	rest := content[3:] // skip opening ---
	idx := strings.Index(rest, "---")
	if idx == -1 {
		return content
	}

	// Return everything after the closing --- (skip the --- and any immediate newline)
	result := rest[idx+3:]
	result = strings.TrimPrefix(result, "\n")
	return result
}
