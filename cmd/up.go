package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(upCmd)
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Builds, (re)creates, and starts containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := pipeline.PullImages(false)
		if err != nil {
			return err
		}
		err = buildCmd.RunE(cmd, args)
		if err != nil {
			return err
		}
		return startCmd.RunE(cmd, args)
	},
}
