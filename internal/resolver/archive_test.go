package resolver_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pierreWagou/lore/internal/resolver"
)

// makeTarGz builds an in-memory .tar.gz with the given files.
// Keys are file paths INCLUDING the root prefix (e.g. "repo-sha/SKILL.md").
func makeTarGz(files map[string]string) []byte {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	for path, content := range files {
		hdr := &tar.Header{
			Name:     path,
			Mode:     0644,
			Size:     int64(len(content)),
			Typeflag: tar.TypeReg,
		}
		_ = tw.WriteHeader(hdr)
		_, _ = tw.Write([]byte(content))
	}
	_ = tw.Close()
	_ = gw.Close()
	return buf.Bytes()
}

// --- ArchiveURL tests ---

func TestArchiveURLGitHub(t *testing.T) {
	h := resolver.Handle{
		Kind:     resolver.KindShorthand,
		Host:     "github.com",
		Owner:    "owner",
		RepoName: "repo",
		Ref:      "main",
	}
	want := "https://github.com/owner/repo/archive/main.tar.gz"
	assert.Equal(t, want, resolver.ArchiveURL(h))
}

func TestArchiveURLGitHubHEAD(t *testing.T) {
	h := resolver.Handle{
		Kind:     resolver.KindShorthand,
		Host:     "github.com",
		Owner:    "owner",
		RepoName: "repo",
		Ref:      "HEAD",
	}
	assert.Contains(t, resolver.ArchiveURL(h), "HEAD.tar.gz")
}

func TestArchiveURLGitLab(t *testing.T) {
	h := resolver.Handle{
		Kind:     resolver.KindHTTPS,
		Host:     "gitlab.com",
		Owner:    "owner",
		RepoName: "repo",
		Ref:      "main",
	}
	url := resolver.ArchiveURL(h)
	assert.Contains(t, url, "/-/archive/main/")
	assert.True(t, strings.HasSuffix(url, ".tar.gz"))
}

func TestArchiveURLGiteaCompatible(t *testing.T) {
	h := resolver.Handle{
		Kind:     resolver.KindHTTPS,
		Host:     "git.company.com",
		Owner:    "owner",
		RepoName: "repo",
		Ref:      "v1.0.0",
	}
	want := "https://git.company.com/owner/repo/archive/v1.0.0.tar.gz"
	assert.Equal(t, want, resolver.ArchiveURL(h))
}

// --- ExtractTarGz tests ---

func TestExtractTarGzBasic(t *testing.T) {
	archive := makeTarGz(map[string]string{
		"repo-abc123/SKILL.md":     "# skill",
		"repo-abc123/README.md":    "readme",
		"repo-abc123/sub/extra.md": "extra",
	})

	files, err := resolver.ExtractTarGz(bytes.NewReader(archive))
	require.NoError(t, err)
	assert.Equal(t, "# skill", string(files["SKILL.md"]))
	assert.Equal(t, "readme", string(files["README.md"]))
	assert.Equal(t, "extra", string(files["sub/extra.md"]))
}

func TestExtractTarGzStripsRootPrefix(t *testing.T) {
	archive := makeTarGz(map[string]string{
		"my-repo-deadbeef1234567890abcdef/SKILL.md": "content",
	})
	files, err := resolver.ExtractTarGz(bytes.NewReader(archive))
	require.NoError(t, err)
	_, withPrefix := files["my-repo-deadbeef1234567890abcdef/SKILL.md"]
	_, withoutPrefix := files["SKILL.md"]
	assert.False(t, withPrefix, "root prefix should be stripped")
	assert.True(t, withoutPrefix, "SKILL.md should be accessible without prefix")
}

// --- FilterBySubPath tests ---

func TestFilterBySubPathEmpty(t *testing.T) {
	files := map[string][]byte{
		"a/SKILL.md": []byte("a"),
		"b/SKILL.md": []byte("b"),
	}
	result := resolver.FilterBySubPath(files, "")
	assert.Len(t, result, 2)
}

func TestFilterBySubPathFilters(t *testing.T) {
	files := map[string][]byte{
		".ai/skills/foo/SKILL.md": []byte("foo"),
		".ai/skills/bar/SKILL.md": []byte("bar"),
		"README.md":               []byte("readme"),
	}
	result := resolver.FilterBySubPath(files, ".ai/skills/foo")
	assert.Len(t, result, 1)
	assert.Equal(t, []byte("foo"), result["SKILL.md"])
}

// --- extractSHAFromURL tests ---

func TestExtractSHAFromGitHubCDNURL(t *testing.T) {
	cdnURL := "https://codeload.github.com/owner/repo/tar.gz/a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6abcd"
	sha := resolver.ExtractSHAFromURL(cdnURL)
	assert.Equal(t, "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6abcd", sha)
}

func TestExtractSHAFromURLNoSHA(t *testing.T) {
	assert.Equal(t, "", resolver.ExtractSHAFromURL("https://example.com/owner/repo/archive/main.tar.gz"))
	assert.Equal(t, "", resolver.ExtractSHAFromURL("not-a-url"))
}

// --- HTTP integration: Fetch via mock server ---

func TestFetchArchiveWithMockServer(t *testing.T) {
	sha := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6abcd"

	archive := makeTarGz(map[string]string{
		"repo-" + sha + "/.ai/skills/my-skill/SKILL.md": "# my skill",
		"repo-" + sha + "/README.md":                    "readme",
	})

	// Mock server that redirects /archive/HEAD.tar.gz → /tar.gz/<sha>
	// then serves the archive at /tar.gz/<sha>.
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/owner/repo/archive/HEAD.tar.gz" {
			http.Redirect(w, r, srv.URL+"/tar.gz/"+sha, http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", "application/x-gzip")
		w.Write(archive)
	}))
	defer srv.Close()

	h := resolver.Handle{
		Kind:     resolver.KindShorthand,
		Host:     srv.Listener.Addr().String(),
		Owner:    "owner",
		RepoName: "repo",
		SubPath:  ".ai/skills/my-skill",
		Ref:      "HEAD",
	}

	result, err := resolver.FetchArchiveForTest(h, "", t.TempDir(), srv.URL)
	require.NoError(t, err)
	assert.Equal(t, sha, result.Commit)
	assert.Equal(t, []byte("# my skill"), result.Files["SKILL.md"])
	_, hasReadme := result.Files["README.md"]
	assert.False(t, hasReadme, "README.md should be filtered out by SubPath")
}

func TestFetchArchive404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	h := resolver.Handle{
		Kind:     resolver.KindShorthand,
		Host:     srv.Listener.Addr().String(),
		Owner:    "owner",
		RepoName: "repo",
		Ref:      "HEAD",
	}
	_, err := resolver.FetchArchiveForTest(h, "", t.TempDir(), srv.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFetchArchive401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	h := resolver.Handle{
		Kind:     resolver.KindShorthand,
		Host:     srv.Listener.Addr().String(),
		Owner:    "owner",
		RepoName: "repo",
		Ref:      "HEAD",
	}
	_, err := resolver.FetchArchiveForTest(h, "", t.TempDir(), srv.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")
}
