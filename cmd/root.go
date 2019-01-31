package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ad-freiburg/gantry"
	"github.com/ad-freiburg/gantry/types"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gantry",
	Short: "gantry is a docker-compose like pipeline tool",
	Long:  `Tool for running pipelines and docker-compose deployments.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if pipeline != nil {
			return nil
		}
		var err error
		ignoredSteps := types.StringSet{}
		for _, step := range stepsToIgnore {
			ignoredSteps[step] = true
		}
		pipeline, err = gantry.NewPipeline(defFile, envFile, ignoredSteps)
		if err != nil {
			return err
		}
		// Check for obvious errors
		if gantry.Verbose {
			log.Print("Check pipeline\n")
		}
		if err = pipeline.Check(); err != nil {
			log.Fatal(err)
		}
		if gantry.ProjectName == "" {
			// If ProjectName was not set, try to calculate it
			if gantry.Verbose {
				log.Print("Calculate project-name\n")
			}
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			gantry.ProjectName = filepath.Base(cwd)
		}
		gantry.ProjectName = strings.Replace(strings.Replace(strings.ToLower(gantry.ProjectName), " ", "_", -1), ".", "", -1)
		pipeline.NetworkName = fmt.Sprintf("%s_gantry", gantry.ProjectName)
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := killCmd.RunE(cmd, args)
		if err != nil {
			return err
		}
		err = rmCmd.RunE(cmd, args)
		if err != nil {
			return err
		}
		return upCmd.RunE(cmd, args)
	},
	Version: gantry.Version,
}

var (
	defFile       string
	envFile       string
	pipeline      *gantry.Pipeline
	stepsToIgnore []string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&defFile, "file", "f", "", fmt.Sprintf("Explicit %s to use", gantry.GantryDef))
	rootCmd.PersistentFlags().StringVarP(&envFile, "env", "e", "", fmt.Sprintf("Explicit %s to use", gantry.GantryEnv))
	rootCmd.PersistentFlags().StringVarP(&gantry.ProjectName, "project-name", "p", "", "Spefify an alternate project name")
	rootCmd.PersistentFlags().BoolVar(&gantry.Verbose, "verbose", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVar(&gantry.ForceWharfer, "force-wharfer", false, "Force usage of wharfer")
	rootCmd.PersistentFlags().StringArrayVarP(&stepsToIgnore, "ignore", "i", []string{}, "Ignore step/service with this name")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
