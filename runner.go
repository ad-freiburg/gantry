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

type Runner interface {
	ImageBuilder(Step, bool) func() error
	ImagePuller(Step) func() error
	ImageExistenceChecker(Step) func() error
	ContainerKiller(Step) func() (int, error)
	ContainerRemover(Step) func() error
	ContainerRunner(Step, Network) func() error
}

// Noop
type NoopRunner struct{}

func (r *NoopRunner) ImageBuilder(Step, bool) func() error {
	return func() error {
		return nil
	}
}
func (r *NoopRunner) ImagePuller(Step) func() error {
	return func() error {
		return nil
	}
}
func (r *NoopRunner) ImageExistenceChecker(Step) func() error {
	return func() error {
		return nil
	}
}
func (r *NoopRunner) ContainerKiller(Step) func() (int, error) {
	return func() (int, error) {
		return 0, nil
	}
}
func (r *NoopRunner) ContainerRemover(Step) func() error {
	return func() error {
		return nil
	}
}
func (r *NoopRunner) ContainerRunner(step Step, n Network) func() error {
	return func() error {
		pipelineLogger.Printf("- Skipping: %s!", step.ColoredContainerName())
		return nil
	}
}

// Local host
type LocalRunner struct {
	prefix string
	stdout io.Writer
	stderr io.Writer
}

func NewLocalRunner(prefix string, stdout io.Writer, stderr io.Writer) *LocalRunner {
	r := &LocalRunner{
		prefix: prefix,
		stdout: stdout,
		stderr: stderr,
	}
	return r
}

func (r *LocalRunner) Exec(args []string) error {
	cmd := exec.Command(getContainerExecutable(), args...)
	stdout := NewPrefixedLogger(r.prefix, log.New(r.stdout, "", log.LstdFlags))
	stderr := NewPrefixedLogger(r.prefix, log.New(r.stderr, "", log.LstdFlags))
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func (r *LocalRunner) Output(args []string) ([]byte, error) {
	cmd := exec.Command(getContainerExecutable(), args...)
	return cmd.Output()
}

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

func (r *LocalRunner) ContainerRunner(step Step, network Network) func() error {
	return func() error {
		if Verbose {
			log.Printf("Run container '%s'", step.ContainerName())
		}
		r.prefix = step.ColoredContainerName()
		r.stdout = step.Meta.Stdout
		r.stderr = step.Meta.Stderr
		return r.Exec(step.RunCommand(fmt.Sprintf("%s", network)))
	}
}

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
			counter += 1
			if err := r.Exec([]string{"kill", scanner.Text()}); err != nil {
				return counter, err
			}
		}
		return counter, scanner.Err()
	}
}

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

func (r *LocalRunner) NetworkCreator(network Network) func() error {
	return func() error {
		if Verbose {
			log.Printf("Create network '%s'", network)
		}
		return r.Exec([]string{"network", "create", fmt.Sprintf("%s", network)})
	}
}

func (r *LocalRunner) NetworkRemover(network Network) func() error {
	return func() error {
		if Verbose {
			log.Printf("Remove network '%s'", network)
		}
		return r.Exec([]string{"network", "rm", fmt.Sprintf("%s", network)})
	}
}
