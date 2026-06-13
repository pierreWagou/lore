package manifest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pierreWagou/lore/internal/manifest"
)

func TestLoadMissing(t *testing.T) {
	m, err := manifest.Load("/tmp/nonexistent/lore.toml")
	require.NoError(t, err)
	assert.Empty(t, m.Dependencies)
	assert.Empty(t, m.Targets)
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, manifest.FileName)

	m := &manifest.Manifest{
		Targets: []string{"opencode", "claude"},
		Dependencies: []manifest.Dependency{
			{Name: "my-skill", Source: "owner/repo/path", Ref: "main"},
		},
	}
	require.NoError(t, manifest.Save(path, m))

	loaded, err := manifest.Load(path)
	require.NoError(t, err)
	assert.Equal(t, m.Targets, loaded.Targets)
	assert.Equal(t, m.Dependencies, loaded.Dependencies)
}

func TestAddDependency(t *testing.T) {
	m := &manifest.Manifest{}
	dep := manifest.Dependency{Name: "skill-a", Source: "owner/repo/a", Ref: "v1"}
	manifest.AddDependency(m, dep)
	assert.Len(t, m.Dependencies, 1)

	// Replace existing.
	updated := manifest.Dependency{Name: "skill-a", Source: "owner/repo/a", Ref: "v2"}
	manifest.AddDependency(m, updated)
	assert.Len(t, m.Dependencies, 1)
	assert.Equal(t, "v2", m.Dependencies[0].Ref)
}

func TestRemoveDependency(t *testing.T) {
	m := &manifest.Manifest{
		Dependencies: []manifest.Dependency{
			{Name: "skill-a"},
			{Name: "skill-b"},
		},
	}
	assert.True(t, manifest.RemoveDependency(m, "skill-a"))
	assert.Len(t, m.Dependencies, 1)
	assert.Equal(t, "skill-b", m.Dependencies[0].Name)

	assert.False(t, manifest.RemoveDependency(m, "nonexistent"))
}

func TestAddTarget(t *testing.T) {
	m := &manifest.Manifest{}
	manifest.AddTarget(m, "opencode")
	manifest.AddTarget(m, "opencode") // duplicate
	manifest.AddTarget(m, "claude")
	assert.Equal(t, []string{"opencode", "claude"}, m.Targets)
}

func TestSaveCreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "dir", manifest.FileName)
	m := &manifest.Manifest{Targets: []string{"opencode"}}
	require.NoError(t, manifest.Save(path, m))
	_, err := os.Stat(path)
	assert.NoError(t, err)
}
