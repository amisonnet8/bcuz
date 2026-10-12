*[日本語](README_ja.md) | **English***

# bcuz documents

**bcuz** is a wrapper that runs a command with a reason, and keeps "when, why, what and the result" as JSONL.

```
bcuz "run the tests" make test
```

## Layout

| Place | Content | For |
|---|---|---|
| [reference/](reference/) | **The source of truth.** The promises to keep (how to use the command, the format of the records) | users and makers |
| [design/](design/) | Why it is decided so, what it does not do, limitations and candidates | makers |
| [examples/](examples/) | Examples of real input and output | users and makers (the answers of the tests) |

### reference/

- [cli.md](reference/cli.md): how to use the command (running, `--init`, `-H`/`--history`, `-h`/`--help`, `--version`), how `.bcuz/` is found, exit codes, messages
- [history.md](reference/history.md): the format of the records (one line of `history.jsonl`)

### design/

- [overview.md](design/overview.md): purpose, when to use it, what it does not do
- [decisions.md](design/decisions.md): decisions and their reasons
- [limitations.md](design/limitations.md): limitations and candidates

### examples/

- [README.md](examples/README.md): the list of examples

## Rules for the documents

- `reference/` is the source of truth. If the implementation differs, it is made to follow `reference/`. When the specification changes, `reference/` is changed first
- The output in `examples/` is character for character the same as the output of the implementation (values that change on every run, such as time and duration, are declared as such in the example)
- The English files (`X.md`) are the source of truth. The Japanese files (`X_ja.md`) are translations, and always have the same structure
