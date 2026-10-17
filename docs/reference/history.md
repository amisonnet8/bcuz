*[日本語](history_ja.md) | **English***

# Record format (`.bcuz/history.jsonl`)

## File

- UTF-8 JSON Lines. **One run = one line**. A line ends with `\n`
- **Append only.** Existing lines are never rewritten or deleted
- A line is written with **one write call**, in append mode (`O_APPEND`), so that lines do not mix even when several bcuz write at the same time
- It is written after the command ends. There is no record of a run in progress

## Shape of a line

```json
{"v":1,"ts":"2026-10-11T21:35:12.345Z","why":"run the tests","cmd":["make","test"],"dir":"backend","exit":0,"ms":8312}
```

| Field | Type | Content |
|---|---|---|
| `v` | number | Version of the format. Now `1` |
| `ts` | string | The time the command was **started**. RFC 3339, with milliseconds, **in UTC** (it ends with `Z`). `-H` shows it in the time zone of the viewer |
| `why` | string | The reason, as received (line breaks included). Never empty |
| `cmd` | array of strings | The command and its arguments, as received. At least one element |
| `dir` | string | Where it ran, as a path relative to the directory that has `.bcuz/`. The separator is `/` on every OS. The top is `"."` |
| `exit` | number | The exit code. 128 + number when ended by a signal. 126 if it could not be started, 127 if it was not found |
| `ms` | number | Milliseconds from start to end (an integer). `0` if it could not be started |

- Fields are written in this order (readers must not rely on the order)
- All seven fields are required. There is no field without a value

## Compatibility

- **Adding a field is fine** (`v` stays). Readers ignore fields they do not know
- Changing the meaning of a field, or removing one, raises `v`
- A reader (`-H`) skips a line whose `v` it does not know, as an unreadable line
- A reader accepts any time zone offset in `ts` (RFC 3339). Records written by earlier versions, with an offset such as `+09:00`, stay as they are and are read as before

## Archiving

bcuz never moves, compresses or deletes records. To keep `history.jsonl` small, move it away yourself. The next run creates a new `history.jsonl`.

```
mv .bcuz/history.jsonl .bcuz/history-2026.jsonl
gzip .bcuz/history-2026.jsonl
gzip -dc .bcuz/history-2026.jsonl.gz | bcuz -H -
```

- `-H` reads only `history.jsonl`. Moved records are read with `-H -` from standard input ([cli.md](cli.md))
- Several files can be read together: `gzip -dc .bcuz/history-*.jsonl.gz | cat - .bcuz/history.jsonl | bcuz -H -`
- Use `gzip -dc` rather than `zcat`: on macOS, `zcat` may not read `.gz`

## Example

[../examples/history.jsonl](../examples/history.jsonl)
