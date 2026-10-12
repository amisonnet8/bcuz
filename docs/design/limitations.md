*[日本語](limitations_ja.md) | **English***

# Limitations and candidates

## Limitations

### Some runs are not recorded
- When bcuz itself is killed with SIGKILL, or the power goes off (the record is written after the command ends)
- When `.bcuz/` is not found (only a warning is printed)
- When bcuz's own arguments are wrong (125)

### Secrets are kept as they are
`cmd` records the arguments as received. A password or a token passed as an argument stays in `history.jsonl`. Pass them through environment variables (environment variables are not recorded). Be especially careful when `.bcuz/` goes into git.

### Shell syntax is not recorded
With `bcuz "reason" make test && make lint`, bcuz sees only `make test` (what follows `&&` is run by the calling shell, outside of bcuz). To record them together, use `bcuz "reason" sh -c 'make test && make lint'`.

### Concurrent writes
A line is written with a single append, so lines do not mix on a local file system. On network file systems (NFS and the like), the atomicity of appends is not guaranteed.

### Windows
- Signals are not forwarded. Ctrl-C does not end bcuz: it waits for the command to end
- The exit code is the value the command returned (there is no conversion to 128 + number)

### Behavior when ended by a signal
When the command is ended by a signal, bcuz **ends normally** with 128 + number. To the parent process it looks like "ended with exit code 130" and not like "ended by a signal". Most shell scripts treat them the same, but strictly they differ.

## Candidates (undecided)

When one is decided, it is written in `reference/` and removed from here.

| Candidate | Content |
|---|---|
| `-n N` | Only the newest N records of the history |
| `--date today`, `yesterday` | Words in place of a date |
| `--here` | Only the records of the current directory (`dir`) |
| `--failed` | Only exit codes other than 0 |
| `--json` | Print the filtered lines as JSONL as they are (to pass to `jq`) |
| Configuration for AI agents | With a Claude Code hook, check that Bash commands have `bcuz` in front |
| Look up the reason afterwards | Match records against `git commit` and the like |
