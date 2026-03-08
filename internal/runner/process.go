package runner

import (
	"bytes"
	"context"
	"os/exec"
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
