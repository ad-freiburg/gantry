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
	Run: func(cmd *cobra.Command, args []string) {
		log.Print("Load pipeline\n")
		p, err := gantry.NewPipeline(defFile, envFile)
		if err != nil {
			log.Fatal(err)
		}

		// Check for obvious errors
		log.Print("Check pipeline\n")
		if err = p.Check(); err != nil {
			log.Fatal(err)
		}

		// Build images
		log.Print("Prepare steps\n")
		if err = p.PrepareImages(); err != nil {
			log.Fatal(err)
		}

		// Execute step after step
		log.Print("Exec steps\n")
		if err = p.ExecuteSteps(); err != nil {
			log.Fatal(err)
		}

	},
}

var (
	defFile string
	envFile string
)

func init() {
	defFileUsage := fmt.Sprintf("Explicit %s to use", gantry.GantryDef)
	envFileUsage := fmt.Sprintf("Explicit %s to use", gantry.GantryEnv)
	rootCmd.PersistentFlags().StringVar(&defFile, "file", "", defFileUsage)
	rootCmd.PersistentFlags().StringVar(&defFile, "f", "", defFileUsage+" (shorthand)")
	rootCmd.PersistentFlags().StringVar(&envFile, "env", "", envFileUsage)
	rootCmd.PersistentFlags().StringVar(&envFile, "e", "", envFileUsage+" (shorthand)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
