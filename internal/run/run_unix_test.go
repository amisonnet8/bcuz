//go:build unix

package run

import (
	"syscall"
	"testing"
)

func init() {
	// End by a signal that the process sends to itself.
	childModes["kill-self"] = func(arg string) int {
		sig := syscall.SIGTERM
		if arg == "kill" {
			sig = syscall.SIGKILL
		}
		_ = syscall.Kill(syscall.Getpid(), sig)
		select {} // wait for the signal to take effect
	}
}

// A command ended by a signal gives 128 plus the signal number, as a shell does.
func TestRunEndedBySignal(t *testing.T) {
	for _, tt := range []struct {
		mode string
		want int
	}{
		{"kill-self:term", 128 + int(syscall.SIGTERM)},
		{"kill-self:kill", 128 + int(syscall.SIGKILL)},
	} {
		res, _, _ := runChild(t, tt.mode)
		if res.Outcome != Ran || res.ExitCode != tt.want {
			t.Errorf("%s: result = %+v, want exit code %d", tt.mode, res, tt.want)
		}
	}
}
