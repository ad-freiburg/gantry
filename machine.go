package gantry // import "github.com/ad-freiburg/gantry"

import "github.com/ad-freiburg/gantry/types"

type Machine struct {
	Host  string
	Roles types.StringSet
	Paths Paths
}

type Paths struct {
	Input   map[string]string
	Output  map[string]string
	Scratch string
}
