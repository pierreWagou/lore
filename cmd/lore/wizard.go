package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/pierreWagou/lore/internal/harness"
	"github.com/pierreWagou/lore/internal/installer"
	"github.com/pierreWagou/lore/internal/manifest"
)

// withHarnessRetry calls fn. If fn returns ErrNoHarnesses it runs the harness
// selection wizard, sets opts.Harnesses, and calls fn again.
func withHarnessRetry(opts *installer.Options, m *manifest.Manifest, mPath string, fn func() error) error {
	err := fn()
	if !errors.As(err, &installer.ErrNoHarnesses{}) {
		return err
	}
	harnesses, wizErr := promptSelectHarnesses(mPath, m)
	if wizErr != nil {
		return wizErr
	}
	opts.Harnesses = harnesses
	return fn()
}

// promptSelectHarnesses displays an interactive wizard to select harness targets.
// Only called when no harness flag was given and lore.toml has no harnesses configured.
// If the user chooses to save, the selection is persisted to lore.toml.
func promptSelectHarnesses(mPath string, m *manifest.Manifest) ([]string, error) {
	all := harness.Names()
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\nno harnesses configured or detected.")
	fmt.Println("\navailable harnesses:")
	for i, name := range all {
		fmt.Printf("  [%d] %s\n", i+1, name)
	}
	fmt.Print("\nselect harnesses to use (e.g. 1,2 or 'all'): ")

	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)

	var selected []string
	if strings.ToLower(line) == "all" {
		selected = all
	} else {
		for _, part := range strings.Split(line, ",") {
			part = strings.TrimSpace(part)
			var idx int
			if _, err := fmt.Sscanf(part, "%d", &idx); err != nil || idx < 1 || idx > len(all) {
				return nil, fmt.Errorf("invalid selection %q — enter numbers like 1,2 or 'all'", part)
			}
			selected = append(selected, all[idx-1])
		}
	}

	if len(selected) == 0 {
		return nil, fmt.Errorf("no harnesses selected")
	}

	if mPath != "" {
		fmt.Print("\nsave selection to lore.toml? [Y/n]: ")
		answer, _ := reader.ReadString('\n')
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(answer)), "n") {
			for _, h := range selected {
				manifest.AddHarness(m, h)
			}
			if err := manifest.Save(mPath, m); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not save lore.toml: %v\n", err)
			} else {
				fmt.Printf("saved harnesses to %s\n", mPath)
			}
		}
	}

	fmt.Println()
	return selected, nil
}
