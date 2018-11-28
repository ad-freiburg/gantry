package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(rmCmd)
}

var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Removes stopped containers",
	Run: func(cmd *cobra.Command, args []string) {
		pipeline.RemoveContainers()
	},
}
