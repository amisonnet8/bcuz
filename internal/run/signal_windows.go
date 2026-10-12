//go:build windows

package run

import "os"

// notified is only Ctrl-C (Ctrl-Break): bcuz receives it and drops it, and
// waits for the command to end. Windows has no signal to forward.
var notified = []os.Signal{os.Interrupt}

func forwarded(os.Signal) bool { return false }

// exitCode is the exit code the command returned, as it is.
func exitCode(st *os.ProcessState) int { return st.ExitCode() }
