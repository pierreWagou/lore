package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/pierreWagou/lore/internal/config"
	"github.com/pierreWagou/lore/internal/lockfile"
	"github.com/pierreWagou/lore/internal/manifest"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed skills",
	Args:  cobra.NoArgs,
	RunE:  runList,
}

var (
	listGlobal  bool
	listProfile string
)

func init() {
	listCmd.Flags().BoolVarP(&listGlobal, "global", "g", false, "list globally installed skills")
	listCmd.Flags().StringVar(&listProfile, "profile", "", "profile to list (global only; defaults to active profile)")
}

func runList(cmd *cobra.Command, args []string) error {
	if listGlobal {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// If a specific profile is requested, show only that one.
		// Otherwise show all profiles.
		if listProfile != "" {
			return listProfileDeps(cfg, listProfile)
		}

		if len(cfg.Profiles) == 0 {
			fmt.Println("no profiles configured.")
			return nil
		}
		for name := range cfg.Profiles {
			if err := listProfileDeps(cfg, name); err != nil {
				return err
			}
		}
		return nil
	}

	// Project-scoped list.
	mPath := manifestPath(false)
	lPath := lockfilePath(false)

	m, err := manifest.Load(mPath)
	if err != nil {
		return err
	}
	lf, err := lockfile.Load(lPath)
	if err != nil {
		return err
	}

	if len(m.Dependencies) == 0 {
		fmt.Println("no skills installed.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSOURCE\tREF\tCOMMIT")
	for _, dep := range m.Dependencies {
		commit := "(not locked)"
		if entry := lockfile.GetEntry(lf, dep.Name); entry != nil {
			commit = entry.Commit[:12]
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", dep.Name, dep.Source, dep.Ref, commit)
	}
	return w.Flush()
}

func listProfileDeps(cfg *config.Config, profileName string) error {
	profile := config.ResolveProfileFromConfig(cfg, profileName)
	if profile == nil {
		return fmt.Errorf("profile %q not found", profileName)
	}

	lPath := globalLockfilePath(profileName)
	lf, err := lockfile.Load(lPath)
	if err != nil {
		return err
	}

	fmt.Printf("[profile: %s]\n", profileName)
	if len(profile.Dependencies) == 0 {
		fmt.Println("  no skills installed.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "  NAME\tSOURCE\tREF\tCOMMIT")
	for _, dep := range profile.Dependencies {
		commit := "(not locked)"
		if entry := lockfile.GetEntry(lf, dep.Name); entry != nil && len(entry.Commit) >= 12 {
			commit = entry.Commit[:12]
		}
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n", dep.Name, dep.Source, dep.Ref, commit)
	}
	return w.Flush()
}
