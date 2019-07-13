package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(downCmd)
}

var downCmd = &cobra.Command{
	Use:   "down [flags] [Service/Step...]",
	Short: "Stop and remove containers, and networks created by `up`",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := killCmd.RunE(cmd, args); err != nil {
			return err
		}
		if err := rmCmd.RunE(cmd, args); err != nil {
			return err
		}
		return pipeline.RemoveNetwork()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {},
}
