package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pierreWagou/lore/internal/auth"
)

// isolate redirects auth's config dir to a temp directory for the duration of t.
func isolate(t *testing.T) {
	t.Helper()
	t.Setenv("LORE_CONFIG_DIR", t.TempDir())
}

func TestStoreAndLoadToken(t *testing.T) {
	isolate(t)

	require.NoError(t, auth.StoreToken("github.com", "ghp_test"))

	tokens, err := auth.ListTokens()
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	assert.Equal(t, "github.com", tokens[0].Host)
	assert.Equal(t, "ghp_test", tokens[0].Token)
}

func TestStoreTokenUpdatesExisting(t *testing.T) {
	isolate(t)

	require.NoError(t, auth.StoreToken("github.com", "old_token"))
	require.NoError(t, auth.StoreToken("github.com", "new_token"))

	tokens, err := auth.ListTokens()
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	assert.Equal(t, "new_token", tokens[0].Token)
}

func TestRemoveToken(t *testing.T) {
	isolate(t)

	require.NoError(t, auth.StoreToken("gitlab.com", "glpat_test"))
	require.NoError(t, auth.RemoveToken("gitlab.com"))

	tokens, err := auth.ListTokens()
	require.NoError(t, err)
	assert.Empty(t, tokens)
}

func TestRemoveTokenNonExistent(t *testing.T) {
	isolate(t)
	// Removing a host that was never added must not error.
	assert.NoError(t, auth.RemoveToken("nonexistent.com"))
}

func TestListTokensEmpty(t *testing.T) {
	isolate(t)

	tokens, err := auth.ListTokens()
	require.NoError(t, err)
	assert.Empty(t, tokens)
}

func TestResolveHTTPSLoreEnvVar(t *testing.T) {
	isolate(t)
	t.Setenv("LORE_GITHUB_COM_TOKEN", "lore_token")
	// Clear competing env vars so only LORE_GITHUB_COM_TOKEN fires.
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")

	authMethod, err := auth.Resolve("https://github.com/owner/repo")
	require.NoError(t, err)
	require.NotNil(t, authMethod)
}

func TestResolveHTTPSGitHubToken(t *testing.T) {
	isolate(t)
	t.Setenv("GITHUB_TOKEN", "github_token")
	t.Setenv("LORE_GITHUB_COM_TOKEN", "")
	t.Setenv("GH_TOKEN", "")

	authMethod, err := auth.Resolve("https://github.com/owner/repo")
	require.NoError(t, err)
	require.NotNil(t, authMethod)
}

func TestResolveHTTPSStoredToken(t *testing.T) {
	isolate(t)
	// Clear env vars so we fall through to stored credentials.
	t.Setenv("LORE_GITLAB_COM_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")

	require.NoError(t, auth.StoreToken("gitlab.com", "glpat_stored"))

	authMethod, err := auth.Resolve("https://gitlab.com/owner/repo")
	require.NoError(t, err)
	require.NotNil(t, authMethod)
}

func TestResolveHTTPSNoAuth(t *testing.T) {
	isolate(t)
	// No env vars, no stored tokens → public repo returns nil auth without error.
	t.Setenv("LORE_EXAMPLE_COM_TOKEN", "")

	authMethod, err := auth.Resolve("https://example.com/owner/repo")
	require.NoError(t, err)
	assert.Nil(t, authMethod)
}
