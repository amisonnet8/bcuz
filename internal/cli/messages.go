package cli

// The texts below are the contract in docs/reference/cli.md. Keep them
// character for character the same as the document.

const usage = `bcuz - run a command with a reason, and record it

usage:
  bcuz [--] <reason> <command> [args...]
  bcuz -H, --history [--date D | --since D] [--until D]
  bcuz --init
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

// fmtNotImplemented is a stand-in for the features that later stages (R1 to
// R3 in dev/roadmap.md) replace. Delete it when the last one is implemented.
const fmtNotImplemented = "bcuz: not implemented yet: %s"

// exitUsage is the exit code of bcuz's own errors (docs/reference/cli.md).
const exitUsage = 125
