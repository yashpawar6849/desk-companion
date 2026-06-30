package pomodoro

import (
	"sync"
	"time"
)

// State represents the current Pomodoro phase.
type State int

const (
	StateWork       State = iota // Focused work session
	StateShortBreak              // Short 5-minute break
	StateLongBreak               // Long 15-minute break after every 4 work sessions
)

func (s State) String() string {
	switch s {
	case StateWork:
		return "Work"
	case StateShortBreak:
		return "Short Break"
	case StateLongBreak:
		return "Long Break"
	default:
		return "Unknown"
	}
}

// Pomodoro manages the work/break state machine.
type Pomodoro struct {
	mu           sync.Mutex
	state        State
	remaining    time.Duration
	workDur      time.Duration
	shortBreak   time.Duration
	longBreak    time.Duration
	completedWork int // number of completed work sessions
	paused       bool
	lastTick     time.Time
}

// New creates a Pomodoro with the given durations, starting in Work state.
func New(work, short, long time.Duration) *Pomodoro {
	return &Pomodoro{
		state:      StateWork,
		remaining:  work,
		workDur:    work,
		shortBreak: short,
		longBreak:  long,
		lastTick:   time.Now(),
	}
}

// Tick advances the timer by one second. Returns (stateChanged, newState).
func (p *Pomodoro) Tick() (bool, State) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.paused {
		return false, p.state
	}

	now := time.Now()
	delta := now.Sub(p.lastTick)
	p.lastTick = now

	if delta > p.remaining {
		// Timer expired — transition to next state.
		next := p.nextState()
		p.state = next
		p.remaining = p.durationFor(next)
		return true, next
	}

	p.remaining -= delta
	return false, p.state
}

// Toggle pauses or resumes the timer.
func (p *Pomodoro) Toggle() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.paused = !p.paused
	if !p.paused {
		p.lastTick = time.Now() // reset to avoid a big delta on resume
	}
}

// Skip advances to the next state immediately.
func (p *Pomodoro) Skip() State {
	p.mu.Lock()
	defer p.mu.Unlock()
	next := p.nextState()
	p.state = next
	p.remaining = p.durationFor(next)
	return next
}

// Reset restarts the current phase from its full duration.
func (p *Pomodoro) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.remaining = p.durationFor(p.state)
	p.lastTick = time.Now()
}

// Snapshot returns a read-only view of the current state.
type Snapshot struct {
	State         State
	Remaining     time.Duration
	Paused        bool
	CompletedWork int
}

func (p *Pomodoro) Snapshot() Snapshot {
	p.mu.Lock()
	defer p.mu.Unlock()
	return Snapshot{
		State:         p.state,
		Remaining:     p.remaining,
		Paused:        p.paused,
		CompletedWork: p.completedWork,
	}
}

func (p *Pomodoro) nextState() State {
	switch p.state {
	case StateWork:
		p.completedWork++
		if p.completedWork%4 == 0 {
			return StateLongBreak
		}
		return StateShortBreak
	default:
		return StateWork
	}
}

func (p *Pomodoro) durationFor(s State) time.Duration {
	switch s {
	case StateWork:
		return p.workDur
	case StateShortBreak:
		return p.shortBreak
	case StateLongBreak:
		return p.longBreak
	default:
		return p.workDur
	}
}
