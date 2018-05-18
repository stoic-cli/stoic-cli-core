package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stoic-cli/stoic-cli-core/engine"
)

func init() {
	runCmd.Flags().SetInterspersed(false)
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a CLI tool",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(args[0], args[1:])
	},
}

func run(toolName string, args []string) error {
	root := viper.GetString("root")
	engine, err := engine.NewWithOptions(engine.EngineOptions{
		Root: root,
	})
	if err != nil {
		return err
	}
	return engine.RunTool(toolName, args)
}
