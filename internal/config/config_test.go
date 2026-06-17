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

func TestDefaultProfileNameAbsent(t *testing.T) {
	t.Setenv("LORE_CONFIG_DIR", t.TempDir())
	assert.Equal(t, "", config.DefaultProfileName())
}

func TestDefaultProfileNamePresent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LORE_CONFIG_DIR", dir)

	content := "default_profile = \"alan\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0644))

	assert.Equal(t, "alan", config.DefaultProfileName())
}

func TestActiveProfileNameNoConfig(t *testing.T) {
	t.Setenv("LORE_CONFIG_DIR", t.TempDir())
	assert.Equal(t, "", config.ActiveProfileName())
}

func TestActiveProfileNameExplicitDefault(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LORE_CONFIG_DIR", dir)

	content := "default_profile = \"alan\"\n\n[profile.alan]\nharnesses = [\"opencode\"]\n\n[profile.other]\nharnesses = [\"claude\"]\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0644))

	// explicit default_profile wins even when multiple profiles exist
	assert.Equal(t, "alan", config.ActiveProfileName())
}

func TestActiveProfileNameSingleProfile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LORE_CONFIG_DIR", dir)

	content := "[profile.solo]\nharnesses = [\"opencode\"]\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0644))

	// only one profile defined and no default_profile — auto-select it
	assert.Equal(t, "solo", config.ActiveProfileName())
}

func TestActiveProfileNameMultipleNoDefault(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LORE_CONFIG_DIR", dir)

	content := "[profile.alpha]\nharnesses = [\"opencode\"]\n\n[profile.beta]\nharnesses = [\"claude\"]\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0644))

	// multiple profiles, no default — cannot auto-select
	assert.Equal(t, "", config.ActiveProfileName())
}

func TestResolveProfileNotFound(t *testing.T) {
	t.Setenv("LORE_CONFIG_DIR", t.TempDir())
	p, err := config.ResolveProfile("nonexistent")
	require.NoError(t, err)
	assert.Nil(t, p)
}

func TestResolveProfileEmpty(t *testing.T) {
	t.Setenv("LORE_CONFIG_DIR", t.TempDir())
	p, err := config.ResolveProfile("")
	require.NoError(t, err)
	assert.Nil(t, p)
}

func TestResolveProfileFound(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LORE_CONFIG_DIR", dir)

	content := `
[profile.alan]
harnesses = ["opencode"]

[profile.alan.harness.opencode]
skills_dir = "/custom/alan/skills"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0644))

	p, err := config.ResolveProfile("alan")
	require.NoError(t, err)
	require.NotNil(t, p)
	assert.Equal(t, []string{"opencode"}, p.Harnesses)
	assert.Equal(t, "/custom/alan/skills", p.Harness["opencode"].SkillsDir)
}

func TestResolveProfileTildeExpansion(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LORE_CONFIG_DIR", dir)

	content := `
[profile.myprofile.harness.opencode]
skills_dir = "~/.config/opencode-myprofile/skills"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0644))

	p, err := config.ResolveProfile("myprofile")
	require.NoError(t, err)
	require.NotNil(t, p)
	// Raw value stored in struct — ExpandHome is applied by the installer when consuming the profile.
	assert.Equal(t, "~/.config/opencode-myprofile/skills", p.Harness["opencode"].SkillsDir)
}
