*[日本語](README_ja.md) | **English***

# bcuz

<p align="center">
  <img src="docs/images/banner.svg" alt="bcuz: run a command with a reason" width="720">
</p>

<p align="center">
  <b>Run a command with a reason, and keep a record of when, why, what and the result.</b>
</p>

<p align="center">
  <a href="https://github.com/amisonnet8/bcuz/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/amisonnet8/bcuz/actions/workflows/ci.yml/badge.svg"></a>
  <a href="https://github.com/amisonnet8/bcuz/releases"><img alt="Release" src="https://img.shields.io/github/v/release/amisonnet8/bcuz"></a>
  <a href="LICENSE"><img alt="License: MIT" src="https://img.shields.io/github/license/amisonnet8/bcuz"></a>
  <img alt="Go" src="https://img.shields.io/github/go-mod/go-version/amisonnet8/bcuz">
  <img alt="Platforms" src="https://img.shields.io/badge/platform-linux%20%7C%20macos%20%7C%20windows-informational">
</p>

<p align="center">
  <a href="#at-a-glance">At a glance</a> ·
  <a href="#features">Features</a> ·
  <a href="#installation">Installation</a> ·
  <a href="#usage">Usage</a> ·
  <a href="#exit-codes">Exit codes</a> ·
  <a href="#documentation">Documentation</a>
</p>

## At a glance

```console
$ bcuz --init
bcuz: created .bcuz/
$ bcuz "run the tests" make test
$ bcuz -H
2026-10-12 06:35:12    0    8.3s  .  run the tests  $ make test
```

Put one reason in front of a command. Later, the history tells you **why** you typed it.

## Features

- 📝 **The reason is kept with the command**: `bcuz "reason" command`. Like `sudo`, you only put it in front
- 🔍 **You know later**: when, why, what and the result (exit code and duration), as one JSON line per run
- 🧩 **Nothing else changes**: output, TTY and exit code pass through as they are. No shell is involved
- 📦 **A single binary, no dependencies**: written in Go. Linux, macOS and Windows
- 🤖 **Good for AI agents too**: read with `bcuz -H` what an agent typed and for what purpose
- 🗂️ **Just a file**: the append-only `.bcuz/history.jsonl`. Use it as it is with `grep`, `jq` and git

## Installation

```bash
go install github.com/amisonnet8/bcuz/cmd/bcuz@latest
```

Pre-built binaries (Linux, macOS and Windows, amd64 and arm64) are on the [Releases](https://github.com/amisonnet8/bcuz/releases) page.

## Usage

```bash
bcuz --init                             # create .bcuz/ here (the first time)
bcuz "run the tests" make test          # run and record
bcuz -- "-v this time" go test -v       # a reason that starts with -
bcuz -H                                 # the history, newest first
bcuz -H --date 2026-10-12               # one day (also --since and --until)
```

```text
bcuz [--] <reason> <command> [args...]   run the command and record it
bcuz -H, --history [--date D | --since D] [--until D]
bcuz --init
bcuz -h, --help
bcuz --version
```

<details>
<summary>What a record looks like (one line of <code>.bcuz/history.jsonl</code>)</summary>

```json
{"v":1,"ts":"2026-10-12T06:35:12.345+09:00","why":"run the tests","cmd":["make","test"],"dir":".","exit":0,"ms":8312}
```

</details>

## Good to know

- Records are appended to the first `.bcuz/history.jsonl` found from the current directory going up through its parents
- No shell is used. `&&` and pipes are interpreted by the shell that called bcuz (use `sh -c '…'` to record them whole)
- Without `.bcuz/`, the command still runs (only a warning)

## Exit codes

| Exit code | Meaning |
|---|---|
| The command's | Passed through as it is |
| 128 + number | Ended by a signal (130 for Ctrl-C) |
| 125 | An error of bcuz itself (not recorded) |
| 126 | Found but cannot be started (recorded) |
| 127 | Not found (recorded) |

## Documentation

- [Command reference](docs/reference/cli.md): output, messages, exit codes
- [Record format](docs/reference/history.md): one line of `history.jsonl`
- [Examples](docs/examples/README.md): real input and output
- [Design](docs/design/overview.md): purpose, reasons for the decisions, limitations

## Development

Open it in the devcontainer, and read `CLAUDE.md` and `dev/roadmap.md`. Build and test with `qsoku check`.

## License

[MIT](LICENSE)
