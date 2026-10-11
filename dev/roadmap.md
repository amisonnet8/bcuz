# 段階（ロードマップ）

`docs/` は完成している（正本）。ここでは、それを実装する順番を決める。**上から順に進める。** 各段階の完了条件を、実際に動かして確かめてから次へ進む。

- 進捗は mtqg の todo（段階ごとに1件。中のタスクは着手前に足す）。今どこかは `mtqg context`
- srwr のテープは、todo ごとに `session` で区切る（`.claude/rules/srwr-usage.md`）
- 「人間の確認」の欄が空の段階は、人間に確認を頼まない（自動のテストで足りる）

| 段階 | 作るもの | 完了条件 | 人間の確認 |
|---|---|---|---|
| **R0 土台** | `cmd/bcuz/main.go`、`internal/cli`（引数の振り分けだけ）、`-h`・`--help`・`--version`。`internal/docs`（文書のリンク切れのテスト）。`internal/doc.go` を消す | `qsoku check` が通る。`bcuz -h` の出力が `docs/reference/cli.md` の使い方のブロックと一字一句同じ（`internal/docs` のテストで照合） | |
| **R1 記録** | `internal/record`：記録の型、1行の書き出し（項目の順を固定）・読み込み、`.bcuz/` の探索、追記（`O_APPEND`・1回の write）、`--init` | `docs/examples/history.jsonl` の各行を読んで書き出すと、元の行と一字一句同じ。入れ子の作業場で探索が最寄りを選ぶ。`--init` の3つの出力（`cli.md`）。同時に追記しても行が混ざらない（`qsoku race`） | |
| **R2 実行** | `internal/run`：コマンドの起動（シェルなし）、入出力の素通し、終了コード（125・126・127・128＋番号）、シグナル（Unix と Windows をビルドタグで分ける）。`internal/cli` の実行の入口と、記録できないときの警告 | 通しテスト（`.claude/rules/testing.md`「bcuz での動作確認」）：終了コードがそのまま返る、126・127 は記録して 125 は記録しない、`.bcuz/` が無くても実行して警告、SIGINT で先に終わらず記録が残る、SIGTERM・SIGHUP の転送、TTY の素通し。`qsoku cross` | 下の「R2 の確認」 |
| **R3 履歴** | `internal/history`：読み込み（途中の最後の行を飛ばす、読めない行は警告）、日付の絞り込み（見る側のタイムゾーン）、表示（時間・引用・改行の置き換え）。`-H`・`--history` と、その誤り | `docs/examples/history-filter.md` の全ブロックが、`internal/docs` のテストで一致（Asia/Tokyo・UTC）。`cli.md` のエラーメッセージの表の、履歴の行がすべてテストにある | |
| **R4 例の照合** | `internal/docs` に、`docs/examples/basic.md`・`errors.md`・`shell.md` の照合を足す（時刻・時間を除いて、出力とメッセージと記録を比べる） | `docs/examples/` のすべての入出力がテストで押さえられている。`README.md` の使い方が `docs/` と合っている | |
| **R5 配布** | `--version` をビルド情報から。`go install github.com/amisonnet8/bcuz/cmd/bcuz@latest` で入ること（`.claude/rules/distribution.md`） | 一時ディレクトリで `go install` して動く。3 OS の CI が通る | タグを打つ・Release を作るのは人間 |

## R2 の確認（人間）

自動のテストで押さえられないのは「本物の端末で、手で Ctrl-C を押したときの感触」だけ。R2 の終わりに、次を作ってから頼む（`.claude/rules/working-with-human.md` 3章）。

- `qsoku try`：`bin/bcuz` をビルドし、一時ディレクトリに `.bcuz/` を作り、`bcuz` が `PATH` に入ったシェルを開く。開いたときに「打つコマンド3つと、見えるはずのもの」を表示し、シェルを閉じると `bcuz -H` の結果を出す
- 打つもの（案）：`bcuz "色" ls --color=auto`（色が付く）、`bcuz "止める" sleep 30` で Ctrl-C（すぐ戻り、終了コード 130）、`bcuz "編集" vi x.txt`（普通に使える）

## 持ち越しの決まり（docs から）

- 出力・メッセージ・終了コードは `docs/reference/cli.md` で確定。**変えるときは、先に人間の了承をもらい、`docs/` を直してから実装する**（`.claude/rules/working-with-human.md` 4章）
- 追加の候補（`-n`・`--failed`・`--json` など）は `docs/design/limitations.md`。この段階表には入れない。やるときは、新しい段階として足す
- bcuz ができたら、この開発自体で bcuz を使うか（CLAUDE.md で Bash に付けさせる）は、R5 のあとに人間が決める
