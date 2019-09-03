package gantry // import "github.com/ad-freiburg/gantry"

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ad-freiburg/gantry/types"
	"github.com/google/shlex"
)

// Service provides a service definition from docker-compose.
type Service struct {
	BuildInfo   BuildInfo                 `json:"build"`
	Command     types.StringOrStringSlice `json:"command"`
	Entrypoint  types.StringOrStringSlice `json:"entrypoint"`
	Image       string                    `json:"image"`
	Ports       []string                  `json:"ports"`
	Volumes     []string                  `json:"volumes"`
	Environment types.StringMap           `json:"environment"`
	DependsOn   types.StringSet           `json:"depends_on"`
	Restart     string                    `json:"restart"`
	Name        string
	Meta        ServiceMeta
	color       int
}

// Step provides an extended service.
type Step struct {
	Service
	Role   string          `json:"role"`
	After  types.StringSet `json:"after"`
	Detach bool            `json:"detach"`
}

// Dependencies returns all steps needed for running s.
func (s Step) Dependencies() types.StringSet {
	r := types.StringSet{}
	for dep := range s.After {
		r[dep] = true
	}
	for dep := range s.DependsOn {
		r[dep] = true
	}
	return r
}

// InitColor initializes the color of s.
func (s *Service) InitColor() {
	s.color = GetNextFriendlyColor()
}

// ColoredName returns the name of s with color applied.
func (s Service) ColoredName() string {
	return ApplyAnsiStyle(s.Name, s.color)
}

// ColoredContainerName returns the container name of s with color applied.
func (s Service) ColoredContainerName() string {
	return ApplyAnsiStyle(s.ContainerName(), s.color)
}

// ImageName returns the name of the image of s.
// The name of the step is used if non is specified.
func (s Service) ImageName() string {
	if s.Image != "" {
		return s.Image
	}
	return strings.ReplaceAll(strings.ToLower(s.Name), " ", "_")
}

// RawContainerName returns the name for a container of s.
func (s Service) RawContainerName() string {
	return strings.ReplaceAll(strings.ToLower(s.Name), " ", "_")
}

// ContainerName returns the name for a container of s prefixed with the
// current project name.
func (s Service) ContainerName() string {
	return fmt.Sprintf("%s_%s", ProjectName, strings.ReplaceAll(strings.ToLower(s.Name), " ", "_"))
}

// IsBuildable returns whether or not the step can be build.
func (s Step) IsBuildable() bool {
	return s.BuildInfo.Dockerfile != "" || s.BuildInfo.Context != ""
}

// BuildCommand returns the command to build a new image for s.
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
		if v == nil {
			t := os.Getenv(k)
			v = &t
		}
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, *v))
	}
	args = append(args, s.BuildInfo.Context)
	return args
}

// RunCommand returns the command to run an instance of step s.
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
	if s.Restart != "" {
		args = append(args, "--restart")
		args = append(args, s.Restart)
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
		if v == nil {
			t := os.Getenv(k)
			v = &t
		}
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, *v))
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

// IsPullable returns whether or not a image is pulled for this step.
func (s Step) IsPullable() bool {
	return !s.IsBuildable()
}

// PullCommand returns the command to pull the image for step s.
func (s Step) PullCommand() []string {
	return []string{"pull", s.ImageName()}
}
