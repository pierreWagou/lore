package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pierreWagou/lore/internal/harness"
	"github.com/pierreWagou/lore/internal/manifest"
)

var harnessesCmd = &cobra.Command{
	Use:   "harnesses",
	Short: "Show configured and detected harnesses",
	Args:  cobra.NoArgs,
	RunE:  runHarnesses,
}

var harnessAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a harness to lore.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runHarnessAdd,
}

var harnessRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a harness from lore.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runHarnessRemove,
}

var harnessAddTeam bool

func init() {
	harnessesCmd.AddCommand(harnessAddCmd)
	harnessesCmd.AddCommand(harnessRemoveCmd)
	harnessAddCmd.Flags().BoolVar(&harnessAddTeam, "team", false, "add as team harness (guest mode only)")
}

func runHarnesses(cmd *cobra.Command, args []string) error {
	mPath := manifestPath(false)
	m, err := manifest.Load(mPath)
	if err != nil {
		return fmt.Errorf("loading lore.toml: %w", err)
	}

	if len(m.Harnesses) > 0 || len(m.TeamHarnesses) > 0 {
		fmt.Println("configured:")
		if len(m.Harnesses) > 0 {
			fmt.Printf("  personal:  %s\n", strings.Join(m.Harnesses, ", "))
		}
		if len(m.TeamHarnesses) > 0 {
			fmt.Printf("  team:      %s\n", strings.Join(m.TeamHarnesses, ", "))
		}
		fmt.Println()
	}

	detected := harness.Detected()
	if len(detected) == 0 {
		fmt.Println("no harnesses detected on this machine.")
		fmt.Printf("available: %s\n", availableHarnessNames())
		return nil
	}

	fmt.Println("detected:")
	for _, a := range detected {
		suffix := ""
		if contains(m.Harnesses, a.Name()) {
			suffix = "  (personal)"
		} else if contains(m.TeamHarnesses, a.Name()) {
			suffix = "  (team)"
		}
		fmt.Printf("  %s%s\n", a.Name(), suffix)
	}
	return nil
}

func runHarnessAdd(cmd *cobra.Command, args []string) error {
	name := args[0]

	if harness.Get(name) == nil {
		return fmt.Errorf("unknown harness %q (available: %s)", name, availableHarnessNames())
	}

	mPath := manifestPath(false)
	m, err := manifest.Load(mPath)
	if err != nil {
		return err
	}

	if harnessAddTeam {
		if !manifest.IsGuest(m) {
			return fmt.Errorf("team harnesses are only valid in guest mode (current mode: %q)", m.Mode)
		}
		if contains(m.TeamHarnesses, name) {
			fmt.Printf("harness %q already in team_harnesses — nothing to do.\n", name)
			return nil
		}
		manifest.AddTeamHarness(m, name)
		fmt.Printf("added %q to team_harnesses\n", name)
	} else {
		if contains(m.Harnesses, name) {
			fmt.Printf("harness %q already in harnesses — nothing to do.\n", name)
			return nil
		}
		manifest.AddHarness(m, name)
		fmt.Printf("added %q to harnesses\n", name)
	}

	if !harness.Get(name).Detect() {
		fmt.Fprintf(os.Stderr, "note: %q is not detected on this machine\n", name)
	}

	return manifest.Save(mPath, m)
}

func runHarnessRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	mPath := manifestPath(false)
	m, err := manifest.Load(mPath)
	if err != nil {
		return err
	}

	removedPersonal := removeFromSlice(&m.Harnesses, name)
	removedTeam := removeFromSlice(&m.TeamHarnesses, name)

	if !removedPersonal && !removedTeam {
		return fmt.Errorf("harness %q not found in lore.toml", name)
	}

	if removedPersonal {
		fmt.Printf("removed %q from harnesses\n", name)
	}
	if removedTeam {
		fmt.Printf("removed %q from team_harnesses\n", name)
	}
	return manifest.Save(mPath, m)
}

// removeFromSlice removes the first occurrence of s from *slice.
// Returns true if removed.
func removeFromSlice(slice *[]string, s string) bool {
	for i, v := range *slice {
		if v == s {
			*slice = append((*slice)[:i], (*slice)[i+1:]...)
			return true
		}
	}
	return false
}
