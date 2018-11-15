package main

import (
	"log"
	"os"

	"github.com/ad-freiburg/gantry"
)

func main() {
	log.Print("Load pipeline\n")
	p, err := gantry.NewPipeline(os.Args[1], os.Args[2])
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
