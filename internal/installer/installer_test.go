package installer_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pierreWagou/lore/internal/installer"
	"github.com/pierreWagou/lore/internal/manifest"
)

func TestComputeContentHash(t *testing.T) {
	files := map[string][]byte{
		"SKILL.md": []byte("# skill content"),
		"extra.md": []byte("extra"),
	}

	h1 := installer.ComputeContentHash(files)
	h2 := installer.ComputeContentHash(files)
	assert.Equal(t, h1, h2, "same files must produce same hash")
	assert.True(t, strings.HasPrefix(h1, "sha256:"), "hash must start with sha256:")

	files2 := map[string][]byte{
		"SKILL.md": []byte("# different content"),
	}
	h3 := installer.ComputeContentHash(files2)
	assert.NotEqual(t, h1, h3)
}

func TestComputeContentHashOrdering(t *testing.T) {
	files := map[string][]byte{
		"z.md": []byte("z"),
		"a.md": []byte("a"),
		"m.md": []byte("m"),
	}
	h1 := installer.ComputeContentHash(files)

	files2 := make(map[string][]byte)
	for k, v := range files {
		files2[k] = v
	}
	h2 := installer.ComputeContentHash(files2)
	assert.Equal(t, h1, h2)
}

func TestInstallLocal(t *testing.T) {
	// Create a local skill directory with a SKILL.md.
	skillDir := t.TempDir()
	skillContent := []byte("# my local skill\n\nDoes something useful.")
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), skillContent, 0644))

	// Project root where lore installs to.
	root := t.TempDir()

	dep := manifest.Dependency{
		Name:   "my-skill",
		Source: skillDir, // local path
		Ref:    "",
	}
	opts := installer.Options{
		Global:    false,
		Harnesses: []string{"opencode"},
		Root:      root,
	}

	sr, err := installer.Install(dep, opts, nil)
	require.NoError(t, err)
	require.Len(t, sr.Results, 1)
	assert.Equal(t, "opencode", sr.Results[0].Harness)
	assert.Equal(t, "my-skill", sr.Results[0].Name)

	// The neutral store must contain SKILL.md.
	neutralPath := filepath.Join(root, ".ai", "skills", "my-skill", "SKILL.md")
	data, err := os.ReadFile(neutralPath)
	require.NoError(t, err)
	assert.Equal(t, skillContent, data)
}

func TestInstallGlobalSingleProfileAutoSelect(t *testing.T) {
	// Config has exactly one profile and no default_profile — it should be auto-selected.
	cfgDir := t.TempDir()
	t.Setenv("LORE_CONFIG_DIR", cfgDir)

	autoSkillsDir := t.TempDir()
	cfgContent := "[profile.only]\nharnesses = [\"opencode\"]\n\n[profile.only.harness.opencode]\nskills_dir = \"" + autoSkillsDir + "\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(cfgDir, "lore.toml"), []byte(cfgContent), 0644))

	skillDir := t.TempDir()
	skillContent := []byte("# skill installed via default profile")
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), skillContent, 0644))

	dep := manifest.Dependency{
		Name:   "default-profile-skill",
		Source: skillDir,
		Ref:    "",
	}
	// No Profile set — should pick up default_profile from config.
	opts := installer.Options{
		Global:    true,
		Harnesses: []string{"opencode"}, // explicit to avoid auto-detect in CI
	}

	sr, err := installer.Install(dep, opts, nil)
	require.NoError(t, err)
	require.Len(t, sr.Results, 1)

	expectedPath := filepath.Join(autoSkillsDir, "default-profile-skill")
	assert.Equal(t, expectedPath, sr.Results[0].Path)

	data, err := os.ReadFile(filepath.Join(expectedPath, "SKILL.md"))
	require.NoError(t, err)
	assert.Equal(t, skillContent, data)
}

func TestInstallGlobalProfile(t *testing.T) {
	// Write a lore.toml with a named profile into a temp LORE_CONFIG_DIR.
	cfgDir := t.TempDir()
	t.Setenv("LORE_CONFIG_DIR", cfgDir)

	profileSkillsDir := t.TempDir()
	cfgContent := "[profile.test]\nharnesses = [\"opencode\"]\n\n[profile.test.harness.opencode]\nskills_dir = \"" + profileSkillsDir + "\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(cfgDir, "lore.toml"), []byte(cfgContent), 0644))

	// Create a local skill directory.
	skillDir := t.TempDir()
	skillContent := []byte("# global skill via --profile")
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), skillContent, 0644))

	dep := manifest.Dependency{
		Name:   "profile-skill",
		Source: skillDir,
		Ref:    "",
	}
	opts := installer.Options{
		Global:    true,
		Harnesses: []string{"opencode"}, // explicit to avoid auto-detect in CI
		Profile:   "test",
	}

	sr, err := installer.Install(dep, opts, nil)
	require.NoError(t, err)
	require.Len(t, sr.Results, 1)

	// Skill must land inside the profile's skills dir.
	expectedPath := filepath.Join(profileSkillsDir, "profile-skill")
	assert.Equal(t, expectedPath, sr.Results[0].Path)

	data, err := os.ReadFile(filepath.Join(expectedPath, "SKILL.md"))
	require.NoError(t, err)
	assert.Equal(t, skillContent, data)
}
