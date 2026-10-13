*[日本語](README_ja.md) | **English***

# Examples

Real input and output. Lines that start with `$` are input, the others are output.

- The time (`ts`, the date and time column) and the duration (`ms`, the duration column) change on every run. Everything else is character for character the same as the output of the implementation (the output of other commands, such as `make test`, is only an illustration)
- The records are in UTC (`ts` ends with `Z`). The time zone of the viewer in the examples is Asia/Tokyo (+09:00), unless a block says `TZ=UTC`
- The workspace of the examples is `/home/me/work` (`.bcuz/` is there)

| File | Content |
|---|---|
| [basic.md](basic.md) | `--init`, running, failing, showing the history |
| [history-filter.md](history-filter.md) | Filtering by date |
| [errors.md](errors.md) | No `.bcuz/`, wrong arguments, command not found |
| [shell.md](shell.md) | Recording a pipe or `&&`, quoting of arguments |
| [history.jsonl](history.jsonl) | A `history.jsonl` after a few days of use. [history-filter.md](history-filter.md) uses these records |
