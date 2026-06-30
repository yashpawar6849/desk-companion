//go:build windows

package notify

import (
	"fmt"
	"os/exec"
	"strings"
)

// WindowsNotifier sends toast notifications via PowerShell.
type WindowsNotifier struct{}

// New returns a WindowsNotifier on Windows.
func New() (Notifier, error) {
	// Quick probe: check that powershell is available.
	if err := exec.Command("powershell", "-Command", "exit 0").Run(); err != nil {
		return Noop{}, fmt.Errorf("powershell not available: %w", err)
	}
	return WindowsNotifier{}, nil
}

// Send fires a Windows toast notification via a PowerShell one-liner.
// Uses the BurntToast-less approach via [Windows.UI.Notifications].
func (w WindowsNotifier) Send(title, body string) error {
	// Escape single quotes for PowerShell string safety.
	title = strings.ReplaceAll(title, "'", "''")
	body = strings.ReplaceAll(body, "'", "''")

	script := fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null

$template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent(
    [Windows.UI.Notifications.ToastTemplateType]::ToastText02)

$template.SelectSingleNode('//text[@id="1"]').InnerText = '%s'
$template.SelectSingleNode('//text[@id="2"]').InnerText = '%s'

$toast = [Windows.UI.Notifications.ToastNotification]::new($template)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('Desk Companion').Show($toast)
`, title, body)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	if out, err := cmd.CombinedOutput(); err != nil {
		// Fallback: use msg command (works on all Windows editions for console alerts)
		return w.fallbackMsg(title, body, string(out))
	}
	return nil
}

// fallbackMsg uses the msg command as a last-resort notification on Windows.
func (w WindowsNotifier) fallbackMsg(title, body, psErr string) error {
	_ = psErr // logged if needed
	// Write a visible alert to the console since msg requires sessions.
	fmt.Printf("\n╔══════════════════════════════╗\n║  🔔 %-24s║\n║  %-28s║\n╚══════════════════════════════╝\n\n", title, body)
	return nil
}
