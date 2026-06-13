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
		Global:  false,
		Targets: []string{"opencode"},
		Root:    root,
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

	// The harness path must be a symlink pointing into the neutral store.
	harnessDest := filepath.Join(root, ".opencode", "skills", "my-skill")
	info, err := os.Lstat(harnessDest)
	require.NoError(t, err)
	assert.Equal(t, os.ModeSymlink, info.Mode()&os.ModeSymlink, "expected symlink at %s", harnessDest)

	// Reading through the symlink must yield the original content.
	linkedPath := filepath.Join(harnessDest, "SKILL.md")
	linked, err := os.ReadFile(linkedPath)
	require.NoError(t, err)
	assert.Equal(t, skillContent, linked)

	// Content hash must be set.
	assert.True(t, strings.HasPrefix(sr.ContentHash, "sha256:"))
}

func TestRemoveDeletesSkillDirs(t *testing.T) {
	root := t.TempDir()

	// Pre-populate neutral store and harness symlinks.
	neutralDir := filepath.Join(root, ".ai", "skills", "my-skill")
	require.NoError(t, os.MkdirAll(neutralDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(neutralDir, "SKILL.md"), []byte("# skill"), 0644))

	for _, harnessDir := range []string{
		filepath.Join(root, ".opencode", "skills"),
		filepath.Join(root, ".claude", "skills"),
	} {
		require.NoError(t, os.MkdirAll(harnessDir, 0755))
		// Simulate what placeProject creates: a symlink.
		rel, _ := filepath.Rel(harnessDir, neutralDir)
		require.NoError(t, os.Symlink(rel, filepath.Join(harnessDir, "my-skill")))
	}

	opts := installer.Options{
		Global:  false,
		Targets: []string{"opencode", "claude"},
		Root:    root,
	}

	require.NoError(t, installer.Remove("my-skill", opts, nil))

	// Harness symlinks must be gone.
	for _, dir := range []string{
		filepath.Join(root, ".opencode", "skills", "my-skill"),
		filepath.Join(root, ".claude", "skills", "my-skill"),
	} {
		_, err := os.Lstat(dir)
		assert.True(t, os.IsNotExist(err), "expected %s to be removed", dir)
	}

	// Neutral store must also be gone.
	_, err := os.Stat(neutralDir)
	assert.True(t, os.IsNotExist(err), "expected neutral store to be removed")
}
