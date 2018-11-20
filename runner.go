package gantry // import "github.com/ad-freiburg/gantry"

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
)

func getContainerExecutable() string {
	if isWharferInstalled() {
		if isUserRoot() || isUserInDockerGroup() {
			return "docker"
		}
		return "wharfer"
	}
	return "docker"
}

func isUserRoot() bool {
	u, err := user.Current()
	if err != nil {
		return false
	}
	return u.Uid == "0"
}

func isUserInDockerGroup() bool {
	u, err := user.Current()
	if err != nil {
		return false
	}
	gids, err := u.GroupIds()
	if err != nil {
		return false
	}
	for _, gid := range gids {
		group, err := user.LookupGroupId(gid)
		if err != nil {
			return false
		}
		if group.Name == "docker" {
			return true
		}
	}
	return false
}

func isWharferInstalled() bool {
	cmd := exec.Command("wharfer", "--version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

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

// Check for existence of Image
type ImageExistenceChecker struct {
	runner Runner
	step   Step
}

func NewImageExistenceChecker(step Step) *ImageExistenceChecker {
	r := &ImageExistenceChecker{runner: step.Runner(), step: step}
	r.runner.SetCommand(getContainerExecutable(), []string{"images", "--format", "{{.Repository}}"})
	return r
}

func (r *ImageExistenceChecker) Exec() error {
	err := r.runner.Exec()
	if err != nil {
		return err
	}

	// Check output
	found := false

	if !found {
		return fmt.Errorf("Image not found '%s'", r.step.ImageName())
	}
	return nil
}

// Build images
type ImageBuilder struct {
	runner Runner
}

func NewImageBuilder(step Step) *ImageBuilder {
	r := &ImageBuilder{runner: step.Runner()}
	r.runner.SetCommand(getContainerExecutable(), []string{"build", "--tag", step.ImageName(), step.BuildInfo.Context})
	return r
}

func (r *ImageBuilder) Exec() error {
	return r.runner.Exec()
}

// Pull images
type ImagePuller struct {
	runner Runner
}

func NewImagePuller(step Step) *ImagePuller {
	r := &ImagePuller{runner: step.Runner()}
	r.runner.SetCommand(getContainerExecutable(), []string{"pull", step.ImageName()})
	return r
}

func (r *ImagePuller) Exec() error {
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
	for _, envvar := range step.Environment {
		args = append(args, "-e", envvar)
	}
	args = append(args, step.ImageName())
	if len(step.Args) > 0 {
		args = append(args, step.Args...)
	}
	r.runner.SetCommand(getContainerExecutable(), args)
	return r
}

func (r *ImageRunner) Exec() error {
	return r.runner.Exec()
}
