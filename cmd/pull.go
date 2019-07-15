package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pullCmd)
}

var pullCmd = &cobra.Command{
	Use:   "pull [flags] [Service/Step...]",
	Short: "Pulls images for services/steps defined in a Compose file, but does not start the containers.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return pipeline.PullImages(true)
	},
}
