package gitwatch

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Status is the result of checking a single repository.
type Status struct {
	Repo    string
	Last    time.Time
	Elapsed time.Duration
	Err     error
}

// Watcher monitors one or more git repositories.
type Watcher struct {
	repos  []string
	mu     sync.Mutex
	cache  map[string]time.Time
}

// NewWatcher creates a Watcher for the given list of repo paths.
func NewWatcher(repos []string) *Watcher {
	return &Watcher{
		repos: repos,
		cache: make(map[string]time.Time),
	}
}

// Check queries each repo for its last commit time and returns a Status slice.
func (w *Watcher) Check() []Status {
	w.mu.Lock()
	defer w.mu.Unlock()

	results := make([]Status, 0, len(w.repos))
	for _, repo := range w.repos {
		t, err := LastCommitTime(repo)
		s := Status{Repo: repo, Err: err}
		if err == nil {
			s.Last = t
			s.Elapsed = time.Since(t)
			w.cache[repo] = t
		} else if cached, ok := w.cache[repo]; ok {
			// Fall back to cached value on transient errors.
			s.Last = cached
			s.Elapsed = time.Since(cached)
			s.Err = nil
		}
		results = append(results, s)
	}
	return results
}

// LastCommitTime shells out to `git log` to get the Unix timestamp of the
// most recent commit in the given repo directory.
func LastCommitTime(repoPath string) (time.Time, error) {
	cmd := exec.Command("git", "-C", repoPath, "log", "-1", "--format=%ct")
	out, err := cmd.Output()
	if err != nil {
		return time.Time{}, fmt.Errorf("git log in %q: %w", repoPath, err)
	}

	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("no commits found in %q", repoPath)
	}

	unixSec, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse timestamp %q: %w", trimmed, err)
	}

	return time.Unix(unixSec, 0), nil
}
