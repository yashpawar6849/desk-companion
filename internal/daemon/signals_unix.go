//go:build !windows

package daemon

import (
	"os"
	"syscall"
)

// termSignals returns the OS signals to trap for graceful shutdown on Unix.
func termSignals() []os.Signal {
	return []os.Signal{os.Interrupt, syscall.SIGTERM}
}
