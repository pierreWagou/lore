package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/pierreWagou/lore/internal/auth"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication tokens",
}

var authAddCmd = &cobra.Command{
	Use:   "add <host> <token>",
	Short: "Store an auth token for a git host",
	Long: `Store a personal access token for a git host.

Examples:
  lore auth add github.com ghp_yourtoken
  lore auth add gitlab.com glpat_yourtoken
  lore auth add git.company.com yourtoken

For GitHub, lore also checks GITHUB_TOKEN, GH_TOKEN, and the gh CLI automatically.`,
	Args: cobra.ExactArgs(2),
	RunE: runAuthAdd,
}

var authListCmd = &cobra.Command{
	Use:   "list",
	Short: "List stored auth tokens",
	Args:  cobra.NoArgs,
	RunE:  runAuthList,
}

var authRemoveCmd = &cobra.Command{
	Use:   "remove <host>",
	Short: "Remove a stored auth token",
	Args:  cobra.ExactArgs(1),
	RunE:  runAuthRemove,
}

func init() {
	authCmd.AddCommand(authAddCmd)
	authCmd.AddCommand(authListCmd)
	authCmd.AddCommand(authRemoveCmd)
}

func runAuthAdd(cmd *cobra.Command, args []string) error {
	host, token := args[0], args[1]
	if err := auth.StoreToken(host, token); err != nil {
		return fmt.Errorf("storing token: %w", err)
	}
	fmt.Printf("token stored for %s\n", host)
	return nil
}

func runAuthList(cmd *cobra.Command, args []string) error {
	tokens, err := auth.ListTokens()
	if err != nil {
		return err
	}
	if len(tokens) == 0 {
		fmt.Println("no stored tokens.")
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "HOST\tTOKEN")
	for _, t := range tokens {
		masked := maskToken(t.Token)
		fmt.Fprintf(w, "%s\t%s\n", t.Host, masked)
	}
	return w.Flush()
}

func runAuthRemove(cmd *cobra.Command, args []string) error {
	if err := auth.RemoveToken(args[0]); err != nil {
		return err
	}
	fmt.Printf("token removed for %s\n", args[0])
	return nil
}

// maskToken shows only the first 4 and last 4 characters.
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
