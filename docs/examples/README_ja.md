*[English](README.md) | **日本語***

# 例

実際の入力と出力。`$` で始まる行が入力、それ以外が出力。

- 時刻（`ts`・日時の欄）と時間（`ms`・時間の欄）は、毎回変わる。それ以外は一字一句、実装の出力と同じにする（`make test` など、ほかのコマンドの出力は、例として示すだけ）
- 記録は UTC（`ts` の末尾が `Z`）。例で見る側のタイムゾーンは Asia/Tokyo（+09:00）。`TZ=UTC` とあるブロックだけが UTC
- 例の作業場は `/home/me/work`（`.bcuz/` はここにある）

| ファイル | 内容 |
|---|---|
| [basic_ja.md](basic_ja.md) | `--init`、実行、失敗、履歴の表示 |
| [history-filter_ja.md](history-filter_ja.md) | 日付で絞り込む |
| [errors_ja.md](errors_ja.md) | `.bcuz/` が無い、引数の誤り、コマンドが見つからない |
| [shell_ja.md](shell_ja.md) | パイプ・`&&` を記録する、引数の引用 |
| [history_ja.jsonl](history_ja.jsonl) | 数日使ったあとの `history.jsonl`。[history-filter_ja.md](history-filter_ja.md) はこの記録を使う |
