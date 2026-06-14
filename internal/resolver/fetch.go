package resolver

import (
	"os"
	"path/filepath"
)

// FetchResult holds the fetched files and the resolved commit SHA.
type FetchResult struct {
	// Files maps paths relative to the skill root (SubPath stripped) to content.
	Files  map[string][]byte
	Commit string
}

// Fetch resolves and downloads the skill referenced by h.
// token is used for HTTPS authentication (pass "" for public repos).
// cacheDir is the root directory for cached archives (e.g. ~/.cache/lore).
func Fetch(h Handle, token, cacheDir string) (FetchResult, error) {
	if h.Kind == KindLocal {
		return fetchLocal(h)
	}
	return fetchArchive(h, token, cacheDir)
}

func fetchLocal(h Handle) (FetchResult, error) {
	result := FetchResult{Files: make(map[string][]byte)}
	err := filepath.Walk(h.SubPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		rel, _ := filepath.Rel(h.SubPath, path)
		result.Files[rel] = data
		return nil
	})
	return result, err
}
