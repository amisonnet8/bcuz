*[English](basic.md) | **日本語***

# 基本

## 初めに `.bcuz/` を作る

```
$ cd /home/me/work
$ bcuz --init
bcuz: created .bcuz/
$ bcuz --init
bcuz: .bcuz/ already exists
```

## 実行する

コマンドの出力はそのまま出る。bcuz は何も足さない。

```
$ bcuz "xxテスト実施" make test
go test ./...
ok      example.com/work/xx    8.102s
$ echo $?
0
```

`.bcuz/history.jsonl` に1行足される。

```json
{"v":1,"ts":"2026-10-11T21:35:12.345Z","why":"xxテスト実施","cmd":["make","test"],"dir":".","exit":0,"ms":8312}
```

## サブディレクトリから打つ

上の `.bcuz/` に記録され、`dir` に場所が残る。

```
$ cd backend
$ bcuz "lint の指摘を確認" golangci-lint run
handler.go:12:2: ineffectual assignment to err (ineffassign)
1 issues:
* ineffassign: 1
$ echo $?
1
```

```json
{"v":1,"ts":"2026-10-11T21:40:03.020Z","why":"lint の指摘を確認","cmd":["golangci-lint","run"],"dir":"backend","exit":1,"ms":2410}
```

## Ctrl-C で止める

```
$ bcuz "ログを眺める" tail -f app.log
...
^C
$ echo $?
130
```

```json
{"v":1,"ts":"2026-10-11T21:41:30.500Z","why":"ログを眺める","cmd":["tail","-f","app.log"],"dir":"backend","exit":130,"ms":65200}
```

## 履歴を見る

新しい順。

```
$ bcuz -H
2026-10-12 06:41:30  130   1m05s  backend  ログを眺める  $ tail -f app.log
2026-10-12 06:40:03    1    2.4s  backend  lint の指摘を確認  $ golangci-lint run
2026-10-12 06:35:12    0    8.3s  .  xxテスト実施  $ make test
```

`-H` と `--history` は同じ。どのサブディレクトリから打っても、同じ履歴が出る。

## 使い方を見る

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

## 理由の後ろのオプションは、コマンドに渡る

```
$ bcuz "サイズを見る" ls -h
```

`-h` は bcuz ではなく `ls` に渡る（bcuz がオプションとして読むのは第1引数だけ）。
