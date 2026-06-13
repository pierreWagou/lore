package harness

import (
	"os"
	"path/filepath"
)

// Claude is the harness adapter for Claude's agent environment.
type Claude struct{}

func (c *Claude) Name() string { return "claude" }

func (c *Claude) GlobalSkillsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "skills")
}

func (c *Claude) ProjectSkillsDir(root string) string {
	return filepath.Join(root, ".claude", "skills")
}

func (c *Claude) Transform(skill Skill) ([]File, error) {
	return passthroughTransform(skill, c.Name())
}

func (c *Claude) Detect() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(home, ".claude"))
	return err == nil
}
