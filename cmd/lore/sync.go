package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

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
)

func init() {
	syncCmd.Flags().BoolVarP(&syncGlobal, "global", "g", false, "sync global skills")
	syncCmd.Flags().StringVar(&syncHarnesses, "harness", "", "comma-separated harnesses to sync")
}

func runSync(cmd *cobra.Command, args []string) error {
	mPath := manifestPath(syncGlobal)
	lPath := lockfilePath(syncGlobal)

	if _, err := os.Stat(mPath); os.IsNotExist(err) {
		return fmt.Errorf("no lore.toml found; run `lore init` first")
	}

	opts := installer.Options{
		Global:    syncGlobal,
		Harnesses: splitHarnesses(syncHarnesses),
		Root:      projectRoot(),
	}

	m, err := manifest.Load(mPath)
	if err != nil {
		return fmt.Errorf("loading lore.toml: %w", err)
	}

	return withHarnessRetry(&opts, m, mPath, func() error {
		return installer.Sync(mPath, lPath, opts)
	})
}
