package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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
		fmt.Print("use these as harnesses? [Y/n] ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(answer)), "n") {
			m.Harnesses = names
		}
	}

	if len(m.Harnesses) == 0 {
		fmt.Printf("available harnesses: %s\n", strings.Join(harness.Names(), ", "))
		fmt.Print("enter harnesses (comma-separated): ")
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		for _, t := range strings.Split(line, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				manifest.AddHarness(m, t)
			}
		}
	}

	if err := manifest.Save(path, m); err != nil {
		return fmt.Errorf("writing lore.toml: %w", err)
	}
	fmt.Printf("created %s\n", path)

	// For project-scope init, update .gitignore with harness dirs.
	if !initGlobal && len(m.Harnesses) > 0 {
		cwd, _ := os.Getwd()
		gitignorePath := filepath.Join(cwd, ".gitignore")
		if err := updateGitignore(gitignorePath, m.Harnesses); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not update .gitignore: %v\n", err)
		} else {
			fmt.Printf("updated .gitignore with harness skill dirs\n")
		}
	}
	return nil
}

// updateGitignore appends harness skill directories to .gitignore if not already present.
func updateGitignore(path string, harnesses []string) error {
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	harnessIgnores := map[string]string{
		"opencode": ".opencode/skills/",
		"claude":   ".claude/skills/",
		"cursor":   ".cursor/rules/",
		"codex":    ".codex/skills/",
	}

	content := string(existing)
	var toAdd []string
	for _, target := range harnesses {
		if entry, ok := harnessIgnores[target]; ok {
			if !strings.Contains(content, entry) {
				toAdd = append(toAdd, entry)
			}
		}
	}

	if len(toAdd) == 0 {
		return nil
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		fmt.Fprintln(f)
	}
	fmt.Fprintln(f, "\n# lore — generated harness skill dirs (do not edit this block)")
	for _, entry := range toAdd {
		fmt.Fprintln(f, entry)
	}
	return nil
}
