package lockfile

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

const FileName = "lore.lock"

// GlobalFileName returns the lockfile filename for a named profile: "lore.<profile>.lock".
// Returns FileName ("lore.lock") when profileName is empty.
func GlobalFileName(profileName string) string {
	if profileName == "" {
		return FileName
	}
	return "lore." + profileName + ".lock"
}

// Lockfile represents the contents of a lore.lock file.
type Lockfile struct {
	Entries []Entry `toml:"entry"`
}

// Entry is a resolved, pinned skill dependency.
type Entry struct {
	Name        string `toml:"name"`
	Source      string `toml:"source"`
	Commit      string `toml:"commit"`
	ContentHash string `toml:"content_hash"`
	ResolvedAt  string `toml:"resolved_at"`
}

// NewEntry creates a new lockfile entry with the current timestamp.
func NewEntry(name, source, commit, contentHash string) Entry {
	return Entry{
		Name:        name,
		Source:      source,
		Commit:      commit,
		ContentHash: contentHash,
		ResolvedAt:  time.Now().UTC().Format(time.RFC3339),
	}
}

// Load reads a lockfile from path. Returns an empty lockfile if the file does not exist.
func Load(path string) (*Lockfile, error) {
	lf := &Lockfile{}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return lf, nil
	}
	if _, err := toml.DecodeFile(path, lf); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return lf, nil
}

// Save writes the lockfile to path, creating parent directories as needed.
func Save(path string, lf *Lockfile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := fmt.Fprintln(f, "# lore.lock — do not edit manually"); err != nil {
		return err
	}
	return toml.NewEncoder(f).Encode(lf)
}

// GetEntry returns the entry with the given name, or nil.
func GetEntry(lf *Lockfile, name string) *Entry {
	for i := range lf.Entries {
		if lf.Entries[i].Name == name {
			return &lf.Entries[i]
		}
	}
	return nil
}

// SetEntry adds or replaces the entry with the given name.
func SetEntry(lf *Lockfile, entry Entry) {
	for i, e := range lf.Entries {
		if e.Name == entry.Name {
			lf.Entries[i] = entry
			return
		}
	}
	lf.Entries = append(lf.Entries, entry)
}

// RemoveEntry removes the entry with the given name. Returns true if found.
func RemoveEntry(lf *Lockfile, name string) bool {
	for i, e := range lf.Entries {
		if e.Name == name {
			lf.Entries = append(lf.Entries[:i], lf.Entries[i+1:]...)
			return true
		}
	}
	return false
}
