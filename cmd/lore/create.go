package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/pierreWagou/lore/internal/installer"
	"github.com/pierreWagou/lore/internal/lockfile"
	"github.com/pierreWagou/lore/internal/manifest"
)

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new project skill in .ai/skills/",
	Long: `Create a new skill in the project's neutral skill store (.ai/skills/<name>/SKILL.md).

The skill is added to lore.toml and installed into all configured harnesses.
Edit .ai/skills/<name>/SKILL.md, then run lore sync to propagate changes
to harnesses that use copies instead of symlinks (e.g. cursor).`,
	Args: cobra.ExactArgs(1),
	RunE: runCreate,
}

func init() {
	createCmd.Flags().StringVarP(&addHarnesses, "harness", "", "", "comma-separated harnesses (e.g. opencode,claude)")
}

func runCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	root := projectRoot()

	skillDir := filepath.Join(root, ".ai", "skills", name)
	skillFile := filepath.Join(skillDir, "SKILL.md")

	if _, err := os.Stat(skillFile); err == nil {
		return fmt.Errorf("skill %q already exists at %s", name, skillFile)
	}

	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return err
	}

	content := fmt.Sprintf("---\nname: %q\ndescription: \"\"\n---\n\n# %s\n\nDescribe what this skill does.\n", name, name)
	if err := os.WriteFile(skillFile, []byte(content), 0644); err != nil {
		return err
	}
	fmt.Printf("created .ai/skills/%s/SKILL.md\n", name)

	mPath := manifestPath(false)
	lPath := lockfilePath(false)

	m, err := manifest.Load(mPath)
	if err != nil {
		return err
	}
	lf, err := lockfile.Load(lPath)
	if err != nil {
		return err
	}

	dep := manifest.Dependency{
		Name:   name,
		Source: filepath.Join(".ai", "skills", name),
		Ref:    "",
	}

	opts := installer.Options{
		Global:    false,
		Harnesses: splitHarnesses(addHarnesses),
		Root:      root,
	}

	sr, err := installer.Install(dep, opts, m)
	if errors.As(err, &installer.ErrNoHarnesses{}) {
		harnesses, wizErr := promptSelectHarnesses(mPath, m)
		if wizErr != nil {
			return wizErr
		}
		opts.Harnesses = harnesses
		sr, err = installer.Install(dep, opts, m)
	}
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}
	for _, r := range sr.Results {
		fmt.Printf("  → %s: %s\n", r.Harness, r.Path)
	}

	manifest.AddDependency(m, dep)
	lockfile.SetEntry(lf, lockfile.NewEntry(name, dep.Source, sr.Commit, sr.ContentHash))

	if err := manifest.Save(mPath, m); err != nil {
		return err
	}
	return lockfile.Save(lPath, lf)
}
