package harness

import (
	"os"
	"path/filepath"

	"github.com/pierreWagou/lore/internal/config"
)

// OpenCode is the harness adapter for the opencode agent.
type OpenCode struct{}

func (o *OpenCode) Name() string { return "opencode" }

func (o *OpenCode) GlobalSkillsDir() string {
	if override := config.SkillsDirOverride(o.Name()); override != "" {
		return override
	}
	// opencode's known location follows the ~/.config/<name>/skills/ convention.
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "opencode", "skills")
}

func (o *OpenCode) ProjectSkillsDir(root string) string {
	return filepath.Join(root, ".opencode", "skills")
}

func (o *OpenCode) Transform(skill Skill) ([]File, error) {
	return passthroughTransform(skill, o.Name())
}

// NeedsTransform returns false — opencode reads SKILL.md natively so a symlink suffices.
func (o *OpenCode) NeedsTransform() bool { return false }

func (o *OpenCode) Detect() bool {
	dir, err := os.UserConfigDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(dir, "opencode"))
	return err == nil
}
