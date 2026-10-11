package cli

import (
	"bytes"
	"fmt"
	"runtime/debug"
	"strings"
	"testing"
)

func run(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code = Main(args, strings.NewReader(""), &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestMain_options(t *testing.T) {
	notImpl := func(what string) string { return fmt.Sprintf(fmtNotImplemented, what) + "\n" }

	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{"no arguments", nil, 125, "", msgMissingArgs + "\n"},
		{"-h", []string{"-h"}, 0, usage, ""},
		{"--help", []string{"--help"}, 0, usage, ""},
		{"-h ignores what follows", []string{"-h", "x"}, 0, usage, ""},
		{"--version", []string{"--version"}, 0, "bcuz v9.9.9\n", ""},
		{"-H", []string{"-H"}, 125, "", notImpl("history")},
		{"--history", []string{"--history"}, 125, "", notImpl("history")},
		{"--init", []string{"--init"}, 125, "", notImpl("init")},

		{"unknown option", []string{"-x"}, 125, "", fmt.Sprintf(fmtUnknownOption, "-x") + "\n"},
		{"unknown long option", []string{"--foo", "ls"}, 125, "", fmt.Sprintf(fmtUnknownOption, "--foo") + "\n"},
		{"lone dash", []string{"-"}, 125, "", fmt.Sprintf(fmtUnknownOption, "-") + "\n"},
		{"reason starting with dash", []string{"-m", "ls"}, 125, "", fmt.Sprintf(fmtUnknownOption, "-m") + "\n"},

		{"only --", []string{"--"}, 125, "", msgMissingArgs + "\n"},
		{"-- then dash reason", []string{"--", "-x reason", "ls"}, 125, "", notImpl("run")},
		{"-- then option-like reason", []string{"--", "-h", "ls"}, 125, "", notImpl("run")},
		{"-- then empty reason", []string{"--", "", "ls"}, 125, "", msgReasonEmpty + "\n"},
		{"-- then reason only", []string{"--", "why"}, 125, "", msgMissingCommand + "\n"},

		{"empty reason", []string{"", "ls"}, 125, "", msgReasonEmpty + "\n"},
		{"blank reason", []string{" \t\n", "ls"}, 125, "", msgReasonEmpty + "\n"},
		{"empty reason without command", []string{""}, 125, "", msgReasonEmpty + "\n"},
		{"reason without command", []string{"why"}, 125, "", msgMissingCommand + "\n"},

		{"command", []string{"why", "make", "test"}, 125, "", notImpl("run")},
		{"-h after the reason belongs to the command", []string{"why", "ls", "-h"}, 125, "", notImpl("run")},
		{"--version after the reason belongs to the command", []string{"why", "x", "--version"}, 125, "", notImpl("run")},
	}

	restore := readBuildInfo
	t.Cleanup(func() { readBuildInfo = restore })
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "v9.9.9"}}, true
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, stdout, stderr := run(t, tt.args...)
			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}
			if stdout != tt.wantStdout {
				t.Errorf("stdout = %q, want %q", stdout, tt.wantStdout)
			}
			if stderr != tt.wantStderr {
				t.Errorf("stderr = %q, want %q", stderr, tt.wantStderr)
			}
		})
	}
}

func TestVersion(t *testing.T) {
	restore := readBuildInfo
	t.Cleanup(func() { readBuildInfo = restore })

	tests := []struct {
		name string
		info *debug.BuildInfo
		ok   bool
		want string
	}{
		{"tagged", &debug.BuildInfo{Main: debug.Module{Version: "v0.1.0"}}, true, "v0.1.0"},
		{"empty version", &debug.BuildInfo{}, true, "(devel)"},
		{"no build info", nil, false, "(devel)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readBuildInfo = func() (*debug.BuildInfo, bool) { return tt.info, tt.ok }
			_, stdout, _ := run(t, "--version")
			if want := "bcuz " + tt.want + "\n"; stdout != want {
				t.Errorf("stdout = %q, want %q", stdout, want)
			}
		})
	}
}
