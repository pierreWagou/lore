package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/pierreWagou/lore/internal/lockfile"
	"github.com/pierreWagou/lore/internal/manifest"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed skills",
	Args:  cobra.NoArgs,
	RunE:  runList,
}

var listGlobal bool

func init() {
	listCmd.Flags().BoolVarP(&listGlobal, "global", "g", false, "list globally installed skills")
}

func runList(cmd *cobra.Command, args []string) error {
	mPath := manifestPath(listGlobal)
	lPath := lockfilePath(listGlobal)

	m, err := manifest.Load(mPath)
	if err != nil {
		return err
	}
	lf, err := lockfile.Load(lPath)
	if err != nil {
		return err
	}

	if len(m.Dependencies) == 0 {
		fmt.Println("no skills installed.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSOURCE\tREF\tCOMMIT")
	for _, dep := range m.Dependencies {
		commit := "(not locked)"
		if entry := lockfile.GetEntry(lf, dep.Name); entry != nil {
			commit = entry.Commit[:12]
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", dep.Name, dep.Source, dep.Ref, commit)
	}
	return w.Flush()
}
