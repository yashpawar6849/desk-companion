//go:build windows

package daemon

import "os"

// termSignals returns the OS signals to trap for graceful shutdown on Windows.
// Windows does not support SIGTERM via os/signal, so we only use os.Interrupt.
func termSignals() []os.Signal {
	return []os.Signal{os.Interrupt}
}
