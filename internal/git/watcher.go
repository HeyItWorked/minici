package git

import (
	"context"
	"time"
)

// Watch polls repoPath every interval and calls onChange when the commit hash changes.
// Runs until ctx is cancelled.
func Watch(repoPath string, interval time.Duration, ctx context.Context, onChange func(hash string)) {
	lastHash, _, _, _ := GetLatestCommit(repoPath)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		// select blocks until one of its cases is ready — like asyncio.wait in Python
		// picks whichever channel fires first: ticker (time to check) or ctx.Done (time to stop)
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			newHash, _, _, err := GetLatestCommit(repoPath)
			if err != nil {
				// intentionally skip bad ticks — transient git errors shouldn't crash the watcher
				// Watch has no return value so errors can't be surfaced; try again next interval
				continue
			}
			// every N seconds check → if hash changed → call onChange
			// if nothing changed, onChange never gets called that tick
			if newHash != lastHash {
				onChange(newHash)
				lastHash = newHash
			}
		}
	}
}
