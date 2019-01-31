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
	RunE: func(cmd *cobra.Command, args []string) error {
		if gantry.Verbose {
			log.Print("(Re)create network\n")
		}
		pipeline.CreateNetwork()
		if gantry.Verbose {
			log.Print("Start container\n")
		}
		return pipeline.ExecuteSteps()
	},
}
