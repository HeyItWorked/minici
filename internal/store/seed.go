package store

import "github.com/liamnguyen/minici/internal/pipeline"

// SeedDemo shoves some fake builds into the database so the dashboard
// isn't empty when someone visits. only seeds if the db has nothing in it.
func SeedDemo(s *SQLiteStore) error {
	// don't double-seed — if there's already data, leave it alone
	existing, err := s.List()
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}

	// these are just hardcoded builds with realistic-ish output.
	// order matters — they get inserted oldest-first so the dashboard
	// shows them newest-first (List() does ORDER BY created_at DESC)

	// build 1 — clean run, everything passes
	s.Save(pipeline.BuildResult{
		Pipeline: "my-app",
		Failed:   false,
		Steps: []pipeline.StepResult{
			{Name: "lint", ExitCode: 0, Stdout: "checking style...\nno issues found\n"},
			{Name: "test", ExitCode: 0, Stdout: "ok  \tgithub.com/user/my-app/internal/api\t0.012s\nok  \tgithub.com/user/my-app/internal/store\t0.008s\n"},
			{Name: "build", ExitCode: 0, Stdout: "compiling...\n"},
		},
	})

	// build 2 — test blows up, auth token expired
	s.Save(pipeline.BuildResult{
		Pipeline: "my-app",
		Failed:   true,
		Steps: []pipeline.StepResult{
			{Name: "lint", ExitCode: 0, Stdout: "checking style...\nno issues found\n"},
			{Name: "test", ExitCode: 1,
				Stdout: "--- FAIL: TestUserAuth (0.02s)\n    auth_test.go:42: expected token to be valid, got expired\nFAIL\tgithub.com/user/my-app/internal/auth\t0.031s\n",
				Stderr: "FAIL\n",
			},
		},
	})

	// build 3 — fixed the auth thing, back to green
	s.Save(pipeline.BuildResult{
		Pipeline: "my-app",
		Failed:   false,
		Steps: []pipeline.StepResult{
			{Name: "lint", ExitCode: 0, Stdout: "checking style...\nno issues found\n"},
			{Name: "test", ExitCode: 0, Stdout: "ok  \tgithub.com/user/my-app/internal/api\t0.011s\nok  \tgithub.com/user/my-app/internal/auth\t0.009s\nok  \tgithub.com/user/my-app/internal/store\t0.007s\n"},
			{Name: "build", ExitCode: 0, Stdout: "compiling...\n"},
		},
	})

	// build 4 — lint catches a missing doc comment
	s.Save(pipeline.BuildResult{
		Pipeline: "my-app",
		Failed:   true,
		Steps: []pipeline.StepResult{
			{Name: "lint", ExitCode: 1,
				Stdout: "checking style...\n",
				Stderr: "internal/api/server.go:47:2: exported function ServeHTTP should have comment or be unexported (golint)\n",
			},
		},
	})

	// build 5 — added the comment, all clear
	s.Save(pipeline.BuildResult{
		Pipeline: "my-app",
		Failed:   false,
		Steps: []pipeline.StepResult{
			{Name: "lint", ExitCode: 0, Stdout: "checking style...\nno issues found\n"},
			{Name: "test", ExitCode: 0, Stdout: "ok  \tgithub.com/user/my-app/internal/api\t0.013s\nok  \tgithub.com/user/my-app/internal/auth\t0.010s\nok  \tgithub.com/user/my-app/internal/store\t0.008s\n"},
			{Name: "build", ExitCode: 0, Stdout: "compiling...\n"},
		},
	})

	return nil
}
