// Package cli reads bcuz's arguments and decides what to do.
package cli

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/amisonnet8/bcuz/internal/history"
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
		return showHistory(args[1:], stdin, stdout, stderr)
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

// parseHistoryArgs reads what follows -H, in any order: "-" to read standard
// input, and --date, --since and --until, each with one date. It returns the
// message of the first mistake, from the left.
func parseHistoryArgs(args []string) (f history.Filter, fromStdin bool, msg string) {
	seen := map[string]bool{}
	for i := 0; i < len(args); i++ {
		name := args[i]
		if seen[name] {
			return f, fromStdin, fmt.Sprintf(fmtDuplicateOption, name)
		}
		var field *string
		switch name {
		case "-":
			seen[name] = true
			fromStdin = true
			continue
		case "--date":
			field = &f.Date
		case "--since":
			field = &f.Since
		case "--until":
			field = &f.Until
		default:
			return f, fromStdin, fmt.Sprintf(fmtUnknownArgument, name)
		}
		seen[name] = true
		if i+1 >= len(args) {
			return f, fromStdin, fmt.Sprintf(fmtMissingDate, name)
		}
		i++
		if history.ParseDate(args[i]) != nil {
			return f, fromStdin, fmt.Sprintf(fmtInvalidDate, args[i])
		}
		*field = args[i]
	}
	if f.Date != "" && (f.Since != "" || f.Until != "") {
		return f, fromStdin, msgDateWithRange
	}
	return f, fromStdin, ""
}

// showHistory prints the history (-H), newest first: the one of the nearest
// .bcuz/, or, with "-", what comes on standard input (no .bcuz/ is needed).
func showHistory(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	f, fromStdin, msg := parseHistoryArgs(args)
	if msg != "" {
		return fail(stderr, msg)
	}
	if fromStdin {
		return printHistory(stdin, placeStdin, f, stdout, stderr)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fail(stderr, msgNoBcuzDirHist)
	}
	root, found := record.FindRoot(cwd)
	if !found {
		return fail(stderr, msgNoBcuzDirHist)
	}

	file, err := os.Open(record.HistoryPath(root))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return 0
		}
		return fail(stderr, fmt.Sprintf(fmtCannotRead, placeHistoryFile, record.Reason(err)))
	}
	defer func() { _ = file.Close() }()
	return printHistory(file, placeHistoryFile, f, stdout, stderr)
}

// printHistory reads the records from in and prints those that pass f,
// newest first. place is what the messages call in.
func printHistory(in io.Reader, place string, f history.Filter, stdout, stderr io.Writer) int {
	recs, skipped, err := history.Read(in)
	if err != nil {
		return fail(stderr, fmt.Sprintf(fmtCannotRead, place, record.Reason(err)))
	}
	for _, n := range skipped {
		write(stderr, fmt.Sprintf(fmtUnreadableLine, place, n)+"\n")
	}
	var out strings.Builder
	for _, r := range history.Select(recs, f) {
		out.WriteString(history.FormatLine(r))
		out.WriteByte('\n')
	}
	write(stdout, out.String())
	return 0
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
