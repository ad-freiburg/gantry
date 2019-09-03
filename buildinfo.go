package gantry // import "github.com/ad-freiburg/gantry"

import (
	"github.com/ad-freiburg/gantry/types"
)

// BuildInfo represents the build-keyword in a docker-compse.yml.
type BuildInfo struct {
	Context    string          `json:"context"`
	Dockerfile string          `json:"dockerfile"`
	Args       types.StringMap `json:"args"`
}
