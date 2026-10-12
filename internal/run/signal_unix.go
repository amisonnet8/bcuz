//go:build unix

package run

import (
	"os"
	"syscall"
)

// notified are the signals bcuz receives while the command runs, so that none
// of them takes bcuz down before the record is written.
var notified = []os.Signal{syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGHUP}

// forwarded reports whether s is passed on to the command. SIGINT and SIGQUIT
// are not: a terminal sends them to the command itself, and it would get them
// twice (docs/design/decisions.md).
func forwarded(s os.Signal) bool {
	return s == syscall.SIGTERM || s == syscall.SIGHUP
}

// exitCode is the exit code of an ended command: 128 plus the number of the
// signal that ended it, as a shell reports it.
func exitCode(st *os.ProcessState) int {
	if ws, ok := st.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return 128 + int(ws.Signal())
	}
	return st.ExitCode()
}
