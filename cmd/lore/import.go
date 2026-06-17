package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/pierreWagou/lore/internal/harness"
	"github.com/pierreWagou/lore/internal/installer"
	"github.com/pierreWagou/lore/internal/lockfile"
	"github.com/pierreWagou/lore/internal/manifest"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import skills from existing harness directories into .ai/skills/",
	Long: `Import skills already present in harness directories (e.g. .claude/skills/)
into lore's neutral skill store (.ai/skills/) and install them for your harnesses.

Only valid in guest mode. Original harness directories are never modified.`,
	Args: cobra.NoArgs,
	RunE: runImport,
}

func runImport(cmd *cobra.Command, args []string) error {
	root := projectRoot()
	mPath := manifestPath(false)
	lPath := lockfilePath(false)

	m, err := manifest.Load(mPath)
	if err != nil {
		return err
	}

	if !manifest.IsGuest(m) {
		return fmt.Errorf("lore import only works in guest mode (current mode: %q)\nuse `lore add` to install skills in keeper mode", m.Mode)
	}

	lf, err := lockfile.Load(lPath)
	if err != nil {
		return err
	}

	fmt.Println("scanning team_harnesses directories...")

	// Use team_harnesses from manifest; fall back to all harness dirs if not configured.
	var adaptersToScan []harness.Adapter
	if len(m.TeamHarnesses) > 0 {
		for _, name := range m.TeamHarnesses {
			if a := harness.Get(name); a != nil {
				adaptersToScan = append(adaptersToScan, a)
			}
		}
	} else {
		adaptersToScan = harness.All()
	}

	imported := 0
	for _, adapter := range adaptersToScan {
		skillsDir := adapter.ProjectSkillsDir(root)
		entries, err := os.ReadDir(skillsDir)
		if err != nil {
			continue // harness dir doesn't exist or unreadable
		}

		for _, entry := range entries {
			// Skip symlinks (already managed by lore) and non-directories.
			if entry.Type()&os.ModeSymlink != 0 {
				continue
			}
			if !entry.IsDir() {
				continue
			}

			name := entry.Name()
			skillMD := filepath.Join(skillsDir, name, "SKILL.md")
			if _, err := os.Stat(skillMD); os.IsNotExist(err) {
				continue // not a skill dir
			}

			// Skip if already in .ai/skills/ or in lore.toml.
			neutralDir := filepath.Join(installer.NeutralSkillsDir(root), name)
			if _, err := os.Stat(neutralDir); err == nil {
				fmt.Printf("  skip %s (already in .ai/skills/)\n", name)
				continue
			}
			if manifest.HasDependency(m, name) {
				fmt.Printf("  skip %s (already in lore.toml)\n", name)
				continue
			}

			fmt.Printf("  found %s/%s → .ai/skills/%s/\n", adapter.Name(), name, name)

			// Copy skill directory to neutral store (non-destructive).
			if err := copyDir(filepath.Join(skillsDir, name), neutralDir); err != nil {
				return fmt.Errorf("copying %s: %w", name, err)
			}

			// Install to all configured harnesses (creates symlinks / copies).
			dep := manifest.Dependency{
				Name:   name,
				Source: filepath.Join(".ai", "skills", name),
				Ref:    "",
			}
			opts := installer.Options{
				Global: false,
				Root:   root,
			}
			sr, installErr := installer.Install(dep, opts, m)
			if installErr != nil {
				fmt.Fprintf(os.Stderr, "  warning: install %s: %v\n", name, installErr)
				continue
			}
			for _, r := range sr.Results {
				fmt.Printf("    → %s: %s\n", r.Harness, r.Path)
			}

			manifest.AddDependency(m, dep)
			lockfile.SetEntry(lf, lockfile.NewEntry(name, dep.Source, sr.Commit, sr.ContentHash))
			imported++
		}
	}

	if imported == 0 {
		fmt.Println("no new skills found to import.")
		return nil
	}

	if err := manifest.Save(mPath, m); err != nil {
		return err
	}
	if err := lockfile.Save(lPath, lf); err != nil {
		return err
	}

	fmt.Printf("\nimported %d skill(s).\n", imported)
	return nil
}

// copyDir copies all files from src to dst recursively.
// dst is created if it does not exist. src is never modified.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		destPath := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}
		return copyFile(path, destPath)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
