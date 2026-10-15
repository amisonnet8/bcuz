*[日本語](cli_ja.md) | **English***

# Command reference

## Synopsis

```
bcuz [--] <reason> <command> [args...]    run the command and record it
bcuz -H [filter]                          show the history (same as --history)
bcuz --init                               create .bcuz/ in the current directory
bcuz -h                                   show the usage (same as --help)
bcuz --version                            show the version
```

### How arguments are read

- **Only the first argument is read as an option.** If it is one of `-H`, `--history`, `--init`, `-h`, `--help` or `--version`, that feature runs
- If the first argument is `--`, it is dropped and the second argument is the reason
- If the first argument starts with `-` and is anything else, it is an error (125). To start a reason with `-`, put `--` before it
- Otherwise the first argument is the reason, and the rest is the command and its arguments. **The arguments of the command are never interpreted** (the `-h` in `bcuz "check" ls -h` goes to `ls`)

## Running a command

```
bcuz "run the tests" make test
```

1. If the reason is empty (or only whitespace), it is an error (125). Nothing is run
2. If there is no command, it is an error (125)
3. `.bcuz/` is looked for (see "How `.bcuz/` is found" below). Not finding it does not stop the run
4. The command is looked up in `PATH` and started. **No shell is used** (`&&`, `|`, `*` and the like are interpreted first by the shell that called bcuz)
5. bcuz waits for the command to end
6. If `.bcuz/` was found, one line is appended ([history.md](history.md))
7. bcuz exits with the exit code of the command

### What passes through

- Standard input, standard output and standard error are given to the command **as they are** (no pipe in between). A TTY stays a TTY. bcuz does not read or keep the output
- Environment variables and the current directory are inherited as they are
- **Ctrl-C** (SIGINT) and SIGQUIT: the terminal sends them to the command too. bcuz itself does not exit on them: it waits for the command to end and records it (they are not forwarded, because they would arrive twice)
- **SIGTERM and SIGHUP**: when they reach bcuz, they are forwarded to the command, and bcuz waits for it to end and records it
- On Windows, Ctrl-C (Ctrl-Break) does not end bcuz itself: it waits for the command to end

### Exit codes

| Exit code | Meaning | Recorded |
|---|---|---|
| The command's | The command ended | yes |
| 128 + signal number | The command was ended by a signal (for example 130 for Ctrl-C) | yes |
| 125 | An error of bcuz itself (bad arguments, empty reason, no command) | no |
| 126 | The command was found but cannot be started (no permission, and so on) | yes |
| 127 | The command was not found | yes |

- 126 and 127 are recorded too, because an attempt to run with a reason was made (`ms` is 0)
- **If recording fails, the exit code is still the command's** (only a warning is printed)

### When `.bcuz/` is not found

The command runs. After it ends, the following is printed to standard error. The exit code is the command's.

```
bcuz: not recorded: no .bcuz/ found (run 'bcuz --init')
```

## How `.bcuz/` is found

- Starting from the current directory and going up through its parents, the **first `.bcuz/` directory found** is used. If none is found up to the root of the file system, it is "not found"
- Records are appended to `history.jsonl` in it. If `.bcuz/` exists but `history.jsonl` does not, it is created
- The boundary of a repository such as git does not stop the search
- The history (`-H`) is found the same way

## `--init`

```
bcuz --init
```

- Creates `.bcuz/` and an empty `.bcuz/history.jsonl` in the current directory
- If the current directory already has `.bcuz/`, nothing is changed, it says so, and it ends with success (0)
- It is created even if a parent has `.bcuz/` (and it says so). From then on, runs below this directory are recorded in the new `.bcuz/`
- `.gitignore` and the like are not created. Whether to put `.bcuz/` in git is up to the user

Output (standard output):

| Case | Output |
|---|---|
| Created | `bcuz: created .bcuz/` |
| Created (a parent has one too) | `bcuz: created .bcuz/ (shadows /home/me/work/.bcuz/)` |
| Already there | `bcuz: .bcuz/ already exists` |

## `-H`, `--history`

```
bcuz -H                                      everything
bcuz -H --date 2026-10-12                    that day only
bcuz -H --since 2026-10-01                   from that day
bcuz -H --until 2026-10-10                   until that day
bcuz -H --since 2026-10-01 --until 2026-10-10
```

### Filters

| Option | Meaning |
|---|---|
| `--date YYYY-MM-DD` | Only runs that **started** on that day |
| `--since YYYY-MM-DD` | From that day on (that day included) |
| `--until YYYY-MM-DD` | Until that day (that day included) |

