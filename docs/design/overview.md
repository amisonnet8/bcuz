*[日本語](overview_ja.md) | **English***

# Overview

## Purpose

When you type a command, keep **why you typed it** together with it.

```
bcuz "run the tests" make test
```

When you look at the history later, you see not only "what was done", but also "why" and "what came of it (exit code and duration)".

The name comes from "because" (bcuz). Like `sudo`, it works by putting it in front.

## When to use it

- When you let an **AI agent** (such as Claude Code) type commands, have it put `bcuz` in front. Later, a human can read with `bcuz -H` what the agent typed and for what purpose
- When a human puts it in front of work that they want to look back on (a deploy, a data migration, a trial in an investigation)

## How it is made

- A single Go binary `bcuz`. No external dependencies
- Linux, macOS and Windows
- Records go to the first `.bcuz/history.jsonl` found from the current directory going up through its parents

```
project/
├── .bcuz/
│   └── history.jsonl    ← appended here
├── backend/             ← typed here, it is still recorded in the .bcuz/ above (dir is "backend")
└── frontend/
```

## What it does not do

- **It does not raise privileges.** The name looks like sudo, but it has nothing of sudo
- **It does not keep output.** Standard output and standard error only pass through
- **It does not go through a shell.** To record a pipe or `&&` as a whole, write `bcuz "reason" sh -c '…'`
- **It does not force recording.** The command runs even without `.bcuz/` (only a warning)
- **It does not rewrite or delete records.** Append only
- It has no configuration file, colors, pager, or output formats such as `--json` (candidates in [limitations.md](limitations.md))
