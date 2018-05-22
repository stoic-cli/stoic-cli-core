package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:              "stoic",
	Short:            "Stoic CLI Hub, for all your CLI needs",
	SilenceErrors:    true,
	SilenceUsage:     true,
	TraverseChildren: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			fmt.Fprintf(os.Stderr, "Error: %s", err)
		}
		os.Exit(1)
	}
}
