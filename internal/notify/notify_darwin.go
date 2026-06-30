//go:build darwin

package notify

import (
	"fmt"
	"os/exec"
	"strings"
)

// DarwinNotifier sends notifications via osascript.
type DarwinNotifier struct{}

// New returns a DarwinNotifier on macOS.
func New() (Notifier, error) {
	if _, err := exec.LookPath("osascript"); err != nil {
		return Noop{}, fmt.Errorf("osascript not found: %w", err)
	}
	return DarwinNotifier{}, nil
}

func (d DarwinNotifier) Send(title, body string) error {
	// Escape double quotes for AppleScript.
	title = strings.ReplaceAll(title, `"`, `\"`)
	body = strings.ReplaceAll(body, `"`, `\"`)

	script := fmt.Sprintf(`display notification "%s" with title "%s"`, body, title)
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("osascript: %w", err)
	}
	return nil
}
