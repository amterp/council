package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/amterp/ra"
)

var (
	installCmd     *ra.Cmd
	installTargets *[]string
)

type installTarget struct {
	install func() (string, error) // returns destination path
}

var targetRegistry = map[string]installTarget{
	"claude": {install: installClaudeSkill},
}

func setupInstallCmd() *ra.Cmd {
	installCmd = ra.NewCmd("install")
	installCmd.SetDescription("Install council integrations")
	installTargets, _ = ra.NewStringSlice("targets").
		SetUsage("Targets to install (e.g., claude)").
		SetVariadic(true).
		Register(installCmd)
	return installCmd
}

func handleInstall() {
	if len(*installTargets) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no targets specified\n")
		fmt.Fprintf(os.Stderr, "Available targets: %s\n", availableTargets())
		os.Exit(1)
	}

	hasErrors := false
	for _, target := range *installTargets {
		t, exists := targetRegistry[target]
		if !exists {
			fmt.Fprintf(os.Stderr, "Error: unknown target '%s'\n", target)
			hasErrors = true
			continue
		}

		destPath, err := t.install()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error installing %s: %v\n", target, err)
			hasErrors = true
			continue
		}

		fmt.Printf("Installed %s skill to %s\n", target, destPath)
	}

	if hasErrors {
		os.Exit(1)
	}
}

func installClaudeSkill() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	destDir := filepath.Join(home, ".claude", "skills", "council-participant")
	destPath := filepath.Join(destDir, "SKILL.md")

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("cannot create directory %s: %w", destDir, err)
	}

	if err := os.WriteFile(destPath, []byte(skillMd), 0644); err != nil {
		return "", fmt.Errorf("cannot write file %s: %w", destPath, err)
	}

	return destPath, nil
}

func availableTargets() string {
	var targets []string
	for name := range targetRegistry {
		targets = append(targets, name)
	}
	return strings.Join(targets, ", ")
}
