# `-H` で標準入力から履歴を読む（`bcuz -H -`）

## 何を足すか、なぜか

- 記録（`.bcuz/history.jsonl`）は増え続ける。アーカイブ機能は「記録は追記だけ（書き換えない・消さない）」に触れるので作らない（q e2b75a2980）
- 代わりに、**`-H` の後に `-` を書くと、履歴を標準入力から読む**ようにする。古い記録は、人が `mv`・`gzip` で逃がし、`gzip -dc … | bcuz -H -` で見る。bcuz は記録を動かさない
- `-H` は `ts` で並べ替えるので、複数のファイルをつないで流しても、新しい順に出る
- `jq` で先に絞ってから流すこともできる（`jq -c 'select(.exit != 0)' .bcuz/history.jsonl | bcuz -H -`）

## 変更前後の例

変更前：

```
$ gzip -dc .bcuz/history-2026.jsonl.gz | bcuz -H -
bcuz: unknown argument: -
$ echo $?
125
```

変更後（絞り込みと一緒に使える。`-` の位置は問わない）：

```
$ gzip -dc .bcuz/history-2026.jsonl.gz | bcuz -H - --date 2026-10-10
2026-10-10 18:02:44    0   35.0s  .  full tests before release  $ make test
2026-10-10 17:55:01    0   950ms  .  check the tags  $ git tag --list 'v0.*'
$ echo $?
0
```

`.bcuz/` の無い場所でも読める（`-` のときは `.bcuz/` を探さない）：

```
$ cd /tmp
$ bcuz -H - < ~/work/old.jsonl
2026-10-12 06:52:10    0    1.2s  .  commit  $ git commit -m 'fix: handle empty reason'
...
```

読めない行・読めないとき（場所は `<stdin>`）：

```
$ bcuz -H - < broken.jsonl
bcuz: warning: <stdin>:12: skipped unreadable line
...
```

標準入力そのものが読めないとき（入出力のエラーなど。まれ）は、今のファイルのときと同じく 125：

```
bcuz: cannot read <stdin>: input/output error
```

`-h` の使い方の行：

```
変更前：  bcuz -H, --history [--date D | --since D] [--until D]
変更後：  bcuz -H, --history [-] [--date D | --since D] [--until D]
```

`-` の無い `bcuz -H` は今のまま（`.bcuz/history.jsonl` を読む。標準入力は読まない）。

## 新しく決めること（推奨どおりで、会話で了承済み）

1. 標準入力を読むのは `-` を書いたときだけ。端末でない標準入力を自動で読む方式にはしない（Claude Code・CI の `bcuz -H` が、入力を待って固まる・空を出すため）
2. `-` のときは `.bcuz/` を探さない（無くても 125 にしない）
3. 場所の名前は `<stdin>`：`bcuz: warning: <stdin>:12: skipped unreadable line`、`bcuz: cannot read <stdin>: <理由>`（`<理由>` は今のファイルのときと同じ書き方）
4. 改行で終わらない最後の行は、ファイルと同じく黙って飛ばす
5. 使い方の行は上のとおり
6. 文書の例は `gzip -dc`（macOS の `zcat` は `.gz` を読まないことがある）
7. `-` は `-H` の後のどこに書いてもよい。2回書くと `bcuz: duplicate option: -`（125。今の重複の決まりのまま）
8. 端末から `bcuz -H -` と打つと、入力の終わり（Ctrl-D）まで待つ（`cat -` と同じ）。文書に書く

## 文書に足すもの

- `docs/reference/cli.md`：使い方、`-H` の節「Reading from standard input」、エラーの表に `<stdin>` の2行
- `docs/reference/history.md`：節「Archiving」（bcuz は記録を動かさない。`mv`・`gzip` してよい。`history.jsonl` が無ければ次の実行で作られる。読むときは `gzip -dc … | bcuz -H -`）
- `docs/examples/history-filter.md`：標準入力から読む例
- `docs/design/decisions.md`：アーカイブを作らず標準入力にした理由、自動で読まない理由

## 了承してほしいこと

1. 上の使い方・出力・メッセージで、`docs/` を直して実装してよいか
