package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pierreWagou/lore/internal/harness"
	"github.com/pierreWagou/lore/internal/manifest"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a lore.toml in the current directory",
	Args:  cobra.NoArgs,
	RunE:  runInit,
}

var initGlobal bool

func init() {
	initCmd.Flags().BoolVarP(&initGlobal, "global", "g", false, "initialise global config (~/.config/lore/lore.toml)")
}

func runInit(cmd *cobra.Command, args []string) error {
	path := manifestPath(initGlobal)

	if _, err := os.Stat(path); err == nil {
		fmt.Fprintf(os.Stderr, "lore.toml already exists at %s\n", path)
		fmt.Fprint(os.Stderr, "overwrite? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(answer)), "y") {
			fmt.Println("aborted.")
			return nil
		}
	}

	m := &manifest.Manifest{}

	// Detect installed harnesses and offer them as defaults.
	detected := harness.Detected()
	if len(detected) > 0 {
		names := make([]string, len(detected))
		for i, a := range detected {
			names[i] = a.Name()
		}
		fmt.Printf("detected harnesses: %s\n", strings.Join(names, ", "))
		fmt.Print("use these as targets? [Y/n] ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(answer)), "n") {
			m.Targets = names
		}
	}

	if len(m.Targets) == 0 {
		fmt.Printf("available harnesses: %s\n", strings.Join(harness.Names(), ", "))
		fmt.Print("enter targets (comma-separated): ")
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		for _, t := range strings.Split(line, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				manifest.AddTarget(m, t)
			}
		}
	}

	if err := manifest.Save(path, m); err != nil {
		return fmt.Errorf("writing lore.toml: %w", err)
	}
	fmt.Printf("created %s\n", path)
	return nil
}
