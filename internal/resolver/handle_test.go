package resolver_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pierreWagou/lore/internal/resolver"
)

func TestParseShorthand(t *testing.T) {
	h, err := resolver.Parse("owner/repo", "main")
	require.NoError(t, err)
	assert.Equal(t, resolver.KindShorthand, h.Kind)
	assert.Equal(t, "https://github.com/owner/repo", h.RepoURL)
	assert.Equal(t, "owner", h.Owner)
	assert.Equal(t, "repo", h.RepoName)
	assert.Equal(t, "", h.SubPath)
	assert.Equal(t, "main", h.Ref)
	assert.True(t, h.ScanAll())
}

func TestParseShorthandWithPath(t *testing.T) {
	h, err := resolver.Parse("owner/repo/path/to/skill", "")
	require.NoError(t, err)
	assert.Equal(t, resolver.KindShorthand, h.Kind)
	assert.Equal(t, "https://github.com/owner/repo", h.RepoURL)
	assert.Equal(t, "path/to/skill", h.SubPath)
	assert.Equal(t, "HEAD", h.Ref)
	assert.False(t, h.ScanAll())
}

func TestParseHTTPS(t *testing.T) {
	h, err := resolver.Parse("https://github.com/owner/repo/tree/v1.0.0/skills/my-skill", "")
	require.NoError(t, err)
	assert.Equal(t, resolver.KindHTTPS, h.Kind)
	assert.Equal(t, "https://github.com/owner/repo", h.RepoURL)
	assert.Equal(t, "v1.0.0", h.Ref)
	assert.Equal(t, "skills/my-skill", h.SubPath)
	assert.Equal(t, "github.com", h.Host)
}

func TestParseHTTPSBare(t *testing.T) {
	h, err := resolver.Parse("https://github.com/owner/repo", "")
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/owner/repo", h.RepoURL)
	assert.Equal(t, "", h.SubPath)
	assert.True(t, h.ScanAll())
}

func TestParseHTTPSWithDotGit(t *testing.T) {
	h, err := resolver.Parse("https://github.com/owner/repo.git", "")
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/owner/repo", h.RepoURL)
	assert.Equal(t, "repo", h.RepoName)
}

func TestParseSSH(t *testing.T) {
	h, err := resolver.Parse("git@github.com:owner/repo.git/path/to/skill", "main")
	require.NoError(t, err)
	assert.Equal(t, resolver.KindSSH, h.Kind)
	assert.Equal(t, "git@github.com:owner/repo.git", h.RepoURL)
	assert.Equal(t, "path/to/skill", h.SubPath)
	assert.Equal(t, "github.com", h.Host)
	assert.Equal(t, "owner", h.Owner)
	assert.Equal(t, "repo", h.RepoName)
}

func TestParseSSHNaked(t *testing.T) {
	h, err := resolver.Parse("git@github.com:owner/repo.git", "")
	require.NoError(t, err)
	assert.Equal(t, "git@github.com:owner/repo.git", h.RepoURL)
	assert.Equal(t, "", h.SubPath)
	assert.True(t, h.ScanAll())
}

func TestParseSSHNoGitSuffix(t *testing.T) {
	_, err := resolver.Parse("git@github.com:owner/repo/path", "")
	assert.Error(t, err)
}

func TestParseLocalRelative(t *testing.T) {
	h, err := resolver.Parse("./local/skill", "")
	require.NoError(t, err)
	assert.Equal(t, resolver.KindLocal, h.Kind)
	assert.NotEmpty(t, h.SubPath) // resolved to absolute
}

func TestParseLocalAbsolute(t *testing.T) {
	h, err := resolver.Parse("/absolute/skill", "")
	require.NoError(t, err)
	assert.Equal(t, resolver.KindLocal, h.Kind)
	assert.Equal(t, "/absolute/skill", h.SubPath)
}

func TestParseShorthandTooFewSegments(t *testing.T) {
	_, err := resolver.Parse("onlyone", "")
	assert.Error(t, err)
}

func TestRefOverridesEmbedded(t *testing.T) {
	h, err := resolver.Parse("https://github.com/owner/repo/tree/main/path", "v2.0.0")
	require.NoError(t, err)
	// Explicit ref from lore.toml overrides the ref in the URL.
	assert.Equal(t, "v2.0.0", h.Ref)
}
