package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pullCmd)
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pulls images for services/steps defined in a Compose file, but does not start the containers.",
	Run: func(cmd *cobra.Command, args []string) {
		pipeline.PullImages(true)
	},
}
