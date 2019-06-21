package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().BoolVar(&forcePull, "pull", false, "Always attempt to pull a newer version of the image.")
}

var (
	forcePull bool
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds all pipeline images",
	RunE: func(cmd *cobra.Command, args []string) error {
		return pipeline.BuildImages(forcePull)
	},
}
