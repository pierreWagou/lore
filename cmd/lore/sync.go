package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/pierreWagou/lore/internal/installer"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Install all skills from lore.toml",
	Args:  cobra.NoArgs,
	RunE:  runSync,
}

var (
	syncGlobal  bool
	syncTargets string
)

func init() {
	syncCmd.Flags().BoolVarP(&syncGlobal, "global", "g", false, "sync global skills")
	syncCmd.Flags().StringVarP(&syncTargets, "target", "t", "", "comma-separated harnesses to sync")
}

func runSync(cmd *cobra.Command, args []string) error {
	mPath := manifestPath(syncGlobal)
	lPath := lockfilePath(syncGlobal)

	if _, err := os.Stat(mPath); os.IsNotExist(err) {
		return fmt.Errorf("no lore.toml found; run `lore init` first")
	}

	opts := installer.Options{
		Global:  syncGlobal,
		Targets: splitTargets(syncTargets),
		Root:    projectRoot(),
	}

	return installer.Sync(mPath, lPath, opts)
}
