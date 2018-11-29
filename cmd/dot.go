package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"bufio"
	"fmt"
	"os"

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
	Run: func(cmd *cobra.Command, args []string) {
		f, err := os.Create(dotOutput)
		if err != nil {
			panic(err)
		}

		defer f.Close()

		pipelines, err := pipeline.Definition.Pipelines()
		if err != nil {
			panic(err)
		}

		w := bufio.NewWriter(f)
		w.WriteString("digraph gantry {\nrankdir=\"BT\"\n")
		for _, step := range pipelines.AllSteps() {
			dependencies, err := step.Dependencies()
			if err != nil {
				panic(err)
			}
			for name, _ := range *dependencies {
				w.WriteString(fmt.Sprintf("%s -> %s\n", step.Name, name))
			}
		}
		w.WriteString("}\n")
		w.Flush()
	},
}
