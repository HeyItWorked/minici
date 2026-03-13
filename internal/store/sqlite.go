package store

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/liamnguyen/minici/internal/pipeline"

	// Blank import registers the "sqlite" driver with database/sql via its init() function.
	// Without this, sql.Open("sqlite", ...) fails with "unknown driver".
	_ "modernc.org/sqlite"
)

// SQLiteStore implements BuildStore using a SQLite database instead of JSON files.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens (or creates) a SQLite database at dbPath.
// Use ":memory:" for an in-memory database (useful in tests).
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// SQLite does NOT enforce foreign keys unless you flip this per-connection.
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		db.Close()
		return nil, err
	}

	err = createTables(db)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &SQLiteStore{db: db}, nil
}

// createTables sets up the schema. IF NOT EXISTS makes this safe to call on every startup.
func createTables(db *sql.DB) error {
	_, err := db.Exec(`
		-- one row per build execution
		CREATE TABLE IF NOT EXISTS builds (
			id         TEXT PRIMARY KEY,
			pipeline   TEXT NOT NULL,
			failed     BOOLEAN NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP  -- not in BuildResult; used for List() ordering
		);

		-- one row per step within a build
		CREATE TABLE IF NOT EXISTS steps (
			id        INTEGER PRIMARY KEY,
			build_id  TEXT NOT NULL REFERENCES builds(id),  -- links step to its parent build
			seq       INTEGER NOT NULL,                      -- preserves slice order from Go
			name      TEXT NOT NULL,
			exit_code INTEGER NOT NULL,
			stdout    TEXT NOT NULL DEFAULT '',
			stderr    TEXT NOT NULL DEFAULT '',
			error     TEXT  -- nullable: nil error in Go = NULL here
		);
	`)
	return err
}

// Save persists a build and all its steps in a single transaction.
func (s *SQLiteStore) Save(result pipeline.BuildResult) (string, error) {
	id := uuid.New().String()

	// --- begin transaction ---
	transaction, err := s.db.Begin()
	if err != nil {
		return "", err
	}
	// Rollback runs when this function exits, no matter what.
	// If Commit() already succeeded, Rollback() is a safe no-op.
	// If we returned early on error, Rollback() undoes all partial inserts.
	defer transaction.Rollback()

	// --- insert build ---
	_, err = transaction.Exec(
		"INSERT INTO builds (id, pipeline, failed) VALUES (?, ?, ?)",
		id, result.Pipeline, result.Failed,
	)
	if err != nil {
		return "", err
	}

	// --- insert each step ---
	for i, step := range result.Steps {
		// nil error → SQL NULL, non-nil → its message string
		var errStr *string
		if step.Err != nil {
			msg := step.Err.Error()
			errStr = &msg
		}

		_, err = transaction.Exec(
			"INSERT INTO steps (build_id, seq, name, exit_code, stdout, stderr, error) VALUES (?, ?, ?, ?, ?, ?, ?)",
			id, i, step.Name, step.ExitCode, step.Stdout, step.Stderr, errStr,
		)
		if err != nil {
			return "", err
		}
	}

	// --- commit transaction ---
	return id, transaction.Commit()
}

// Get retrieves a single build and its steps by ID.
// Depends on: querySteps (to fetch steps from the steps table)
func (s *SQLiteStore) Get(id string) (pipeline.BuildResult, error) {
	var result pipeline.BuildResult

	// --- fetch build ---
	// QueryRow = "I expect exactly one row". Returns a single row, not a cursor.
	row := s.db.QueryRow("SELECT pipeline, failed FROM builds WHERE id = ?", id)
	err := row.Scan(&result.Pipeline, &result.Failed)
	if err != nil {
		return pipeline.BuildResult{}, err
	}

	// --- fetch steps ---
	steps, err := s.querySteps(id)
	if err != nil {
		return pipeline.BuildResult{}, err
	}
	result.Steps = steps

	return result, nil
}

// List returns all builds with their steps, newest first.
// Depends on: querySteps (to fetch steps for each build)
//
// Flow: query all builds → for each build, query its steps → combine into []BuildResult
// The "id" column is fetched only to pass to querySteps — it's the glue between the two queries.
func (s *SQLiteStore) List() ([]pipeline.BuildResult, error) {
	// Query = "I expect multiple rows". Returns a cursor we iterate with rows.Next().
	rows, err := s.db.Query("SELECT id, pipeline, failed FROM builds ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	// rows holds a database cursor/lock — Close() releases it.
	// defer ensures it runs even if we return early on error.
	defer rows.Close()

	var results []pipeline.BuildResult
	// rows.Next() advances the cursor one row at a time.
	// Returns false when there are no more rows (like bufio.Scanner.Scan).
	for rows.Next() {
		var id string
		var result pipeline.BuildResult

		err = rows.Scan(&id, &result.Pipeline, &result.Failed)
		if err != nil {
			return nil, err
		}

		// --- fetch steps for this build ---
		steps, err := s.querySteps(id)
		if err != nil {
			return nil, err
		}
		result.Steps = steps

		results = append(results, result)
	}
	// rows.Err() catches errors that happened during iteration (e.g. network issues)
	return results, rows.Err()
}

// Close closes the database connection.
func (s *SQLiteStore) Close() error {
	err := s.db.Close()
	return err
}

// querySteps fetches all steps for a given build, ordered by seq.
// Used by: Get (single build) and List (all builds)
// This loop looks similar to List's loop — both follow the standard database/sql
// iteration pattern (Query → rows.Next → Scan → append). They scan different types
// so a shared helper isn't worth the complexity.
func (s *SQLiteStore) querySteps(buildID string) ([]pipeline.StepResult, error) {
	rows, err := s.db.Query(
		"SELECT name, exit_code, stdout, stderr, error FROM steps WHERE build_id = ? ORDER BY seq",
		buildID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []pipeline.StepResult
	for rows.Next() {
		var step pipeline.StepResult
		// sql.NullString handles nullable columns — has .Valid (was it NULL?) and .String (the value)
		var errStr sql.NullString

		err = rows.Scan(&step.Name, &step.ExitCode, &step.Stdout, &step.Stderr, &errStr)
		if err != nil {
			return nil, err
		}

		// reconstitute error from its string — loses type info, fine for display/history
		if errStr.Valid {
			step.Err = errors.New(errStr.String)
		}

		steps = append(steps, step)
	}
	return steps, rows.Err()
}
