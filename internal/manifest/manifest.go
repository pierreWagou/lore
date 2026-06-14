package manifest

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const FileName = "lore.toml"

// Manifest represents the contents of a lore.toml file.
type Manifest struct {
	Mode         string       `toml:"mode,omitempty"` // "guest" | "keeper" (auto-detected on init)
	Harnesses    []string     `toml:"harnesses"`
	Dependencies []Dependency `toml:"dependencies"`
}

// Dependency is a single skill dependency entry.
type Dependency struct {
	Name   string `toml:"name"`
	Source string `toml:"source"`
	Ref    string `toml:"ref"`
}

// Load reads a manifest from path. Returns an empty manifest if the file does not exist.
func Load(path string) (*Manifest, error) {
	m := &Manifest{}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return m, nil
	}
	if _, err := toml.DecodeFile(path, m); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return m, nil
}

// Save writes the manifest to path, creating parent directories as needed.
func Save(path string, m *Manifest) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(m)
}

// AddDependency adds or replaces the dependency with the given name.
func AddDependency(m *Manifest, dep Dependency) {
	for i, d := range m.Dependencies {
		if d.Name == dep.Name {
			m.Dependencies[i] = dep
			return
		}
	}
	m.Dependencies = append(m.Dependencies, dep)
}

// RemoveDependency removes the dependency with the given name. Returns true if found.
func RemoveDependency(m *Manifest, name string) bool {
	for i, d := range m.Dependencies {
		if d.Name == name {
			m.Dependencies = append(m.Dependencies[:i], m.Dependencies[i+1:]...)
			return true
		}
	}
	return false
}

// HasDependency reports whether a dependency with the given name exists.
func HasDependency(m *Manifest, name string) bool {
	for _, d := range m.Dependencies {
		if d.Name == name {
			return true
		}
	}
	return false
}

// AddHarness adds a harness to the manifest if not already present.
func AddHarness(m *Manifest, harness string) {
	for _, h := range m.Harnesses {
		if h == harness {
			return
		}
	}
	m.Harnesses = append(m.Harnesses, harness)
}

// IsGuest reports whether the manifest is in guest mode.
// In guest mode lore adapts to existing harness dirs; .ai/skills/ is ephemeral (gitignored).
// In keeper mode (default) .ai/skills/ is the committed source of truth.
func IsGuest(m *Manifest) bool { return m.Mode == "guest" }
