package gantry // import "github.com/ad-freiburg/gantry"

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"os/user"
)

func getContainerExecutable() string {
	if ForceWharfer {
		return "wharfer"
	}
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

// Runner represents generic container runners.
type Runner interface {
	ImageBuilder(Step, bool) func() error
	ImagePuller(Step) func() error
	ImageExistenceChecker(Step) func() error
	ContainerKiller(Step) func() (int, error)
	ContainerRemover(Step) func() error
	ContainerRunner(Step, Network) func() error
}

// NoopRunner is a runner that does nothing.
type NoopRunner struct{}

// ImageBuilder returns a function to build the image for the given step.
func (r *NoopRunner) ImageBuilder(Step, bool) func() error {
	return func() error {
		return nil
	}
}

// ImagePuller returns a function to pull the image for the given step.
func (r *NoopRunner) ImagePuller(Step) func() error {
	return func() error {
		return nil
	}
}

// ImageExistenceChecker returns a function which checks if the image for the given step exists.
func (r *NoopRunner) ImageExistenceChecker(Step) func() error {
	return func() error {
		return nil
	}
}

// ContainerKiller returns a function to kill the container for the given step.
func (r *NoopRunner) ContainerKiller(Step) func() (int, error) {
	return func() (int, error) {
		return 0, nil
	}
}

// ContainerRemover returns a function to remove the container for the given step.
func (r *NoopRunner) ContainerRemover(Step) func() error {
	return func() error {
		return nil
	}
}

// ContainerRunner returns a function to run the given step.
func (r *NoopRunner) ContainerRunner(step Step, n Network) func() error {
	return func() error {
		pipelineLogger.Printf("- Skipping: %s!", step.ColoredContainerName())
		return nil
	}
}

// LocalRunner creates functions running on localhost.
type LocalRunner struct {
	prefix string
	stdout io.Writer
	stderr io.Writer
}

// NewLocalRunner returns a LocalRunner using provided defaults.
func NewLocalRunner(prefix string, stdout io.Writer, stderr io.Writer) *LocalRunner {
	r := &LocalRunner{
		prefix: prefix,
		stdout: stdout,
		stderr: stderr,
	}
	return r
}

// Exec executes given arguments with the containerExecutable.
func (r *LocalRunner) Exec(args []string) error {
	cmd := exec.Command(getContainerExecutable(), args...)
	stdout := NewPrefixedLogger(r.prefix, log.New(r.stdout, "", log.LstdFlags))
	stderr := NewPrefixedLogger(r.prefix, log.New(r.stderr, "", log.LstdFlags))
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

// Output executes given arguments with the containerExecutable and returns the output.
func (r *LocalRunner) Output(args []string) ([]byte, error) {
	cmd := exec.Command(getContainerExecutable(), args...)
	return cmd.Output()
}

// ImageBuilder returns a function to build the image for the given step.
func (r *LocalRunner) ImageBuilder(step Step, pull bool) func() error {
	return func() error {
		if Verbose {
			log.Printf("Build image for '%s'", step.ContainerName())
		}
		r.prefix = step.ColoredContainerName()
		r.stdout = step.Meta.Stdout
		r.stderr = step.Meta.Stderr
		return r.Exec(step.BuildCommand(pull))
	}
}

// ImagePuller retunrs a function to pull the image for the given step.
func (r *LocalRunner) ImagePuller(step Step) func() error {
	return func() error {
		if Verbose {
			log.Printf("Pull image for '%s'", step.ContainerName())
		}
		r.prefix = step.ColoredContainerName()
		r.stdout = step.Meta.Stdout
		r.stderr = step.Meta.Stderr
		return r.Exec(step.PullCommand())
	}
}

// ImageExistenceChecker returns a function which checks if the image for the given step exists.
func (r *LocalRunner) ImageExistenceChecker(step Step) func() error {
	return func() error {
		if Verbose {
			log.Printf("Check image ('%s') existence for '%s'", step.ImageName(), step.ContainerName())
		}
		r.prefix = step.ColoredContainerName()
		r.stdout = step.Meta.Stdout
		r.stderr = step.Meta.Stderr
		out, err := r.Output([]string{"images", "--format", "{{.ID}};{{.Repository}}", step.ImageName()})
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(bytes.NewReader(out))
		scanner.Split(bufio.ScanWords)
		count := 0
		for scanner.Scan() {
			count++
		}
		if err := scanner.Err(); err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("Image not found '%s'", step.ImageName())
		}
		return nil
	}
}

// ContainerKiller returns a function to kill the container for the given step.
func (r *LocalRunner) ContainerKiller(step Step) func() (int, error) {
	return func() (int, error) {
		var counter int
		if Verbose {
			log.Printf("Kill container '%s'", step.ContainerName())
		}
		r.prefix = step.ColoredContainerName()
		r.stdout = step.Meta.Stdout
		r.stderr = step.Meta.Stderr
		out, err := r.Output([]string{"ps", "-q", "--filter", "name=" + step.ContainerName()})
		if err != nil {
			return counter, err
		}
		scanner := bufio.NewScanner(bytes.NewReader(out))
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			counter++
			if err := r.Exec([]string{"kill", scanner.Text()}); err != nil {
				return counter, err
			}
		}
		return counter, scanner.Err()
	}
}

// ContainerRemover returns a function to remove the container for the given step.
func (r *LocalRunner) ContainerRemover(step Step) func() error {
	return func() error {
		if Verbose {
			log.Printf("Remove container '%s'", step.ContainerName())
		}
		r.prefix = step.ColoredContainerName()
		r.stdout = step.Meta.Stdout
		r.stderr = step.Meta.Stderr
		out, err := r.Output([]string{"ps", "-a", "-q", "--filter", "name=" + step.ContainerName()})
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(bytes.NewReader(out))
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			if err := r.Exec([]string{"rm", scanner.Text()}); err != nil {
				return err
			}
		}
		return scanner.Err()
	}
}

// ContainerRunner returns a function to run the given step.
func (r *LocalRunner) ContainerRunner(step Step, network Network) func() error {
	return func() error {
		if Verbose {
			log.Printf("Run container '%s'", step.ContainerName())
		}
		r.prefix = step.ColoredContainerName()
		r.stdout = step.Meta.Stdout
		r.stderr = step.Meta.Stderr
		return r.Exec(step.RunCommand(string(network)))
	}
}

// NetworkCreator returns a function to create the given network.
func (r *LocalRunner) NetworkCreator(network Network) func() error {
	return func() error {
		if Verbose {
			log.Printf("Create network '%s'", network)
		}
		return r.Exec([]string{"network", "create", string(network)})
	}
}

// NetworkRemover returns a function to remove the given network.
func (r *LocalRunner) NetworkRemover(network Network) func() error {
	return func() error {
		if Verbose {
			log.Printf("Remove network '%s'", network)
		}
		return r.Exec([]string{"network", "rm", string(network)})
	}
}
