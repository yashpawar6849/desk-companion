// Package notify provides a cross-platform desktop notification interface.
// Platform-specific implementations live in notify_windows.go,
// notify_darwin.go, and notify_linux.go.
package notify

import "fmt"

// Notifier sends a desktop notification.
type Notifier interface {
	Send(title, body string) error
}

// Noop is a no-op notifier used when no platform implementation is available.
type Noop struct{}

func (Noop) Send(_, _ string) error { return nil }

// LogNotifier prints notifications to stdout (useful for testing).
type LogNotifier struct{}

func (LogNotifier) Send(title, body string) error {
	fmt.Printf("🔔 [%s] %s\n", title, body)
	return nil
}
