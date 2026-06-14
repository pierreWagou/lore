package installer

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pierreWagou/lore/internal/auth"
	"github.com/pierreWagou/lore/internal/harness"
	"github.com/pierreWagou/lore/internal/lockfile"
	"github.com/pierreWagou/lore/internal/manifest"
	"github.com/pierreWagou/lore/internal/resolver"
)

// ErrNoHarnesses is returned when no harness targets can be resolved —
// no --harness flag, no lore.toml harnesses, and no harnesses auto-detected.
type ErrNoHarnesses struct{}

func (ErrNoHarnesses) Error() string { return "no harnesses configured or detected" }

// Options controls installation behaviour.
type Options struct {
	Global    bool     // install into global harness dirs (no .ai/skills/ neutral store)
	Harnesses []string // explicit harness names (overrides manifest + auto-detect)
	Root      string   // project root directory (for project-scoped installs)
}

// Result describes a single installed skill placement.
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

// neutralSkillsDir returns the .ai/skills directory for the given project root.
func neutralSkillsDir(root string) string {
	return filepath.Join(root, ".ai", "skills")
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

	token := auth.ResolveToken(h.RepoURL)

	fetchResult, err := resolver.Fetch(h, token, DefaultCacheDir())
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

	var results []Result
	if opts.Global {
		results, err = placeGlobal(skill, adapters)
	} else {
		results, err = placeProject(skill, adapters, opts.Root)
	}
	if err != nil {
		return SkillResult{}, err
	}

	return SkillResult{
		Results:     results,
		Commit:      fetchResult.Commit,
		ContentHash: ComputeContentHash(fetchResult.Files),
	}, nil
}

// placeGlobal installs a skill directly into each harness's global skills directory.
func placeGlobal(skill harness.Skill, adapters []harness.Adapter) ([]Result, error) {
	var results []Result
	for _, adapter := range adapters {
		skillsDir := adapter.GlobalSkillsDir()
		files, err := adapter.Transform(skill)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", adapter.Name(), err)
		}
		destDir := filepath.Join(skillsDir, skill.Name)
		if err := writeFilesToDir(destDir, files); err != nil {
			return nil, err
		}
		results = append(results, Result{
			Name:    skill.Name,
			Harness: adapter.Name(),
			Path:    destDir,
		})
	}
	return results, nil
}

// placeProject writes a skill to the neutral .ai/skills/<name>/ store, then
// creates a symlink (or transformed copy) in each harness's project skills directory.
func placeProject(skill harness.Skill, adapters []harness.Adapter, root string) ([]Result, error) {
	neutralDir := filepath.Join(neutralSkillsDir(root), skill.Name)

	// Write to neutral store (idempotent — safe even if source == neutralDir).
	if err := writeRawFiles(neutralDir, skill.Files); err != nil {
		return nil, fmt.Errorf("writing to .ai/skills/%s: %w", skill.Name, err)
	}

	var results []Result
	for _, adapter := range adapters {
		skillsDir := adapter.ProjectSkillsDir(root)
		destDir := filepath.Join(skillsDir, skill.Name)

		if adapter.NeedsTransform() {
			// Transform and write a copy — symlink not suitable.
			files, err := adapter.Transform(skill)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", adapter.Name(), err)
			}
			if err := writeFilesToDir(destDir, files); err != nil {
				return nil, err
			}
		} else {
			// Create a relative symlink: harness/skills/<name> → ../../.ai/skills/<name>
			if err := os.MkdirAll(skillsDir, 0755); err != nil {
				return nil, err
			}
			if err := createSymlink(destDir, neutralDir); err != nil {
				return nil, fmt.Errorf("symlinking %s for %s: %w", skill.Name, adapter.Name(), err)
			}
		}

		results = append(results, Result{
			Name:    skill.Name,
			Harness: adapter.Name(),
			Path:    destDir,
		})
	}
	return results, nil
}

// createSymlink creates a relative symlink at target pointing to source.
// Removes any existing file/symlink at target first.
func createSymlink(target, source string) error {
	rel, err := filepath.Rel(filepath.Dir(target), source)
	if err != nil {
		return err
	}
	_ = os.Remove(target)
	return os.Symlink(rel, target)
}

// writeRawFiles writes a map of file contents to dir, creating it if needed.
func writeRawFiles(dir string, files map[string][]byte) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	for path, content := range files {
		dest := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(dest, content, 0644); err != nil {
			return err
		}
	}
	return nil
}

// writeFilesToDir writes transformed harness.File slice to dir.
func writeFilesToDir(dir string, files []harness.File) error {
	for _, f := range files {
		dest := filepath.Join(dir, f.Path)
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(dest, f.Content, 0644); err != nil {
			return err
		}
	}
	return nil
}

// Remove uninstalls a skill from all relevant harness directories and the neutral store.
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
	// For project-scope installs, also remove from the neutral store.
	if !opts.Global {
		neutralDir := filepath.Join(neutralSkillsDir(opts.Root), name)
		if err := os.RemoveAll(neutralDir); err != nil {
			return err
		}
	}
	return nil
}

// Sync installs all dependencies from the manifest, updating the lockfile.
func Sync(manifestPath, lockfilePath string, opts Options) error {
	m, err := manifest.Load(manifestPath)
	if err != nil {
		return err
	}
	lf, err := lockfile.Load(lockfilePath)
	if err != nil {
		return err
	}

	syncOpts := opts
	if len(syncOpts.Harnesses) == 0 && len(m.Harnesses) > 0 {
		syncOpts.Harnesses = m.Harnesses
	}

	for _, dep := range m.Dependencies {
		sr, installErr := Install(dep, syncOpts, m)
		if installErr != nil {
			return fmt.Errorf("installing %s: %w", dep.Name, installErr)
		}
		lockfile.SetEntry(lf, lockfile.NewEntry(dep.Name, dep.Source, sr.Commit, sr.ContentHash))
		fmt.Printf("installed %s\n", dep.Name)
		for _, r := range sr.Results {
			fmt.Printf("  → %s: %s\n", r.Harness, r.Path)
		}
	}

	return lockfile.Save(lockfilePath, lf)
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
	if len(opts.Harnesses) > 0 {
		return adaptersByNames(opts.Harnesses)
	}
	if m != nil && len(m.Harnesses) > 0 {
		return adaptersByNames(m.Harnesses)
	}
	detected := harness.Detected()
	if len(detected) == 0 {
		return nil, ErrNoHarnesses{}
	}
	return detected, nil
}

func adaptersByNames(names []string) ([]harness.Adapter, error) {
	var adapters []harness.Adapter
	for _, name := range names {
		a := harness.Get(name)
		if a == nil {
			return nil, fmt.Errorf("unknown harness %q (available: %s)", name, strings.Join(harness.Names(), ", "))
		}
		adapters = append(adapters, a)
	}
	return adapters, nil
}
