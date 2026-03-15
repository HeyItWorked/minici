package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/liamnguyen/minici/internal/api"
	"github.com/liamnguyen/minici/internal/pipeline"
	"github.com/liamnguyen/minici/internal/store"
)

func main() {
	// open (or create) the database — same call as before, just not throwaway anymore
	s, err := store.NewSQLiteStore("data/minici.db")
	if err != nil {
		fmt.Println("error creating store:", err)
		os.Exit(1)
	}
	defer s.Close()

	// load pipeline config from working directory
	p, err := pipeline.Load("pipeline.yaml")
	if err != nil {
		fmt.Println("error loading pipeline:", err)
		os.Exit(1)
	}

	// wire up the server — store for persistence, pipeline for triggering builds
	srv := api.NewServer(s, p, ".")
	router := srv.SetupRouter()

	// start serving — blocks until the process is killed
	fmt.Println("minici running on http://localhost:8080")
	err = http.ListenAndServe(":8080", router)
	fmt.Println("server failed:", err)
	os.Exit(1)
}
