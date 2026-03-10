package main

import (
	"fmt"
	"time"

	"github.com/liamnguyen/minici/internal/pipeline"
)

func main() {
	p, err := pipeline.Load("pipeline.yaml")
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	result := pipeline.Run(p, ".", 30*time.Second)
	fmt.Printf("pipeline: %s  failed: %v\n", result.Pipeline, result.Failed)
	for _, step := range result.Steps {
		fmt.Printf("  [%s] exit=%d\n", step.Name, step.ExitCode)
		if step.Stdout != "" {
			fmt.Printf("    stdout: %s\n", step.Stdout)
		}
		if step.Err != nil {
			fmt.Printf("    err: %v\n", step.Err)
		}
	}
}
