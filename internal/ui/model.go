package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/user/desk-companion/internal/gitwatch"
	"github.com/user/desk-companion/internal/notify"
	"github.com/user/desk-companion/internal/pomodoro"
)

// tickMsg is sent every second to drive the TUI update loop.
type tickMsg time.Time

// stateChangeMsg is sent when the Pomodoro transitions.
type stateChangeMsg pomodoro.State

// ── Styles ──────────────────────────────────────────────────────────────────

var (
	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#5C4AE4")).
			Padding(0, 2).
			MarginBottom(1)

	styleBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#5C4AE4")).
			Padding(0, 1)

	styleWork = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6B6B"))

	styleBreak = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#51CF66"))

	styleDim = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	stylePaused = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAB005")).
			Bold(true)

	styleASCII = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#868E96")).
			MarginLeft(2)

	styleAlert = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true)

	styleGood = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#51CF66"))

	styleHelp = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#444444")).
			MarginTop(1)
)

// ── Model ────────────────────────────────────────────────────────────────────

// Model is the Bubble Tea model for the TUI.
type Model struct {
	pom      *pomodoro.Pomodoro
	watcher  *gitwatch.Watcher
	notifier notify.Notifier
	nudge    time.Duration

	gitStatuses []gitwatch.Status
	nudgesSent  map[string]time.Time

	width  int
	height int
}

// New creates a Model wired to a Pomodoro, git watcher, and notifier.
func New(
	pom *pomodoro.Pomodoro,
	watcher *gitwatch.Watcher,
	notifier notify.Notifier,
	nudge time.Duration,
) Model {
	return Model{
		pom:        pom,
		watcher:    watcher,
		notifier:   notifier,
		nudge:      nudge,
		nudgesSent: make(map[string]time.Time),
	}
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Init starts the tick loop.
func (m Model) Init() tea.Cmd {
	return tick()
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case " ":
			m.pom.Toggle()
		case "s":
			newState := m.pom.Skip()
			title, body := pomodoroNotif(newState)
			_ = m.notifier.Send(title, body)
		case "r":
			m.pom.Reset()
		}

	case tickMsg:
		// Advance Pomodoro.
		if changed, newState := m.pom.Tick(); changed {
			title, body := pomodoroNotif(newState)
			_ = m.notifier.Send(title, body)
		}

		// Refresh git statuses every minute.
		if time.Now().Second() == 0 {
			m.gitStatuses = m.watcher.Check()
			for _, s := range m.gitStatuses {
				if s.Err != nil || s.Elapsed <= m.nudge {
					continue
				}
				last, ok := m.nudgesSent[s.Repo]
				if !ok || time.Since(last) > m.nudge {
					msg := fmt.Sprintf("No commit in %s in %s", s.Repo, formatDur(s.Elapsed))
					_ = m.notifier.Send("⏰ Commit nudge", msg)
					m.nudgesSent[s.Repo] = time.Now()
				}
			}
		}

		return m, tick()
	}

	return m, nil
}

// View renders the TUI.
func (m Model) View() string {
	snap := m.pom.Snapshot()
	tod := CurrentTimeOfDay()

	// ── Header ──
	header := styleTitle.Render("  🖥️  Desk Companion  ")

	// ── Pomodoro panel ──
	pomLabel := styleStateLabel(snap.State)
	countdown := formatDur(snap.Remaining)
	countdownStr := fmt.Sprintf("  %s  %s", pomLabel, countdown)
	if snap.Paused {
		countdownStr += "  " + stylePaused.Render("(paused)")
	}
	sessions := styleDim.Render(fmt.Sprintf("  Completed sessions: %d", snap.CompletedWork))
	pomPanel := styleBox.Render(
		lipgloss.JoinVertical(lipgloss.Left, countdownStr, sessions),
	)

	// ── Git panel ──
	var gitLines []string
	if len(m.gitStatuses) == 0 {
		gitLines = append(gitLines, styleDim.Render("  No repos watched"))
	}
	for _, s := range m.gitStatuses {
		if s.Err != nil {
			gitLines = append(gitLines, fmt.Sprintf("  %-20s %s", s.Repo, styleAlert.Render("⚠ "+s.Err.Error())))
			continue
		}
		durStr := formatDur(s.Elapsed)
		indicator := styleGood.Render("●")
		if s.Elapsed > m.nudge {
			indicator = styleAlert.Render("⚠")
		}
		gitLines = append(gitLines, fmt.Sprintf("  %s %-22s last commit: %s ago", indicator, s.Repo, durStr))
	}
	gitPanel := styleBox.Render(
		lipgloss.JoinVertical(lipgloss.Left, append(
			[]string{styleDim.Render("  Git repos")},
			gitLines...,
		)...),
	)

	// ── ASCII art panel ──
	asciiPanel := styleASCII.Render(
		SceneLabel(tod) + "\n" + Scene(tod),
	)

	// ── Clock ──
	clock := styleDim.Render("  " + time.Now().Format("Mon 02 Jan · 15:04:05"))

	// ── Help bar ──
	help := styleHelp.Render("  [space] pause/resume  [s] skip  [r] reset  [q] quit")

	left := lipgloss.JoinVertical(lipgloss.Left,
		header,
		pomPanel,
		"",
		gitPanel,
		"",
		clock,
		help,
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, "   ", asciiPanel)
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func formatDur(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

func styleStateLabel(s pomodoro.State) string {
	switch s {
	case pomodoro.StateWork:
		return styleWork.Render("💪 Work")
	case pomodoro.StateShortBreak:
		return styleBreak.Render("🍵 Short Break")
	case pomodoro.StateLongBreak:
		return styleBreak.Render("🎉 Long Break")
	default:
		return s.String()
	}
}

func pomodoroNotif(s pomodoro.State) (title, body string) {
	switch s {
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
