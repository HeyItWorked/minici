package pipeline

import (
	"time"
	"github.com/liamnguyen/minici/internal/runner"
)

type StepResult struct {
	Name     string
	ExitCode int
	Stdout   string
	Stderr   string
	Err      error
}

type BuildResult struct {
	Pipeline string
	Steps    []StepResult
	Failed   bool
}

func Run(p *Pipeline, workdir string, timeout time.Duration) BuildResult {
	var results []StepResult
	failed := false

	for _, step := range p.Steps {
		stdout, stderr, code, err := runner.RunCommand("bash", []string{"-c", step.Command}, timeout)

		// accumulate every step result regardless of outcome — caller needs the full picture
		results = append(results, StepResult{Name: step.Name, ExitCode: code, Stdout: stdout, Stderr: stderr, Err: err})

		// err != nil catches crashes (binary not found etc.)
		// code != 0 catches failures like `go test` failing — RunCommand returns err=nil for non-zero exits
		if err != nil || code != 0 {
			failed = true
			break // stop on first failure
		}
	}

	return BuildResult{Pipeline: p.Name, Steps: results, Failed: failed}
}
