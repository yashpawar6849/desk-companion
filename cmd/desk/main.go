package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/user/desk-companion/internal/daemon"
	"github.com/user/desk-companion/internal/gitwatch"
	"github.com/user/desk-companion/internal/notify"
	"github.com/user/desk-companion/internal/pomodoro"
	"github.com/user/desk-companion/internal/ui"
)

func main() {
	// ── Flags ────────────────────────────────────────────────────────────────
	var (
		reposFlag    = flag.String("repo", ".", "Comma-separated list of git repo paths to watch")
		workFlag     = flag.Duration("work", 25*time.Minute, "Pomodoro work duration")
		shortFlag    = flag.Duration("short-break", 5*time.Minute, "Pomodoro short break duration")
		longFlag     = flag.Duration("long-break", 15*time.Minute, "Pomodoro long break duration")
		nudgeFlag    = flag.Duration("nudge", 30*time.Minute, "Commit nudge threshold")
		daemonFlag   = flag.Bool("daemon", false, "Run as a headless background daemon (no TUI)")
		versionFlag  = flag.Bool("version", false, "Print version and exit")
	)
	flag.Parse()

	if *versionFlag {
		fmt.Println("desk-companion v0.1.0")
		os.Exit(0)
	}

	repos := parseRepos(*reposFlag)

	// ── Notifier ─────────────────────────────────────────────────────────────
	notifier, err := notify.New()
	if err != nil {
		log.Printf("Warning: desktop notifications unavailable (%v) — falling back to console output", err)
		notifier = notify.LogNotifier{}
	}

	// ── Daemon mode ───────────────────────────────────────────────────────────
	if *daemonFlag {
		cfg := daemon.Config{
			Repos:        repos,
			CommitNudge:  *nudgeFlag,
			WorkDuration: *workFlag,
			ShortBreak:   *shortFlag,
			LongBreak:    *longFlag,
			TickInterval: time.Second,
		}
		if err := daemon.Run(context.Background(), cfg); err != nil {
			log.Fatal(err)
		}
		return
	}

	// ── TUI mode (default) ────────────────────────────────────────────────────
	pom := pomodoro.New(*workFlag, *shortFlag, *longFlag)
	watcher := gitwatch.NewWatcher(repos)

	// Pre-populate git statuses before first render.
	_ = watcher.Check()

	model := ui.New(pom, watcher, notifier, *nudgeFlag)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

// parseRepos splits a comma-separated repo flag into a cleaned slice.
func parseRepos(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		out = []string{"."}
	}
	return out
}
