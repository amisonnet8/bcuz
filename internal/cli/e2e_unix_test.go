//go:build unix

package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/amisonnet8/bcuz/internal/record"
)

func init() {
	// Catch SIGTERM and SIGHUP, say which came, and end with 42.
	// Like "wait", and a SIGTERM is announced ("got terminated") instead of
	// ending it. Signals reach bcuz's loop in the order of their numbers
	// (SIGINT before SIGTERM), so once this has been announced, a SIGINT sent
	// before it has been dealt with.
	childModes["echo"] = func(string) int {
		ch := make(chan os.Signal, 2)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		go func() {
			for s := range ch {
				if s == os.Interrupt {
					os.Exit(130)
				}
				fmt.Println("got", s)
			}
		}()
		fmt.Println("ready")
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			fmt.Println("pong")
		}
		return 0
	}
	childModes["handle"] = func(string) int {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
		fmt.Println("ready")
		s := <-ch
		fmt.Println("got", s)
		return 42
	}
}

var buildOnce sync.Once
var buildErr error

// bcuzBinary builds the real bcuz once per test run.
func bcuzBinary(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		testBinDir, buildErr = os.MkdirTemp("", "bcuz-e2e-")
		if buildErr != nil {
			return
		}
		out := filepath.Join(testBinDir, "bcuz")
		cmd := exec.Command("go", "build", "-o", out, "github.com/amisonnet8/bcuz/cmd/bcuz")
		cmd.Dir = pkgDir // inside the module; tests have moved to a temporary directory
		if b, err := cmd.CombinedOutput(); err != nil {
			buildErr = fmt.Errorf("go build: %w\n%s", err, b)
		}
	})
	if buildErr != nil {
		t.Fatal(buildErr)
	}
	return filepath.Join(testBinDir, "bcuz")
}

// bcuzProc is a running bcuz whose child is this test binary.
type bcuzProc struct {
	cmd   *exec.Cmd
	stdin io.WriteCloser
	out   *bufio.Reader
	work  string
}

// start runs `bcuz why <test binary>` in a new workspace, in its own process
// group, with the child in the given mode.
func start(t *testing.T, mode string) *bcuzProc {
	t.Helper()
	work := workspace(t)

	cmd := exec.Command(bcuzBinary(t), "why", self(t))
	cmd.Dir = work
	cmd.Env = append(os.Environ(), childEnv+"="+mode)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	p := &bcuzProc{cmd: cmd, stdin: stdin, out: bufio.NewReader(stdout), work: work}

	// Never leave a process behind, and never hang the test run.
	timer := time.AfterFunc(30*time.Second, func() { _ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) })
	t.Cleanup(func() {
		timer.Stop()
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		_ = cmd.Wait()
	})
	return p
}

// line reads the next line of the standard output.
func (p *bcuzProc) line(t *testing.T) string {
	t.Helper()
	s, err := p.out.ReadString('\n')
	if err != nil {
		t.Fatalf("reading the output: %v (got %q)", err, s)
	}
	return strings.TrimSuffix(s, "\n")
}

// wait waits for bcuz to end and returns its exit code. It leaves the input
// open: closing it would let a child that waits for it end by itself, and the
// signal sent before might lose the race.
func (p *bcuzProc) wait(t *testing.T) int {
	t.Helper()
	err := p.cmd.Wait()
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		t.Fatal(err)
	}
	if ws, ok := ee.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		t.Fatalf("bcuz itself was ended by %v", ws.Signal())
	}
	return ee.ExitCode()
}

func (p *bcuzProc) lastRecord(t *testing.T) record.Record {
	t.Helper()
	recs := records(t, p.work)
	if len(recs) != 1 {
		t.Fatalf("records = %d, want 1", len(recs))
	}
	return recs[0]
}

// A SIGINT that reaches only bcuz neither ends bcuz nor reaches the command.
func TestSIGINTToBcuzAlone(t *testing.T) {
	p := start(t, "echo")
	if got := p.line(t); got != "ready" {
		t.Fatalf("first line = %q", got)
	}

	if err := p.cmd.Process.Signal(syscall.SIGINT); err != nil {
		t.Fatal(err)
	}
	// The SIGTERM that follows is forwarded, and the child announces it. By
	// then bcuz has dealt with the SIGINT, so a forwarded one would have
	// ended the child already.
	if err := p.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if got := p.line(t); got != "got terminated" {
		t.Fatalf("after SIGINT and SIGTERM: %q, want \"got terminated\"", got)
	}
	// Both are still alive: the child answers, and bcuz passes it on.
	if _, err := io.WriteString(p.stdin, "ping\n"); err != nil {
		t.Fatal(err)
	}
	if got := p.line(t); got != "pong" {
		t.Fatalf("after SIGINT: %q, want pong", got)
	}

	// Now let the child end by itself.
	_ = p.stdin.Close()
	if code := p.wait(t); code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if r := p.lastRecord(t); r.Exit != 0 {
		t.Errorf("record = %+v", r)
	}
}

