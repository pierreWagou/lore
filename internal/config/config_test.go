package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pierreWagou/lore/internal/config"
)

func TestLoadMissing(t *testing.T) {
	t.Setenv("LORE_CONFIG_DIR", t.TempDir())
	c, err := config.Load()
	require.NoError(t, err)
	assert.Nil(t, c.Harness)
}

func TestLoadWithOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LORE_CONFIG_DIR", dir)

	content := `
[harness.claude]
skills_dir = "~/.claude/skills"

[harness.opencode]
skills_dir = "/custom/path"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0644))

	c, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "~/.claude/skills", c.Harness["claude"].SkillsDir)
	assert.Equal(t, "/custom/path", c.Harness["opencode"].SkillsDir)
}

func TestSkillsDirOverrideAbsent(t *testing.T) {
	t.Setenv("LORE_CONFIG_DIR", t.TempDir())
	assert.Equal(t, "", config.SkillsDirOverride("opencode"))
}

func TestSkillsDirOverridePresent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LORE_CONFIG_DIR", dir)

	content := "[harness.claude]\nskills_dir = \"/custom/claude\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0644))

	assert.Equal(t, "/custom/claude", config.SkillsDirOverride("claude"))
	assert.Equal(t, "", config.SkillsDirOverride("opencode"))
}

func TestSkillsDirOverrideTildeExpansion(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LORE_CONFIG_DIR", dir)

	content := "[harness.myharness]\nskills_dir = \"~/.config/myharness/skills\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0644))

	result := config.SkillsDirOverride("myharness")
	assert.True(t, filepath.IsAbs(result), "expanded path should be absolute")
	assert.NotContains(t, result, "~")
}
