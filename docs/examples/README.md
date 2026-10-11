# 例

実際の入力と出力。`$` で始まる行が入力、それ以外が出力。

- 時刻（`ts`・日時の欄）と時間（`ms`・時間の欄）は、毎回変わる。それ以外は一字一句、実装の出力と同じにする
- 例のタイムゾーンは Asia/Tokyo（+09:00）
- 例の作業場は `/home/me/work`（`.bcuz/` はここにある）

| ファイル | 内容 |
|---|---|
| [basic.md](basic.md) | `--init`、実行、失敗、履歴の表示 |
| [history-filter.md](history-filter.md) | 日付で絞り込む |
| [errors.md](errors.md) | `.bcuz/` が無い、引数の誤り、コマンドが見つからない |
| [shell.md](shell.md) | パイプ・`&&` を記録する、引数の引用 |
| [history.jsonl](history.jsonl) | 数日使ったあとの `history.jsonl`。[history-filter.md](history-filter.md) はこの記録を使う |
