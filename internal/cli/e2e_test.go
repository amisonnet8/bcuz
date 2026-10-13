package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/amisonnet8/bcuz/internal/record"
)

// The commands that bcuz runs in these tests are the test binary itself.
// childEnv selects what it does ("exit:3", "wait", ...); see childMain. No
// external command (sh, sleep) is used: Windows has none of them.
const childEnv = "BCUZ_TEST_CHILD"

// childModes holds what a child can do. The OS-specific ones are added by
// e2e_unix_test.go.
var childModes = map[string]func(arg string) int{
	"exit": func(arg string) int {
		n, _ := strconv.Atoi(arg)
		return n
	},
	// Print the arguments, one per line.
	"echo-args": func(string) int {
		for _, a := range os.Args[1:] {
			fmt.Println(a)
		}
		return 0
	},
	// Write the argument to the standard error output.
	"stderr": func(arg string) int {
		fmt.Fprintln(os.Stderr, arg)
		return 0
	},
	"sleep": func(arg string) int {
		ms, _ := strconv.Atoi(arg)
		time.Sleep(time.Duration(ms) * time.Millisecond)
		return 0
	},
	// Copy the standard input to the standard output.
	"cat": func(string) int {
		_, _ = io.Copy(os.Stdout, os.Stdin)
		return 0
	},
	// Say "ready", then answer each input line with "pong" until the input
	// ends. A SIGINT ends it with 130, the code of a shell for it. It is
	// received here, not left to the default: a process started in the
	// background has SIGINT ignored, and would not end.
	"wait": func(string) int {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		go func() {
			<-ch
			os.Exit(130)
		}()
		fmt.Println("ready")
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			fmt.Println("pong")
		}
		return 0
	},
	"istty": func(string) int {
		fmt.Printf("stdin=%v stdout=%v stderr=%v\n", isTerminal(os.Stdin), isTerminal(os.Stdout), isTerminal(os.Stderr))
		return 0
	},
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

var (
	testBinDir string // where the built bcuz is (e2e_unix_test.go)
	pkgDir     string // this package's directory, before any test changes directory
)

func TestMain(m *testing.M) {
	if mode := os.Getenv(childEnv); mode != "" {
		name, arg, _ := strings.Cut(mode, ":")
		f, ok := childModes[name]
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown child mode %q\n", mode)
			os.Exit(99)
		}
		os.Exit(f(arg))
	}
	pkgDir, _ = os.Getwd()
	code := m.Run()
	if testBinDir != "" {
		_ = os.RemoveAll(testBinDir)
	}
	os.Exit(code)
}

func self(t *testing.T) string {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return exe
}

