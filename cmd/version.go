package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"fmt"

	"github.com/ad-freiburg/gantry"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Gantry",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(gantry.Version)
	},
}
