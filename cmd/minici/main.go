package main

import (
	"fmt"
	"time"

	"github.com/liamnguyen/minici/internal/runner"
)

func main() {
	stdout, stderr, code, err := runner.RunCommand("go", []string{"version"}, 10*time.Second)
	fmt.Printf("stdout=%q stderr=%q code=%d err=%v\n", stdout, stderr, code, err)
}
