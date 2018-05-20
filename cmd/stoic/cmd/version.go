package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stoic-cli/stoic-cli-core/engine"
	"github.com/stoic-cli/stoic-cli-core/tool"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information for tools that have been setup",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return version(args)
	},
}

func version(args []string) error {
	root := viper.GetString("root")
	engine, err := engine.NewWithOptions(engine.EngineOptions{
		Root: root,
	})
	if err != nil {
		return err
	}

	tools := engine.Tools()
	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Name() < tools[j].Name()
	})

	out := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	for _, t := range tools {
		version := t.Version()
		if version == tool.NullVersion {
			continue
		}

		name := t.Name()
		endpoint := strings.TrimPrefix(t.Endpoint().String(), "https://")

		channel := t.Channel()
		if channel != tool.DefaultChannel {
			endpoint = fmt.Sprintf("%v@%v", endpoint, channel)
		}
		fmt.Fprintf(out, "%v\t%v\t%v%v\n", name, version, endpoint, channel)
	}
	out.Flush()

	return nil
}
