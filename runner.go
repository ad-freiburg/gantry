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

type Executable interface {
	Exec() error
	Output() ([]byte, error)
}

type Runner interface {
	Executable
	SetCommand(name string, args []string)
}

// Local host
type LocalRunner struct {
	name   string
	args   []string
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

func (r *LocalRunner) Exec() error {
	cmd := exec.Command(r.name, r.args...)
	stdout := NewPrefixedLogger(r.prefix, log.New(r.stdout, "", log.LstdFlags))
	stderr := NewPrefixedLogger(r.prefix, log.New(r.stderr, "", log.LstdFlags))
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func (r *LocalRunner) Output() ([]byte, error) {
	cmd := exec.Command(r.name, r.args...)
	return cmd.Output()
}

func (r *LocalRunner) SetCommand(name string, args []string) {
	r.name = name
	r.args = args
}

func NewImageBuilder(step Step, pull bool) func() error {
	return func() error {
		if Verbose {
			log.Printf("Build image for '%s'", step.ContainerName())
		}
		r := step.Runner()
		r.SetCommand(getContainerExecutable(), step.BuildCommand(pull))
		return r.Exec()
	}
}

func NewImagePuller(step Step) func() error {
	return func() error {
		if Verbose {
			log.Printf("Pull image for '%s'", step.ContainerName())
		}
		r := step.Runner()
		r.SetCommand(getContainerExecutable(), step.PullCommand())
		return r.Exec()
	}
}

func NewContainerRunner(step Step, network string) func() error {
	return func() error {
		if Verbose {
			log.Printf("Run container '%s'", step.ContainerName())
		}
		r := step.Runner()
		r.SetCommand(getContainerExecutable(), step.RunCommand(network))
		return r.Exec()
	}
}

func NewContainerKiller(step Step) func() (int, error) {
	return func() (int, error) {
		var counter int
		if Verbose {
			log.Printf("Kill container '%s'", step.ContainerName())
		}
		r := step.Runner()
		r.SetCommand(getContainerExecutable(), []string{"ps", "-q", "--filter", "name=" + step.ContainerName()})
		out, err := r.Output()
		if err != nil {
			return counter, err
		}
		scanner := bufio.NewScanner(bytes.NewReader(out))
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			counter += 1
			k := step.Runner()
			k.SetCommand(getContainerExecutable(), []string{"kill", scanner.Text()})
			if err := k.Exec(); err != nil {
				return counter, err
			}
		}
		return counter, scanner.Err()
	}
}

func NewImageExistenceChecker(step Step) func() error {
	return func() error {
		if Verbose {
			log.Printf("Check image ('%s') existence for '%s'", step.ImageName(), step.ContainerName())
		}
		r := step.Runner()
		r.SetCommand(getContainerExecutable(), []string{"images", "--format", "{{.ID}};{{.Repository}}", step.ImageName()})
		out, err := r.Output()
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

func NewContainerRemover(step Step) func() error {
	return func() error {
		if Verbose {
			log.Printf("Remove container '%s'", step.ContainerName())
		}
		r := step.Runner()
		r.SetCommand(getContainerExecutable(), []string{"ps", "-a", "-q", "--filter", "name=" + step.ContainerName()})
		out, err := r.Output()
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(bytes.NewReader(out))
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			k := step.Runner()
			k.SetCommand(getContainerExecutable(), []string{"rm", scanner.Text()})
			if err := k.Exec(); err != nil {
				return err
			}
		}
		return scanner.Err()
	}
}

func NewNetworkCreator(p Pipeline) func() error {
	return func() error {
		if Verbose {
			log.Printf("Create network '%s'", p.NetworkName)
		}
		r := p.Runner()
		r.SetCommand(getContainerExecutable(), []string{"network", "create", p.NetworkName})
		return r.Exec()
	}
}

func NewNetworkRemover(p Pipeline) func() error {
	return func() error {
		if Verbose {
			log.Printf("Remove network '%s'", p.NetworkName)
		}
		r := p.Runner()
		r.SetCommand(getContainerExecutable(), []string{"network", "rm", p.NetworkName})
		return r.Exec()
	}
}
