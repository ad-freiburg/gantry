package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stepsCmd)
}

var stepsCmd = &cobra.Command{
	Use: "steps",
	RunE: func(cmd *cobra.Command, args []string) error {
		pipelines, err := pipeline.Definition.Pipelines()
		if err != nil {
			return err
		}

		for _, step := range pipelines.AllSteps() {
			fmt.Println(step.Name())
		}
		return nil
	},
}
