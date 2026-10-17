package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const goodLine = `{"v":1,"ts":"2026-10-12T06:35:12.345+09:00","why":"xxテスト実施","cmd":["make","test"],"dir":".","exit":0,"ms":8312}` + "\n"

// historyWorkspace makes a workspace whose history.jsonl is content, and
// works in it. The viewer's zone is fixed to +09:00.
func historyWorkspace(t *testing.T, content string) string {
	t.Helper()
	old := time.Local
	time.Local = time.FixedZone("JST", 9*3600)
	t.Cleanup(func() { time.Local = old })

	work := t.TempDir()
	if err := os.Mkdir(filepath.Join(work, ".bcuz"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(work, ".bcuz", "history.jsonl"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(work)
	return work
}

const goodOut = "2026-10-12 06:35:12    0    8.3s  .  xxテスト実施  $ make test\n"

func TestHistoryShows(t *testing.T) {
	historyWorkspace(t, goodLine)
	for _, opt := range []string{"-H", "--history"} {
		code, stdout, stderr := runMain(t, opt)
		if code != 0 || stdout != goodOut || stderr != "" {
			t.Errorf("%s: code %d, stdout %q, stderr %q", opt, code, stdout, stderr)
		}
	}
}

func TestHistoryFromSubdirectory(t *testing.T) {
	work := historyWorkspace(t, goodLine)
	sub := filepath.Join(work, "sub")
	if err := os.Mkdir(sub, 0o750); err != nil {
		t.Fatal(err)
	}
	t.Chdir(sub)
	code, stdout, stderr := runMain(t, "-H")
	if code != 0 || stdout != goodOut || stderr != "" {
		t.Errorf("code %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

func TestHistoryArgumentErrors(t *testing.T) {
	historyWorkspace(t, goodLine)
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"invalid date", []string{"-H", "--date", "2026-02-30"}, fmt.Sprintf(fmtInvalidDate, "2026-02-30")},
		{"bad form", []string{"-H", "--since", "2026-1-2"}, fmt.Sprintf(fmtInvalidDate, "2026-1-2")},
		{"not a date", []string{"-H", "--until", "yesterday"}, fmt.Sprintf(fmtInvalidDate, "yesterday")},
		{"date with since", []string{"-H", "--date", "2026-10-12", "--since", "2026-10-01"}, msgDateWithRange},
		{"since with date", []string{"-H", "--since", "2026-10-01", "--date", "2026-10-12"}, msgDateWithRange},
		{"date with until", []string{"-H", "--date", "2026-10-12", "--until", "2026-10-13"}, msgDateWithRange},
		{"duplicate date", []string{"-H", "--date", "2026-10-12", "--date", "2026-10-12"}, fmt.Sprintf(fmtDuplicateOption, "--date")},
		{"duplicate since", []string{"-H", "--since", "2026-10-01", "--since", "2026-10-02"}, fmt.Sprintf(fmtDuplicateOption, "--since")},
		{"duplicate until", []string{"--history", "--until", "2026-10-01", "--until", "2026-10-02"}, fmt.Sprintf(fmtDuplicateOption, "--until")},
		{"unknown argument", []string{"-H", "--foo"}, fmt.Sprintf(fmtUnknownArgument, "--foo")},
		{"a bare date", []string{"-H", "2026-10-12"}, fmt.Sprintf(fmtUnknownArgument, "2026-10-12")},
		{"equals form", []string{"-H", "--date=2026-10-12"}, fmt.Sprintf(fmtUnknownArgument, "--date=2026-10-12")},
		{"missing date after --date", []string{"-H", "--date"}, fmt.Sprintf(fmtMissingDate, "--date")},
		{"missing date after --since", []string{"-H", "--since"}, fmt.Sprintf(fmtMissingDate, "--since")},
		{"missing date after --until", []string{"-H", "--until"}, fmt.Sprintf(fmtMissingDate, "--until")},
		{"duplicate -", []string{"-H", "-", "--date", "2026-10-12", "-"}, fmt.Sprintf(fmtDuplicateOption, "-")},
		{"unknown after -", []string{"-H", "-", "--foo"}, fmt.Sprintf(fmtUnknownArgument, "--foo")},
		{"bad date with -", []string{"-H", "-", "--date", "-"}, fmt.Sprintf(fmtInvalidDate, "-")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, stdout, stderr := runMain(t, tt.args...)
			if code != 125 || stdout != "" || stderr != tt.want+"\n" {
				t.Errorf("code %d, stdout %q, stderr %q; want 125, empty, %q", code, stdout, stderr, tt.want+"\n")
			}
		})
	}
}

func TestHistoryFilterAndEmpty(t *testing.T) {
	historyWorkspace(t, goodLine)
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"the day", []string{"-H", "--date", "2026-10-12"}, goodOut},
		{"another day", []string{"-H", "--date", "2026-10-11"}, ""},
		{"since", []string{"-H", "--since", "2026-10-12"}, goodOut},
		{"until", []string{"-H", "--until", "2026-10-11"}, ""},
		{"since after until", []string{"-H", "--since", "2026-10-13", "--until", "2026-10-01"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, stdout, stderr := runMain(t, tt.args...)
			if code != 0 || stdout != tt.want || stderr != "" {
				t.Errorf("code %d, stdout %q, stderr %q; want 0, %q, empty", code, stdout, stderr, tt.want)
			}
		})
	}
}

func TestHistoryNoBcuzDir(t *testing.T) {
	t.Chdir(t.TempDir())
	code, stdout, stderr := runMain(t, "-H")
	if code != 125 || stdout != "" || stderr != msgNoBcuzDirHist+"\n" {
		t.Errorf("code %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

func TestHistoryFileMissingIsEmpty(t *testing.T) {
	work := historyWorkspace(t, "")
	if err := os.Remove(filepath.Join(work, ".bcuz", "history.jsonl")); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runMain(t, "-H")
	if code != 0 || stdout != "" || stderr != "" {
		t.Errorf("code %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

func TestHistoryCannotRead(t *testing.T) {
	// A directory in place of the file: it cannot be read on any OS.
	work := historyWorkspace(t, "")
	path := filepath.Join(work, ".bcuz", "history.jsonl")
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0o750); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runMain(t, "-H")
	prefix := fmt.Sprintf(fmtCannotRead, placeHistoryFile, "")
	if code != 125 || stdout != "" || !strings.HasPrefix(stderr, prefix) || strings.Count(stderr, "\n") != 1 {
		t.Errorf("code %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

func TestHistoryUnreadableLines(t *testing.T) {
	// Line 2 is broken; the last line has no newline (a write in progress)
	// and is skipped without a warning.
	partial := strings.TrimSuffix(goodLine, "\n")
	historyWorkspace(t, goodLine+"not json\n"+goodLine+partial)
	code, stdout, stderr := runMain(t, "-H")
	wantErr := fmt.Sprintf(fmtUnreadableLine, placeHistoryFile, 2) + "\n"
	if code != 0 || stdout != goodOut+goodOut || stderr != wantErr {
		t.Errorf("code %d, stdout %q, stderr %q; want 0, two lines, %q", code, stdout, stderr, wantErr)
	}
}

// runMainWithInput is runMain with in as standard input.
func runMainWithInput(t *testing.T, in io.Reader, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code = Main(args, in, &out, &errOut)
	return code, out.String(), errOut.String()
}

// stdinPlace works where there is no .bcuz/, with the viewer's zone fixed
// to +09:00: "-" must not need one.
func stdinPlace(t *testing.T) {
	t.Helper()
	old := time.Local
	time.Local = time.FixedZone("JST", 9*3600)
	t.Cleanup(func() { time.Local = old })
	t.Chdir(t.TempDir())
}

func TestHistoryFromStdin(t *testing.T) {
	stdinPlace(t)
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"-H", []string{"-H", "-"}, goodOut},
		{"--history", []string{"--history", "-"}, goodOut},
		{"- before a filter", []string{"-H", "-", "--date", "2026-10-12"}, goodOut},
		{"- after a filter", []string{"-H", "--since", "2026-10-12", "-"}, goodOut},
		{"- between filters", []string{"-H", "--since", "2026-10-01", "-", "--until", "2026-10-12"}, goodOut},
		{"filtered out", []string{"-H", "-", "--date", "2026-10-11"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, stdout, stderr := runMainWithInput(t, strings.NewReader(goodLine), tt.args...)
			if code != 0 || stdout != tt.want || stderr != "" {
				t.Errorf("code %d, stdout %q, stderr %q; want 0, %q, empty", code, stdout, stderr, tt.want)
			}
		})
	}
}

// Joined files come in any order: the newest is shown first.
func TestHistoryFromStdinJoined(t *testing.T) {
	stdinPlace(t)
	newer := strings.Replace(goodLine, "2026-10-12T06:35:12.345+09:00", "2026-10-13T00:00:00.000Z", 1)
	code, stdout, stderr := runMainWithInput(t, strings.NewReader(newer+goodLine), "-H", "-")
	want := "2026-10-13 09:00:00    0    8.3s  .  xxテスト実施  $ make test\n" + goodOut
	if code != 0 || stdout != want || stderr != "" {
		t.Errorf("code %d, stdout %q, stderr %q; want 0, %q", code, stdout, stderr, want)
	}
	code, stdout, _ = runMainWithInput(t, strings.NewReader(goodLine+newer), "-H", "-")
	if code != 0 || stdout != want {
		t.Errorf("reversed input: code %d, stdout %q; want %q", code, stdout, want)
	}
}

// With "-", the history file of the workspace is not read.
func TestHistoryFromStdinNotTheFile(t *testing.T) {
	historyWorkspace(t, goodLine)
	code, stdout, stderr := runMainWithInput(t, strings.NewReader(""), "-H", "-")
	if code != 0 || stdout != "" || stderr != "" {
		t.Errorf("code %d, stdout %q, stderr %q; want 0, empty, empty", code, stdout, stderr)
	}
}

func TestHistoryFromStdinUnreadableLines(t *testing.T) {
	stdinPlace(t)
	partial := strings.TrimSuffix(goodLine, "\n")
	in := strings.NewReader(goodLine + "not json\n" + goodLine + partial)
	code, stdout, stderr := runMainWithInput(t, in, "-H", "-")
	wantErr := "bcuz: warning: <stdin>:2: skipped unreadable line\n"
	if code != 0 || stdout != goodOut+goodOut || stderr != wantErr {
		t.Errorf("code %d, stdout %q, stderr %q; want 0, two lines, %q", code, stdout, stderr, wantErr)
	}
}

// failingReader fails the first read, as standard input does on an I/O
// error, and remembers that it was read.
type failingReader struct{ read bool }

func (r *failingReader) Read([]byte) (int, error) {
	r.read = true
	return 0, &os.PathError{Op: "read", Path: "/dev/stdin", Err: errors.New("input/output error")}
}

func TestHistoryFromStdinCannotRead(t *testing.T) {
	stdinPlace(t)
	code, stdout, stderr := runMainWithInput(t, &failingReader{}, "-H", "-")
	wantErr := "bcuz: cannot read <stdin>: input/output error\n"
	if code != 125 || stdout != "" || stderr != wantErr {
		t.Errorf("code %d, stdout %q, stderr %q; want 125, empty, %q", code, stdout, stderr, wantErr)
	}
}

// Without "-", standard input is never read, even when it is not a terminal.
func TestHistoryWithoutDashIgnoresStdin(t *testing.T) {
	historyWorkspace(t, goodLine)
	in := &failingReader{}
	code, stdout, stderr := runMainWithInput(t, in, "-H")
	if code != 0 || stdout != goodOut || stderr != "" || in.read {
		t.Errorf("code %d, stdout %q, stderr %q, stdin read %v", code, stdout, stderr, in.read)
	}
}

// What bcuz writes is what -H shows: the newest first, with the exit code.
func TestHistoryShowsWhatWasRecorded(t *testing.T) {
	workspace(t)
	exe := self(t)
	runChild(t, "exit:0", "", "first run", exe)
	runChild(t, "exit:3", "", "second\nrun", exe, "a b")

	code, stdout, stderr := runMain(t, "-H")
	lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
	if code != 0 || stderr != "" || len(lines) != 2 {
		t.Fatalf("code %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	if !strings.Contains(lines[0], "    3  ") || !strings.Contains(lines[0], "  second run  $ ") || !strings.HasSuffix(lines[0], " 'a b'") {
		t.Errorf("newest line = %q", lines[0])
	}
	if !strings.Contains(lines[1], "    0  ") || !strings.Contains(lines[1], "  first run  $ ") {
		t.Errorf("older line = %q", lines[1])
	}
}
