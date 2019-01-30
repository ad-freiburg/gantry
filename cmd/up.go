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
	Run: func(cmd *cobra.Command, args []string) {
		pipeline.PullImages(false)
		buildCmd.Run(cmd, args)
		startCmd.Run(cmd, args)
	},
}
