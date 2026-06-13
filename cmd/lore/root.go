package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/pierreWagou/lore/internal/installer"
	"github.com/pierreWagou/lore/internal/manifest"
)

var rootCmd = &cobra.Command{
	Use:   "lore",
	Short: "Agent skills package manager",
	Long: `lore manages AI agent skills across harnesses (opencode, claude, cursor, codex).

Skills are fetched from git repositories and installed in each harness's
native format. A lore.toml manifest tracks dependencies; lore.lock pins
exact commit SHAs for reproducibility.`,
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(targetsCmd)
	rootCmd.AddCommand(authCmd)
}

// manifestPath returns the path to lore.toml for the current scope.
func manifestPath(global bool) string {
	if global {
		return filepath.Join(installer.DefaultConfigDir(), manifest.FileName)
	}
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, manifest.FileName)
}

// lockfilePath returns the path to lore.lock for the current scope.
func lockfilePath(global bool) string {
	if global {
		return filepath.Join(installer.DefaultConfigDir(), "lore.lock")
	}
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, "lore.lock")
}

// projectRoot returns the project root for the current scope.
func projectRoot() string {
	cwd, _ := os.Getwd()
	return cwd
}
