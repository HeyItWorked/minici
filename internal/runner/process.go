package runner

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

// RunCommand runs cmd with args, kills it after timeout, and returns its output.
// err is only non-nil for unexpected failures (binary not found, etc.) — a non-zero
// exit code is returned as exitCode with err=nil.
func RunCommand(cmd string, args []string, timeout time.Duration) (stdout, stderr string, exitCode int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout) // cancel kills process after timeout
	defer cancel()

	c := exec.CommandContext(ctx, cmd, args...) // process tied to ctx — dies when ctx cancels

	// wire up buffers so stdout and stderr are captured separately
	var outBuf, errBuf bytes.Buffer
	c.Stdout = &outBuf
	c.Stderr = &errBuf

	err = c.Run()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok { // "if err is secretly an *exec.ExitError"
			// process ran but exited with a non-zero code — expected, not a crash
			exitCode = exitErr.ExitCode()
			err = nil
		} else {
			// real failure: binary not found, timeout, permission denied, etc.
			return
		}
	}

	stdout = outBuf.String()
	stderr = errBuf.String()

	return
}

// RunStreaming runs cmd and writes each output line to out as it arrives.
// Each line is prefixed with a timestamp: [15:04:05] line content
// Unlike RunCommand, output is not buffered — lines appear as the process produces them.
func RunStreaming(cmd string, args []string, timeout time.Duration, out io.Writer) (exitCode int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout) // kills process after timeout
	defer cancel()

	c := exec.CommandContext(ctx, cmd, args...) // process tied to ctx — dies when ctx cancels

	stdoutPipe, err := c.StdoutPipe() // live connection to process stdout
	if err != nil {
		return
	}
	stderrPipe, err := c.StderrPipe() // live connection to process stderr
	if err != nil {
		return
	}

	c.Start() // start without blocking — we need to read pipes while it runs

	var wg sync.WaitGroup

	scanPipe := func(pipe io.ReadCloser) { // closure — captures wg and out from outer scope
		defer wg.Done()                    // decrement counter when this goroutine finishes
		scanner := bufio.NewScanner(pipe)  // wraps pipe to read one line at a time
		for scanner.Scan() {               // Scan() advances to next line, returns false when pipe closes
			fmt.Fprintf(out, "[%s] %s\n", time.Now().Format("15:04:05"), scanner.Text())
		}
	}

	wg.Add(2)               // expecting 2 goroutines to finish
	go scanPipe(stdoutPipe) // read stdout concurrently
	go scanPipe(stderrPipe) // read stderr concurrently

	wg.Wait()      // block until both goroutines finish draining the pipes
	err = c.Wait() // then wait for the process to fully exit

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok { // "if err is secretly an *exec.ExitError"
			exitCode = exitErr.ExitCode()              // process exited non-zero — expected, not a crash
			err = nil
		} else {
			return // real failure: binary not found, timeout, permission denied, etc.
		}
	}

	return
}
