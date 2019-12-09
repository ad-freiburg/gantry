package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(preprocessorCmd)
}

var preprocessorCmd = &cobra.Command{
	Use:   "preprocessor",
	Short: "Preprocessor functionality, has subcommands",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("missing sub-command")
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {},
}
