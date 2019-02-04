package gantry // import "github.com/ad-freiburg/gantry"

import (
	"fmt"
	"os"
	"strings"

	"github.com/ad-freiburg/gantry/types"
)

type Service struct {
	BuildInfo   BuildInfo                 `json:"build"`
	Command     types.StringOrStringSlice `json:"command"`
	Entrypoint  types.StringOrStringSlice `json:"entrypoint"`
	Image       string                    `json:"image"`
	Ports       []string                  `json:"ports"`
	Volumes     []string                  `json:"volumes"`
	Environment []string                  `json:"environment"`
	DependsOn   types.StringSet           `json:"depends_on"`
	name        string
	color       int
}

type Step struct {
	Service
	Role   string          `json:"role"`
	After  types.StringSet `json:"after"`
	Detach bool            `json:"detach"`
}

func (s Step) Dependencies() (*types.StringSet, error) {
	r := types.StringSet{}
	for dep, _ := range s.After {
		r[dep] = true
	}
	for dep, _ := range s.DependsOn {
		r[dep] = true
	}
	return &r, nil
}

func (s *Service) SetName(name string) {
	s.name = name
	s.color = GetNextFriendlyColor()
}

func (s Service) Name() string {
	return s.name
}

func (s Service) ColoredName() string {
	return ApplyStyle(s.name, s.color)
}

func (s Service) ColoredContainerName() string {
	return ApplyStyle(s.ContainerName(), s.color)
}

func (s Service) ImageName() string {
	if s.Image != "" {
		return s.Image
	}
	return strings.Replace(strings.ToLower(s.name), " ", "_", -1)
}

func (s Service) RawContainerName() string {
	return strings.Replace(strings.ToLower(s.name), " ", "_", -1)
}

func (s Service) ContainerName() string {
	return fmt.Sprintf("%s_%s", ProjectName, strings.Replace(strings.ToLower(s.name), " ", "_", -1))
}

func (s Service) Runner() Runner {
	r := NewLocalRunner(s.ColoredContainerName(), os.Stdout, os.Stderr)
	return r
}
