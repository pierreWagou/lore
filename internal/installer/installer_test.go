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

	// Create a temp dir to act as the harness root.
	harnessRoot := t.TempDir()

	dep := manifest.Dependency{
		Name:   "my-skill",
		Source: skillDir, // local path
		Ref:    "",
	}
	opts := installer.Options{
		Global:  false,
		Targets: []string{"opencode"},
		Root:    harnessRoot,
	}

	sr, err := installer.Install(dep, opts, nil)
	require.NoError(t, err)
	require.Len(t, sr.Results, 1)
	assert.Equal(t, "opencode", sr.Results[0].Harness)
	assert.Equal(t, "my-skill", sr.Results[0].Name)

	// The SKILL.md must exist in the harness skills directory.
	installedPath := filepath.Join(harnessRoot, ".opencode", "skills", "my-skill", "SKILL.md")
	data, err := os.ReadFile(installedPath)
	require.NoError(t, err)
	assert.Equal(t, skillContent, data)

	// Content hash must be set.
	assert.True(t, strings.HasPrefix(sr.ContentHash, "sha256:"))
}

func TestRemoveDeletesSkillDirs(t *testing.T) {
	// Pre-create skill directories for two harnesses.
	root := t.TempDir()
	for _, harness := range []string{"opencode", "claude"} {
		var skillsDir string
		switch harness {
		case "opencode":
			skillsDir = filepath.Join(root, ".opencode", "skills", "my-skill")
		case "claude":
			skillsDir = filepath.Join(root, ".claude", "skills", "my-skill")
		}
		require.NoError(t, os.MkdirAll(skillsDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte("# skill"), 0644))
	}

	opts := installer.Options{
		Global:  false,
		Targets: []string{"opencode", "claude"},
		Root:    root,
	}

	require.NoError(t, installer.Remove("my-skill", opts, nil))

	// Both skill directories must be gone.
	for _, dir := range []string{
		filepath.Join(root, ".opencode", "skills", "my-skill"),
		filepath.Join(root, ".claude", "skills", "my-skill"),
	} {
		_, err := os.Stat(dir)
		assert.True(t, os.IsNotExist(err), "expected %s to be removed", dir)
	}
}
