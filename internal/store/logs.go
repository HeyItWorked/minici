package store

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// LogStore manages log files for builds.
// Methods on this struct (WriteLog, TailLog) use l.dir to know where files live —
// same pattern as SQLiteStore using s.db. The struct is the object, methods are its behavior.
type LogStore struct {
	dir string
}

// NewLogStore creates a LogStore and ensures the log directory exists.
func NewLogStore(dir string) (*LogStore, error) {
	// MkdirAll creates the directory and any missing parents (like mkdir -p)
	// 0755 = owner can read/write/execute, everyone else can read/execute
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}
	return &LogStore{dir: dir}, nil
}

// WriteLog appends a single log line to <dir>/<buildID>.log
func (l *LogStore) WriteLog(buildID string, line string) error {
	path := filepath.Join(l.dir, buildID+".log")

	// Flags are bitwise OR'd into one int: append + create-if-missing + write-only
	// 0644 = owner read/write, others read-only
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close() // Go's version of Python's "with open() as f" — auto-closes when function exits

	_, err = fmt.Fprintln(file, line)
	return err
}

// TailLog streams a build's log file to out, then watches for new lines (like tail -f).
// Depends on: WriteLog (writes the lines that TailLog reads)
// Stops when ctx is cancelled (e.g. client disconnects or build finishes).
// TODO: implement catch-up read + polling loop (see issue #13)
func (l *LogStore) TailLog(buildID string, ctx context.Context, out io.Writer) error {
	return nil
}
