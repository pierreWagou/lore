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

var (
	initGlobal bool
	initMode   string
)

func init() {
	initCmd.Flags().BoolVarP(&initGlobal, "global", "g", false, "initialise global config (~/.config/lore/lore.toml)")
	initCmd.Flags().StringVar(&initMode, "mode", "", `mode: "guest" (adapt to existing harness dirs) or "keeper" (lore-first, default)`)
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

	// Resolve mode: explicit flag > auto-detect.
	mode := initMode
	if mode == "" {
		if !initGlobal && hasExistingHarnessDirs(projectRoot()) {
			mode = "guest"
		} else {
			mode = "keeper"
		}
	}
	if mode != "guest" && mode != "keeper" {
		return fmt.Errorf("invalid mode %q: must be \"guest\" or \"keeper\"", mode)
	}
	m.Mode = mode
	fmt.Printf("mode: %s\n", mode)

	// Detect installed harnesses and offer them as defaults.
	detected := harness.Detected()
	if len(detected) > 0 {
		names := make([]string, len(detected))
		for i, a := range detected {
			names[i] = a.Name()
		}
		fmt.Printf("detected harnesses: %s\n", strings.Join(names, ", "))
		fmt.Print("use these? [Y/n] ")
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
		for _, h := range strings.Split(line, ",") {
			h = strings.TrimSpace(h)
			if h != "" {
				manifest.AddHarness(m, h)
			}
		}
	}

	if err := manifest.Save(path, m); err != nil {
		return fmt.Errorf("writing lore.toml: %w", err)
	}
	fmt.Printf("created %s\n", path)

	// Set up exclusions — scope depends on mode.
	if !initGlobal {
		root := projectRoot()
		harnessEntries := harnessIgnoreEntries(m.Harnesses)

		switch mode {
		case "guest":
			// Guest: all lore artifacts are local-only — use .git/info/exclude.
			entries := append([]string{".ai/skills/"}, harnessEntries...)
			if err := updateGitExclude(root, entries); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not update .git/info/exclude: %v\n", err)
			} else {
				fmt.Println("updated .git/info/exclude (local-only, not committed)")
			}
		case "keeper":
			// Keeper: harness dirs are generated, gitignore them; .ai/skills/ is committed.
			if err := updateGitignore(filepath.Join(root, ".gitignore"), m.Harnesses); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not update .gitignore: %v\n", err)
			} else {
				fmt.Println("updated .gitignore with harness skill dirs")
			}
		}
	}
	return nil
}

// hasExistingHarnessDirs returns true if any harness project skill dir already exists.
func hasExistingHarnessDirs(root string) bool {
	for _, a := range harness.All() {
		if _, err := os.Stat(a.ProjectSkillsDir(root)); err == nil {
			return true
		}
	}
	return false
}

// harnessIgnoreEntries returns the .gitignore / .git/info/exclude entries for each harness.
func harnessIgnoreEntries(harnesses []string) []string {
	known := map[string]string{
		"opencode": ".opencode/skills/",
		"claude":   ".claude/skills/",
		"cursor":   ".cursor/rules/",
		"codex":    ".codex/skills/",
	}
	var entries []string
	for _, h := range harnesses {
		if e, ok := known[h]; ok {
			entries = append(entries, e)
		}
	}
	return entries
}

// updateGitignore appends harness skill directories to .gitignore if not already present.
func updateGitignore(path string, harnesses []string) error {
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	content := string(existing)
	var toAdd []string
	for _, entry := range harnessIgnoreEntries(harnesses) {
		if !strings.Contains(content, entry) {
			toAdd = append(toAdd, entry)
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
	for _, e := range toAdd {
		fmt.Fprintln(f, e)
	}
	return nil
}

// updateGitExclude appends entries to .git/info/exclude (local-only, never committed).
func updateGitExclude(projectRoot string, entries []string) error {
	excludePath := filepath.Join(projectRoot, ".git", "info", "exclude")
	existing, err := os.ReadFile(excludePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	content := string(existing)
	var toAdd []string
	for _, e := range entries {
		if !strings.Contains(content, e) {
			toAdd = append(toAdd, e)
		}
	}
	if len(toAdd) == 0 {
		return nil
	}
	f, err := os.OpenFile(excludePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		fmt.Fprintln(f)
	}
	fmt.Fprintln(f, "\n# lore (guest mode) — local-only, not committed")
	for _, e := range toAdd {
		fmt.Fprintln(f, e)
	}
	return nil
}
