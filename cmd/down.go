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
	RunE: func(cmd *cobra.Command, args []string) error {
		err := killCmd.RunE(cmd, args)
		if err != nil {
			return err
		}
		err = rmCmd.RunE(cmd, args)
		if err != nil {
			return err
		}
		if gantry.Verbose {
			log.Print("Remove network\n")
		}
		return pipeline.RemoveNetwork()
	},
}
