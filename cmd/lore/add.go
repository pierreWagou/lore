package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pierreWagou/lore/internal/auth"
	"github.com/pierreWagou/lore/internal/harness"
	"github.com/pierreWagou/lore/internal/installer"
	"github.com/pierreWagou/lore/internal/lockfile"
	"github.com/pierreWagou/lore/internal/manifest"
	"github.com/pierreWagou/lore/internal/resolver"
	"github.com/pierreWagou/lore/internal/scanner"
)

var addCmd = &cobra.Command{
	Use:   "add <source>",
	Short: "Add and install a skill",
	Long: `Add a skill to the manifest and install it.

Source formats:
  owner/repo/path/to/skill        GitHub shorthand
  owner/repo                      scan repo for all skills
  https://github.com/owner/repo/tree/main/path/to/skill
  git@github.com:owner/repo.git/path/to/skill
  ./local/path/to/skill`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

var (
	addGlobal  bool
	addTargets string
	addName    string
	addRef     string
	addAll     bool
)

func init() {
	addCmd.Flags().BoolVarP(&addGlobal, "global", "g", false, "install globally")
	addCmd.Flags().StringVarP(&addTargets, "target", "t", "", "comma-separated harnesses (e.g. opencode,claude)")
	addCmd.Flags().StringVarP(&addName, "name", "n", "", "skill name (defaults to last path segment)")
	addCmd.Flags().StringVarP(&addRef, "ref", "r", "", "git ref: branch, tag, or SHA (default: HEAD)")
	addCmd.Flags().BoolVar(&addAll, "all", false, "install all skills found without prompting")
}

func runAdd(cmd *cobra.Command, args []string) error {
	source := args[0]
	h, err := resolver.Parse(source, addRef)
	if err != nil {
		return err
	}

	targets := splitTargets(addTargets)
	opts := installer.Options{
		Global:  addGlobal,
		Targets: targets,
		Root:    projectRoot(),
	}

	mPath := manifestPath(addGlobal)
	lPath := lockfilePath(addGlobal)

	m, err := manifest.Load(mPath)
	if err != nil {
		return err
	}
	lf, err := lockfile.Load(lPath)
	if err != nil {
		return err
	}

	// Scan mode: no explicit subpath — discover all skills in the repo.
	if h.ScanAll() {
		return runAddScan(h, opts, m, lf, mPath, lPath)
	}

	// Direct install of a single skill.
	name := addName
	if name == "" {
		name = inferName(h.SubPath, source)
	}
	return installOne(name, source, h.Ref, opts, m, lf, mPath, lPath)
}

func runAddScan(h resolver.Handle, opts installer.Options, m *manifest.Manifest, lf *lockfile.Lockfile, mPath, lPath string) error {
	fmt.Printf("scanning %s for skills...\n", h.RepoURL)

	authMethod, err := auth.Resolve(h.RepoURL)
	if err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	fetchResult, err := resolver.Fetch(h, authMethod, installer.DefaultCacheDir())
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	dirs := scanner.Scan(fetchResult.Files)
	if len(dirs) == 0 {
		return fmt.Errorf("no skills (SKILL.md files) found in %s", h.RepoURL)
	}

	// Build skill entries for each found directory.
	var candidates []candidate
	for _, dir := range dirs {
		var src string
		if dir == "" {
			src = h.Raw
		} else {
			src = buildSource(h, dir)
		}
		candidates = append(candidates, candidate{
			name:   filepath.Base(dir),
			source: src,
		})
	}

	selected := candidates
	if !addAll && len(candidates) > 1 {
		selected, err = promptSelectSkills(candidates)
		if err != nil {
			return err
		}
	}

	if len(selected) == 0 {
		fmt.Println("no skills selected.")
		return nil
	}

	for _, c := range selected {
		name := addName
		if name == "" || len(selected) > 1 {
			name = c.name
		}
		if err := installOne(name, c.source, h.Ref, opts, m, lf, mPath, lPath); err != nil {
			return err
		}
	}
	return nil
}

func installOne(name, source, ref string, opts installer.Options, m *manifest.Manifest, lf *lockfile.Lockfile, mPath, lPath string) error {
	dep := manifest.Dependency{
		Name:   name,
		Source: source,
		Ref:    ref,
	}
	if dep.Ref == "" {
		dep.Ref = "HEAD"
	}

	fmt.Printf("installing %s from %s...\n", name, source)

	sr, err := installer.Install(dep, opts, m)
	if err != nil {
		return fmt.Errorf("install %s: %w", name, err)
	}

	for _, r := range sr.Results {
		fmt.Printf("  → %s: %s\n", r.Harness, r.Path)
	}

	manifest.AddDependency(m, dep)
	lockfile.SetEntry(lf, lockfile.NewEntry(name, source, sr.Commit, sr.ContentHash))

	if err := manifest.Save(mPath, m); err != nil {
		return err
	}
	return lockfile.Save(lPath, lf)
}

type candidate struct {
	name   string
	source string
}

func promptSelectSkills(candidates []candidate) ([]candidate, error) {
	fmt.Printf("found %d skills:\n", len(candidates))
	for i, c := range candidates {
		fmt.Printf("  [%d] %s (%s)\n", i+1, c.name, c.source)
	}
	fmt.Print("select skills to install (e.g. 1,3 or 'all'): ")

	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)

	if strings.ToLower(line) == "all" {
		return candidates, nil
	}

	var selected []candidate
	for _, part := range strings.Split(line, ",") {
		part = strings.TrimSpace(part)
		var idx int
		if _, err := fmt.Sscanf(part, "%d", &idx); err != nil || idx < 1 || idx > len(candidates) {
			return nil, fmt.Errorf("invalid selection %q", part)
		}
		selected = append(selected, candidates[idx-1])
	}
	return selected, nil
}

func buildSource(h resolver.Handle, subPath string) string {
	switch h.Kind {
	case resolver.KindSSH:
		return fmt.Sprintf("%s/%s", h.RepoURL[:len(h.RepoURL)-4], subPath) // strip .git, add path
	case resolver.KindHTTPS:
		return fmt.Sprintf("%s/%s", h.RepoURL, subPath)
	default:
		return fmt.Sprintf("%s/%s/%s", h.Owner, h.RepoName, subPath)
	}
}

func inferName(subPath, fallback string) string {
	if subPath != "" {
		return filepath.Base(subPath)
	}
	parts := strings.Split(fallback, "/")
	return parts[len(parts)-1]
}

func splitTargets(s string) []string {
	if s == "" {
		return nil
	}
	var targets []string
	for _, t := range strings.Split(s, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			targets = append(targets, t)
		}
	}
	return targets
}

// Re-export for use in other command files.
func availableHarnessNames() string {
	return strings.Join(harness.Names(), ", ")
}
