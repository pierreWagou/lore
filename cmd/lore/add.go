package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pierreWagou/lore/internal/auth"
	"github.com/pierreWagou/lore/internal/config"
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
	addGlobal    bool
	addHarnesses string
	addName      string
	addRef       string
	addAll       bool
	addSkillsDir string
	addProfile   string
)

func init() {
	addCmd.Flags().BoolVarP(&addGlobal, "global", "g", false, "install globally")
	addCmd.Flags().StringVar(&addHarnesses, "harness", "", "comma-separated harnesses (e.g. opencode,claude)")
	addCmd.Flags().StringVarP(&addName, "name", "n", "", "skill name (defaults to last path segment)")
	addCmd.Flags().StringVarP(&addRef, "ref", "r", "", "git ref: branch, tag, or SHA (default: HEAD)")
	addCmd.Flags().BoolVar(&addAll, "all", false, "install all skills found without prompting")
	addCmd.Flags().StringVar(&addSkillsDir, "skills-dir", "", "install into this directory instead of the harness default (global installs only)")
	addCmd.Flags().StringVar(&addProfile, "profile", "", "use a named profile from ~/.config/lore/lore.toml (global installs only)")
}

func runAdd(cmd *cobra.Command, args []string) error {
	source := args[0]
	h, err := resolver.Parse(source, addRef)
	if err != nil {
		return err
	}

	harnesses := splitHarnesses(addHarnesses)
	opts := installer.Options{
		Global:    addGlobal,
		Harnesses: harnesses,
		Root:      projectRoot(),
		SkillsDir: addSkillsDir,
		Profile:   addProfile,
	}

	// Scan mode: no explicit subpath — discover all skills in the repo.
	if h.ScanAll() {
		if addGlobal {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			profileName := resolveActiveProfile(addProfile, cfg)
			lPath := globalLockfilePath(profileName)
			lf, err := lockfile.Load(lPath)
			if err != nil {
				return err
			}
			return runAddScanGlobal(h, opts, cfg, profileName, lf, lPath)
		}
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
		return runAddScan(h, opts, m, lf, mPath, lPath)
	}

	// Direct install of a single skill.
	name := addName
	if name == "" {
		name = inferName(h.SubPath, source)
	}

	if addGlobal {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		profileName := resolveActiveProfile(addProfile, cfg)
		lPath := globalLockfilePath(profileName)
		lf, err := lockfile.Load(lPath)
		if err != nil {
			return err
		}
		return installOneGlobal(name, source, h.Ref, opts, cfg, profileName, lf, lPath)
	}

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
	return installOne(name, source, h.Ref, opts, m, lf, mPath, lPath)
}

// scanCandidates fetches a repo, scans it for skills, and optionally prompts the user
// to select a subset. Returns the selected candidates.
func scanCandidates(h resolver.Handle) ([]candidate, error) {
	fmt.Printf("scanning %s for skills...\n", h.RepoURL)

	token := auth.ResolveToken(h.RepoURL)
	fetchResult, err := resolver.Fetch(h, token, installer.DefaultCacheDir())
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}

	dirs := scanner.Scan(fetchResult.Files)
	if len(dirs) == 0 {
		return nil, fmt.Errorf("no skills (SKILL.md files) found in %s", h.RepoURL)
	}

	var candidates []candidate
	for _, dir := range dirs {
		var src string
		if dir == "" {
			src = h.Raw
		} else {
			src = buildSource(h, dir)
		}
		candidates = append(candidates, candidate{name: filepath.Base(dir), source: src})
	}

	if addAll || len(candidates) == 1 {
		return candidates, nil
	}
	return promptSelectSkills(candidates)
}

func runAddScan(h resolver.Handle, opts installer.Options, m *manifest.Manifest, lf *lockfile.Lockfile, mPath, lPath string) error {
	selected, err := scanCandidates(h)
	if err != nil {
		return err
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

func runAddScanGlobal(h resolver.Handle, opts installer.Options, cfg *config.Config, profileName string, lf *lockfile.Lockfile, lPath string) error {
	selected, err := scanCandidates(h)
	if err != nil {
		return err
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
		if err := installOneGlobal(name, c.source, h.Ref, opts, cfg, profileName, lf, lPath); err != nil {
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

	var sr installer.SkillResult
	err := withHarnessRetry(&opts, m, mPath, func() error {
		var installErr error
		sr, installErr = installer.Install(dep, opts, m)
		return installErr
	})
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

func installOneGlobal(name, source, ref string, opts installer.Options, cfg *config.Config, profileName string, lf *lockfile.Lockfile, lPath string) error {
	dep := manifest.Dependency{Name: name, Source: source, Ref: ref}
	if dep.Ref == "" {
		dep.Ref = "HEAD"
	}

	fmt.Printf("installing %s from %s...\n", name, source)

	var sr installer.SkillResult
	err := withHarnessRetryGlobal(&opts, func() error {
		var installErr error
		sr, installErr = installer.Install(dep, opts, nil)
		return installErr
	})
	if err != nil {
		return fmt.Errorf("install %s: %w", name, err)
	}

	for _, r := range sr.Results {
		fmt.Printf("  → %s: %s\n", r.Harness, r.Path)
	}

	config.AddDependency(cfg, profileName, dep)
	lockfile.SetEntry(lf, lockfile.NewEntry(name, source, sr.Commit, sr.ContentHash))

	if err := config.Save(cfg); err != nil {
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
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(h.RepoURL, ".git"), subPath)
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

func splitHarnesses(s string) []string {
	if s == "" {
		return nil
	}
	var harnesses []string
	for _, h := range strings.Split(s, ",") {
		h = strings.TrimSpace(h)
		if h != "" {
			harnesses = append(harnesses, h)
		}
	}
	return harnesses
}

// Re-export for use in other command files.
func availableHarnessNames() string {
	return strings.Join(harness.Names(), ", ")
}
