*[日本語](shell_ja.md) | **English***

# Shell syntax and quoting

## `&&` and pipes work outside of bcuz

bcuz does not go through a shell. In the next example bcuz runs and records only `make test`. `make lint` is run by the shell that called bcuz, after bcuz (it is not recorded).

```
$ bcuz "tests and lint" make test && make lint
```

To record everything together, wrap it in `sh -c`.

```
$ bcuz "record the whole pipe" sh -c 'grep -c TODO *.go | sort -t: -k2 -n'
main.go:0
cli.go:2
```

```json
{"v":1,"ts":"2026-10-12T06:50:00.000+09:00","why":"record the whole pipe","cmd":["sh","-c","grep -c TODO *.go | sort -t: -k2 -n"],"dir":".","exit":0,"ms":120}
```

## Arguments are kept as received

```
$ bcuz "commit" git commit -m "fix: handle empty reason"
```

```json
{"v":1,"ts":"2026-10-12T06:52:10.000+09:00","why":"commit","cmd":["git","commit","-m","fix: handle empty reason"],"dir":".","exit":0,"ms":1200}
```

The history shows it in a form that a shell can run (an argument that contains anything other than ASCII letters, digits and `-_./=:,+@%` is wrapped in `'…'`).

```
2026-10-12 06:52:10    0    1.2s  .  commit  $ git commit -m 'fix: handle empty reason'
```

`*` is expanded first by the shell. An argument passed without expansion is shown wrapped in `'…'`.

```
$ bcuz "check the tags" git tag --list 'v0.*'
```

```
2026-10-10 17:55:01    0   950ms  .  check the tags  $ git tag --list 'v0.*'
```

## Line breaks in the reason

Line breaks and tabs in the reason stay as they are in the record, and become one space in the display of the history.

```
$ bcuz "first migration run
count only" ./migrate --dry-run
```

```json
{"v":1,"ts":"2026-10-12T07:10:00.000+09:00","why":"first migration run\ncount only","cmd":["./migrate","--dry-run"],"dir":".","exit":0,"ms":3050}
```

```
2026-10-12 07:10:00    0    3.0s  .  first migration run count only  $ ./migrate --dry-run
```
