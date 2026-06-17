package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/pierreWagou/lore/internal/config"
	"github.com/pierreWagou/lore/internal/installer"
	"github.com/pierreWagou/lore/internal/lockfile"
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

// projectRoot returns the current working directory, used as the project root.
func projectRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot determine working directory: %v\n", err)
		os.Exit(1)
	}
	return cwd
}

// manifestPath returns the path to lore.toml for the current scope.
// For global scope this is ~/.config/lore/lore.toml (the merged config+manifest).
// For project scope this is <cwd>/lore.toml.
func manifestPath(global bool) string {
	if global {
		return filepath.Join(installer.DefaultConfigDir(), manifest.FileName)
	}
	return filepath.Join(projectRoot(), manifest.FileName)
}

// lockfilePath returns the path to lore.lock for project-scoped installs.
// For global installs use globalLockfilePath instead.
func lockfilePath(global bool) string {
	if global {
		return filepath.Join(installer.DefaultConfigDir(), lockfile.FileName)
	}
	return filepath.Join(projectRoot(), lockfile.FileName)
}

// globalLockfilePath returns the per-profile lockfile path: ~/.config/lore/lore.<profile>.lock.
func globalLockfilePath(profileName string) string {
	return filepath.Join(installer.DefaultConfigDir(), lockfile.GlobalFileName(profileName))
}

// resolveActiveProfile returns the active profile name for a global command, given an
// explicit --profile flag value and an already-loaded Config.
// Priority: explicit flag > config.ActiveProfileNameFromConfig.
func resolveActiveProfile(flagValue string, cfg *config.Config) string {
	if flagValue != "" {
		return flagValue
	}
	return config.ActiveProfileNameFromConfig(cfg)
}
