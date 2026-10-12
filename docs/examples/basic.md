*[日本語](basic_ja.md) | **English***

# Basics

## Create `.bcuz/` first

```
$ cd /home/me/work
$ bcuz --init
bcuz: created .bcuz/
$ bcuz --init
bcuz: .bcuz/ already exists
```

## Run a command

The output of the command appears as it is. bcuz adds nothing.

```
$ bcuz "run the tests" make test
go test ./...
ok      example.com/work/xx    8.102s
$ echo $?
0
```

One line is added to `.bcuz/history.jsonl`.

```json
{"v":1,"ts":"2026-10-12T06:35:12.345+09:00","why":"run the tests","cmd":["make","test"],"dir":".","exit":0,"ms":8312}
```

## From a subdirectory

It is recorded in the `.bcuz/` above, and `dir` keeps the place.

```
$ cd backend
$ bcuz "check the lint findings" golangci-lint run
handler.go:12:2: ineffectual assignment to err (ineffassign)
1 issues:
* ineffassign: 1
$ echo $?
1
```

```json
{"v":1,"ts":"2026-10-12T06:40:03.020+09:00","why":"check the lint findings","cmd":["golangci-lint","run"],"dir":"backend","exit":1,"ms":2410}
```

## Stop with Ctrl-C

```
$ bcuz "watch the log" tail -f app.log
...
^C
$ echo $?
130
```

```json
{"v":1,"ts":"2026-10-12T06:41:30.500+09:00","why":"watch the log","cmd":["tail","-f","app.log"],"dir":"backend","exit":130,"ms":65200}
```

## See the history

Newest first.

```
$ bcuz -H
2026-10-12 06:41:30  130   1m05s  backend  watch the log  $ tail -f app.log
2026-10-12 06:40:03    1    2.4s  backend  check the lint findings  $ golangci-lint run
2026-10-12 06:35:12    0    8.3s  .  run the tests  $ make test
```

`-H` and `--history` are the same. The same history appears from any subdirectory.

## See the usage

```
$ bcuz -h
bcuz - run a command with a reason, and record it

usage:
  bcuz [--] <reason> <command> [args...]
  bcuz -H, --history [--date D | --since D] [--until D]
  bcuz --init
  bcuz -h, --help
  bcuz --version

dates are YYYY-MM-DD. records go to the nearest .bcuz/history.jsonl.
```

## Options after the reason go to the command

```
$ bcuz "check the size" ls -h
```

`-h` goes to `ls`, not to bcuz (bcuz reads only the first argument as an option).
