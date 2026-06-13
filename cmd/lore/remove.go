package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pierreWagou/lore/internal/installer"
	"github.com/pierreWagou/lore/internal/lockfile"
	"github.com/pierreWagou/lore/internal/manifest"
)

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a skill from the manifest and uninstall it",
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

var removeGlobal bool

func init() {
	removeCmd.Flags().BoolVarP(&removeGlobal, "global", "g", false, "remove from global install")
}

func runRemove(cmd *cobra.Command, args []string) error {
	name := args[0]
	mPath := manifestPath(removeGlobal)
	lPath := lockfilePath(removeGlobal)

	m, err := manifest.Load(mPath)
	if err != nil {
		return err
	}

	if !manifest.HasDependency(m, name) {
		return fmt.Errorf("skill %q not found in lore.toml", name)
	}

	opts := installer.Options{
		Global: removeGlobal,
		Root:   projectRoot(),
	}

	if err := installer.Remove(name, opts, m); err != nil {
		return fmt.Errorf("uninstall %s: %w", name, err)
	}

	manifest.RemoveDependency(m, name)
	if err := manifest.Save(mPath, m); err != nil {
		return err
	}

	lf, err := lockfile.Load(lPath)
	if err != nil {
		return err
	}
	lockfile.RemoveEntry(lf, name)
	if err := lockfile.Save(lPath, lf); err != nil {
		return err
	}

	fmt.Printf("removed %s\n", name)
	return nil
}
