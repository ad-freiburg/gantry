package gantry // import "github.com/ad-freiburg/gantry"

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/google/shlex"
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
		args := []string{"build", "--tag", step.ImageName()}
		if step.BuildInfo.Dockerfile != "" {
			args = append(args, "--file", filepath.Join(step.BuildInfo.Context, step.BuildInfo.Dockerfile))
		}
		if step.BuildInfo.Context == "" {
			step.BuildInfo.Context = "."
		}
		if pull {
			args = append(args, "--pull")
		}
		args = append(args, step.BuildInfo.Context)
		r := step.Runner()
		r.SetCommand(getContainerExecutable(), args)
		return r.Exec()
	}
}

func NewImagePuller(step Step) func() error {
	return func() error {
		r := step.Runner()
		r.SetCommand(getContainerExecutable(), []string{"pull", step.ImageName()})
		return r.Exec()
	}
}

func NewContainerRunner(step Step, network string) func() error {
	return func() error {
		r := step.Runner()
		args := []string{
			"run",
			"--name", step.ContainerName(),
			"--network", network,
			"--network-alias", step.RawContainerName(),
			"--network-alias", step.ContainerName(),
		}
		if step.Detach {
			args = append(args, "-d")
		} else {
			args = append(args, "--rm")
		}
		for _, port := range step.Ports {
			args = append(args, "-p", port)
		}
		for _, volume := range step.Volumes {
			// Resolve relative paths
			var err error
			parts := strings.SplitN(volume, ":", 2)
			parts[0], err = filepath.Abs(parts[0])
			if err != nil {
				return err
			}
			args = append(args, "-v", strings.Join(parts, ":"))
		}
		for _, envvar := range step.Environment {
			args = append(args, "-e", envvar)
		}
		// Determine entrypoint and arguments
		callerArgs := make([]string, 0)
		if len(step.Entrypoint) > 0 {
			if len(step.Entrypoint) > 1 {
				args = append(args, "--entrypoint", step.Entrypoint[0])
				callerArgs = append(callerArgs, step.Entrypoint[1:]...)
			} else {
				tokens, _ := shlex.Split(step.Entrypoint[0])
				args = append(args, "--entrypoint", tokens[0])
				callerArgs = append(callerArgs, tokens[1:]...)
			}
		}
		// Add command
		if len(step.Command) > 0 {
			if len(step.Command) > 1 {
				callerArgs = append(callerArgs, step.Command...)
			} else {
				tokens, _ := shlex.Split(step.Command[0])
				callerArgs = append(callerArgs, tokens...)
			}
		}
		args = append(args, step.ImageName())
		if len(callerArgs) > 0 {
			args = append(args, callerArgs...)
		}
		r.SetCommand(getContainerExecutable(), args)
		return r.Exec()
	}
}

func NewContainerKiller(step Step) func() error {
	return func() error {
		r := step.Runner()
		r.SetCommand(getContainerExecutable(), []string{"ps", "-q", "--filter", "name=" + step.ContainerName()})
		out, err := r.Output()
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(bytes.NewReader(out))
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			k := step.Runner()
			k.SetCommand(getContainerExecutable(), []string{"kill", scanner.Text()})
			if err := k.Exec(); err != nil {
				return err
			}
		}
		if err := scanner.Err(); err != nil {
			return err
		}
		return nil
	}
}

func NewImageExistenceChecker(step Step) func() error {
	return func() error {
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

func NewOldContainerRemover(step Step) func() error {
	return func() error {
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
		if err := scanner.Err(); err != nil {
			return err
		}
		return nil
	}
}

func NewNetworkCreator(p Pipeline) func() error {
	return func() error {
		r := p.Runner()
		r.SetCommand(getContainerExecutable(), []string{"network", "create", p.NetworkName})
		return r.Exec()
	}
}

func NewNetworkRemover(p Pipeline) func() error {
	return func() error {
		r := p.Runner()
		r.SetCommand(getContainerExecutable(), []string{"network", "rm", p.NetworkName})
		return r.Exec()
	}
}
