package cli

// The texts below are the contract in docs/reference/cli.md. Keep them
// character for character the same as the document.

const usage = `bcuz - run a command with a reason, and record it

usage:
  bcuz [--] <reason> <command> [args...]
  bcuz -H, --history [-] [--date D | --since D] [--until D]
  bcuz --init
  bcuz --completion <shell>
  bcuz -h, --help
  bcuz --version

dates are YYYY-MM-DD. records go to the nearest .bcuz/history.jsonl.
`

const (
	msgMissingArgs    = "bcuz: missing reason and command (see 'bcuz -h')"
	msgReasonEmpty    = "bcuz: reason is empty"
	msgMissingCommand = "bcuz: missing command after reason"
	fmtUnknownOption  = "bcuz: unknown option: %s (use '--' before a reason starting with '-')"

	msgCreated        = "bcuz: created .bcuz/"
	fmtCreatedShadows = "bcuz: created .bcuz/ (shadows %s)"
	msgAlreadyExists  = "bcuz: .bcuz/ already exists"
	fmtCannotCreate   = "bcuz: cannot create .bcuz/: %s"

	fmtCommandNotFound = "bcuz: command not found: %s"
	fmtCannotRun       = "bcuz: cannot run %s: %s"
	msgNoBcuzDirRun    = "bcuz: not recorded: no .bcuz/ found (run 'bcuz --init')"
	fmtNotRecorded     = "bcuz: warning: not recorded: %s"
)

// The messages of -H.
const (
	fmtInvalidDate     = "bcuz: invalid date: %s"
	msgDateWithRange   = "bcuz: --date cannot be used with --since or --until"
	fmtDuplicateOption = "bcuz: duplicate option: %s"
	fmtUnknownArgument = "bcuz: unknown argument: %s"
	fmtMissingDate     = "bcuz: missing date after %s"
	msgNoBcuzDirHist   = "bcuz: no .bcuz/ found (run 'bcuz --init')"
	fmtCannotRead      = "bcuz: cannot read %s: %s"
	fmtUnreadableLine  = "bcuz: warning: %s:%d: skipped unreadable line"
)

// The messages of --completion. Unknown arguments after the shell use
// fmtUnknownArgument, as after -H.
const (
	msgMissingShell = "bcuz: missing shell after --completion (bash, zsh, fish, powershell)"
	fmtUnknownShell = "bcuz: unknown shell: %s (bash, zsh, fish, powershell)"
)

// The places -H reads from, as messages name them.
const (
	placeHistoryFile = ".bcuz/history.jsonl"
	placeStdin       = "<stdin>"
)

// exitUsage is the exit code of bcuz's own errors (docs/reference/cli.md).
const exitUsage = 125
