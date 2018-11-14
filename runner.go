package gantry // import "github.com/ad-freiburg/gantry"

import (
	"os"
	"os/exec"
)

type Executable interface {
	Exec() error
}

type Runner interface {
	Executable
	SetCommand(name string, args []string)
}

// Local host
type LocalRunner struct {
	name string
	args []string
}

func NewLocalRunner() *LocalRunner {
	r := &LocalRunner{}
	return r
}

func (r *LocalRunner) Exec() error {
	cmd := exec.Command(r.name, r.args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (r *LocalRunner) SetCommand(name string, args []string) {
	r.name = name
	r.args = args
}

// Run arbitray command on machine
type CommandRunner struct {
	runner Runner
}

func NewCommandRunner(runner Runner) *CommandRunner {
	r := &CommandRunner{runner: runner}
	return r
}

func (r *CommandRunner) Exec() error {
	return r.runner.Exec()
}

// Build images
type ImageBuilder struct {
	runner Runner
}

func NewImageBuilder(step Step) *ImageBuilder {
	r := &ImageBuilder{runner: step.Runner()}
	r.runner.SetCommand("wharfer", []string{"build", "--tag", step.ImageName(), step.Context})
	return r
}

func (r *ImageBuilder) Exec() error {
	return r.runner.Exec()
}

// Run image using wharfer or docker
type ImageRunner struct {
	runner Runner
}

func NewImageRunner(step Step) *ImageRunner {
	r := &ImageRunner{runner: step.Runner()}
	args := []string{"run"}
	for _, port := range step.Ports {
		args = append(args, "-p", port)
	}
	for _, volume := range step.Volumes {
		args = append(args, "-v", volume)
	}
	args = append(args, step.ImageName())
	r.runner.SetCommand("wharfer", args)
	return r
}

func (r *ImageRunner) Exec() error {
	return r.runner.Exec()
}
