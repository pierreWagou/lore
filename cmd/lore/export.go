package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/pierreWagou/lore/internal/harness"
	"github.com/pierreWagou/lore/internal/installer"
	"github.com/pierreWagou/lore/internal/manifest"
)

var exportCmd = &cobra.Command{
	Use:   "export [skill-name]",
	Short: "Export skills from .ai/skills/ to harness directories",
	Long: `Export skills from the neutral store (.ai/skills/) into one or more harness
directories, applying the harness's native format transformation.

Only valid in guest mode — push locally-imported skills into a committed harness
directory, or produce skills in a specific harness's format for sharing.

In keeper mode, .ai/skills/ is already the source of truth; use lore sync instead.

Harness resolution order: --harness flag > lore.toml harnesses > auto-detect.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runExport,
}

var (
	exportGlobal  bool
	exportHarness string
	exportAll     bool
)

func init() {
	exportCmd.Flags().BoolVarP(&exportGlobal, "global", "g", false, "export to global harness dirs")
	exportCmd.Flags().StringVar(&exportHarness, "harness", "", "comma-separated target harnesses (default: manifest harnesses or auto-detect)")
	exportCmd.Flags().BoolVar(&exportAll, "all", false, "export all skills in .ai/skills/")
}

func runExport(cmd *cobra.Command, args []string) error {
	if !exportAll && len(args) == 0 {
		return fmt.Errorf("specify a skill name or use --all")
	}

	root := projectRoot()
	mPath := manifestPath(exportGlobal)

	m, err := manifest.Load(mPath)
	if err != nil {
		return err
	}

	if !manifest.IsGuest(m) {
		return fmt.Errorf("lore export only works in guest mode (current mode: %q)\nin keeper mode, .ai/skills/ is the committed source of truth — use `lore sync` instead", m.Mode)
	}

	// Resolve target adapters (same priority chain as all other commands).
	adapters, err := resolveExportAdapters(exportHarness, m)
	if err != nil {
		return err
	}

	// Collect skill names to export.
	var skillNames []string
	if exportAll {
		skillNames, err = listNeutralSkills(root)
		if err != nil {
			return err
		}
		if len(skillNames) == 0 {
			fmt.Println("no skills found in .ai/skills/")
			return nil
		}
	} else {
		skillNames = []string{args[0]}
	}

	for _, name := range skillNames {
		neutralDir := filepath.Join(installer.NeutralSkillsDir(root), name)
		if _, err := os.Stat(neutralDir); os.IsNotExist(err) {
			return fmt.Errorf("skill %q not found in .ai/skills/", name)
		}

		files, err := readSkillFiles(neutralDir)
		if err != nil {
			return fmt.Errorf("reading %s: %w", name, err)
		}

		skill := harness.Skill{Name: name, Files: files}

		for _, adapter := range adapters {
			outFiles, err := adapter.Transform(skill)
			if err != nil {
				return fmt.Errorf("%s/%s: %w", adapter.Name(), name, err)
			}

			var skillsDir string
			if exportGlobal {
				skillsDir = adapter.GlobalSkillsDir()
			} else {
				skillsDir = adapter.ProjectSkillsDir(root)
			}
			destDir := filepath.Join(skillsDir, name)

			for _, f := range outFiles {
				dest := filepath.Join(destDir, f.Path)
				if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
					return err
				}
				if err := os.WriteFile(dest, f.Content, 0644); err != nil {
					return err
				}
			}
			fmt.Printf("exported %s → %s\n", name, destDir)
		}
	}
	return nil
}

// resolveExportAdapters resolves harness adapters for export using the same
// priority chain as install: flag > manifest > auto-detect.
func resolveExportAdapters(flag string, m *manifest.Manifest) ([]harness.Adapter, error) {
	if flag != "" {
		return installer.AdaptersByNames(splitHarnesses(flag))
	}
	if m != nil && len(m.Harnesses) > 0 {
		return installer.AdaptersByNames(m.Harnesses)
	}
	detected := harness.Detected()
	if len(detected) == 0 {
		return nil, fmt.Errorf("no harnesses detected; use --harness to specify one")
	}
	return detected, nil
}

// listNeutralSkills returns all skill names in .ai/skills/.
func listNeutralSkills(root string) ([]string, error) {
	neutralBase := installer.NeutralSkillsDir(root)
	entries, err := os.ReadDir(neutralBase)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// readSkillFiles reads all files from a skill directory into a map.
func readSkillFiles(skillDir string) (map[string][]byte, error) {
	files := make(map[string][]byte)
	err := filepath.Walk(skillDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(skillDir, path)
		files[rel] = data
		return nil
	})
	return files, err
}
