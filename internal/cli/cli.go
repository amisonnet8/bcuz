// Package cli reads bcuz's arguments and decides what to do.
package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/amisonnet8/bcuz/internal/record"
	"github.com/amisonnet8/bcuz/internal/run"
)

// Main runs bcuz with the arguments after the program name and returns the
// exit code. Only the first argument is read as an option: the arguments of
// the command are never interpreted (docs/reference/cli.md).
func Main(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return fail(stderr, msgMissingArgs)
	}

	switch args[0] {
	case "-h", "--help":
		write(stdout, usage)
		return 0
	case "--version":
		write(stdout, "bcuz "+version()+"\n")
		return 0
	case "-H", "--history":
		return fail(stderr, fmt.Sprintf(fmtNotImplemented, "history"))
	case "--init":
		return runInit(stdout, stderr)
	case "--":
		args = args[1:]
		if len(args) == 0 {
			return fail(stderr, msgMissingArgs)
		}
	default:
		if strings.HasPrefix(args[0], "-") {
			return fail(stderr, fmt.Sprintf(fmtUnknownOption, args[0]))
		}
	}

	// args[0] is the reason, the rest is the command and its arguments.
	if strings.TrimSpace(args[0]) == "" {
		return fail(stderr, msgReasonEmpty)
	}
	if len(args) < 2 {
		return fail(stderr, msgMissingCommand)
	}
	return execute(args, stdin, stdout, stderr)
}

// execute runs args[1:] with the reason args[0], records it, and returns the
// exit code of the command. A failure to record only warns: the exit code is
// the command's (docs/reference/cli.md).
func execute(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	reason, command := args[0], args[1:]

	cwd, cwdErr := os.Getwd()
	root, found := "", false
	if cwdErr == nil {
		root, found = record.FindRoot(cwd)
	}

	res := run.Run(command, stdin, stdout, stderr)

	switch res.Outcome {
	case run.NotFound:
		write(stderr, fmt.Sprintf(fmtCommandNotFound, command[0])+"\n")
	case run.CannotRun:
		write(stderr, fmt.Sprintf(fmtCannotRun, command[0], res.Reason)+"\n")
	}

	switch {
	case cwdErr != nil:
		write(stderr, fmt.Sprintf(fmtNotRecorded, record.Reason(cwdErr))+"\n")
	case !found:
		write(stderr, msgNoBcuzDirRun+"\n")
	default:
		if err := appendRecord(root, cwd, reason, command, res); err != nil {
			write(stderr, fmt.Sprintf(fmtNotRecorded, record.Reason(err))+"\n")
		}
	}
	return res.ExitCode
}

func appendRecord(root, cwd, reason string, command []string, res run.Result) error {
	dir, err := record.RelDir(root, cwd)
	if err != nil {
		return err
	}
	return record.Append(root, record.Record{
		Time: res.Start,
		Why:  reason,
		Cmd:  command,
		Dir:  dir,
		Exit: res.ExitCode,
		MS:   res.Duration.Milliseconds(),
	})
}

// runInit makes .bcuz/ in the current directory (--init).
func runInit(stdout, stderr io.Writer) int {
	cwd, err := os.Getwd()
	if err != nil {
		return fail(stderr, fmt.Sprintf(fmtCannotCreate, err))
	}
	res, err := record.Init(cwd)
	if err != nil {
		return fail(stderr, fmt.Sprintf(fmtCannotCreate, err))
	}
	switch res.Status {
	case record.CreatedShadowing:
		shadowed := filepath.Join(res.Shadowed, record.DirName) + string(filepath.Separator)
		write(stdout, fmt.Sprintf(fmtCreatedShadows, shadowed)+"\n")
	case record.AlreadyExists:
		write(stdout, msgAlreadyExists+"\n")
	default:
		write(stdout, msgCreated+"\n")
	}
	return 0
}

// fail writes one error line and returns the exit code of bcuz's own errors.
func fail(w io.Writer, msg string) int {
	write(w, msg+"\n")
	return exitUsage
}

// write ignores the error: there is nowhere else to report it.
func write(w io.Writer, s string) {
	_, _ = io.WriteString(w, s)
}
