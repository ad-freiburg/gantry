package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"log"

	"github.com/ad-freiburg/gantry"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(downCmd)
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop and remove containers, and networks created by `up`",
	Run: func(cmd *cobra.Command, args []string) {
		killCmd.Run(cmd, args)
		rmCmd.Run(cmd, args)
		if gantry.Verbose {
			log.Print("Remove network\n")
		}
		pipeline.RemoveNetwork()
	},
}
