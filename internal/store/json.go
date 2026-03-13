package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/liamnguyen/minici/internal/pipeline"
)

// BuildStore defines the interface for persisting build results.
// Any type with these three methods satisfies the interface — no explicit declaration needed.
// This lets us swap JSONStore for SQLiteStore later without changing the rest of the app.
type BuildStore interface {
	Save(result pipeline.BuildResult) (string, error)
	Get(id string) (pipeline.BuildResult, error)
	List() ([]pipeline.BuildResult, error)
}

// JSONStore implements BuildStore using JSON files on disk.
// Each build is stored as <dir>/<uuid>.json
type JSONStore struct {
	dir string // directory where build files are stored
}

// NewJSONStore creates a JSONStore and ensures the storage directory exists.
func NewJSONStore(dir string) (*JSONStore, error) {
	// MkdirAll is like mkdir -p — creates all parent dirs, no error if already exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &JSONStore{dir: dir}, nil
}

// Save marshals the build result to JSON and writes it to disk with a generated UUID as the filename.
func (s *JSONStore) Save(result pipeline.BuildResult) (string, error) {
	uid := uuid.New().String()
	data, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	err = os.WriteFile(filepath.Join(s.dir, uid+".json"), data, 0644)
	if err != nil {
		return "", err
	}
	return uid, nil
}

// Get reads and unmarshals a single build result by ID.
func (s *JSONStore) Get(id string) (pipeline.BuildResult, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, id+".json"))
	if err != nil {
		return pipeline.BuildResult{}, err
	}
	var result pipeline.BuildResult
	if err := json.Unmarshal(data, &result); err != nil {
		return pipeline.BuildResult{}, err
	}
	return result, nil
}

// List returns all build results stored in the directory.
// Skips files that fail to parse — non-.json files are ignored entirely.
func (s *JSONStore) List() ([]pipeline.BuildResult, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	var results []pipeline.BuildResult
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".json")
		result, err := s.Get(id)
		if err != nil {
			continue // skip unreadable files
		}
		results = append(results, result)
	}
	return results, nil
}