- Dates are compared as dates in **the time zone of the viewer** (`ts` is converted to the local time zone first, and then its date is taken)
- `--date` cannot be used with `--since` or `--until` (125)
- Writing the same option twice is an error (125)
- A date in the wrong form, or one that does not exist (`2026-02-30`), is an error (125)
- If `--since` is after `--until`, nothing is printed and it succeeds (0)

### Display

- One line per record on standard output, **newest first**
- If there is no record, nothing is printed and it succeeds (0)

```
2026-10-12 06:35:12    0    8.3s  backend  run the tests  $ make test
```

| Column | Content |
|---|---|
| Date and time | `ts` converted to the local time zone, as `YYYY-MM-DD HH:MM:SS` |
| Exit code | Right-aligned, 3 characters |
| Duration | Right-aligned, 6 characters. See "Duration" below |
| Place | `dir` (`.` for the top) |
| Reason | `why`. Line breaks and tabs are replaced by one space |
| Command | `$ ` followed by `cmd`, rewritten so that a shell can run it (below) |

Columns are separated by two spaces. Place and reason are not padded to a common width.

#### Duration

| `ms` | Shown as |
|---|---|
| less than 1000 | `123ms` |
| less than 60000 | `8.3s`, `35.0s` (always one decimal, truncated) |
| less than 3600000 | `2m05s` |
| more | `1h02m` |

#### Command

- Arguments are joined with one space
- An argument made only of ASCII letters, digits and `-_./=:,+@%` is shown as it is. Any other (containing a space, a quote, `$` and so on, or the empty string) is wrapped in `'…'`, and each `'` inside becomes `'\''`
- Example: `["git","commit","-m","fix: a b"]` → `git commit -m 'fix: a b'`

### Unreadable lines

- A last line that does not end with a newline is skipped as a write in progress (no warning)
- A line that cannot be read as JSON, or that lacks a required field, is skipped and a warning goes to standard error. The exit code stays 0

```
bcuz: warning: .bcuz/history.jsonl:12: skipped unreadable line
```

### When `.bcuz/` is not found

```
bcuz: no .bcuz/ found (run 'bcuz --init')
```

It is printed to standard error, and bcuz ends with 125.

### When the history cannot be read

If `history.jsonl` does not exist, it is treated as having no records (nothing is printed, and it ends with 0). If it exists but cannot be read, a message goes to standard error and bcuz ends with 125.

```
bcuz: cannot read .bcuz/history.jsonl: permission denied
```

## `-h`, `--help`

Prints the usage to standard output and ends with 0.

```
bcuz - run a command with a reason, and record it

usage:
  bcuz [--] <reason> <command> [args...]
  bcuz -H, --history [--date D | --since D] [--until D]
  bcuz --init
  bcuz -h, --help
  bcuz --version

dates are YYYY-MM-DD. records go to the nearest .bcuz/history.jsonl.
```

## `--version`

```
bcuz v0.1.1
```

The version comes from the build information (the tag that `go install …@v0.1.1` installs). When it is not available (a local build), it is `bcuz (devel)`.

## Error messages

All go to standard error as one line that starts with `bcuz: `.

| Case | Message | Exit code |
|---|---|---|
| No arguments | `bcuz: missing reason and command (see 'bcuz -h')` | 125 |
| Empty reason | `bcuz: reason is empty` | 125 |
| No command | `bcuz: missing command after reason` | 125 |
| Unknown option | `bcuz: unknown option: -x (use '--' before a reason starting with '-')` | 125 |
| Bad date | `bcuz: invalid date: 2026-02-30` | 125 |
| `--date` with a range | `bcuz: --date cannot be used with --since or --until` | 125 |
| Duplicate option | `bcuz: duplicate option: --since` | 125 |
| Unknown argument after `-H` | `bcuz: unknown argument: --foo` | 125 |
| Missing date | `bcuz: missing date after --date` | 125 |
| No `.bcuz/` (run) | `bcuz: not recorded: no .bcuz/ found (run 'bcuz --init')` | the command's |
| No `.bcuz/` (`-H`) | `bcuz: no .bcuz/ found (run 'bcuz --init')` | 125 |
| History cannot be read (`-H`) | `bcuz: cannot read .bcuz/history.jsonl: permission denied` | 125 |
| Unreadable line (`-H`) | `bcuz: warning: .bcuz/history.jsonl:12: skipped unreadable line` | 0 |
| Command not found | `bcuz: command not found: mkae` | 127 |
| Cannot run | `bcuz: cannot run ./build.sh: permission denied` | 126 |
| Cannot record | `bcuz: warning: not recorded: <reason>` | the command's |
| `--init` cannot create | `bcuz: cannot create .bcuz/: <reason>` | 125 |
