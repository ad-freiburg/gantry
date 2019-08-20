package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(dotCmd)
	dotCmd.Flags().StringVar(&dotOutput, "output", "gantry.dot", "File to store .dot output")
	dotCmd.Flags().BoolVar(&hideIgnored, "hide-ignored", false, "Hide ignored steps in .dot output")
}

var (
	dotOutput   string
	hideIgnored bool
)

var dotCmd = &cobra.Command{
	Use:   "dot [flags] [Service/Step...]",
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
		if _, err := w.WriteString("digraph gantry {\nrankdir=\"BT\"\n"); err != nil {
			log.Printf("Error writing graph header: %s", err)
		}
		for _, step := range pipelines.AllSteps() {
			if hideIgnored && step.Meta.Ignore {
				continue
			}
			sName := strings.ReplaceAll(step.Name, "-", "_")
			// Display services as ellipse, and steps as rectangle
			shape := "rectangle"
			style := "solid"
			if step.Detach {
				shape = "ellipse"
			}
			if step.Meta.Ignore {
				style = "dashed"
			}
			if _, err := w.WriteString(fmt.Sprintf("%s [label=\"%s\", shape=%s, style=%s]\n", sName, step.Name, shape, style)); err != nil {
				log.Printf("Error writing node: %s", err)
			}
			for name := range step.Dependencies() {
				if hideIgnored && pipeline.Definition.Steps[name].Meta.Ignore {
					continue
				}
				if _, err := w.WriteString(fmt.Sprintf("%s -> %s\n", sName, strings.ReplaceAll(name, "-", "_"))); err != nil {
					log.Printf("Error writing connection: %s", err)
				}
			}
		}
		if _, err := w.WriteString("}\n"); err != nil {
			log.Printf("Error writing graph tail: %s", err)
		}
		w.Flush()
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {},
}
