package auth

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/pierreWagou/lore/internal/config"
)

// HostToken is a stored credential for a git host.
type HostToken struct {
	Host  string `toml:"host"`
	Token string `toml:"token"`
}

type credentials struct {
	Hosts []HostToken `toml:"hosts"`
}

func credentialsPath() string {
	return filepath.Join(config.LoreConfigDir(), "credentials.toml")
}

func loadCredentials() (*credentials, error) {
	creds := &credentials{}
	path := credentialsPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return creds, nil
	}
	if _, err := toml.DecodeFile(path, creds); err != nil {
		return nil, fmt.Errorf("reading credentials: %w", err)
	}
	return creds, nil
}

func saveCredentials(creds *credentials) error {
	path := credentialsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(creds)
}

// StoreToken stores an auth token for a host.
func StoreToken(host, token string) error {
	creds, err := loadCredentials()
	if err != nil {
		return err
	}
	for i, h := range creds.Hosts {
		if h.Host == host {
			creds.Hosts[i].Token = token
			return saveCredentials(creds)
		}
	}
	creds.Hosts = append(creds.Hosts, HostToken{Host: host, Token: token})
	return saveCredentials(creds)
}

// RemoveToken removes the stored token for a host.
func RemoveToken(host string) error {
	creds, err := loadCredentials()
	if err != nil {
		return err
	}
	for i, h := range creds.Hosts {
		if h.Host == host {
			creds.Hosts = append(creds.Hosts[:i], creds.Hosts[i+1:]...)
			return saveCredentials(creds)
		}
	}
	return nil
}

// ListTokens returns all stored host tokens.
func ListTokens() ([]HostToken, error) {
	creds, err := loadCredentials()
	if err != nil {
		return nil, err
	}
	return creds.Hosts, nil
}

// ResolveToken returns the raw HTTPS token for repoURL, or "" for public repos
// or SSH URLs.
func ResolveToken(repoURL string) string {
	if strings.HasPrefix(repoURL, "git@") || strings.HasPrefix(repoURL, "ssh://") {
		return ""
	}
	return resolveHTTPSToken(repoURL)
}

func resolveHTTPSToken(repoURL string) string {
	host := extractHost(repoURL)

	envKey := "LORE_" + strings.ToUpper(strings.NewReplacer(".", "_", "-", "_").Replace(host)) + "_TOKEN"
	if token := os.Getenv(envKey); token != "" {
		return token
	}

	if host == "github.com" {
		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			return token
		}
		if token := os.Getenv("GH_TOKEN"); token != "" {
			return token
		}
		if token := ghCliToken(); token != "" {
			return token
		}
		if token := ghHostsToken(); token != "" {
			return token
		}
	}

	creds, err := loadCredentials()
	if err == nil {
		for _, h := range creds.Hosts {
			if h.Host == host {
				return h.Token
			}
		}
	}
	return ""
}

func extractHost(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Hostname()
}

func ghCliToken() string {
	out, err := exec.Command("gh", "auth", "token").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func ghHostsToken() string {
	data, err := os.ReadFile(filepath.Join(config.XDGConfigHome(), "gh", "hosts.yml"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "oauth_token:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "oauth_token:"))
		}
	}
	return ""
}
