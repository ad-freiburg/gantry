package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/ad-freiburg/gantry"
	"github.com/ad-freiburg/gantry/types"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gantry [flags] [Service/Step...]",
	Args:  cobra.ArbitraryArgs,
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
		selectedSteps := types.StringSet{}
		for _, step := range args {
			selectedSteps[step] = true
		}
		env := types.MappingWithEquals{}
		for _, v := range environment {
			parts := strings.SplitN(v, "=", 2)
			if len(parts) == 1 {
				env[parts[0]] = nil
			} else {
				env[parts[0]] = &parts[1]
			}
		}
		pipeline, err = gantry.NewPipeline(defFile, envFile, env, ignoredSteps, selectedSteps)
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
		if gantry.ProjectName == "" && pipeline.Environment.ProjectName != "" {
			gantry.ProjectName = pipeline.Environment.ProjectName
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
		gantry.ProjectName = strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(gantry.ProjectName), " ", "_"), ".", "")
		pipeline.NetworkName = fmt.Sprintf("%s_gantry", gantry.ProjectName)
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := pipeline.KillContainers(true); err != nil {
			return err
		}
		if err := pipeline.RemoveContainers(true); err != nil {
			return err
		}
		return upCmd.RunE(cmd, args)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		pipeline.CleanUp(syscall.Signal(0))
	},
	Version:                gantry.Version,
	BashCompletionFunction: bashCompletionFunc,
}

const (
	bashCompletionFunc = `__gantry_get_steps()
{
    local gantry_output out
    if gantry_output=$(gantry steps 2>/dev/null); then
        out=($(echo "${gantry_output}" | awk '{print $1}'))
        COMPREPLY=( $( compgen -W "${out[*]}" -- "$cur" ) )
    fi
}
`
)

var (
	defFile       string
	envFile       string
	pipeline      *gantry.Pipeline
	stepsToIgnore []string
	environment   []string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&defFile, "file", "f", "", fmt.Sprintf("Explicit %s to use", gantry.GantryDef))
	rootCmd.PersistentFlags().StringVarP(&envFile, "global-environment", "g", "", fmt.Sprintf("Explicit %s to use", gantry.GantryEnv))
	rootCmd.PersistentFlags().StringVarP(&gantry.ProjectName, "project-name", "p", "", "Spefify an alternate project name")
	rootCmd.PersistentFlags().BoolVar(&gantry.Verbose, "verbose", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVar(&gantry.ForceWharfer, "force-wharfer", false, "Force usage of wharfer")
	rootCmd.PersistentFlags().StringArrayVarP(&stepsToIgnore, "ignore", "i", []string{}, "Ignore step/service with this name")
	rootCmd.PersistentFlags().StringArrayVarP(&environment, "env", "e", []string{}, "Set environment variables")
	rootCmd.PersistentFlags().SetAnnotation("file", cobra.BashCompFilenameExt, []string{".yaml", ".yml"})
	rootCmd.PersistentFlags().SetAnnotation("global-environment", cobra.BashCompFilenameExt, []string{".yaml", ".yml"})
	rootCmd.PersistentFlags().SetAnnotation("ignore", cobra.BashCompCustom, []string{"__gantry_get_steps"})
	go signalHandler()
}

func signalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c)
	for s := range c {
		switch s {
		case syscall.SIGINT:
			pipeline.CleanUp(s)
			os.Exit(1)
		case syscall.SIGKILL:
			pipeline.CleanUp(s)
			os.Exit(1)
		case syscall.SIGCHLD:
		default:
			log.Printf("%q\n", s)
		}
	}
}

// Execute is the main entrypoint for using gantry commands.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
