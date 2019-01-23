package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(killCmd)
}

var killCmd = &cobra.Command{
	Use:   "kill",
	Short: "Force stop containers",
	Run: func(cmd *cobra.Command, args []string) {
		if verbose {
			log.Print("Kill container\n")
		}
		pipeline.KillContainers()
		if verbose {
			log.Print("Remove network\n")
		}
	},
}
