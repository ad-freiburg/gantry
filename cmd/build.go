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
	Run: func(cmd *cobra.Command, args []string) {
		if gantry.Verbose {
			log.Print("(Re)build images\n")
		}
		pipeline.BuildImages(forcePull)
	},
}
