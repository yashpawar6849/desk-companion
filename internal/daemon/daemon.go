package daemon

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/user/desk-companion/internal/gitwatch"
	"github.com/user/desk-companion/internal/notify"
	"github.com/user/desk-companion/internal/pomodoro"
)

// Config holds runtime configuration for the daemon.
type Config struct {
	Repos         []string
	CommitNudge   time.Duration // alert if no commit within this duration
	WorkDuration  time.Duration
	ShortBreak    time.Duration
	LongBreak     time.Duration
	TickInterval  time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Repos:        []string{"."},
		CommitNudge:  30 * time.Minute,
		WorkDuration: 25 * time.Minute,
		ShortBreak:   5 * time.Minute,
		LongBreak:    15 * time.Minute,
		TickInterval: time.Second,
	}
}

// Run starts the daemon loop. It blocks until the context is cancelled or a
// termination signal is received.
func Run(ctx context.Context, cfg Config) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Intercept OS signals for graceful shutdown.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, termSignals()...)
	go func() {
		select {
		case sig := <-sigCh:
			fmt.Fprintf(os.Stderr, "\nReceived signal %s — shutting down...\n", sig)
			cancel()
		case <-ctx.Done():
		}
	}()

	notifier, err := notify.New()
	if err != nil {
		log.Printf("Warning: notifications unavailable: %v", err)
		notifier = notify.Noop{}
	}

	watcher := gitwatch.NewWatcher(cfg.Repos)
	pom := pomodoro.New(cfg.WorkDuration, cfg.ShortBreak, cfg.LongBreak)

	ticker := time.NewTicker(cfg.TickInterval)
	defer ticker.Stop()

	// Track nudge state so we don't spam.
	nudgesSent := make(map[string]time.Time)

	fmt.Println("🖥️  Desk Companion started. Press Ctrl+C to quit.")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Goodbye! 👋")
			return nil

		case t := <-ticker.C:
			// --- Pomodoro tick ---
			if stateChanged, newState := pom.Tick(); stateChanged {
				title, body := pomodoroNotification(newState)
				if nerr := notifier.Send(title, body); nerr != nil {
					log.Printf("notify error: %v", nerr)
				}
			}

			// --- Git watcher tick (check every minute) ---
			if t.Second() == 0 {
				statuses := watcher.Check()
				for _, s := range statuses {
					if s.Err != nil {
						continue
					}
					if s.Elapsed > cfg.CommitNudge {
						last, alerted := nudgesSent[s.Repo]
						// Re-alert at most once per CommitNudge interval.
						if !alerted || time.Since(last) > cfg.CommitNudge {
							msg := fmt.Sprintf("No commit in %s — time to push something!", formatDuration(s.Elapsed))
							if nerr := notifier.Send("⏰ Commit nudge", msg); nerr != nil {
								log.Printf("notify error: %v", nerr)
							}
							nudgesSent[s.Repo] = time.Now()
							fmt.Printf("[git] %s: %s\n", s.Repo, msg)
						}
					}
				}
			}
		}
	}
}

func pomodoroNotification(state pomodoro.State) (title, body string) {
	switch state {
	case pomodoro.StateShortBreak:
		return "🍵 Short break!", "5 minutes — stretch, breathe, hydrate."
	case pomodoro.StateLongBreak:
		return "🎉 Long break!", "15 minutes — you earned it. Step away."
	case pomodoro.StateWork:
		return "💪 Back to work!", "Focus time. You've got this."
	default:
		return "Pomodoro", "Timer update."
	}
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	m := (d % time.Hour) / time.Minute
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
