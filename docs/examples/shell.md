# シェルの構文と引用

## `&&`・パイプは、bcuz の外で動く

bcuz はシェルを通さない。次の例で bcuz が実行・記録するのは `make test` だけ。`make lint` は、呼んだシェルが bcuz の後に実行する（記録されない）。

```
$ bcuz "テストと lint" make test && make lint
```

まとめて記録するなら、`sh -c` で包む。

```
$ bcuz "パイプごと記録" sh -c 'grep -c TODO *.go | sort -t: -k2 -n'
main.go:0
cli.go:2
```

```json
{"v":1,"ts":"2026-10-12T06:50:00.000+09:00","why":"パイプごと記録","cmd":["sh","-c","grep -c TODO *.go | sort -t: -k2 -n"],"dir":".","exit":0,"ms":120}
```

## 引数は、受け取ったまま残る

```
$ bcuz "コミット" git commit -m "fix: handle empty reason"
```

```json
{"v":1,"ts":"2026-10-12T06:52:10.000+09:00","why":"コミット","cmd":["git","commit","-m","fix: handle empty reason"],"dir":".","exit":0,"ms":1200}
```

履歴では、シェルで打てる形に直して出す（英数字と `-_./=:,+@%` 以外を含む引数は `'…'` で囲む）。

```
2026-10-12 06:52:10    0    1.2s  .  コミット  $ git commit -m 'fix: handle empty reason'
```

`*` はシェルが先に展開する。展開させずに渡した引数は、`'…'` で囲んで表示される。

```
$ bcuz "タグを確認" git tag --list 'v0.*'
```

```
2026-10-10 17:55:01    0   950ms  .  タグを確認  $ git tag --list 'v0.*'
```

## 理由の改行

理由の改行・タブは、記録にはそのまま残り、履歴の表示では空白1つになる。

```
$ bcuz "移行の1回目
件数だけ確認" ./migrate --dry-run
```

```json
{"v":1,"ts":"2026-10-12T07:10:00.000+09:00","why":"移行の1回目\n件数だけ確認","cmd":["./migrate","--dry-run"],"dir":".","exit":0,"ms":3050}
```

```
2026-10-12 07:10:00    0    3.0s  .  移行の1回目 件数だけ確認  $ ./migrate --dry-run
```
