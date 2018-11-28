package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().BoolVar(&forceBuildPull, "force", false, "Always attempt to pull/build a newer version of the image.")
}

var (
	forceBuildPull bool
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds all pipeline images",
	Run: func(cmd *cobra.Command, args []string) {
		pipeline.PrepareImages(forceBuildPull)
	},
}
