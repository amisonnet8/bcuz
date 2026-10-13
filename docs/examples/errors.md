*[日本語](errors_ja.md) | **English***

# Failures and mistakes

## No `.bcuz/`

The command runs. A warning appears after it ends. The exit code is the command's.

```
$ cd /tmp/scratch
$ bcuz "try it out" echo hello
hello
bcuz: not recorded: no .bcuz/ found (run 'bcuz --init')
$ echo $?
0
```

The history cannot be seen either.

```
$ bcuz -H
bcuz: no .bcuz/ found (run 'bcuz --init')
$ echo $?
125
```

## Command not found (recorded)

```
$ cd /home/me/work
$ bcuz "build" mkae build
bcuz: command not found: mkae
$ echo $?
127
```

```json
{"v":1,"ts":"2026-10-11T21:45:00.000Z","why":"build","cmd":["mkae","build"],"dir":".","exit":127,"ms":0}
```

## Cannot be started (recorded)

```
$ bcuz "try the script" ./build.sh
bcuz: cannot run ./build.sh: permission denied
$ echo $?
126
```

## Wrong arguments (not recorded)

```
$ bcuz
bcuz: missing reason and command (see 'bcuz -h')
$ echo $?
125
$ bcuz "tests"
bcuz: missing command after reason
$ bcuz "  " make test
bcuz: reason is empty
$ bcuz -x make test
bcuz: unknown option: -x (use '--' before a reason starting with '-')
```

To start a reason with `-`, put `--` before it.

```
$ bcuz -- "-v this time" go test -v ./...
```

```json
{"v":1,"ts":"2026-10-11T22:00:00.000Z","why":"-v this time","cmd":["go","test","-v","./..."],"dir":".","exit":0,"ms":9100}
```

## `--init` where a parent has `.bcuz/` too

```
$ cd /home/me/work/frontend
$ bcuz --init
bcuz: created .bcuz/ (shadows /home/me/work/.bcuz/)
```

From then on, runs below `frontend/` are recorded in `frontend/.bcuz/`.
