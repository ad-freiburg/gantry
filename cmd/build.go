package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds all pipeline images",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		// Build images
		log.Print("Prepare steps\n")
		if err = pipeline.PrepareImages(); err != nil {
			log.Fatal(err)
		}
	},
}
