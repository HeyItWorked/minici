package main

import (
	"fmt"

	"github.com/liamnguyen/minici/internal/pipeline"
)

func main() {
	p, err := pipeline.Load("pipeline.yaml")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("pipeline: %s\n", p.Name)
	for _, step := range p.Steps {
		fmt.Printf("  step: %s → %s\n", step.Name, step.Command)
	}
}
