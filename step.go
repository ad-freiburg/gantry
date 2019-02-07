package gantry // import "github.com/ad-freiburg/gantry"

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ad-freiburg/gantry/types"
	"github.com/google/shlex"
)

type Service struct {
	BuildInfo   BuildInfo                 `json:"build"`
	Command     types.StringOrStringSlice `json:"command"`
	Entrypoint  types.StringOrStringSlice `json:"entrypoint"`
	Image       string                    `json:"image"`
	Ports       []string                  `json:"ports"`
	Volumes     []string                  `json:"volumes"`
	Environment map[string]string         `json:"environment"`
	DependsOn   types.StringSet           `json:"depends_on"`
	Name        string
	color       int
}

type Step struct {
	Service
	Role   string          `json:"role"`
	After  types.StringSet `json:"after"`
	Detach bool            `json:"detach"`
}

func (s Step) Dependencies() *types.StringSet {
	r := types.StringSet{}
	for dep, _ := range s.After {
		r[dep] = true
	}
	for dep, _ := range s.DependsOn {
		r[dep] = true
	}
	return &r
}

func (s *Service) InitColor() {
	s.color = GetNextFriendlyColor()
}

func (s Service) ColoredName() string {
	return ApplyStyle(s.Name, s.color)
}

func (s Service) ColoredContainerName() string {
	return ApplyStyle(s.ContainerName(), s.color)
}

func (s Service) ImageName() string {
	if s.Image != "" {
		return s.Image
	}
	return strings.Replace(strings.ToLower(s.Name), " ", "_", -1)
}

func (s Service) RawContainerName() string {
	return strings.Replace(strings.ToLower(s.Name), " ", "_", -1)
}

func (s Service) ContainerName() string {
	return fmt.Sprintf("%s_%s", ProjectName, strings.Replace(strings.ToLower(s.Name), " ", "_", -1))
}

func (s Service) Runner() Runner {
	r := NewLocalRunner(s.ColoredContainerName(), os.Stdout, os.Stderr)
	return r
}

func (s Step) BuildCommand(pull bool) []string {
	args := []string{"build", "--tag", s.ImageName()}
	if s.BuildInfo.Dockerfile != "" {
		args = append(args, "--file", filepath.Join(s.BuildInfo.Context, s.BuildInfo.Dockerfile))
	}
	if s.BuildInfo.Context == "" {
		s.BuildInfo.Context = "."
	}
	if pull {
		args = append(args, "--pull")
	}
	for k, v := range s.BuildInfo.Args {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, v))
	}
	args = append(args, s.BuildInfo.Context)
	return args
}

func (s Step) RunCommand(network string) []string {
	args := []string{
		"run",
		"--name", s.ContainerName(),
		"--network", network,
		"--network-alias", s.RawContainerName(),
		"--network-alias", s.ContainerName(),
	}
	if s.Detach {
		args = append(args, "-d")
	} else {
		args = append(args, "--rm")
	}
	for _, port := range s.Ports {
		args = append(args, "-p", port)
	}
	for _, volume := range s.Volumes {
		// Resolve relative paths
		parts := strings.SplitN(volume, ":", 2)
		parts[0], _ = filepath.Abs(parts[0])
		args = append(args, "-v", strings.Join(parts, ":"))
	}
	for k, v := range s.Environment {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	// Determine entrypoint and arguments
	callerArgs := make([]string, 0)
	if len(s.Entrypoint) > 0 {
		if len(s.Entrypoint) > 1 {
			args = append(args, "--entrypoint", s.Entrypoint[0])
			callerArgs = append(callerArgs, s.Entrypoint[1:]...)
		} else {
			tokens, _ := shlex.Split(s.Entrypoint[0])
			args = append(args, "--entrypoint", tokens[0])
			callerArgs = append(callerArgs, tokens[1:]...)
		}
	}
	// Add command
	if len(s.Command) > 0 {
		if len(s.Command) > 1 {
			callerArgs = append(callerArgs, s.Command...)
		} else {
			tokens, _ := shlex.Split(s.Command[0])
			callerArgs = append(callerArgs, tokens...)
		}
	}
	args = append(args, s.ImageName())
	if len(callerArgs) > 0 {
		args = append(args, callerArgs...)
	}
	return args
}

func (s Step) PullCommand() []string {
	return []string{"pull", s.ImageName()}
}
