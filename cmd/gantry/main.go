package main

import (
	"fmt"
	"os"

	"github.com/ad-freiburg/gantry/pipeline"
)

func main() {
	p, err := pipeline.NewPipeline(os.Args[1], os.Args[2])
	if err != nil {
		panic(err)
	}
	fmt.Printf("Pipeline:  %#v\n", p)
	if err = p.Check(); err != nil {
		fmt.Println(err.Error())
	}
}
