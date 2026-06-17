package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// HarnessConfig holds per-harness overrides set by the user.
type HarnessConfig struct {
	// SkillsDir overrides the global skills directory for this harness.
	// Supports ~ expansion. Leave empty to use the harness default.
	SkillsDir string `toml:"skills_dir"`
}

// ProfileHarnessConfig holds per-harness overrides within a named profile.
type ProfileHarnessConfig struct {
	// SkillsDir overrides the global skills directory for this harness within the profile.
	// Supports ~ expansion.
	SkillsDir string `toml:"skills_dir"`
}

// Profile is a named set of harness overrides in the lore config.
//
// Example config.toml:
//
//	[profile.alan]
//	harnesses = ["opencode"]
//
//	[profile.alan.harness.opencode]
//	skills_dir = "~/.config/opencode-alan/skills"
type Profile struct {
	// Harnesses limits which harnesses are active when this profile is used.
	// If empty, harness resolution falls back to the manifest / auto-detect.
	Harnesses []string `toml:"harnesses,omitempty"`
	// Harness holds per-harness overrides for this profile.
	Harness map[string]ProfileHarnessConfig `toml:"harness,omitempty"`
}

// Config is the lore user configuration (~/.config/lore/config.toml).
type Config struct {
	DefaultProfile string                   `toml:"default_profile"`
	Harness        map[string]HarnessConfig `toml:"harness"`
	Profiles       map[string]Profile       `toml:"profile"`
}

// configDir returns the lore config directory, respecting LORE_CONFIG_DIR.
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

// configPath returns the full path to config.toml.
func configPath() string {
	return filepath.Join(configDir(), "config.toml")
}

// Load reads the lore config file. Returns an empty Config if the file does not exist.
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

// SkillsDirOverride returns the user-configured global skills directory for harnessName.
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

// ResolveProfile returns the named profile from the lore config.
// Returns (nil, nil) when name is "" or the profile does not exist.
// Errors loading the config are propagated to the caller.
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

// DefaultProfileName returns the default_profile value from the lore config.
// Returns "" if unset or the config cannot be read.
func DefaultProfileName() string {
	c, err := Load()
	if err != nil {
		return ""
	}
	return c.DefaultProfile
}

// ActiveProfileName returns the profile name that should be used for a global
// install when no explicit --profile flag was given. Resolution order:
//  1. default_profile from config.toml (explicit)
//  2. the sole profile name when exactly one profile is defined (implicit)
//  3. "" — no active profile
func ActiveProfileName() string {
	c, err := Load()
	if err != nil {
		return ""
	}
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
