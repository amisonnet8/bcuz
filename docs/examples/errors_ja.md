*[English](errors.md) | **日本語***

# 失敗と誤り

## `.bcuz/` が無い

コマンドは実行される。終わったあとに警告が出る。終了コードはコマンドのもの。

```
$ cd /tmp/scratch
$ bcuz "試しに動かす" echo hello
hello
bcuz: not recorded: no .bcuz/ found (run 'bcuz --init')
$ echo $?
0
```

履歴も見られない。

```
$ bcuz -H
bcuz: no .bcuz/ found (run 'bcuz --init')
$ echo $?
125
```

## コマンドが見つからない（記録する）

```
$ cd /home/me/work
$ bcuz "ビルド" mkae build
bcuz: command not found: mkae
$ echo $?
127
```

```json
{"v":1,"ts":"2026-10-11T21:45:00.000Z","why":"ビルド","cmd":["mkae","build"],"dir":".","exit":127,"ms":0}
```

## 起動できない（記録する）

```
$ bcuz "スクリプトを試す" ./build.sh
bcuz: cannot run ./build.sh: permission denied
$ echo $?
126
```

## 引数の誤り（記録しない）

```
$ bcuz
bcuz: missing reason and command (see 'bcuz -h')
$ echo $?
125
$ bcuz "テスト"
bcuz: missing command after reason
$ bcuz "  " make test
bcuz: reason is empty
$ bcuz -x make test
bcuz: unknown option: -x (use '--' before a reason starting with '-')
```

理由を `-` で始めたいときは `--` を付ける。

```
$ bcuz -- "-v を付けて再確認" go test -v ./...
```

```json
{"v":1,"ts":"2026-10-11T22:00:00.000Z","why":"-v を付けて再確認","cmd":["go","test","-v","./..."],"dir":".","exit":0,"ms":9100}
```

## 親にも `.bcuz/` があるところで `--init`

```
$ cd /home/me/work/frontend
$ bcuz --init
bcuz: created .bcuz/ (shadows /home/me/work/.bcuz/)
```

以後、`frontend/` より下で打ったものは `frontend/.bcuz/` に記録される。
