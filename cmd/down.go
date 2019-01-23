package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(downCmd)
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop and remove containers, and networks",
	Run: func(cmd *cobra.Command, args []string) {
		killCmd.Run(cmd, args)
	},
}
