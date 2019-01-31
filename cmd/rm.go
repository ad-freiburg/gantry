package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"log"

	"github.com/ad-freiburg/gantry"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(rmCmd)
}

var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Removes stopped containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		if gantry.Verbose {
			log.Print("Remove container\n")
		}
		return pipeline.RemoveContainers()
	},
}
