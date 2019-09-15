package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolVar(&noFollow, "no-follow", false, "Do not follow log output.")
}

var (
	noFollow bool
)

var logsCmd = &cobra.Command{
	Use:   "logs [flags] [Service/Step...]",
	Short: "View output from containers.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return pipeline.Logs(!noFollow)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {},
}
