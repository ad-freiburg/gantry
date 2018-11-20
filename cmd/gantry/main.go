package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/ad-freiburg/gantry"
)

var (
	defFile   string
	envFile   string
	gantrydef string
	gantryenv string
)

func init() {
	gantrydef = "Gantrydef"
	gantryenv = "Gantryenv"
	defFileUsage := fmt.Sprintf("Explicit %s to use", gantrydef)
	envFileUsage := fmt.Sprintf("Explicit %s to use", gantryenv)
	flag.StringVar(&defFile, "file", gantrydef, defFileUsage)
	flag.StringVar(&defFile, "f", gantrydef, defFileUsage+" (shorthand)")
	flag.StringVar(&envFile, "env", gantryenv, envFileUsage)
	flag.StringVar(&envFile, "e", gantryenv, envFileUsage+" (shorthand)")
}

func main() {
	flag.Parse()

	log.Print("Load pipeline\n")
	p, err := gantry.NewPipeline(defFile, envFile)
	if err != nil {
		log.Fatal(err)
	}

	// Check for obvious errors
	log.Print("Check pipeline\n")
	if err = p.Check(); err != nil {
		log.Fatal(err)
	}

	// Build images
	log.Print("Prepare steps\n")
	if err = p.PrepareImages(); err != nil {
		log.Fatal(err)
	}

	// Execute step after step
	log.Print("Exec steps\n")
	if err = p.ExecuteSteps(); err != nil {
		log.Fatal(err)
	}
}
