package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/pierreWagou/lore/internal/manifest"
)

// HarnessConfig holds per-harness overrides set by the user.
// Used both at the top-level [harness.<name>] and within profiles [profile.<p>.harness.<name>].
// Supports ~ expansion in SkillsDir.
type HarnessConfig struct {
	// SkillsDir overrides the global skills directory for this harness.
	// Leave empty to use the harness default.
	SkillsDir string `toml:"skills_dir"`
}

// Profile is a named set of harness overrides and skill dependencies in the global lore.toml.
//
// Example lore.toml:
//
//	[profile.alan]
//	harnesses = ["opencode"]
//
//	[profile.alan.harness.opencode]
//	skills_dir = "~/.config/opencode-alan/skills"
//
//	[[profile.alan.dependencies]]
//	name   = "standup"
//	source = "alan-eu/alan-skills/skills/standup"
//	ref    = "main"
type Profile struct {
	// Harnesses limits which harnesses are active when this profile is used.
	// If empty, harness resolution falls back to auto-detect.
	Harnesses []string `toml:"harnesses,omitempty"`
	// Harness holds per-harness overrides for this profile.
	Harness map[string]HarnessConfig `toml:"harness,omitempty"`
	// Dependencies lists the skills installed under this profile.
	Dependencies []manifest.Dependency `toml:"dependencies,omitempty"`
}

// Config is the global lore configuration (~/.config/lore/lore.toml).
// Project-scoped lore.toml files use manifest.Manifest instead.
type Config struct {
	DefaultProfile string                   `toml:"default_profile"`
	Harness        map[string]HarnessConfig `toml:"harness"`
	Profiles       map[string]Profile       `toml:"profile"`
}

// LoreConfigDir returns the lore config directory.
// LORE_CONFIG_DIR overrides the default for testing and custom setups.
func LoreConfigDir() string {
	return configDir()
}

func configDir() string {
	if override := os.Getenv("LORE_CONFIG_DIR"); override != "" {
		return override
	}
	return filepath.Join(XDGConfigHome(), "lore")
}

// configPath returns the full path to the global lore.toml.
func configPath() string {
	return filepath.Join(configDir(), "lore.toml")
}

// Load reads the global lore.toml. Returns an empty Config if the file does not exist.
func Load() (*Config, error) {
	c := &Config{}
	path := configPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return c, nil
	}
	if _, err := toml.DecodeFile(path, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Save writes the global config back to lore.toml, creating parent directories as needed.
func Save(c *Config) error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(c)
}

// SkillsDirOverride returns the user-configured global skills directory for harnessName
// from the top-level [harness.<name>] section (not profile-specific).
// Returns "" if no override is set. Expands ~ in paths.
// Errors reading the config are silently ignored — callers fall back to defaults.
func SkillsDirOverride(harnessName string) string {
	c, err := Load()
	if err != nil || c.Harness == nil {
		return ""
	}
	hc, ok := c.Harness[harnessName]
	if !ok || hc.SkillsDir == "" {
		return ""
	}
	return ExpandHome(hc.SkillsDir)
}

// ResolveProfile returns the named profile from the config.
// Returns (nil, nil) when name is "" or the profile does not exist.
func ResolveProfile(name string) (*Profile, error) {
	if name == "" {
		return nil, nil
	}
	c, err := Load()
	if err != nil {
		return nil, err
	}
	p, ok := c.Profiles[name]
	if !ok {
		return nil, nil
	}
	return &p, nil
}

// ResolveProfileFromConfig returns the named profile from an already-loaded Config.
// Returns nil when name is "" or the profile does not exist.
func ResolveProfileFromConfig(c *Config, name string) *Profile {
	if name == "" || c.Profiles == nil {
		return nil
	}
	p, ok := c.Profiles[name]
	if !ok {
		return nil
	}
	return &p
}

// ActiveProfileName returns the profile name to use for a global install when no
// explicit --profile flag was given. Resolution order:
//  1. default_profile from config (explicit)
//  2. the sole profile name when exactly one profile is defined (implicit)
//  3. "" — no active profile
func ActiveProfileName() string {
	c, err := Load()
	if err != nil {
		return ""
	}
	return activeProfileNameFromConfig(c)
}

// ActiveProfileNameFromConfig is the same as ActiveProfileName but operates on an
// already-loaded Config, avoiding a second disk read.
func ActiveProfileNameFromConfig(c *Config) string {
	return activeProfileNameFromConfig(c)
}

func activeProfileNameFromConfig(c *Config) string {
	if c.DefaultProfile != "" {
		return c.DefaultProfile
	}
	if len(c.Profiles) == 1 {
		for name := range c.Profiles {
			return name
		}
	}
	return ""
}

// AddDependency adds or replaces a dependency in the named profile.
// Creates the profile if it does not exist.
func AddDependency(c *Config, profileName string, dep manifest.Dependency) {
	if c.Profiles == nil {
		c.Profiles = make(map[string]Profile)
	}
	p := c.Profiles[profileName]
	for i, d := range p.Dependencies {
		if d.Name == dep.Name {
			p.Dependencies[i] = dep
			c.Profiles[profileName] = p
			return
		}
	}
	p.Dependencies = append(p.Dependencies, dep)
	c.Profiles[profileName] = p
}

// RemoveDependency removes a dependency from the named profile. Returns true if found.
func RemoveDependency(c *Config, profileName, name string) bool {
	if c.Profiles == nil {
		return false
	}
	p, ok := c.Profiles[profileName]
	if !ok {
		return false
	}
	for i, d := range p.Dependencies {
		if d.Name == name {
			p.Dependencies = append(p.Dependencies[:i], p.Dependencies[i+1:]...)
			c.Profiles[profileName] = p
			return true
		}
	}
	return false
}

// HasDependency reports whether a dependency with the given name exists in the named profile.
func HasDependency(c *Config, profileName, name string) bool {
	if c.Profiles == nil {
		return false
	}
	p, ok := c.Profiles[profileName]
	if !ok {
		return false
	}
	for _, d := range p.Dependencies {
		if d.Name == name {
			return true
		}
	}
	return false
}

// ExpandHome replaces a leading ~ with the user's home directory.
func ExpandHome(path string) string {
	if !strings.HasPrefix(path, "~/") && path != "~" {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[2:])
}

// XDGConfigHome returns the XDG config home directory.
// Respects $XDG_CONFIG_HOME; falls back to ~/.config on all platforms.
// Unlike os.UserConfigDir(), this returns ~/.config on macOS rather than
// ~/Library/Application Support/, which is what most developer tools expect.
func XDGConfigHome() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config")
}