// Ctrl-C at a terminal reaches the whole process group: the command ends
// with 130, and bcuz stays to record it and to end with 130.
func TestCtrlCReachesTheGroup(t *testing.T) {
	p := start(t, "wait")
	if got := p.line(t); got != "ready" {
		t.Fatalf("first line = %q", got)
	}

	if err := syscall.Kill(-p.cmd.Process.Pid, syscall.SIGINT); err != nil {
		t.Fatal(err)
	}
	if code := p.wait(t); code != 130 {
		t.Errorf("exit code = %d, want 130", code)
	}
	if r := p.lastRecord(t); r.Exit != 130 {
		t.Errorf("record = %+v, want exit 130", r)
	}
}

func TestSIGQUITIsNotFatalToBcuz(t *testing.T) {
	p := start(t, "handle")
	if got := p.line(t); got != "ready" {
		t.Fatalf("first line = %q", got)
	}
	// SIGQUIT goes to bcuz alone: it is dropped, not forwarded. Then SIGHUP
	// reaches the child, which proves bcuz is alive and the child got no SIGQUIT.
	if err := p.cmd.Process.Signal(syscall.SIGQUIT); err != nil {
		t.Fatal(err)
	}
	if err := p.cmd.Process.Signal(syscall.SIGHUP); err != nil {
		t.Fatal(err)
	}
	if got := p.line(t); got != "got hangup" {
		t.Errorf("child saw %q, want \"got hangup\"", got)
	}
	if code := p.wait(t); code != 42 {
		t.Errorf("exit code = %d, want 42", code)
	}
}

// SIGTERM and SIGHUP that reach bcuz are passed on to the command.
func TestSignalsAreForwarded(t *testing.T) {
	for _, tt := range []struct {
		sig  syscall.Signal
		want string
	}{
		{syscall.SIGTERM, "got terminated"},
		{syscall.SIGHUP, "got hangup"},
	} {
		t.Run(tt.want, func(t *testing.T) {
			p := start(t, "handle")
			if got := p.line(t); got != "ready" {
				t.Fatalf("first line = %q", got)
			}
			if err := p.cmd.Process.Signal(tt.sig); err != nil {
				t.Fatal(err)
			}
			if got := p.line(t); got != tt.want {
				t.Errorf("child saw %q, want %q", got, tt.want)
			}
			if code := p.wait(t); code != 42 {
				t.Errorf("exit code = %d, want the command's 42", code)
			}
			if r := p.lastRecord(t); r.Exit != 42 {
				t.Errorf("record = %+v", r)
			}
		})
	}
}

// A command that is ended by a signal gives 128 plus its number, recorded.
func TestEndedBySignal(t *testing.T) {
	p := start(t, "wait")
	if got := p.line(t); got != "ready" {
		t.Fatalf("first line = %q", got)
	}
	if err := p.cmd.Process.Signal(syscall.SIGTERM); err != nil { // forwarded, ends the child
		t.Fatal(err)
	}
	if code := p.wait(t); code != 143 {
		t.Errorf("exit code = %d, want 143", code)
	}
	if r := p.lastRecord(t); r.Exit != 143 {
		t.Errorf("record = %+v, want exit 143", r)
	}
}

// With a terminal, the command sees a terminal: bcuz puts no pipe between.
func TestTerminalPassesThrough(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("script -qec is the util-linux form")
	}
	script, err := exec.LookPath("script")
	if err != nil {
		t.Skip("script is not installed")
	}
	work := workspace(t)

	cmd := exec.Command(script, "-qec", bcuzBinary(t)+" why "+self(t), "/dev/null")
	cmd.Dir = work
	cmd.Env = append(os.Environ(), childEnv+"=istty")
	cmd.Stdin = strings.NewReader("")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("script could not allocate a terminal here: %v: %s", err, out)
	}
	if !strings.Contains(string(out), "stdin=true stdout=true stderr=true") {
		t.Errorf("output = %q, want the command to see a terminal on all three", out)
	}
	if recs := records(t, work); len(recs) != 1 || recs[0].Exit != 0 {
		t.Errorf("records = %+v", recs)
	}
}

// The real bcuz reads a file given as standard input with "-", where there
// is no .bcuz/.
func TestHistoryStdinFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "old.jsonl")
	if err := os.WriteFile(path, []byte(goodLine), 0o600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	cmd := exec.Command(bcuzBinary(t), "-H", "-")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "TZ=UTC")
	cmd.Stdin = f
	out, err := cmd.Output()
	want := "2026-10-11 21:35:12    0    8.3s  .  xxテスト実施  $ make test\n"
	if err != nil || string(out) != want {
		t.Errorf("err %v, stdout %q; want %q", err, out, want)
	}
}

// Without "-", bcuz -H ends even when standard input is a pipe that is never
// closed, as with an AI agent or CI.
func TestHistoryDoesNotWaitForStdin(t *testing.T) {
	work := workspace(t)
	cmd := exec.Command(bcuzBinary(t), "-H")
	cmd.Dir = work
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stdin.Close() }()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("bcuz -H: %v", err)
		}
	case <-time.After(10 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("bcuz -H waited for standard input")
	}
}
