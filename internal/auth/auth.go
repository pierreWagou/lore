package auth

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	gogithttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	gogitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"

	"github.com/go-git/go-git/v5/plumbing/transport"
)

// HostToken is a stored credential for a git host.
type HostToken struct {
	Host  string `toml:"host"`
	Token string `toml:"token"`
}

type credentials struct {
	Hosts []HostToken `toml:"hosts"`
}

// configDir returns the lore config directory.
// LORE_CONFIG_DIR overrides the default for testing and custom setups.
func configDir() string {
	if override := os.Getenv("LORE_CONFIG_DIR"); override != "" {
		return override
	}
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "lore")
}

func credentialsPath() string {
	return filepath.Join(configDir(), "credentials.toml")
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

// Resolve returns the appropriate go-git auth method for the given repo URL.
// Returns nil auth (no error) for public repos that don't need authentication.
func Resolve(repoURL string) (transport.AuthMethod, error) {
	if strings.HasPrefix(repoURL, "git@") || strings.HasPrefix(repoURL, "ssh://") {
		return resolveSSH()
	}
	return resolveHTTPS(repoURL)
}

func resolveSSH() (transport.AuthMethod, error) {
	// 1. Try SSH agent
	auth, err := gogitssh.NewSSHAgentAuth("git")
	if err == nil {
		return auth, nil
	}

	// 2. Try common key files in ~/.ssh
	home, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return nil, fmt.Errorf("SSH auth: no agent (%v) and cannot find home dir: %w", err, homeErr)
	}
	for _, name := range []string{"id_ed25519", "id_rsa", "id_ecdsa", "id_dsa"} {
		keyPath := filepath.Join(home, ".ssh", name)
		if _, statErr := os.Stat(keyPath); statErr != nil {
			continue
		}
		keyAuth, keyErr := gogitssh.NewPublicKeysFromFile("git", keyPath, "")
		if keyErr == nil {
			return keyAuth, nil
		}
	}

	return nil, fmt.Errorf("SSH auth: no agent available and no usable key file found in ~/.ssh (tried id_ed25519, id_rsa, id_ecdsa, id_dsa)")
}

func resolveHTTPS(repoURL string) (transport.AuthMethod, error) {
	token := resolveHTTPSToken(repoURL)
	if token == "" {
		return nil, nil
	}
	return httpAuth(token), nil
}

// ResolveToken returns the raw HTTPS token for repoURL, or "" for public repos
// or SSH URLs. Uses the same resolution chain as Resolve.
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

func httpAuth(token string) transport.AuthMethod {
	return &gogithttp.BasicAuth{
		Username: "git",
		Password: token,
	}
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
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(dir, "gh", "hosts.yml"))
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
