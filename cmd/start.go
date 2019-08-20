package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start [flags] [Service/Step...]",
	Short: "Starts containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := pipeline.CreateNetwork(); err != nil {
			log.Printf("Error creating network: %s", err)
		}
		return pipeline.ExecuteSteps()
	},
}
