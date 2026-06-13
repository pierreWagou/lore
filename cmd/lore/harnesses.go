package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pierreWagou/lore/internal/harness"
)

var harnessesCmd = &cobra.Command{
	Use:   "harnesses",
	Short: "Detect installed harnesses",
	Args:  cobra.NoArgs,
	RunE:  runHarnesses,
}

func runHarnesses(cmd *cobra.Command, args []string) error {
	detected := harness.Detected()
	if len(detected) == 0 {
		fmt.Println("no harnesses detected.")
		fmt.Printf("available: %s\n", availableHarnessNames())
		return nil
	}
	fmt.Println("detected harnesses:")
	for _, a := range detected {
		fmt.Printf("  %s\n", a.Name())
	}
	return nil
}
