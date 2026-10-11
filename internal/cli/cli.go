// Package cli reads bcuz's arguments and decides what to do.
package cli

import (
	"fmt"
	"io"
	"strings"
)

// Main runs bcuz with the arguments after the program name and returns the
// exit code. Only the first argument is read as an option: the arguments of
// the command are never interpreted (docs/reference/cli.md).
func Main(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	_ = stdin // used by the stages that run the command

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
		return fail(stderr, fmt.Sprintf(fmtNotImplemented, "init"))
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
	return fail(stderr, fmt.Sprintf(fmtNotImplemented, "run"))
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
