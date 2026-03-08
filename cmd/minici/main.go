package main

import (
	"fmt"
	"os"
	"time"

	"github.com/liamnguyen/minici/internal/runner"
)

func main() {
	// RunCommand — buffered, returns everything at once
	stdout, stderr, code, err := runner.RunCommand("go", []string{"version"}, 10*time.Second)
	fmt.Printf("stdout=%q stderr=%q code=%d err=%v\n", stdout, stderr, code, err)

	// RunStreaming — lines printed to terminal as they arrive, each timestamped
	fmt.Println("\n[streaming]")
	code, err = runner.RunStreaming(
		"bash",
		[]string{"-c", "echo line1; sleep 1; echo line2; sleep 1; echo line3"},
		10*time.Second,
		os.Stdout,
	)
	fmt.Printf("[done] code=%d err=%v\n", code, err)
}
