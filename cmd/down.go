package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(downCmd)
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop and remove containers, and networks created by `up`",
	RunE: func(cmd *cobra.Command, args []string) error {
		// This is a docker-compose only command, do not try to clean up TempDirs
		pipeline.Environment.TempDirNoAutoClean = true
		if err := killCmd.RunE(cmd, args); err != nil {
			return err
		}
		if err := rmCmd.RunE(cmd, args); err != nil {
			return err
		}
		return pipeline.RemoveNetwork()
	},
}
