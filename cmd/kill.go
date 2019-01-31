package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"log"

	"github.com/ad-freiburg/gantry"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(killCmd)
}

var killCmd = &cobra.Command{
	Use:   "kill",
	Short: "Force stop containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		if gantry.Verbose {
			log.Print("Kill container\n")
		}
		return pipeline.KillContainers()
	},
}
