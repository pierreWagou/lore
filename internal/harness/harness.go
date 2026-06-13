package harness

import (
	"fmt"
	"path/filepath"

	"github.com/pierreWagou/lore/internal/config"
)

// defaultGlobalSkillsDir returns the conventional global skills directory for a harness.
// Convention: $XDG_CONFIG_HOME/<harness-name>/skills/ (defaults to ~/.config/<name>/skills/).
// Adapters with a known non-standard location override this in their own GlobalSkillsDir().
func defaultGlobalSkillsDir(harnessName string) string {
	return filepath.Join(config.XDGConfigHome(), harnessName, "skills")
}

// File is a file to be written to a harness's skill directory.
type File struct {
	Path    string
	Content []byte
}

// Skill is a fetched skill ready to be installed.
type Skill struct {
	Name  string
	Files map[string][]byte // path relative to skill root → content
}

// Adapter adapts skills to a target harness's native format and directory layout.
type Adapter interface {
	Name() string
	GlobalSkillsDir() string
	ProjectSkillsDir(root string) string
	// Transform converts skill files to the harness's native format.
	Transform(skill Skill) ([]File, error)
	// NeedsTransform reports whether this harness requires content transformation.
	// When false, lore can use a symlink instead of copying for project-scope installs.
	NeedsTransform() bool
	Detect() bool
}

var registry = []Adapter{
	&OpenCode{},
	&Claude{},
}

// All returns all registered harness adapters.
func All() []Adapter {
	return registry
}

// Get returns the adapter with the given name, or nil if unknown.
func Get(name string) Adapter {
	for _, a := range registry {
		if a.Name() == name {
			return a
		}
	}
	return nil
}

// Detected returns all harness adapters whose Detect() returns true.
func Detected() []Adapter {
	var found []Adapter
	for _, a := range registry {
		if a.Detect() {
			found = append(found, a)
		}
	}
	return found
}

// Names returns the names of all registered adapters.
func Names() []string {
	names := make([]string, len(registry))
	for i, a := range registry {
		names[i] = a.Name()
	}
	return names
}

// passthroughTransform is the default transform for SKILL.md-native harnesses.
// Content is passed through as-is (including any frontmatter).
func passthroughTransform(skill Skill, harnessName string) ([]File, error) {
	content, ok := skill.Files["SKILL.md"]
	if !ok {
		return nil, fmt.Errorf("%s: SKILL.md not found in skill %q", harnessName, skill.Name)
	}
	return []File{{Path: "SKILL.md", Content: content}}, nil
}
