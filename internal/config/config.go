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

// Config is the lore user configuration (~/.config/lore/config.toml).
type Config struct {
	Harness map[string]HarnessConfig `toml:"harness"`
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
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "lore")
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
	return expandHome(hc.SkillsDir)
}

// expandHome replaces a leading ~ with the user's home directory.
func expandHome(path string) string {
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
