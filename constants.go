package gantry // import "github.com/ad-freiburg/gantry"

import (
	"log"
	"os"
)

// DockerCompose stores the default name of a docker compose file.
const DockerCompose string = "docker-compose.yml"

// GantryDef stores the default name of a gantry definition.
const GantryDef string = "gantry.def.yml"

// GantryEnv stores the default name of a gantry environment.
const GantryEnv string = "gantry.env.yml"

var (
	gantryLogger *PrefixedLogger
	// Version of the program
	Version = "no-version"
	// Verbose is a global verbosity flag
	Verbose = false
	// ProjectName stores the global prefix for networks and containers.
	ProjectName = ""
	// ForceWharfer is a global flag to force the usage of wharfer even
	// if the user could use docker directly.
	ForceWharfer = false
)

func init() {
	pipelineLogger = NewPrefixedLogger(
		ApplyStyle("gantry", STYLE_BOLD),
		log.New(os.Stderr, "", log.LstdFlags),
	)
}
