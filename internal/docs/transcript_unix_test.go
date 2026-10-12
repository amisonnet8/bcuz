//go:build unix

package docs

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/amisonnet8/bcuz/internal/history"
	"github.com/amisonnet8/bcuz/internal/record"
)

// The examples show commands such as make, git and tail. The test runs
// them as fakes: symlinks, named like the command, to the test binary
// itself. When the binary is started through one of them (the environment
// has fakeExitEnv), TestMain prints the output the document shows and exits
// with the code the document shows, and nothing else runs. What is checked
// is bcuz: what it adds, its messages, its exit codes and its records.
const (
	fakeExitEnv = "BCUZ_DOC_FAKE_EXIT"
	fakeOutEnv  = "BCUZ_DOC_FAKE_OUT"
)

func TestMain(m *testing.M) {
	if code, ok := os.LookupEnv(fakeExitEnv); ok {
		_, _ = os.Stdout.WriteString(os.Getenv(fakeOutEnv))
		n, err := strconv.Atoi(code)
		if err != nil {
			n = 2
		}
		os.Exit(n)
	}
	os.Exit(m.Run())
}

// fakeNames are the commands of the examples that exist. The misspelled
// make of errors.md is not among them: it must not be found (127).
var fakeNames = []string{"make", "golangci-lint", "tail", "echo", "git", "go", "sh", "ls"}

const (
	docWork    = "/home/me/work"
	docScratch = "/tmp/scratch"
)

// step is a command of a transcript and the lines printed under it.
type step struct {
	cmd string
	out []string
}

// parseSteps splits a transcript. A command that opens a quote goes on in
// the next lines.
func parseSteps(text string) []step {
	var steps []step
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	for i := 0; i < len(lines); i++ {
		cmd, ok := strings.CutPrefix(lines[i], "$ ")
		if !ok {
			if len(steps) > 0 {
				steps[len(steps)-1].out = append(steps[len(steps)-1].out, lines[i])
			}
			continue
		}
		for {
			if _, closed := tokenize(cmd); closed || i+1 >= len(lines) {
				break
			}
			i++
			cmd += "\n" + lines[i]
		}
		steps = append(steps, step{cmd: cmd})
	}
	return steps
}

// tokenize splits a command line at spaces, with '…' and "…" for quoting.
// ok is false when a quote is left open.
func tokenize(s string) (tokens []string, ok bool) {
	var cur strings.Builder
	inToken := false
	var quote rune
	for _, c := range s {
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			} else {
				cur.WriteRune(c)
			}
		case c == '\'' || c == '"':
			quote, inToken = c, true
		case c == ' ' || c == '\t' || c == '\n':
			if inToken {
				tokens = append(tokens, cur.String())
				cur.Reset()
				inToken = false
			}
		default:
			cur.WriteRune(c)
			inToken = true
		}
	}
	if quote != 0 {
		return nil, false
	}
	if inToken {
		tokens = append(tokens, cur.String())
	}
	return tokens, true
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		in     string
		want   []string
		wantOK bool
	}{
		{`bcuz "a b" make test`, []string{"bcuz", "a b", "make", "test"}, true},
		{`sh -c 'grep -c TODO *.go | sort'`, []string{"sh", "-c", "grep -c TODO *.go | sort"}, true},
		{`bcuz "  " make`, []string{"bcuz", "  ", "make"}, true},
		{`bcuz "" x`, []string{"bcuz", "", "x"}, true},
		{"bcuz \"a\nb\" c", []string{"bcuz", "a\nb", "c"}, true},
		{`bcuz "open`, nil, false},
	}
	for _, tt := range tests {
		got, ok := tokenize(tt.in)
		if ok != tt.wantOK || strings.Join(got, "\x00") != strings.Join(tt.want, "\x00") || len(got) != len(tt.want) {
			t.Errorf("tokenize(%q) = %q, %v; want %q, %v", tt.in, got, ok, tt.want, tt.wantOK)
		}
	}
}

// docSession is a workspace in which the commands of a document run, one
// after another, the way a person would type them.
type docSession struct {
	t        *testing.T
	work     string
	scratch  string
	paths    []string // real paths shown as docWork (longest first)
	cwd      string
	lastCode int
	last     *record.Record // the record the latest command wrote
}

