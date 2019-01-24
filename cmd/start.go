package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"log"

	"github.com/ad-freiburg/gantry"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts containers",
	Run: func(cmd *cobra.Command, args []string) {
		if gantry.Verbose {
			log.Print("(Re)create network\n")
		}
		pipeline.CreateNetwork()
		if gantry.Verbose {
			log.Print("Start container\n")
		}
		pipeline.ExecuteSteps()
	},
}
