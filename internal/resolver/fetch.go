package resolver

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
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
		if err != nil || info.IsDir() {
			return err
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

// --- go-git helpers kept for future bare-server fallback ---

func fetchGoGit(h Handle, auth transport.AuthMethod, cacheDir string) (FetchResult, error) {
	repoDir := filepath.Join(cacheDir, "repos", h.Host, h.Owner, h.RepoName)

	repo, err := openOrClone(repoDir, h, auth)
	if err != nil {
		return FetchResult{}, fmt.Errorf("fetching %s: %w", h.RepoURL, err)
	}

	hash, err := resolveRef(repo, h.Ref)
	if err != nil {
		return FetchResult{}, fmt.Errorf("resolving ref %q in %s: %w", h.Ref, h.RepoURL, err)
	}

	commit, err := repo.CommitObject(hash)
	if err != nil {
		return FetchResult{}, fmt.Errorf("reading commit %s: %w", hash, err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return FetchResult{}, err
	}

	result := FetchResult{
		Files:  make(map[string][]byte),
		Commit: hash.String(),
	}

	if h.SubPath != "" {
		subtree, subErr := tree.Tree(h.SubPath)
		if subErr != nil {
			return FetchResult{}, fmt.Errorf("path %q not found in %s@%s: %w", h.SubPath, h.RepoURL, h.Ref, subErr)
		}
		tree = subtree
	}

	if err := tree.Files().ForEach(func(f *object.File) error {
		content, readErr := f.Contents()
		if readErr != nil {
			return readErr
		}
		result.Files[f.Name] = []byte(content)
		return nil
	}); err != nil {
		return FetchResult{}, err
	}

	return result, nil
}

func openOrClone(repoDir string, h Handle, auth transport.AuthMethod) (*git.Repository, error) {
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		return cloneRepo(h.RepoURL, repoDir, auth)
	}

	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		_ = os.RemoveAll(repoDir)
		return cloneRepo(h.RepoURL, repoDir, auth)
	}

	if !looksLikeSHA(h.Ref) {
		fetchErr := repo.Fetch(&git.FetchOptions{
			RemoteName: "origin",
			Auth:       auth,
			Force:      true,
		})
		if fetchErr != nil && fetchErr != git.NoErrAlreadyUpToDate {
			_ = fetchErr
		}
	}

	return repo, nil
}

func cloneRepo(repoURL, destDir string, auth transport.AuthMethod) (*git.Repository, error) {
	if err := os.MkdirAll(filepath.Dir(destDir), 0755); err != nil {
		return nil, err
	}
	return git.PlainClone(destDir, false, &git.CloneOptions{
		URL:  repoURL,
		Auth: auth,
	})
}

func resolveRef(repo *git.Repository, ref string) (plumbing.Hash, error) {
	if ref == "" || ref == "HEAD" {
		if remoteHead, err := repo.Reference(plumbing.NewRemoteReferenceName("origin", "HEAD"), true); err == nil {
			return remoteHead.Hash(), nil
		}
		head, err := repo.Head()
		if err != nil {
			return plumbing.ZeroHash, err
		}
		return head.Hash(), nil
	}

	if looksLikeSHA(ref) {
		hash := plumbing.NewHash(ref)
		if _, err := repo.CommitObject(hash); err != nil {
			return plumbing.ZeroHash, fmt.Errorf("commit %s not found: %w", ref, err)
		}
		return hash, nil
	}

	if branchRef, err := repo.Reference(plumbing.NewRemoteReferenceName("origin", ref), true); err == nil {
		return branchRef.Hash(), nil
	}

	if tagRef, err := repo.Tag(ref); err == nil {
		if tagObj, err := repo.TagObject(tagRef.Hash()); err == nil {
			return tagObj.Target, nil
		}
		return tagRef.Hash(), nil
	}

	return plumbing.ZeroHash, fmt.Errorf("ref %q not found as branch, tag, or commit SHA", ref)
}

func looksLikeSHA(ref string) bool {
	if len(ref) < 7 || len(ref) > 40 {
		return false
	}
	for _, c := range ref {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
