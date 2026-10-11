# 日付で絞り込む

[history.jsonl](history.jsonl) の記録（10月10日〜12日）があるとき。

## 全部

```
$ bcuz -H
2026-10-12 06:52:10    0    1.2s  .  コミット  $ git commit -m 'fix: handle empty reason'
2026-10-12 06:50:00    0   120ms  .  パイプごと記録  $ sh -c 'grep -c TODO *.go | sort -t: -k2 -n'
2026-10-12 06:45:00  127     0ms  .  ビルド  $ mkae build
2026-10-12 06:41:30  130   1m05s  backend  ログを眺める  $ tail -f app.log
2026-10-12 06:40:03    1    2.4s  backend  lint の指摘を確認  $ golangci-lint run
2026-10-12 06:35:12    0    8.3s  .  xxテスト実施  $ make test
2026-10-11 09:15:30    0    4.0s  frontend  依存を入れ直す  $ npm ci
2026-10-10 18:02:44    0   35.0s  .  リリース前の全テスト  $ make test
2026-10-10 17:55:01    0   950ms  .  タグを確認  $ git tag --list 'v0.*'
```

## その日だけ

```
$ bcuz -H --date 2026-10-10
2026-10-10 18:02:44    0   35.0s  .  リリース前の全テスト  $ make test
2026-10-10 17:55:01    0   950ms  .  タグを確認  $ git tag --list 'v0.*'
```

## その日から

```
$ bcuz -H --since 2026-10-11
2026-10-12 06:52:10    0    1.2s  .  コミット  $ git commit -m 'fix: handle empty reason'
2026-10-12 06:50:00    0   120ms  .  パイプごと記録  $ sh -c 'grep -c TODO *.go | sort -t: -k2 -n'
2026-10-12 06:45:00  127     0ms  .  ビルド  $ mkae build
2026-10-12 06:41:30  130   1m05s  backend  ログを眺める  $ tail -f app.log
2026-10-12 06:40:03    1    2.4s  backend  lint の指摘を確認  $ golangci-lint run
2026-10-12 06:35:12    0    8.3s  .  xxテスト実施  $ make test
2026-10-11 09:15:30    0    4.0s  frontend  依存を入れ直す  $ npm ci
```

## 期間

```
$ bcuz -H --since 2026-10-10 --until 2026-10-11
2026-10-11 09:15:30    0    4.0s  frontend  依存を入れ直す  $ npm ci
2026-10-10 18:02:44    0   35.0s  .  リリース前の全テスト  $ make test
2026-10-10 17:55:01    0   950ms  .  タグを確認  $ git tag --list 'v0.*'
```

## 該当なし

何も出さず、成功（0）。

```
$ bcuz -H --date 2026-09-01
$ echo $?
0
```

## 誤り

```
$ bcuz -H --date 2026-02-30
bcuz: invalid date: 2026-02-30
$ echo $?
125
$ bcuz -H --date 2026-10-12 --since 2026-10-01
bcuz: --date cannot be used with --since or --until
$ echo $?
125
```

## タイムゾーン

日時の欄と日付の比較は、**見る側**のタイムゾーンで行う。`ts` が `2026-10-12T06:35:12.345+09:00` の記録は、UTC の環境（`TZ=UTC`）では次のように見え、10月11日の記録として扱われる。

```
$ TZ=UTC bcuz -H --date 2026-10-11
2026-10-11 21:52:10    0    1.2s  .  コミット  $ git commit -m 'fix: handle empty reason'
2026-10-11 21:50:00    0   120ms  .  パイプごと記録  $ sh -c 'grep -c TODO *.go | sort -t: -k2 -n'
2026-10-11 21:45:00  127     0ms  .  ビルド  $ mkae build
2026-10-11 21:41:30  130   1m05s  backend  ログを眺める  $ tail -f app.log
2026-10-11 21:40:03    1    2.4s  backend  lint の指摘を確認  $ golangci-lint run
2026-10-11 21:35:12    0    8.3s  .  xxテスト実施  $ make test
2026-10-11 00:15:30    0    4.0s  frontend  依存を入れ直す  $ npm ci
```
