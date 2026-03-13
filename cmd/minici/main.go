package main

import (
	"fmt"

	"github.com/liamnguyen/minici/internal/pipeline"
	"github.com/liamnguyen/minici/internal/store"
)

func main() {
	s, err := store.NewSQLiteStore("data/minici.db")
	if err != nil {
		fmt.Println("error creating store:", err)
		return
	}

	// save a fake build result
	result := pipeline.BuildResult{
		Pipeline: "my-app",
		Failed:   false,
		Steps: []pipeline.StepResult{
			{Name: "test", ExitCode: 0, Stdout: "ok"},
			{Name: "build", ExitCode: 0},
		},
	}

	id, err := s.Save(result)
	if err != nil {
		fmt.Println("error saving:", err)
		return
	}
	fmt.Println("saved:", id)

	// get it back by id
	got, err := s.Get(id)
	if err != nil {
		fmt.Println("error getting:", err)
		return
	}
	fmt.Printf("got: pipeline=%s failed=%v steps=%d\n", got.Pipeline, got.Failed, len(got.Steps))

	// list all
	all, err := s.List()
	if err != nil {
		fmt.Println("error listing:", err)
		return
	}
	fmt.Printf("total builds: %d\n", len(all))
}
