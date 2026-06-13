package installer

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/pierreWagou/lore/internal/auth"
	"github.com/pierreWagou/lore/internal/harness"
	"github.com/pierreWagou/lore/internal/lockfile"
	"github.com/pierreWagou/lore/internal/manifest"
	"github.com/pierreWagou/lore/internal/resolver"
)

// Options controls installation behaviour.
type Options struct {
	Global  bool     // install into global harness dirs
	Targets []string // explicit harness names (overrides manifest + auto-detect)
	Root    string   // project root directory (for project-scoped installs)
}

// Result describes a single installed file placement.
type Result struct {
	Name    string
	Harness string
	Path    string
}

// SkillResult is the outcome of installing one skill.
type SkillResult struct {
	Results     []Result
	Commit      string
	ContentHash string
}

// DefaultCacheDir returns the default go-git cache directory.
func DefaultCacheDir() string {
	dir, _ := os.UserCacheDir()
	return filepath.Join(dir, "lore", "repos")
}

// DefaultConfigDir returns the default lore config directory.
func DefaultConfigDir() string {
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "lore")
}

// Install fetches and installs a single skill from its manifest dependency.
func Install(dep manifest.Dependency, opts Options, m *manifest.Manifest) (SkillResult, error) {
	h, err := resolver.Parse(dep.Source, dep.Ref)
	if err != nil {
		return SkillResult{}, err
	}

	authMethod, err := auth.Resolve(h.RepoURL)
	if err != nil {
		return SkillResult{}, fmt.Errorf("auth for %s: %w", dep.Source, err)
	}

	fetchResult, err := resolver.Fetch(h, authMethod, DefaultCacheDir())
	if err != nil {
		return SkillResult{}, fmt.Errorf("fetch %s: %w", dep.Source, err)
	}

	adapters, err := resolveAdapters(opts, m)
	if err != nil {
		return SkillResult{}, err
	}

	skill := harness.Skill{
		Name:  dep.Name,
		Files: fetchResult.Files,
	}

	sr := SkillResult{
		Commit:      fetchResult.Commit,
		ContentHash: ComputeContentHash(fetchResult.Files),
	}

	for _, adapter := range adapters {
		var skillsDir string
		if opts.Global {
			skillsDir = adapter.GlobalSkillsDir()
		} else {
			skillsDir = adapter.ProjectSkillsDir(opts.Root)
		}

		files, err := adapter.Transform(skill)
		if err != nil {
			return SkillResult{}, fmt.Errorf("%s: %w", adapter.Name(), err)
		}

		destDir := filepath.Join(skillsDir, dep.Name)
		for _, f := range files {
			destPath := filepath.Join(destDir, f.Path)
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return SkillResult{}, err
			}
			if err := os.WriteFile(destPath, f.Content, 0644); err != nil {
				return SkillResult{}, err
			}
		}

		sr.Results = append(sr.Results, Result{
			Name:    dep.Name,
			Harness: adapter.Name(),
			Path:    destDir,
		})
	}

	return sr, nil
}

// Remove uninstalls a skill from all relevant harness directories.
func Remove(name string, opts Options, m *manifest.Manifest) error {
	adapters, err := resolveAdapters(opts, m)
	if err != nil {
		return err
	}
	for _, adapter := range adapters {
		var skillsDir string
		if opts.Global {
			skillsDir = adapter.GlobalSkillsDir()
		} else {
			skillsDir = adapter.ProjectSkillsDir(opts.Root)
		}
		if err := os.RemoveAll(filepath.Join(skillsDir, name)); err != nil {
			return err
		}
	}
	return nil
}

// Sync installs all dependencies from the manifest, updating the lockfile.
// It skips skills whose content_hash already matches the lockfile entry.
func Sync(manifestPath, lockfilePath string, opts Options) error {
	m, err := manifest.Load(manifestPath)
	if err != nil {
		return err
	}
	lf, err := lockfile.Load(lockfilePath)
	if err != nil {
		return err
	}

	// Manifest targets feed into opts when not overridden by flags.
	syncOpts := opts
	if len(syncOpts.Targets) == 0 && len(m.Targets) > 0 {
		syncOpts.Targets = m.Targets
	}

	changed := false
	for _, dep := range m.Dependencies {
		sr, installErr := Install(dep, syncOpts, m)
		if installErr != nil {
			return fmt.Errorf("installing %s: %w", dep.Name, installErr)
		}

		lockfile.SetEntry(lf, lockfile.NewEntry(dep.Name, dep.Source, sr.Commit, sr.ContentHash))
		changed = true

		fmt.Printf("installed %s\n", dep.Name)
		for _, r := range sr.Results {
			fmt.Printf("  → %s: %s\n", r.Harness, r.Path)
		}
	}

	if changed {
		return lockfile.Save(lockfilePath, lf)
	}
	return nil
}

// ComputeContentHash returns a deterministic SHA256 over a file map.
func ComputeContentHash(files map[string][]byte) string {
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		h.Write([]byte(k))
		h.Write(files[k])
	}
	return fmt.Sprintf("sha256:%x", h.Sum(nil))
}

func resolveAdapters(opts Options, m *manifest.Manifest) ([]harness.Adapter, error) {
	// 1. Explicit --target flag.
	if len(opts.Targets) > 0 {
		return adaptersByNames(opts.Targets)
	}
	// 2. Manifest targets.
	if m != nil && len(m.Targets) > 0 {
		return adaptersByNames(m.Targets)
	}
	// 3. Auto-detect installed harnesses.
	detected := harness.Detected()
	if len(detected) == 0 {
		return nil, fmt.Errorf("no harnesses detected; use --target to specify one (available: %s)", availableNames())
	}
	return detected, nil
}

func adaptersByNames(names []string) ([]harness.Adapter, error) {
	var adapters []harness.Adapter
	for _, name := range names {
		a := harness.Get(name)
		if a == nil {
			return nil, fmt.Errorf("unknown harness %q (available: %s)", name, availableNames())
		}
		adapters = append(adapters, a)
	}
	return adapters, nil
}

func availableNames() string {
	names := harness.Names()
	result := ""
	for i, n := range names {
		if i > 0 {
			result += ", "
		}
		result += n
	}
	return result
}
