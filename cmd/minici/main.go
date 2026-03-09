package main

import (
	"fmt"

	"github.com/liamnguyen/minici/internal/git"
)

func main() {
	files, err := git.ChangedFiles(".", "48b85c6", "8348d6b")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("changed files:")
	for _, f := range files {
		fmt.Println(" ", f)
	}
}
