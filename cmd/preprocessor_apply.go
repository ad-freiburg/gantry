package cmd // import "github.com/ad-freiburg/gantry/cmd"

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"syscall"

	"github.com/ad-freiburg/gantry"
	"github.com/ad-freiburg/gantry/preprocessor"
	"github.com/ad-freiburg/gantry/types"
	"github.com/spf13/cobra"
)

func init() {
	preprocessorCmd.AddCommand(preprocessorApplyCmd)
}

var preprocessorApplyCmd = &cobra.Command{
	Use:   "apply file",
	Short: "Prints the result of applying the preprocessor to the given file",
	Args:  cobra.ExactArgs(1),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		defFile = args[0]
		ignoredSteps := types.StringSet{}
		for _, step := range stepsToIgnore {
			ignoredSteps[step] = true
		}
		selectedSteps := types.StringSet{}
		for _, step := range args {
			selectedSteps[step] = true
		}
		env := types.StringMap{}
		for _, v := range environment {
			parts := strings.SplitN(v, "=", 2)
			if len(parts) == 1 {
				env[parts[0]] = nil
			} else {
				env[parts[0]] = &parts[1]
			}
		}
		environment, err := gantry.NewPipelineEnvironment(envFile, env, ignoredSteps, selectedSteps)
		if err != nil {
			if e, ok := err.(*os.PathError); ok && e.Err != syscall.ENOENT {
				return err
			}
		}

		file, err := os.Open(defFile)
		if err != nil {
			return err
		}
		defer file.Close()
		data, err := ioutil.ReadAll(file)
		if err != nil {
			return err
		}
		// Apply environment to yaml
		preproc, err := preprocessor.NewPreprocessor()
		if err != nil {
			return err
		}
		data, err = preproc.Process(data, environment)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", data)
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {},
}
