//go:build linux

package notify

import (
	"fmt"
	"os/exec"
)

// LinuxNotifier sends notifications via notify-send.
type LinuxNotifier struct{}

// New returns a LinuxNotifier on Linux.
func New() (Notifier, error) {
	if _, err := exec.LookPath("notify-send"); err != nil {
		return Noop{}, fmt.Errorf("notify-send not found: %w", err)
	}
	return LinuxNotifier{}, nil
}

func (l LinuxNotifier) Send(title, body string) error {
	cmd := exec.Command("notify-send", "--app-name=Desk Companion", title, body)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("notify-send: %w", err)
	}
	return nil
}
