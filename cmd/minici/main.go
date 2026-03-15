package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/liamnguyen/minici/internal/api"
	"github.com/liamnguyen/minici/internal/pipeline"
	"github.com/liamnguyen/minici/internal/store"
)

// set DEMO=1 to pre-seed the database and disable the trigger endpoint.
// intended for public-facing deployments where we don't want anyone
// running real commands on the host.

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

	// if DEMO=1, seed some fake builds so the dashboard has stuff to show
	demo := os.Getenv("DEMO") == "1"
	if demo {
		if err := store.SeedDemo(s); err != nil {
			fmt.Println("error seeding demo data:", err)
			os.Exit(1)
		}
		fmt.Println("demo mode — seeded builds, trigger disabled")
	}

	// wire up the server — store for persistence, pipeline for triggering builds
	srv := api.NewServer(s, p, ".", demo)
	router := srv.SetupRouter()

	// start serving — blocks until the process is killed
	fmt.Println("minici running on http://localhost:8080")
	err = http.ListenAndServe(":8080", router)
	fmt.Println("server failed:", err)
	os.Exit(1)
}
