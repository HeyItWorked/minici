package main

import (
	"context"
	"fmt"
	"time"

	"github.com/liamnguyen/minici/internal/git"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("watching for new commits (10s)...")
	git.Watch(".", 2*time.Second, ctx, func(hash string) {
		fmt.Println("new commit:", hash)
	})
	fmt.Println("done")
}