// workspace makes tmp/work with a .bcuz/, moves into it and returns it.
func workspace(t *testing.T) string {
	t.Helper()
	work := filepath.Join(t.TempDir(), "work")
	if err := os.MkdirAll(filepath.Join(work, ".bcuz"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(record.HistoryPath(work), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(work)
	return work
}

// records reads the history of the workspace.
func records(t *testing.T, work string) []record.Record {
	t.Helper()
	b, err := os.ReadFile(record.HistoryPath(work))
	if err != nil {
		t.Fatal(err)
	}
	var out []record.Record
	for i, line := range bytes.Split(bytes.TrimSuffix(b, []byte("\n")), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		r, err := record.Unmarshal(line)
		if err != nil {
			t.Fatalf("line %d: %v", i+1, err)
		}
		out = append(out, r)
	}
	return out
}

// runChild runs bcuz in-process with the child mode and returns what it did.
func runChild(t *testing.T, mode, stdin string, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	t.Setenv(childEnv, mode)
	var out, errOut bytes.Buffer
	code = Main(args, strings.NewReader(stdin), &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestExecuteExitCodeIsTheCommands(t *testing.T) {
	for _, want := range []int{0, 1, 3, 42, 255} {
		t.Run(strconv.Itoa(want), func(t *testing.T) {
			work := workspace(t)
			exe := self(t)

			code, stdout, stderr := runChild(t, "exit:"+strconv.Itoa(want), "", "why", exe)
			if code != want || stdout != "" || stderr != "" {
				t.Errorf("code %d (want %d), stdout %q, stderr %q", code, want, stdout, stderr)
			}
			recs := records(t, work)
			if len(recs) != 1 {
				t.Fatalf("records = %d, want 1", len(recs))
			}
			r := recs[0]
			if r.Why != "why" || len(r.Cmd) != 1 || r.Cmd[0] != exe || r.Dir != "." || r.Exit != want {
				t.Errorf("record = %+v", r)
			}
		})
	}
}

// Everything after the reason belongs to the command, whatever it looks like.
func TestExecuteArgumentsAreNotInterpreted(t *testing.T) {
	work := workspace(t)
	exe := self(t)
	args := []string{"-h", "--version", "-H", "--init", "--", "a b", "$HOME", ""}

	// A reason that starts with "-" needs "--".
	code, stdout, stderr := runChild(t, "echo-args", "", append([]string{"--", "-v again", exe}, args...)...)
	if code != 0 || stderr != "" {
		t.Fatalf("code %d, stderr %q", code, stderr)
	}
	if want := strings.Join(args, "\n") + "\n"; stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
	recs := records(t, work)
	if len(recs) != 1 {
		t.Fatalf("records = %d, want 1", len(recs))
	}
	if recs[0].Why != "-v again" || strings.Join(recs[0].Cmd, "\x00") != exe+"\x00"+strings.Join(args, "\x00") {
		t.Errorf("record = %+v", recs[0])
	}
}

func TestExecuteStreamsPassThrough(t *testing.T) {
	t.Run("stdin to stdout", func(t *testing.T) {
		workspace(t)
		code, stdout, stderr := runChild(t, "cat", "hello\nworld", "why", self(t))
		if code != 0 || stdout != "hello\nworld" || stderr != "" {
			t.Errorf("code %d, stdout %q, stderr %q", code, stdout, stderr)
		}
	})
	t.Run("stderr stays on stderr", func(t *testing.T) {
		workspace(t)
		code, stdout, stderr := runChild(t, "stderr:oops", "", "why", self(t))
		if code != 0 || stdout != "" || stderr != "oops\n" {
			t.Errorf("code %d, stdout %q, stderr %q", code, stdout, stderr)
		}
	})
}

func TestExecuteWithoutBcuzDir(t *testing.T) {
	t.Chdir(t.TempDir())

	code, stdout, stderr := runChild(t, "exit:7", "", "why", self(t))
	if code != 7 {
		t.Errorf("code = %d, want 7 (the command's)", code)
	}
	if stdout != "" || stderr != msgNoBcuzDirRun+"\n" {
		t.Errorf("stdout %q, stderr %q; want the warning only", stdout, stderr)
	}
}

func TestExecuteNotFound(t *testing.T) {
	work := workspace(t)
	const name = "bcuz-no-such-command-xyz"

	code, stdout, stderr := runMain(t, "why", name, "arg")
	if want := fmt.Sprintf(fmtCommandNotFound, name) + "\n"; code != 127 || stdout != "" || stderr != want {
		t.Errorf("code %d, stdout %q, stderr %q (want %q)", code, stdout, stderr, want)
	}
	recs := records(t, work)
	if len(recs) != 1 {
		t.Fatalf("records = %d, want 1: 127 is recorded", len(recs))
	}
	if r := recs[0]; r.Exit != 127 || r.MS != 0 || strings.Join(r.Cmd, " ") != name+" arg" {
		t.Errorf("record = %+v", r)
	}
}

func TestExecuteCannotRun(t *testing.T) {
	work := workspace(t)
	// Not an executable: no permission on Unix; on Windows an .exe that is
	// not a program (a name without an extension would be "not found").
	name := "build.sh"
	if runtime.GOOS == "windows" {
		name = "build.exe"
	}
	script := filepath.Join(work, name)
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho hi\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := runMain(t, "why", script)
	prefix := fmt.Sprintf(fmtCannotRun, script, "")
	if code != 126 || stdout != "" || !strings.HasPrefix(stderr, prefix) || strings.Count(stderr, "\n") != 1 {
		t.Errorf("code %d, stdout %q, stderr %q (want one line starting with %q)", code, stdout, stderr, prefix)
	}
	if runtime.GOOS != "windows" && stderr != prefix+"permission denied\n" {
		t.Errorf("stderr = %q", stderr)
	}
	recs := records(t, work)
	if len(recs) != 1 || recs[0].Exit != 126 || recs[0].MS != 0 {
		t.Errorf("records = %+v, want one with exit 126, ms 0", recs)
	}
}

// bcuz's own errors (125) are not recorded.
func TestExecuteUsageErrorsAreNotRecorded(t *testing.T) {
	work := workspace(t)
	for _, args := range [][]string{
		nil, {"-x", "ls"}, {"  ", "ls"}, {"why"}, {"--"},
	} {
		if code, _, _ := runMain(t, args...); code != 125 {
			t.Errorf("%q: code = %d, want 125", args, code)
		}
	}
	if recs := records(t, work); len(recs) != 0 {
		t.Errorf("records = %+v, want none", recs)
	}
}

func TestExecuteRecordsWhereAndWhen(t *testing.T) {
	work := workspace(t)
	sub := filepath.Join(work, "sub", "deeper")
	if err := os.MkdirAll(sub, 0o750); err != nil {
		t.Fatal(err)
	}
	t.Chdir(sub)

	before := time.Now().Add(-time.Millisecond) // the record has milliseconds only
	code, _, stderr := runChild(t, "sleep:150", "", "why", self(t))
	after := time.Now()
	if code != 0 || stderr != "" {
		t.Fatalf("code %d, stderr %q", code, stderr)
	}

	recs := records(t, work)
	if len(recs) != 1 {
		t.Fatalf("records = %d, want 1", len(recs))
	}
	r := recs[0]
	if r.Dir != "sub/deeper" {
		t.Errorf("dir = %q, want sub/deeper", r.Dir)
	}
	if r.Time.Before(before) || r.Time.After(after) {
		t.Errorf("ts = %v, want between %v and %v", r.Time, before, after)
	}
	// ts is when the command was started, not when it ended: the command ran
	// for 150ms, so the start is at least that long before the end.
	if r.Time.Add(150 * time.Millisecond).After(after) {
		t.Errorf("ts = %v is not the start: the command ended by %v and ran for 150ms", r.Time, after)
	}
	if r.MS < 150 || time.Duration(r.MS)*time.Millisecond > after.Sub(before) {
		t.Errorf("ms = %d, want at least 150 and at most %v", r.MS, after.Sub(before))
	}
}

// ts is written in UTC whatever the time zone of the machine is, and -H shows
// it in the time zone of the viewer (docs/reference/history.md).
func TestExecuteWritesUTC(t *testing.T) {
	old := time.Local
	time.Local = time.FixedZone("JST", 9*60*60)
	t.Cleanup(func() { time.Local = old })

	work := workspace(t)
	if code, _, stderr := runChild(t, "exit:0", "", "why", self(t)); code != 0 || stderr != "" {
		t.Fatalf("code %d, stderr %q", code, stderr)
	}
	raw, err := os.ReadFile(record.HistoryPath(work))
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`^\{"v":1,"ts":"(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z)"`).FindSubmatch(raw)
	if m == nil {
		t.Fatalf("ts is not in UTC: %s", raw)
	}

	// The same instant, seen from Tokyo: nine hours later than the text.
	utc, err := time.Parse(time.RFC3339, string(m[1]))
	if err != nil {
		t.Fatal(err)
	}
	_, stdout, _ := runMain(t, "-H")
	if want := utc.In(time.Local).Format("2006-01-02 15:04:05"); !strings.HasPrefix(stdout, want) {
		t.Errorf("-H shows %q, want it to start with %q", stdout, want)
	}
}

// Failing to record never changes the exit code (docs/reference/cli.md).
func TestExecuteRecordFailureKeepsExitCode(t *testing.T) {
	work := workspace(t)
	history := record.HistoryPath(work)
	if err := os.Remove(history); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(history, 0o750); err != nil { // a directory cannot be appended to
		t.Fatal(err)
	}

	code, stdout, stderr := runChild(t, "exit:5", "", "why", self(t))
	if code != 5 || stdout != "" {
		t.Errorf("code %d, stdout %q; want the command's 5", code, stdout)
	}
	prefix := fmt.Sprintf(fmtNotRecorded, "")
	if !strings.HasPrefix(stderr, prefix) || strings.Count(stderr, "\n") != 1 {
		t.Errorf("stderr = %q, want one warning starting with %q", stderr, prefix)
	}
}

func TestExecuteNothingExtraOnSuccess(t *testing.T) {
	workspace(t)
	// bcuz adds nothing of its own when all goes well.
	_, _, stderr := runChild(t, "exit:0", "", "why", self(t))
	if stderr != "" {
		t.Errorf("stderr = %q", stderr)
	}
}
