package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pierreWagou/lore/internal/config"
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

var (
	removeGlobal  bool
	removeProfile string
)

func init() {
	removeCmd.Flags().BoolVarP(&removeGlobal, "global", "g", false, "remove from global install")
	removeCmd.Flags().StringVar(&removeProfile, "profile", "", "profile to remove from (global only)")
}

func runRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	if removeGlobal {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		profileName := resolveActiveProfile(removeProfile, cfg)
		if profileName == "" {
			return fmt.Errorf("no profile active — set default_profile in lore.toml or use --profile")
		}
		if !config.HasDependency(cfg, profileName, name) {
			return fmt.Errorf("skill %q not found in profile %q", name, profileName)
		}

		opts := installer.Options{
			Global:  true,
			Profile: profileName,
		}
		if err := installer.Remove(name, opts, nil); err != nil {
			return fmt.Errorf("uninstall %s: %w", name, err)
		}

		config.RemoveDependency(cfg, profileName, name)
		if err := config.Save(cfg); err != nil {
			return err
		}

		lPath := globalLockfilePath(profileName)
		lf, err := lockfile.Load(lPath)
		if err != nil {
			return err
		}
		lockfile.RemoveEntry(lf, name)
		if err := lockfile.Save(lPath, lf); err != nil {
			return err
		}

		fmt.Printf("removed %s from profile %q\n", name, profileName)
		return nil
	}

	// Project-scoped remove.
	mPath := manifestPath(false)
	lPath := lockfilePath(false)

	m, err := manifest.Load(mPath)
	if err != nil {
		return err
	}

	if !manifest.HasDependency(m, name) {
		return fmt.Errorf("skill %q not found in lore.toml", name)
	}

	opts := installer.Options{
		Global: false,
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
