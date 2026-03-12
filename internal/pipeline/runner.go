package pipeline

import (
	"sync"
	"time"

	"github.com/liamnguyen/minici/internal/runner"
	"golang.org/x/sync/errgroup"
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

	i := 0
	for i < len(p.Steps) {
		step := p.Steps[i]

		if !step.Parallel {
			var stdout, stderr string
  			var code int
  			var err error

			if step.Image != "" {
				// RunInContainer
 				stdout, stderr, code, err = runner.RunInContainer(step.Image, workdir, []string{"bash", "-c", step.Command}, timeout)

			} else {
				// sequential step — run and wait before moving on
				stdout, stderr, code, err = runner.RunCommand("bash", []string{"-c", step.Command}, timeout)
			}


			// accumulate every result regardless of outcome — caller needs the full picture
			results = append(results, StepResult{Name: step.Name, ExitCode: code, Stdout: stdout, Stderr: stderr, Err: err})

			// err != nil catches crashes, code != 0 catches failures like `go test` failing
			if err != nil || code != 0 {
				failed = true
				break
			}
			i++
		} else {
			// collect consecutive parallel steps into a batch
			var batch []Step
			for i < len(p.Steps) && p.Steps[i].Parallel {
				batch = append(batch, p.Steps[i])
				i++
			}

			// errgroup launches each step as a goroutine and collects errors — like asyncio.gather() in Python
			g := errgroup.Group{}
			var mu sync.Mutex // protects results slice from concurrent writes

			for _, s := range batch {
				step := s // capture loop var — each goroutine needs its own copy, not a shared reference
				g.Go(func() error {
					var stdout, stderr string
  					var code int
  					var err error

					if step.Image != "" {
						// RunInContainer
						stdout, stderr, code, err = runner.RunInContainer(step.Image, workdir, []string{"bash", "-c", step.Command}, timeout)

					} else {
						// no image — run locally
						stdout, stderr, code, err = runner.RunCommand("bash", []string{"-c", step.Command}, timeout)
					}

					mu.Lock()
					results = append(results, StepResult{Name: step.Name, ExitCode: code, Stdout: stdout, Stderr: stderr, Err: err})
					mu.Unlock()

					if err != nil || code != 0 {
						return err
					}
					return nil
				})
			}

			// Wait blocks until all goroutines finish, returns first error if any
			if err := g.Wait(); err != nil {
				failed = true
				break
			}
		}
	}

	return BuildResult{Pipeline: p.Name, Steps: results, Failed: failed}
}
