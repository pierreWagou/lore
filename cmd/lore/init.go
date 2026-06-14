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
	reader := bufio.NewReader(os.Stdin)

	if _, err := os.Stat(path); err == nil {
		fmt.Fprintf(os.Stderr, "lore.toml already exists at %s\n", path)
		fmt.Fprint(os.Stderr, "overwrite? [y/N] ")
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

	root := projectRoot()

	// Guest mode: detect team harnesses (existing committed dirs, source only).
	if mode == "guest" && !initGlobal {
		teamFound := detectExistingHarnesses(root)
		if len(teamFound) > 0 {
			fmt.Printf("\nteam harnesses found (committed source, never modified by lore):\n")
			for _, name := range teamFound {
				fmt.Printf("  %s\n", name)
			}
			fmt.Print("use these as team_harnesses? [Y/n] ")
			answer, _ := reader.ReadString('\n')
			if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(answer)), "n") {
				for _, name := range teamFound {
					manifest.AddTeamHarness(m, name)
				}
			}
		}
	}

	// Personal harnesses: where lore installs skills for you.
	fmt.Println()
	detected := harness.Detected()

	// In guest mode, exclude already-selected team harnesses from personal suggestions.
	var suggestions []harness.Adapter
	for _, a := range detected {
		if !contains(m.TeamHarnesses, a.Name()) {
			suggestions = append(suggestions, a)
		}
	}

	if len(suggestions) > 0 {
		names := adapterNames(suggestions)
		fmt.Printf("personal harnesses (your install targets): %s\n", strings.Join(names, ", "))
		fmt.Print("use these? [Y/n] ")
		answer, _ := reader.ReadString('\n')
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(answer)), "n") {
			m.Harnesses = names
		}
	}

	if len(m.Harnesses) == 0 {
		fmt.Printf("available harnesses: %s\n", strings.Join(harness.Names(), ", "))
		fmt.Print("enter personal harnesses (comma-separated): ")
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
	fmt.Printf("\ncreated %s\n", path)

	// Set up exclusions.
	if !initGlobal {
		switch mode {
		case "guest":
			// Exclude personal harness dirs, .ai/skills/, and lore.lock.
			// Team harness dirs are NOT excluded — they are the team's committed source.
			entries := append(
				[]string{".ai/skills/", "lore.lock"},
				harnessIgnoreEntries(m.Harnesses)...,
			)
			if err := updateGitExclude(root, entries); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not update .git/info/exclude: %v\n", err)
			} else {
				fmt.Println("updated .git/info/exclude (local-only, not committed)")
			}
		case "keeper":
			// Exclude generated harness dirs in .gitignore. .ai/skills/ is committed.
			if err := updateGitignore(filepath.Join(root, ".gitignore"), m.Harnesses); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not update .gitignore: %v\n", err)
			} else {
				fmt.Println("updated .gitignore with harness skill dirs")
			}
		}
	}
	return nil
}

// detectExistingHarnesses returns names of harnesses whose project skill dir exists.
func detectExistingHarnesses(root string) []string {
	var found []string
	for _, a := range harness.All() {
		if _, err := os.Stat(a.ProjectSkillsDir(root)); err == nil {
			found = append(found, a.Name())
		}
	}
	return found
}

// hasExistingHarnessDirs returns true if any harness project skill dir exists.
func hasExistingHarnessDirs(root string) bool {
	return len(detectExistingHarnesses(root)) > 0
}

// adapterNames extracts names from a slice of adapters.
func adapterNames(adapters []harness.Adapter) []string {
	names := make([]string, len(adapters))
	for i, a := range adapters {
		names[i] = a.Name()
	}
	return names
}

// contains reports whether slice contains s.
func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// harnessIgnoreEntries returns gitignore entries for the given harness names.
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
