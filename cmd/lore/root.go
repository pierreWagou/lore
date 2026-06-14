package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/pierreWagou/lore/internal/installer"
	"github.com/pierreWagou/lore/internal/manifest"
)

var rootCmd = &cobra.Command{
	Use:     "lore",
	Version: Version,
	Short:   "Agent skills package manager",
	Long: `lore manages AI agent skills across harnesses (opencode, claude, cursor, codex).

Skills are fetched from git repositories and installed in each harness's
native format. A lore.toml manifest tracks dependencies; lore.lock pins
exact commit SHAs for reproducibility.`,
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(harnessesCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)
}

// mustGetwd returns the current working directory or exits with an error message.
func mustGetwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot determine working directory: %v\n", err)
		os.Exit(1)
	}
	return cwd
}

// manifestPath returns the path to lore.toml for the current scope.
func manifestPath(global bool) string {
	if global {
		return filepath.Join(installer.DefaultConfigDir(), manifest.FileName)
	}
	return filepath.Join(mustGetwd(), manifest.FileName)
}

// lockfilePath returns the path to lore.lock for the current scope.
func lockfilePath(global bool) string {
	if global {
		return filepath.Join(installer.DefaultConfigDir(), "lore.lock")
	}
	return filepath.Join(mustGetwd(), "lore.lock")
}

// projectRoot returns the project root directory.
func projectRoot() string {
	return mustGetwd()
}
