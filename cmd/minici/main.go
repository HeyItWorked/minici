package main

import (
	"fmt"
	"time"

	"github.com/liamnguyen/minici/internal/runner"
)

func main() {
	stdout, stderr, code, err := runner.RunInContainer(
		"golang:1.22",
		"/workspaces/minici",
		[]string{"go", "version"},
		30*time.Second,
	)
	fmt.Printf("exit=%d err=%v\n", code, err)
	fmt.Printf("stdout: %s\n", stdout)
	if stderr != "" {
		fmt.Printf("stderr: %s\n", stderr)
	}
}
