package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"log"

	"github.com/ad-freiburg/gantry"
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
		if gantry.Verbose {
			log.Print("(Re)build images\n")
		}
		return pipeline.BuildImages(forcePull)
	},
}
