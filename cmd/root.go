package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"fmt"
	"log"
	"os"

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
		if verbose {
			log.Print("Check pipeline\n")
		}
		if err = pipeline.Check(); err != nil {
			log.Fatal(err)
		}
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
	verbose  bool
	pipeline *gantry.Pipeline
)

func init() {
	defFileUsage := fmt.Sprintf("Explicit %s to use", gantry.GantryDef)
	envFileUsage := fmt.Sprintf("Explicit %s to use", gantry.GantryEnv)
	rootCmd.PersistentFlags().StringVar(&defFile, "file", "", defFileUsage)
	rootCmd.PersistentFlags().StringVar(&defFile, "f", "", defFileUsage+" (shorthand)")
	rootCmd.PersistentFlags().StringVar(&envFile, "env", "", envFileUsage)
	rootCmd.PersistentFlags().StringVar(&envFile, "e", "", envFileUsage+" (shorthand)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Verbose output")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
