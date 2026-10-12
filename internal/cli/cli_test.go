package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
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

func TestInit(t *testing.T) {
	// Each case works in a new directory tree: tmp/work/sub.
	setup := func(t *testing.T) (work, sub string) {
		t.Helper()
		work = filepath.Join(t.TempDir(), "work")
		sub = filepath.Join(work, "sub")
		if err := os.MkdirAll(sub, 0o750); err != nil {
			t.Fatal(err)
		}
		return work, sub
	}

	t.Run("created, then already exists", func(t *testing.T) {
		work, _ := setup(t)
		t.Chdir(work)

		code, stdout, stderr := run(t, "--init")
		if code != 0 || stdout != msgCreated+"\n" || stderr != "" {
			t.Errorf("first: code %d, stdout %q, stderr %q", code, stdout, stderr)
		}
		fi, err := os.Stat(filepath.Join(work, ".bcuz", "history.jsonl"))
		if err != nil || fi.Size() != 0 {
			t.Errorf("history.jsonl: %v, %v; want an empty file", fi, err)
		}

		code, stdout, stderr = run(t, "--init")
		if code != 0 || stdout != msgAlreadyExists+"\n" || stderr != "" {
			t.Errorf("second: code %d, stdout %q, stderr %q", code, stdout, stderr)
		}
	})

	t.Run("shadows the .bcuz/ of a parent", func(t *testing.T) {
		work, sub := setup(t)
		if err := os.Mkdir(filepath.Join(work, ".bcuz"), 0o750); err != nil {
			t.Fatal(err)
		}
		t.Chdir(sub)
		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		parent := filepath.Join(filepath.Dir(wd), ".bcuz") + string(filepath.Separator)

		code, stdout, stderr := run(t, "--init")
		want := fmt.Sprintf(fmtCreatedShadows, parent) + "\n"
		if code != 0 || stdout != want || stderr != "" {
			t.Errorf("code %d, stdout %q (want %q), stderr %q", code, stdout, want, stderr)
		}
		if _, err := os.Stat(filepath.Join(sub, ".bcuz", "history.jsonl")); err != nil {
			t.Error(err)
		}
	})

	t.Run("cannot create", func(t *testing.T) {
		work, _ := setup(t)
		if err := os.WriteFile(filepath.Join(work, ".bcuz"), nil, 0o600); err != nil {
			t.Fatal(err)
		}
		t.Chdir(work)

		code, stdout, stderr := run(t, "--init")
		if code != 125 || stdout != "" {
			t.Errorf("code %d, stdout %q", code, stdout)
		}
		prefix := strings.TrimSuffix(fmtCannotCreate, "%s")
		if !strings.HasPrefix(stderr, prefix) || !strings.HasSuffix(stderr, "\n") || strings.Count(stderr, "\n") != 1 {
			t.Errorf("stderr = %q, want one line starting with %q", stderr, prefix)
		}
	})
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
		{"devel", &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, true, "(devel)"},
		{"pseudo-version", &debug.BuildInfo{Main: debug.Module{Version: "v0.0.0-20261011232240-19408631a724+dirty"}}, true, "(devel)"},
		{"tag with dirty", &debug.BuildInfo{Main: debug.Module{Version: "v0.1.0+dirty"}}, true, "v0.1.0+dirty"},
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
