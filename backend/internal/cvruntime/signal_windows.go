//go:build windows

package cvruntime

import "os"

// gracefulSignal on Windows falls back to an immediate kill because Windows
// has no SIGTERM equivalent for arbitrary child processes.
func gracefulSignal(p *os.Process) error {
	return p.Kill()
}
