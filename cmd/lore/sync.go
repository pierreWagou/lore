package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/pierreWagou/lore/internal/config"
	"github.com/pierreWagou/lore/internal/installer"
	"github.com/pierreWagou/lore/internal/manifest"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Install all skills from lore.toml",
	Args:  cobra.NoArgs,
	RunE:  runSync,
}

var (
	syncGlobal    bool
	syncHarnesses string
	syncSkillsDir string
	syncProfile   string
)

func init() {
	syncCmd.Flags().BoolVarP(&syncGlobal, "global", "g", false, "sync global skills")
	syncCmd.Flags().StringVar(&syncHarnesses, "harness", "", "comma-separated harnesses to sync")
	syncCmd.Flags().StringVar(&syncSkillsDir, "skills-dir", "", "install into this directory instead of the harness default (global installs only)")
	syncCmd.Flags().StringVar(&syncProfile, "profile", "", "use a named profile from lore.toml (global installs only)")
}

func runSync(cmd *cobra.Command, args []string) error {
	opts := installer.Options{
		Global:    syncGlobal,
		Harnesses: splitHarnesses(syncHarnesses),
		Root:      projectRoot(),
		SkillsDir: syncSkillsDir,
		Profile:   syncProfile,
	}

	if syncGlobal {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading lore.toml: %w", err)
		}
		profileName := resolveActiveProfile(syncProfile, cfg)
		if profileName == "" {
			return fmt.Errorf("no profile active — set default_profile in lore.toml or use --profile")
		}
		profile := config.ResolveProfileFromConfig(cfg, profileName)
		if profile == nil {
			return fmt.Errorf("profile %q not found in lore.toml", profileName)
		}
		if len(profile.Dependencies) == 0 {
			fmt.Printf("profile %q has no dependencies.\n", profileName)
			return nil
		}
		lPath := globalLockfilePath(profileName)
		return withHarnessRetryGlobal(&opts, cfg, func() error {
			return installer.SyncDeps(profile.Dependencies, lPath, opts)
		})
	}

	// Project-scoped sync.
	mPath := manifestPath(false)
	lPath := lockfilePath(false)

	if _, err := os.Stat(mPath); os.IsNotExist(err) {
		return fmt.Errorf("no lore.toml found; run `lore init` first")
	}

	m, err := manifest.Load(mPath)
	if err != nil {
		return fmt.Errorf("loading lore.toml: %w", err)
	}

	return withHarnessRetry(&opts, m, mPath, func() error {
		return installer.Sync(mPath, lPath, opts)
	})
}
