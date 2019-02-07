package gantry // import "github.com/ad-freiburg/gantry"

import (
	"github.com/ad-freiburg/gantry/types"
)

type BuildInfo struct {
	Context    string                  `json:"context"`
	Dockerfile string                  `json:"dockerfile"`
	Args       types.MappingWithEquals `json:"args"`
}
