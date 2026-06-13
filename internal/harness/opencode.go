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
	return filepath.Join(config.XDGConfigHome(), "opencode", "skills")
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
	_, err := os.Stat(filepath.Join(config.XDGConfigHome(), "opencode"))
	return err == nil
}
