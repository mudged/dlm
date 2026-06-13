//go:build !windows

package cvruntime

import (
	"os"
	"syscall"
)

// gracefulSignal sends SIGTERM to p so the child can clean up before the
// WaitDelay SIGKILL escalation fires (see run.go).
func gracefulSignal(p *os.Process) error {
	return p.Signal(syscall.SIGTERM)
}
