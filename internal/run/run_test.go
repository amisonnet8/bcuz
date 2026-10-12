package run

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"
)

// The command is the test binary itself (see internal/cli/e2e_test.go).
const childEnv = "BCUZ_RUN_TEST_CHILD"

// childModes is extended by run_unix_test.go.
var childModes = map[string]func(arg string) int{
	"exit": func(arg string) int {
		n, _ := strconv.Atoi(arg)
		return n
	},
	"sleep": func(arg string) int {
		ms, _ := strconv.Atoi(arg)
		time.Sleep(time.Duration(ms) * time.Millisecond)
		return 0
	},
	"print": func(arg string) int {
		fmt.Println(arg)
		fmt.Fprintln(os.Stderr, "err:"+arg)
		return 0
	},
}

func TestMain(m *testing.M) {
	if mode := os.Getenv(childEnv); mode != "" {
		name, arg := mode, ""
		for i := range mode {
			if mode[i] == ':' {
				name, arg = mode[:i], mode[i+1:]
				break
			}
		}
		f, ok := childModes[name]
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown child mode %q\n", mode)
			os.Exit(99)
		}
		os.Exit(f(arg))
	}
	os.Exit(m.Run())
}

func exe(t *testing.T) string {
	t.Helper()
	p, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func runChild(t *testing.T, mode string, args ...string) (Result, string, string) {
	t.Helper()
	t.Setenv(childEnv, mode)
	var out, errOut bytes.Buffer
	res := Run(append([]string{exe(t)}, args...), nil, &out, &errOut)
	return res, out.String(), errOut.String()
}

func TestRunExitCode(t *testing.T) {
	for _, want := range []int{0, 1, 2, 42, 255} {
		res, _, _ := runChild(t, "exit:"+strconv.Itoa(want))
		if res.Outcome != Ran || res.ExitCode != want {
			t.Errorf("exit:%d: result = %+v", want, res)
		}
	}
}

func TestRunStreams(t *testing.T) {
	res, stdout, stderr := runChild(t, "print:hello")
	if res.ExitCode != 0 || stdout != "hello\n" || stderr != "err:hello\n" {
		t.Errorf("result %+v, stdout %q, stderr %q", res, stdout, stderr)
	}
}

func TestRunTime(t *testing.T) {
	before := time.Now()
	res, _, _ := runChild(t, "sleep:100")
	after := time.Now()

	if res.Start.Before(before) || res.Start.After(after) {
		t.Errorf("Start = %v, want between %v and %v", res.Start, before, after)
	}
	if res.Duration < 100*time.Millisecond || res.Duration > after.Sub(before) {
		t.Errorf("Duration = %v, want at least 100ms and at most %v", res.Duration, after.Sub(before))
	}
}

func TestRunNotFound(t *testing.T) {
	for _, name := range []string{
		"bcuz-no-such-command-xyz",
		filepath.Join(t.TempDir(), "no-such-file"),
	} {
		res := Run([]string{name, "a"}, nil, nil, nil)
		if res.Outcome != NotFound || res.ExitCode != 127 || res.Duration != 0 {
			t.Errorf("%q: result = %+v", name, res)
		}
		if res.Start.IsZero() {
			t.Errorf("%q: Start is not set", name)
		}
	}
}

func TestRunCannotRun(t *testing.T) {
	path := filepath.Join(t.TempDir(), "script")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o600); err != nil { // not executable
		t.Fatal(err)
	}

	res := Run([]string{path}, nil, nil, nil)
	if res.Outcome != CannotRun || res.ExitCode != 126 || res.Duration != 0 || res.Reason == "" {
		t.Errorf("result = %+v", res)
	}
	if runtime.GOOS != "windows" && res.Reason != "permission denied" {
		t.Errorf("Reason = %q, want permission denied", res.Reason)
	}
}
