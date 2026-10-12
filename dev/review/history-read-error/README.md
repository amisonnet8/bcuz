# `-H` で履歴を読めないとき

## 何を足すか、なぜか

- `.bcuz/` はあるが `history.jsonl` を読めない（権限が無い、ディレクトリになっている、など）ときの出力が、`docs/reference/cli.md` に無い
- 何も出さずに 0 で終わると、履歴が空なのか読めないのか区別できない。**メッセージを出して 125 で終わる**ことを提案する
- `history.jsonl` が**無い**ときは、これまでどおり「1件も無い」として、何も出さず 0 で終わる（`.bcuz/` を作った直後の `--init` は空のファイルを作るので、ふつうは起きない）

## 変更前後の例

変更前（決まりが無い）：

```
$ chmod 000 .bcuz/history.jsonl
$ bcuz -H
$ echo $?
0
```

変更後：

```
$ chmod 000 .bcuz/history.jsonl
$ bcuz -H
bcuz: cannot read .bcuz/history.jsonl: permission denied
$ echo $?
125
```

## 新しく決めること

- 文言：`bcuz: cannot read .bcuz/history.jsonl: <理由>`（`--init` の `cannot create .bcuz/: <理由>` と同じ形）。**推奨**
- `<理由>` は、記録できないときの `<理由>` と同じ書き方（`permission denied` など。パスや `open …:` は含めない）
- 終了コードは 125（bcuz 自身のエラー）

## 了承してほしいこと

1. 上のメッセージと 125 で、`docs/reference/cli.md` のエラーの表に1行足してよいか
