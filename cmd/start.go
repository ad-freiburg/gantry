package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start [flags] [Service/Step...]",
	Short: "Starts containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		pipeline.CreateNetwork()
		return pipeline.ExecuteSteps()
	},
}
