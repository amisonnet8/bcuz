*[日本語](decisions_ja.md) | **English***

# Decisions and their reasons

## Specification

### The reason comes first, the command after it
`bcuz "reason" make test`. The arguments of a command may be any number and any shape, so putting the command after the reason and saying "everything else is the command" is the least ambiguous. Like sudo, it is used by adding it in front of an existing command.

### Only the first argument is read as an option
So that bcuz does not take the arguments of the command (such as the `-h` of `ls -h`). `--` is needed only when the reason starts with `-`.

### Shell completion is `--completion <shell>`, not `completion <shell>`
Other tools have a `completion` subcommand, but bcuz has no subcommands: a first argument that is not an option is the reason. Taking the word `completion` would change what `bcuz completion make` does. The scripts are fixed text and do not ask bcuz for candidates, because its options are few and fixed, and a hidden option for that would be one more word read in the first argument. The four scripts are checked against each other by running the real shells in the tests.

### `-h` is help, `-H` is history
`-h` is help in almost every tool. If it showed the history, a person who typed it expecting help would get the history.

### The place to record is the first `.bcuz/` found going up
The history is separate for each project, and the same history is used from a subdirectory. The lookup is the same as in git and similar tools, so there is little to remember. Where the command was typed is kept in `dir`.

### Run even without `.bcuz/` (only a warning)
If a wrapper stops the work, people stop using it. A missing record is noticed through the warning on standard error (AI agents can read it too).

### Provide `--init`
`mkdir .bcuz` works too, but the user should not have to know what to create. `.gitignore` is not created (whether to include it depends on the project).

### The exit code is the command's, as it is
So that adding `bcuz` does not change what scripts and CI decide. bcuz's own errors follow the conventions of `env`, `timeout` and others (125, 126, 127), so that they are unlikely to collide with the exit codes of commands.

### A failure to record does not change the exit code
Recording is a by-product. Even if recording is impossible (the disk is full, for example), the result of the command is not changed.

### 126 and 127 are recorded, 125 is not
126 and 127 mean "an attempt to run with a reason was made", which is worth keeping. 125 is a mistake in using bcuz, and the intention to run is not clear.

### Signals
The Ctrl-C of a terminal reaches the whole foreground process group (bcuz and the command). If bcuz ended first on it, the record would not be left. So bcuz does not end on SIGINT or SIGQUIT, and waits for the command to end. It does not forward them (they would arrive twice). SIGTERM and SIGHUP are often sent to bcuz only, so they are forwarded to the command.
A command ended by a signal returns 128 + number, as a shell does. bcuz does not kill itself again with the same signal (simplicity first).

### Standard input and output are passed without a pipe
So that commands that assume a TTY (colored output, interaction, `vim` and so on) work the same as without bcuz.

## Record format

### JSONL, one line per run, written afterwards
Append only is hard to break, and `grep` and `jq` can read it as it is. It is written after the end, so the exit code and the duration are in the same line.

### `cmd` is an array
So that arguments with spaces or quotes are kept intact. It is rewritten into a form a shell can run only when displayed.

### `ts` is the start time, in UTC
It keeps "when it was typed". The end time is known from `ts` + `ms`. The data is kept in UTC, so that it means the same thing wherever it is read and sorts the same way as text. The time zone is applied only when the history is displayed (the viewer's time zone).

### Include `ms`
With one field, slow commands (tests, builds) can be found afterwards.

### `dir` is relative to `.bcuz/`
An absolute path contains where the repository is and the name of the user, which gets in the way when shared.

### Include `v`
To tell records apart from old ones when the format changes.

## Displaying the history

### Newest first
What people want to see is usually what happened recently.

### Dates are compared in the time zone of the viewer
So that "today" means the viewer's today.

### No colors, pager or alignment
To keep it minimal. `bcuz -H | less` is enough.

### No archive feature: moved records are read from standard input (`-H -`)
Moving, compressing or deleting records would break "append only". People move old records away themselves (`mv`, `gzip`), and `-H -` reads them from standard input: `gzip -dc … | bcuz -H -`. The records are put in order by `ts`, so several files can be joined, and `jq` can narrow them down first.

### Standard input is read only with `-`
If `-H` read standard input whenever it is not a terminal, `bcuz -H` run by an AI agent or in CI would wait for input that never comes, or show an empty history instead of `history.jsonl`.

## How it is made

### A single Go binary, no external dependencies
Easy to install (`go install`). It works the same on the three OSes. JSON, starting processes and signals are covered by the standard library.

### Messages are in English
As the messages of a CLI, they can be read in any environment and are easy to search for. The documents are in English, with a Japanese translation next to each (`_ja.md`).
