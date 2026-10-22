# 命名規則

## 製品名

| 対象 | 表記 | 備考 |
|---|---|---|
| 製品・コマンド | **bcuz** | 常に小文字。文頭でも `bcuz`。because から |
| 記録の置き場 | `.bcuz/`・`.bcuz/history.jsonl` | 名前を変えない（`docs/reference/cli.md`） |
| オプション | `-H`・`--history`、`--init`、`--completion`、`-h`・`--help`、`--version`、`--date`・`--since`・`--until` | `docs/reference/cli.md` にあるものだけ |

## 用語（日本語と英語の対応）

文書・コメント・メッセージで、同じものを別の言葉で呼ばない。

| 日本語 | 英語（コード・メッセージ） | 意味 |
|---|---|---|
| 理由 | reason（JSON のフィールドは `why`） | コマンドに添える理由。メッセージでは `reason` |
| コマンド | command（`cmd`） | 理由の後ろの、実行するものと引数 |
| 記録 | record | `history.jsonl` の1行＝1回の実行 |
| 履歴 | history | 記録の集まり。`-H` で見るもの |
| 記録先 | `.bcuz/` | 親へたどって最初に見つかるもの |
| 作業場 | workspace（Go の名前では root） | `.bcuz/` のあるディレクトリ。`dir` はここからの相対パス |
| 終了コード | exit code（`exit`） | |
| 実行時間 | duration（`ms`） | |
| 絞り込み | filter | `--date`・`--since`・`--until` |

## JSON のフィールド名

- `v`・`ts`・`why`・`cmd`・`dir`・`exit`・`ms`。**確定。** 変えない（`docs/reference/history.md`）
- 足すときは、短い小文字の名前に揃え、了承をもらってから先に `docs/reference/history.md` に書く（`.claude/rules/working-with-human.md` 4章）

## メッセージ

- 英語。すべて `bcuz: ` で始まる1行。警告は `bcuz: warning: `
- 文言は `docs/reference/cli.md` の表のとおり。Go では `internal/cli` の定数（またはフォーマット文字列の定数）にまとめ、テストも同じ定数を使う

## Go のコード

- パッケージ名は短い小文字1語：`cli`・`record`・`run`・`history`・`docs`
- 用語表の英語をそのまま型名・関数名に使う（例：`Record`、`FindRoot`、`Append`、`Filter`、`FormatLine`）
- `cmd/bcuz/main.go` は `os.Exit(cli.Main(…))` だけにする
- OS ごとのファイルは `_unix.go`・`_windows.go` で終え、先頭に `//go:build unix`・`//go:build windows`
