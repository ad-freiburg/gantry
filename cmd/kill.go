package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(killCmd)
}

var killCmd = &cobra.Command{
	Use:   "kill",
	Short: "Force stop containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		return pipeline.KillContainers()
	},
}
