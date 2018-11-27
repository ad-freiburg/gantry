package cmd // import "github.com/ad-freiburg/gantry/cmd"

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds all pipeline images",
	Run: func(cmd *cobra.Command, args []string) {

	},
}
