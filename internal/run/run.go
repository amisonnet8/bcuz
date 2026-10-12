// Package run starts a command and waits for it. It does not know about
// records or messages: it returns how the command ended, when it started and
// how long it ran (docs/reference/cli.md, "実行").
package run

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"time"
)

// Outcome says how the command came to an end.
type Outcome int

const (
	// Ran means the command started and ended (by itself or by a signal).
	Ran Outcome = iota
	// NotFound means the command was not found. The exit code is 127.
	NotFound
	// CannotRun means the command was found but could not be started (no
	// permission, not an executable). The exit code is 126.
	CannotRun
)

// Exit codes of a command that did not start.
const (
	exitCannotRun = 126
	exitNotFound  = 127
)

// Result is what happened to the command.
type Result struct {
	Outcome  Outcome
	ExitCode int
	// Start is the time just before the command was started.
	Start time.Time
	// Duration runs from just before the start to the end of the command. It
	// is zero when the command did not start.
	Duration time.Duration
	// Reason is the plain cause (such as "permission denied") when Outcome is
	// CannotRun.
	Reason string
}

// Run starts args[0] with args[1:], without a shell, and waits for it. The
// standard streams are given to the command as they are: an *os.File is
// passed on as the file itself, so a terminal stays a terminal.
//
// While the command runs, SIGINT and SIGQUIT are received and dropped (the
// terminal sends them to the command as well), and SIGTERM and SIGHUP are
// forwarded to the command (signal_unix.go).
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) Result {
	//nolint:gosec // running the command that the user typed is the purpose of bcuz
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdin, stdout, stderr

	// Subscribe before the start, so that no signal between the start and the
	// first receive takes bcuz down.
	sigs := make(chan os.Signal, 8)
	signal.Notify(sigs, notified...)
	defer signal.Stop(sigs)

	start := time.Now()
	if err := cmd.Start(); err != nil {
		return startFailure(err, start)
	}

	done := make(chan struct{})
	go func() {
		for {
			select {
			case s := <-sigs:
				if forwarded(s) {
					_ = cmd.Process.Signal(s) // an ended process is not an error here
				}
			case <-done:
				return
			}
		}
	}()

	_ = cmd.Wait() // the end of the command is read from ProcessState
	elapsed := time.Since(start)
	close(done)

	return Result{
		Outcome:  Ran,
		ExitCode: exitCode(cmd.ProcessState),
		Start:    start,
		Duration: elapsed,
	}
}

// startFailure turns the error of Start into a result.
func startFailure(err error, start time.Time) Result {
	if errors.Is(err, exec.ErrNotFound) || errors.Is(err, fs.ErrNotExist) {
		return Result{Outcome: NotFound, ExitCode: exitNotFound, Start: start}
	}
	return Result{Outcome: CannotRun, ExitCode: exitCannotRun, Start: start, Reason: reason(err)}
}

// reason returns the cause of an error without the operation and the path.
func reason(err error) string {
	var pe *fs.PathError
	if errors.As(err, &pe) {
		return pe.Err.Error()
	}
	var ee *exec.Error
	if errors.As(err, &ee) {
		return ee.Err.Error()
	}
	return err.Error()
}
