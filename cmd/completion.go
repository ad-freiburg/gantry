package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(completionCmd)
}

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates completion scripts for the specified shell (bash or zsh)",
	Args:  cobra.MaximumNArgs(1),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		switch {
		case len(args) == 0:
			return rootCmd.GenBashCompletion(os.Stdout)
		case args[0] == "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		default:
			return rootCmd.GenBashCompletion(os.Stdout)
		}
	},
}
