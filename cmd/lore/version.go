package main

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Populated at build time via -ldflags.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the lore version",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("lore %s %s built with %s at %s\n", Version, Commit, runtime.Version(), Date)
	},
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate an autocompletion script for lore for the specified shell.
See each sub-command's help for details on how to use the generated script.`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			_ = cmd.Root().GenBashCompletion(cmd.OutOrStdout())
		case "zsh":
			_ = cmd.Root().GenZshCompletion(cmd.OutOrStdout())
		case "fish":
			_ = cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
		case "powershell":
			_ = cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
		}
	},
}
