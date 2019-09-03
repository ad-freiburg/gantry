package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(upCmd)
}

var upCmd = &cobra.Command{
	Use:   "up [flags] [Service/Step...]",
	Short: "Builds, (re)creates, and starts containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := pipeline.PullImages(false); err != nil {
			return err
		}
		if err := buildCmd.RunE(cmd, args); err != nil {
			return err
		}
		return startCmd.RunE(cmd, args)
	},
}
