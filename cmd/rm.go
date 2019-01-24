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
	Run: func(cmd *cobra.Command, args []string) {
		if gantry.Verbose {
			log.Print("Remove container\n")
		}
		pipeline.RemoveContainers()
	},
}