// maskTime hides what changes at each run: the date and time column, and the
// duration column of a history line.
var maskTime = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(  [ 0-9]{3}  )[ 0-9a-z.]{6}(  )`)

func mask(text string) string {
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		lines[i] = maskTime.ReplaceAllString(l, "<time>${1}<duration>${2}")
	}
	return strings.Join(lines, "\n")
}

// newDocSession makes the workspace of the examples (docWork, with
// backend/, frontend/, an executable migrate and a build.sh that is not
// executable). withBcuz tells whether docWork already has .bcuz/.
func newDocSession(t *testing.T, withBcuz bool) *docSession {
	t.Helper()
	tokyoZone(t)

	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	work := filepath.Join(t.TempDir(), "work")
	for _, d := range []string{"backend", "frontend"} {
		if err := os.MkdirAll(filepath.Join(work, d), 0o750); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(exe, filepath.Join(work, "migrate")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(work, "build.sh"), []byte("#!/bin/sh\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if withBcuz {
		if err := os.Mkdir(filepath.Join(work, record.DirName), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(record.HistoryPath(work), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	bin := t.TempDir()
	for _, name := range fakeNames {
		if err := os.Symlink(exe, filepath.Join(bin, name)); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin)

	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })

	s := &docSession{t: t, work: work, scratch: t.TempDir(), cwd: work}
	s.paths = []string{work}
	if real, err := filepath.EvalSymlinks(work); err == nil && real != work {
		s.paths = []string{real, work}
	}
	return s
}

// mapDir turns a directory of the document into the real one.
func (s *docSession) mapDir(dir string) string {
	switch {
	case dir == docWork || strings.HasPrefix(dir, docWork+"/"):
		dir = filepath.Join(s.work, strings.TrimPrefix(dir, docWork))
	case dir == docScratch:
		dir = s.scratch
	case !filepath.IsAbs(dir):
		dir = filepath.Join(s.cwd, dir)
	}
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		s.t.Fatalf("cd: %s is not a directory", dir)
	}
	return dir
}

func (s *docSession) shown(text string) string {
	for _, p := range s.paths {
		text = strings.ReplaceAll(text, p, docWork)
	}
	return strings.ReplaceAll(text, s.scratch, docScratch)
}

func readRecords(t *testing.T, root string) []record.Record {
	t.Helper()
	b, err := os.ReadFile(record.HistoryPath(root))
	if err != nil {
		return nil
	}
	recs, skipped, err := history.Read(bytes.NewReader(b))
	if err != nil || len(skipped) != 0 {
		t.Fatalf("history of %s: %v, skipped %v", root, err, skipped)
	}
	return recs
}

// fakeOutput is what a fake command prints: the lines under the command that
// bcuz does not print itself. ^C is what the terminal echoes.
func fakeOutput(lines []string) string {
	var out strings.Builder
	for _, l := range lines {
		if strings.HasPrefix(l, "bcuz: ") || l == "^C" {
			continue
		}
		out.WriteString(l + "\n")
	}
	return out.String()
}

// runBcuz runs bcuz with tokens[1:] in s.cwd. Fake commands print out and
// exit with code.
func (s *docSession) runBcuz(tokens []string, out string, code int) string {
	s.t.Helper()
	if err := os.Chdir(s.cwd); err != nil {
		s.t.Fatal(err)
	}

	before := s.rootRecords()
	_ = os.Setenv(fakeExitEnv, strconv.Itoa(code))
	_ = os.Setenv(fakeOutEnv, out)
	got, stdout, stderr := runCLI(tokens[1:]...)
	_ = os.Unsetenv(fakeExitEnv)
	_ = os.Unsetenv(fakeOutEnv)
	s.lastCode = got

	after := s.rootRecords()
	if after.root == before.root && len(after.recs) == len(before.recs)+1 {
		r := after.recs[len(after.recs)-1]
		s.last = &r
		s.checkRecordedAsGiven(tokens[1:], r)
	}
	return s.shown(stdout + stderr)
}

type rootRecords struct {
	root string
	recs []record.Record
}

func (s *docSession) rootRecords() rootRecords {
	cwd, err := os.Getwd()
	if err != nil {
		s.t.Fatal(err)
	}
	root, ok := record.FindRoot(cwd)
	if !ok {
		return rootRecords{}
	}
	return rootRecords{root, readRecords(s.t, root)}
}

// checkRecordedAsGiven: the reason and the command are in the record just as
// they were given, whatever they look like (docs/reference/cli.md).
func (s *docSession) checkRecordedAsGiven(args []string, r record.Record) {
	s.t.Helper()
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}
	if len(args) < 2 {
		return
	}
	if r.Why != args[0] || strings.Join(r.Cmd, "\x00") != strings.Join(args[1:], "\x00") || len(r.Cmd) != len(args)-1 {
		s.t.Errorf("recorded why %q cmd %q; given %q", r.Why, r.Cmd, args)
	}
}

// runSteps runs a transcript and returns what it prints, in the same form.
func (s *docSession) runSteps(text string) string {
	s.t.Helper()
	steps := parseSteps(text)
	var out strings.Builder
	for i, st := range steps {
		out.WriteString("$ " + st.cmd + "\n")
		if st.cmd == "echo $?" {
			fmt.Fprintf(&out, "%d\n", s.lastCode)
			continue
		}
		tokens, ok := tokenize(st.cmd)
		if !ok || len(tokens) == 0 {
			s.t.Fatalf("cannot read the command: %q", st.cmd)
		}
		// Whatever follows && is run by the shell that called bcuz, not by bcuz.
		for j, tok := range tokens {
			if tok == "&&" {
				tokens = tokens[:j]
				break
			}
		}
		switch tokens[0] {
		case "cd":
			if len(tokens) != 2 {
				s.t.Fatalf("cd needs one argument: %q", st.cmd)
			}
			s.cwd = s.mapDir(tokens[1])
		case "bcuz":
			code := 0
			if i+1 < len(steps) && steps[i+1].cmd == "echo $?" && len(steps[i+1].out) > 0 {
				n, err := strconv.Atoi(steps[i+1].out[0])
				if err != nil {
					s.t.Fatalf("not an exit code: %q", steps[i+1].out[0])
				}
				code = n
			}
			out.WriteString(s.runBcuz(tokens, fakeOutput(st.out), code))
		default:
			s.t.Fatalf("unsupported command: %q", st.cmd)
		}
	}
	return out.String()
}

var historyLine = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}  `)

