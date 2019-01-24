package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ad-freiburg/gantry"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gantry",
	Short: "gantry is a docker-compose like pipeline tool",
	Long:  `Tool for running pipelines and docker-compose deployments.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if pipeline != nil {
			return
		}
		var err error
		pipeline, err = gantry.NewPipeline(defFile, envFile)
		if err != nil {
			log.Fatal(err)
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
				log.Fatal(err)
			}
			gantry.ProjectName = filepath.Base(cwd)
		}
		gantry.ProjectName = strings.Replace(strings.Replace(strings.ToLower(gantry.ProjectName), " ", "_", -1), ".", "", -1)
		pipeline.NetworkName = fmt.Sprintf("%s_gantry", gantry.ProjectName)
	},
	Run: func(cmd *cobra.Command, args []string) {
		killCmd.Run(cmd, args)
		rmCmd.Run(cmd, args)
		upCmd.Run(cmd, args)
	},
	Version: gantry.Version,
}

var (
	defFile  string
	envFile  string
	pipeline *gantry.Pipeline
)

func init() {
	defFileUsage := fmt.Sprintf("Explicit %s to use", gantry.GantryDef)
	envFileUsage := fmt.Sprintf("Explicit %s to use", gantry.GantryEnv)
	projectNameUsage := "Spefify an alternate project name"
	rootCmd.PersistentFlags().StringVar(&defFile, "file", "", defFileUsage)
	rootCmd.PersistentFlags().StringVar(&defFile, "f", "", defFileUsage+" (shorthand)")
	rootCmd.PersistentFlags().StringVar(&envFile, "env", "", envFileUsage)
	rootCmd.PersistentFlags().StringVar(&envFile, "e", "", envFileUsage+" (shorthand)")
	rootCmd.PersistentFlags().StringVar(&gantry.ProjectName, "project-name", "", projectNameUsage)
	rootCmd.PersistentFlags().StringVar(&gantry.ProjectName, "p", "", projectNameUsage)
	rootCmd.PersistentFlags().BoolVar(&gantry.Verbose, "verbose", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVar(&gantry.ForceWharfer, "force-wharfer", false, "Force usage of wharfer")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
