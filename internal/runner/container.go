package runner

import (
	"time"

)

// RunInContainer runs cmd inside a Docker container using the Docker CLI.
// Mounts workdir to /workspace inside the container and sets it as the working directory.
func RunInContainer(image, workdir string, cmd []string, timeout time.Duration) (stdout, stderr string, exitCode int, err error) {
	// build args for: docker run --rm -v workdir:/workspace -w /workspace image cmd...
	// --rm removes the container after exit so stopped containers don't accumulate
	// workdir:/workspace mounts the local dir into the container at /workspace
	args := []string{"run", "--rm", "-v", workdir + ":/workspace", "-w", "/workspace", image}

	// cmd... spreads the slice into individual args at the end
	args = append(args, cmd...)

	stdout, stderr, exitCode, err = RunCommand("docker", args, timeout)
	return
}
