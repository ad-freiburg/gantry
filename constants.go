package gantry // import "github.com/ad-freiburg/gantry"

import (
	"log"
	"os"
)

const DockerCompose string = "docker-compose.yml"
const GantryDef string = "gantry.def.yml"
const GantryEnv string = "gantry.env.yml"
const Version = "0.0.1a1"

var (
	gantryLogger *PrefixedLogger
)

func init() {
	pipelineLogger = NewPrefixedLogger(
		ApplyStyle("gantry", STYLE_BOLD),
		log.New(os.Stderr, "", log.LstdFlags),
	)
}
