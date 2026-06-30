# Desk Companion

A terminal-based background daemon that keeps you company while you work. Tracks time since your last commit, runs a Pomodoro timer with ambient ASCII art, and nudges you with desktop notifications when it matters.

Built in Go to learn long-running process design, signal handling, and cross-platform process invocation.

---

## Features

| Feature | Details |
|---|---|
| **Daemon core** | Ticker loop, graceful `Ctrl+C` / `SIGTERM` shutdown via `context` |
| **Git watcher** | Tracks time since last commit, alerts when threshold exceeded |
| **Pomodoro timer** | Work → Short Break → Long Break state machine (standard 25/5/15) |
| **Desktop notifications** | Windows (PowerShell toast), Linux (`notify-send`), macOS (`osascript`) |
| **Bubble Tea TUI** | Live countdown, git status, time-of-day ASCII art |

---

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git (must be in PATH)

### Install & Run

```bash
# Clone / navigate to the project
cd desk-companion

# Fetch dependencies
go mod tidy

# Run (TUI mode, watches current directory repo)
go run ./cmd/desk

# Watch a specific repo
go run ./cmd/desk --repo /path/to/your/project

# Watch multiple repos
go run ./cmd/desk --repo /path/one,/path/two

# Build a binary
go build -o desk ./cmd/desk
./desk
```

---

## Flags

| Flag | Default | Description |
|---|---|---|
| `--repo` | `.` | Comma-separated git repo paths to watch |
| `--work` | `25m` | Pomodoro work session duration |
| `--short-break` | `5m` | Short break duration |
| `--long-break` | `15m` | Long break duration |
| `--nudge` | `30m` | Alert if no commit within this interval |
| `--daemon` | `false` | Run headless (no TUI, prints to stdout) |
| `--version` | — | Print version and exit |

---

## TUI Keybinds

| Key | Action |
|---|---|
| `Space` | Pause / resume Pomodoro |
| `s` | Skip to next phase |
| `r` | Reset current phase |
| `q` / `Ctrl+C` | Quit |

---

## Project Structure

```
desk-companion/
├── cmd/
│   └── desk/
│       └── main.go          # wiring, flag parsing, starts the daemon or TUI
├── internal/
│   ├── daemon/
│   │   └── daemon.go        # ticker loop, signal handling, lifecycle
│   ├── gitwatch/
│   │   └── gitwatch.go      # last-commit-time logic
│   ├── notify/
│   │   ├── notify.go        # Notifier interface + Noop/Log impls
│   │   ├── notify_windows.go # PowerShell toast
│   │   ├── notify_darwin.go  # osascript
│   │   └── notify_linux.go   # notify-send
│   ├── pomodoro/
│   │   └── pomodoro.go      # work/break state machine
│   └── ui/
│       ├── ascii.go         # time-of-day ASCII art scenes
│       └── model.go         # Bubble Tea model/update/view
├── go.mod
└── README.md
```

---

## How It Works

### Daemon Core
Uses `time.NewTicker` for 1-second resolution. `signal.Notify` listens for `os.Interrupt` and `syscall.SIGTERM`, cancelling a `context.Context` that propagates shutdown to all goroutines.

### Git Watcher
Shells out to `git -C <path> log -1 --format=%ct` (Unix timestamp of last commit). Results are cached so transient `git` failures don't lose state.

### Desktop Notifications
Dispatched via Go's build tag convention (`_windows.go`, `_darwin.go`, `_linux.go`). On Windows, uses `[Windows.UI.Notifications.ToastNotificationManager]` via PowerShell with a console-print fallback.

### Pomodoro
Classic 25/5/15 state machine. After every 4 work sessions, a long break fires. The timer is thread-safe (sync.Mutex) and used by both the TUI and headless daemon.

### TUI
Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea). A `tea.Tick` command fires every second, advancing the Pomodoro and refreshing the view. Git repos are re-queried once per minute.

---

## Extending

- **Build status ticker** — add a `buildwatch` package that polls GitHub Actions via their REST API (`/repos/{owner}/{repo}/actions/runs`) and prints pass/fail to the TUI.
- **Config file** — add `~/.config/desk-companion/config.toml` support for persistent defaults.
- **Multiple themes** — swap `lipgloss` color palettes per time of day.
