package resolver

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// httpClient is used for all archive downloads. A timeout prevents hangs on slow/broken servers.
var httpClient = &http.Client{Timeout: 60 * time.Second}

type platform int

const (
	platformGitHubCompat platform = iota // github.com, Gitea, Forgejo — archive/{ref}.tar.gz
	platformGitLab                       // gitlab.com — /-/archive/{ref}/{repo}-{ref}.tar.gz
)

func detectPlatform(host string) platform {
	if host == "gitlab.com" || strings.Contains(host, "gitlab") {
		return platformGitLab
	}
	return platformGitHubCompat
}

// ArchiveURL returns the tarball download URL for a given Handle.
func ArchiveURL(h Handle) string {
	ref := h.Ref
	if ref == "" || ref == "HEAD" {
		ref = "HEAD"
	}
	switch detectPlatform(h.Host) {
	case platformGitLab:
		return fmt.Sprintf("https://%s/%s/%s/-/archive/%s/%s-%s.tar.gz",
			h.Host, h.Owner, h.RepoName, ref, h.RepoName, ref)
	default:
		return fmt.Sprintf("https://%s/%s/%s/archive/%s.tar.gz",
			h.Host, h.Owner, h.RepoName, ref)
	}
}

// fetchArchive downloads a repo archive, extracts it (optionally filtered to h.SubPath),
// and caches the result by commit SHA.
func fetchArchive(h Handle, token, cacheDir string) (FetchResult, error) {
	return fetchArchiveWithBase(h, token, cacheDir, "")
}

// fetchArchiveWithBase is the testable version of fetchArchive.
// baseURL overrides the scheme+host of the archive URL (used in tests with mock servers).
func fetchArchiveWithBase(h Handle, token, cacheDir, baseURL string) (FetchResult, error) {
	aURL := ArchiveURL(h)
	if baseURL != "" {
		aURL = rebaseURL(aURL, baseURL)
	}

	// Step 1: HEAD to resolve SHA from redirect (best-effort, no body download).
	sha := resolveSHAViaHead(aURL, token)

	// Step 2: Cache hit by SHA.
	if sha != "" {
		if cached, ok := loadFromCache(cacheDir, h.Host, h.Owner, h.RepoName, sha); ok {
			return FetchResult{
				Files:  filterBySubPath(cached, h.SubPath),
				Commit: sha,
			}, nil
		}
	}

	// Step 3: Download the archive.
	allFiles, finalURL, err := downloadAndExtract(aURL, token)
	if err != nil {
		return FetchResult{}, err
	}

	// Refine SHA from GET response final URL if HEAD didn't give us one.
	if sha == "" {
		sha = extractSHAFromURL(finalURL)
	}

	// Step 4: Cache by SHA.
	if sha != "" {
		_ = saveToCache(cacheDir, h.Host, h.Owner, h.RepoName, sha, allFiles)
	}

	return FetchResult{
		Files:  filterBySubPath(allFiles, h.SubPath),
		Commit: sha,
	}, nil
}

// resolveSHAViaHead sends a HEAD request and extracts the commit SHA from the// final redirect URL. Returns "" if SHA cannot be determined.
func resolveSHAViaHead(archiveURL, token string) string {
	req, err := http.NewRequest(http.MethodHead, archiveURL, nil)
	if err != nil {
		return ""
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return ""
	}
	resp.Body.Close()
	if resp.Request != nil {
		return extractSHAFromURL(resp.Request.URL.String())
	}
	return ""
}

// downloadAndExtract GETs the archive URL and returns all extracted files plus
// the final URL (used to derive the commit SHA).
func downloadAndExtract(archiveURL, token string) (map[string][]byte, string, error) {
	req, err := http.NewRequest(http.MethodGet, archiveURL, nil)
	if err != nil {
		return nil, "", err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("fetching archive: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		host := ""
		if resp.Request != nil {
			host = resp.Request.URL.Hostname()
		}
		return nil, "", fmt.Errorf("authentication failed (HTTP %d); use `lore auth add %s <token>` or set a token env var",
			resp.StatusCode, host)
	case http.StatusNotFound:
		return nil, "", fmt.Errorf("archive not found at %s", archiveURL)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected HTTP %d from %s", resp.StatusCode, archiveURL)
	}

	finalURL := ""
	if resp.Request != nil {
		finalURL = resp.Request.URL.String()
	}

	files, err := ExtractTarGz(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("extracting archive: %w", err)
	}
	return files, finalURL, nil
}

// ExtractTarGz reads a .tar.gz stream and returns all regular files as a map
// of path → content. The first path component (repo root prefix) is stripped.
func ExtractTarGz(r io.Reader) (map[string][]byte, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("reading gzip stream: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	files := make(map[string][]byte)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading tar entry: %w", err)
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}

		// Strip root prefix: "{repo}-{sha}/" or "{repo}-{ref}/"
		name := header.Name
		if idx := strings.Index(name, "/"); idx != -1 {
			name = name[idx+1:]
		} else {
			continue
		}
		if name == "" {
			continue
		}

		content, err := io.ReadAll(tr)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", header.Name, err)
		}
		files[name] = content
	}
	return files, nil
}

// FilterBySubPath filters a file map to files under subPath and strips the prefix.
// Returns the original map unchanged when subPath is empty.
func FilterBySubPath(files map[string][]byte, subPath string) map[string][]byte {
	return filterBySubPath(files, subPath)
}

func filterBySubPath(files map[string][]byte, subPath string) map[string][]byte {
	if subPath == "" {
		return files
	}
	prefix := strings.TrimSuffix(subPath, "/") + "/"
	result := make(map[string][]byte)
	for path, content := range files {
		if strings.HasPrefix(path, prefix) {
			result[strings.TrimPrefix(path, prefix)] = content
		}
	}
	return result
}

// ExtractSHAFromURL looks for a 40-char hex SHA in the path segments of rawURL.
// Works for GitHub's codeload.github.com redirect URLs.
func ExtractSHAFromURL(rawURL string) string {
	return extractSHAFromURL(rawURL)
}

func extractSHAFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	for _, part := range strings.Split(strings.Trim(u.Path, "/"), "/") {
		if len(part) == 40 && isHexString(part) {
			return part
		}
	}
	return ""
}

func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// rebaseURL replaces the scheme and host of src with those from base.
// Used in tests to redirect archive requests to a mock server.
func rebaseURL(src, base string) string {
	srcU, err := url.Parse(src)
	if err != nil {
		return src
	}
	baseU, err := url.Parse(base)
	if err != nil {
		return src
	}
	srcU.Scheme = baseU.Scheme
	srcU.Host = baseU.Host
	return srcU.String()
}

// --- Cache ---

func archiveCacheDir(cacheDir, host, owner, repo, sha string) string {
	return filepath.Join(cacheDir, "archives", host, owner, repo, sha)
}

func loadFromCache(cacheDir, host, owner, repo, sha string) (map[string][]byte, bool) {
	dir := archiveCacheDir(cacheDir, host, owner, repo, sha)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, false
	}
	files := make(map[string][]byte)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		files[filepath.ToSlash(rel)] = content
		return nil
	})
	if err != nil {
		return nil, false
	}
	return files, true
}

func saveToCache(cacheDir, host, owner, repo, sha string, files map[string][]byte) error {
	dir := archiveCacheDir(cacheDir, host, owner, repo, sha)
	for path, content := range files {
		dest := filepath.Join(dir, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(dest, content, 0644); err != nil {
			return err
		}
	}
	return nil
}
