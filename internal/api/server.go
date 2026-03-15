// server.go — HTTP layer for minici. Exposes build data over REST.
//
// Reused from: babel-shelf/go-bookshelf (handlers.go + main.go)
// Same pattern: handler reads request → talks to store → writes JSON response.
// respondJSON and setupRouter are near-identical copy-paste.
//
// What changed:
//   - Store is a struct field, not a package-level var (minici already uses struct-based stores)
//   - IDs are strings (UUIDs) not ints — no strconv.Atoi needed
//   - No Update/Delete — CI builds are immutable
//   - TriggerBuild runs the pipeline then saves, instead of decoding a JSON body
//
// ~60% reused structure, ~40% minici-specific changes.
package api

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"time"

	"github.com/liamnguyen/minici/internal/pipeline"
	"github.com/liamnguyen/minici/internal/store"
)

// embeds static/ directory into the binary at compile time — no external files needed at runtime
//
//go:embed static
var staticFiles embed.FS

// Server holds dependencies that handlers need.
// Same idea as LogStore holding dir or SQLiteStore holding db.
type Server struct {
	store    *store.SQLiteStore
	pipeline *pipeline.Pipeline
	workdir  string
}

// NewServer creates a Server with its dependencies wired in.
func NewServer(s *store.SQLiteStore, p *pipeline.Pipeline, workdir string) *Server {
	return &Server{store: s, pipeline: p, workdir: workdir}
}

// respondJSON writes a status code and JSON body to the response.
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status) // must come before writing body
	json.NewEncoder(w).Encode(data)
}

// SetupRouter maps URL patterns to handler methods.
func (s *Server) SetupRouter() http.Handler {
	// Go 1.22+ lets you put the HTTP method right in the pattern
	mux := http.NewServeMux()

	// serve the embedded HTML dashboard at root
	// fs.Sub strips the "static" prefix so index.html is served at / not /static/
	staticSub, _ := fs.Sub(staticFiles, "static")
	mux.Handle("GET /", http.FileServer(http.FS(staticSub)))

	mux.HandleFunc("GET /builds", s.ListBuilds)
	mux.HandleFunc("GET /builds/{id}", s.GetBuild)
	mux.HandleFunc("POST /builds", s.TriggerBuild)

	return mux
}

// ListBuilds — GET /builds
func (s *Server) ListBuilds(w http.ResponseWriter, r *http.Request) {
	// no input needed — just ask the store for everything
	builds, err := s.store.List()
	if err != nil {
		http.Error(w, "could not list builds", http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, builds)
}

// GetBuild — GET /builds/{id}
func (s *Server) GetBuild(w http.ResponseWriter, r *http.Request) {
	// ID is a string (UUID) — no strconv.Atoi like babel-shelf's int IDs
	id := r.PathValue("id")
	build, err := s.store.Get(id)
	if err != nil {
		http.Error(w, "build not found", http.StatusNotFound)
		return
	}
	respondJSON(w, http.StatusOK, build)
}

// TriggerBuild — POST /builds
func (s *Server) TriggerBuild(w http.ResponseWriter, r *http.Request) {
	// unlike babel-shelf's CreateBook, no JSON body to decode —
	// we run the pipeline and the result IS the thing we save
	result := pipeline.Run(s.pipeline, s.workdir, 5*time.Minute)

	id, err := s.store.Save(result)
	if err != nil {
		http.Error(w, "could not save build", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}