// checkDocument runs the blocks of a document in order. A transcript is run
// and compared as a whole. A json block is the record the command before it
// wrote; a history line is that record as -H shows it. The date, time and
// duration are not compared (they change at each run).
func (s *docSession) checkDocument(md string) {
	t := s.t
	t.Helper()
	n := 0
	for i, b := range allBlocks(md) {
		switch {
		case b.lang == "json":
			n++
			s.checkJSON(i+1, b.text)
		case b.lang == "" && strings.HasPrefix(b.text, "$ "):
			n++
			want := strings.ReplaceAll(b.text, "^C\n", "")
			if got := s.runSteps(b.text); mask(got) != mask(want) {
				t.Errorf("block %d does not match.\n--- got\n%s--- want\n%s", i+1, mask(got), mask(want))
			}
		case b.lang == "" && historyLine.MatchString(b.text):
			n++
			if s.last == nil {
				t.Fatalf("block %d: no record to show", i+1)
			}
			got := history.FormatLine(*s.last) + "\n"
			if mask(got) != mask(b.text) {
				t.Errorf("block %d does not match.\n--- got\n%s--- want\n%s", i+1, mask(got), mask(b.text))
			}
		default:
			t.Errorf("block %d: a kind of block this test does not know:\n%s", i+1, b.text)
		}
	}
	if n == 0 {
		t.Error("no blocks")
	}
}

func (s *docSession) checkJSON(n int, text string) {
	t := s.t
	t.Helper()
	want, err := record.Unmarshal([]byte(strings.TrimSuffix(text, "\n")))
	if err != nil {
		t.Errorf("block %d: %v", n, err)
		return
	}
	if s.last == nil {
		t.Errorf("block %d: no record was written before it", n)
		return
	}
	got := *s.last
	if got.Why != want.Why || got.Dir != want.Dir || got.Exit != want.Exit ||
		len(got.Cmd) != len(want.Cmd) || strings.Join(got.Cmd, "\x00") != strings.Join(want.Cmd, "\x00") {
		t.Errorf("block %d: record\n got: why %q cmd %q dir %q exit %d\nwant: why %q cmd %q dir %q exit %d",
			n, got.Why, got.Cmd, got.Dir, got.Exit, want.Why, want.Cmd, want.Dir, want.Exit)
	}
}

// basic.md starts in a workspace without .bcuz/. The others start with it.
// The Japanese translations are the same sessions with Japanese reasons.
func TestExamplesRunAsDocumented(t *testing.T) {
	for _, tc := range []struct {
		file     string
		withBcuz bool
	}{
		{"basic.md", false},
		{"errors.md", true},
		{"shell.md", true},
		{"basic_ja.md", false},
		{"errors_ja.md", true},
		{"shell_ja.md", true},
	} {
		t.Run(tc.file, func(t *testing.T) {
			s := newDocSession(t, tc.withBcuz)
			s.checkDocument(readRepoFile(t, "docs/examples/"+tc.file))
		})
	}
}

// The first example of README.md is the same session.
func TestReadmeExample(t *testing.T) {
	md := readRepoFile(t, "README.md")
	blocks := allBlocks(md)
	if len(blocks) == 0 || !strings.HasPrefix(blocks[0].text, "$ ") {
		t.Fatal("README.md does not start with an example")
	}
	s := newDocSession(t, false)
	want := blocks[0].text
	if got := s.runSteps(want); mask(got) != mask(want) {
		t.Errorf("README example does not match.\n--- got\n%s--- want\n%s", mask(got), mask(want))
	}
}

// The usage lines of README.md are lines of cli.md or of bcuz -h.
func TestReadmeUsageLines(t *testing.T) {
	// The 書式 block and the usage text of bcuz -h are both in cli_ja.md (the
	// README is in Japanese).
	known := readRepoFile(t, "docs/reference/cli_ja.md")
	var lines []string
	for _, b := range allBlocks(readRepoFile(t, "README.md")) {
		if strings.HasPrefix(b.text, "bcuz ") {
			lines = strings.Split(strings.TrimSuffix(b.text, "\n"), "\n")
		}
	}
	if len(lines) == 0 {
		t.Fatal("no usage block in README.md")
	}
	for _, l := range lines {
		form := strings.TrimRight(regexp.MustCompile(`\s{2,}.*$`).ReplaceAllString(l, ""), " ")
		if !strings.Contains(known, "\n"+form) && !strings.Contains(known, "  "+form) {
			t.Errorf("README.md usage line %q (%q) is not in docs/reference/cli.md", l, form)
		}
	}
}
