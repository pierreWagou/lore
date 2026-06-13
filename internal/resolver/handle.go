package resolver

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// HandleKind identifies the type of source handle.
type HandleKind int

const (
	KindShorthand HandleKind = iota // owner/repo[/path...]
	KindHTTPS                       // https://host/owner/repo[/tree/ref/path]
	KindSSH                         // git@host:owner/repo.git[/path]
	KindLocal                       // ./path or /abs/path
)

// Handle is a parsed source reference for a skill.
type Handle struct {
	Raw      string
	Kind     HandleKind
	RepoURL  string // full git clone URL (empty for local)
	SubPath  string // path within the repo or filesystem (empty = scan all)
	Ref      string // branch, tag, or SHA
	Host     string // git host (e.g. github.com)
	Owner    string
	RepoName string
}

// ScanAll reports whether the handle targets an entire repo (no explicit subpath).
func (h Handle) ScanAll() bool {
	return h.Kind != KindLocal && h.SubPath == ""
}

// Parse parses a source handle string into a Handle.
// ref overrides any ref embedded in the handle (e.g. from lore.toml).
func Parse(raw, ref string) (Handle, error) {
	h := Handle{Raw: raw, Ref: ref}
	if h.Ref == "" {
		h.Ref = "HEAD"
	}
	switch {
	case strings.HasPrefix(raw, "./"), strings.HasPrefix(raw, "../"), strings.HasPrefix(raw, "/"):
		return parseLocal(h, raw)
	case strings.HasPrefix(raw, "git@"):
		return parseSSH(h, raw)
	case strings.HasPrefix(raw, "https://"), strings.HasPrefix(raw, "http://"):
		return parseHTTPS(h, raw)
	default:
		return parseShorthand(h, raw)
	}
}

func parseLocal(h Handle, raw string) (Handle, error) {
	abs, err := filepath.Abs(raw)
	if err != nil {
		return h, fmt.Errorf("resolving local path %q: %w", raw, err)
	}
	h.Kind = KindLocal
	h.SubPath = abs
	return h, nil
}

func parseSSH(h Handle, raw string) (Handle, error) {
	// git@github.com:owner/repo.git[/path/to/skill]
	h.Kind = KindSSH

	atIdx := strings.Index(raw, "@")
	colonIdx := strings.Index(raw, ":")
	if atIdx == -1 || colonIdx == -1 || colonIdx < atIdx {
		return h, fmt.Errorf("invalid SSH handle %q: expected git@host:owner/repo.git[/path]", raw)
	}
	h.Host = raw[atIdx+1 : colonIdx]
	rest := raw[colonIdx+1:] // owner/repo.git[/path]

	var repoSlug string
	if idx := strings.Index(rest, ".git/"); idx != -1 {
		repoSlug = rest[:idx]
		h.SubPath = rest[idx+5:]
	} else if strings.HasSuffix(rest, ".git") {
		repoSlug = strings.TrimSuffix(rest, ".git")
	} else {
		return h, fmt.Errorf("SSH handle must include .git suffix: %q", raw)
	}

	h.RepoURL = fmt.Sprintf("git@%s:%s.git", h.Host, repoSlug)
	parts := strings.SplitN(repoSlug, "/", 2)
	h.Owner = parts[0]
	if len(parts) > 1 {
		h.RepoName = parts[1]
	}
	return h, nil
}

func parseHTTPS(h Handle, raw string) (Handle, error) {
	h.Kind = KindHTTPS
	u, err := url.Parse(raw)
	if err != nil {
		return h, fmt.Errorf("parsing HTTPS URL %q: %w", raw, err)
	}
	h.Host = u.Host

	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(parts) < 2 {
		return h, fmt.Errorf("HTTPS URL must include owner/repo: %q", raw)
	}
	h.Owner = parts[0]
	h.RepoName = strings.TrimSuffix(parts[1], ".git")
	h.RepoURL = fmt.Sprintf("https://%s/%s/%s", h.Host, h.Owner, h.RepoName)

	if len(parts) > 2 {
		// https://github.com/owner/repo/tree/<ref>/path...
		if parts[2] == "tree" && len(parts) > 3 {
			// URL-embedded ref is used only when no explicit ref was provided.
			if h.Ref == "HEAD" {
				h.Ref = parts[3]
			}
			if len(parts) > 4 {
				h.SubPath = strings.Join(parts[4:], "/")
			}
		} else {
			h.SubPath = strings.Join(parts[2:], "/")
		}
	}
	return h, nil
}

func parseShorthand(h Handle, raw string) (Handle, error) {
	h.Kind = KindShorthand
	h.Host = "github.com"

	parts := strings.Split(raw, "/")
	if len(parts) < 2 {
		return h, fmt.Errorf("shorthand must be at least owner/repo, got: %q", raw)
	}
	h.Owner = parts[0]
	h.RepoName = parts[1]
	h.RepoURL = fmt.Sprintf("https://github.com/%s/%s", h.Owner, h.RepoName)
	if len(parts) > 2 {
		h.SubPath = strings.Join(parts[2:], "/")
	}
	return h, nil
}
