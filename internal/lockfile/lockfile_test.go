package lockfile_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pierreWagou/lore/internal/lockfile"
)

func TestLoadMissing(t *testing.T) {
	lf, err := lockfile.Load("/tmp/nonexistent/lore.lock")
	require.NoError(t, err)
	assert.Empty(t, lf.Entries)
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, lockfile.FileName)

	lf := &lockfile.Lockfile{
		Entries: []lockfile.Entry{
			lockfile.NewEntry("my-skill", "owner/repo/path", "abc123", "sha256:deadbeef"),
		},
	}
	require.NoError(t, lockfile.Save(path, lf))

	loaded, err := lockfile.Load(path)
	require.NoError(t, err)
	require.Len(t, loaded.Entries, 1)
	assert.Equal(t, "my-skill", loaded.Entries[0].Name)
	assert.Equal(t, "abc123", loaded.Entries[0].Commit)
}

func TestSetAndGetEntry(t *testing.T) {
	lf := &lockfile.Lockfile{}

	e1 := lockfile.NewEntry("skill-a", "src", "sha1", "hash1")
	lockfile.SetEntry(lf, e1)
	assert.Len(t, lf.Entries, 1)

	// Replace existing.
	e2 := lockfile.NewEntry("skill-a", "src", "sha2", "hash2")
	lockfile.SetEntry(lf, e2)
	assert.Len(t, lf.Entries, 1)
	assert.Equal(t, "sha2", lf.Entries[0].Commit)

	got := lockfile.GetEntry(lf, "skill-a")
	require.NotNil(t, got)
	assert.Equal(t, "sha2", got.Commit)

	assert.Nil(t, lockfile.GetEntry(lf, "nonexistent"))
}

func TestRemoveEntry(t *testing.T) {
	lf := &lockfile.Lockfile{
		Entries: []lockfile.Entry{
			{Name: "a"}, {Name: "b"},
		},
	}
	assert.True(t, lockfile.RemoveEntry(lf, "a"))
	assert.Len(t, lf.Entries, 1)
	assert.Equal(t, "b", lf.Entries[0].Name)
	assert.False(t, lockfile.RemoveEntry(lf, "nonexistent"))
}

func TestGlobalFileName(t *testing.T) {
	assert.Equal(t, "lore.lock", lockfile.GlobalFileName(""))
	assert.Equal(t, "lore.wagou.lock", lockfile.GlobalFileName("wagou"))
	assert.Equal(t, "lore.alan.lock", lockfile.GlobalFileName("alan"))
}
