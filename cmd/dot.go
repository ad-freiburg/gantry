package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(dotCmd)
	dotCmd.Flags().StringVar(&dotOutput, "output", "gantry.dot", "File to store .dot output")
}

var (
	dotOutput string
)

var dotCmd = &cobra.Command{
	Use:   "dot",
	Short: "Generates a .dot file for graph visualisation",
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Create(dotOutput)
		if err != nil {
			return err
		}

		defer f.Close()

		pipelines, err := pipeline.Definition.Pipelines()
		if err != nil {
			return err
		}

		w := bufio.NewWriter(f)
		w.WriteString("digraph gantry {\nrankdir=\"BT\"\n")
		for _, step := range pipelines.AllSteps() {
			sName := strings.Replace(step.Name(), "-", "_", -1)
			// Display services as ellipse, and steps as rectangle
			shape := "ellipse"
			if !step.Detach {
				shape = "rectangle"
			}
			w.WriteString(fmt.Sprintf("%s [label=\"%s\", shape=%s]\n", sName, step.Name(), shape))
			for name, _ := range *step.Dependencies() {
				w.WriteString(fmt.Sprintf("%s -> %s\n", sName, strings.Replace(name, "-", "_", -1)))
			}
		}
		w.WriteString("}\n")
		w.Flush()
		return nil
	},
}
