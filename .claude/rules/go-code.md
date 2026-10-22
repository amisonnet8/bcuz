# Go のコード（bcuz 固有の決まり）

一般的な lint・整形・テストは `.claude/rules/testing.md` と `.golangci.yaml` に任せる。ここには bcuz 固有の決まりだけを書く。

## 依存

- **外部依存ゼロ。** `encoding/json`・`os/exec`・`os/signal`・`time`・`runtime/debug` で足りる

## 引数

- **オプションとして読むのは第1引数だけ**（`docs/reference/cli.md`「引数の読み方」）。`flag` パッケージは使わない（理由の後ろのコマンドの引数を横取りするため）。`-H` の後の絞り込みも、自前で読む
- bcuz 自身のエラーは 125。メッセージは `docs/reference/cli.md` の表と一字一句同じ

## 補完（`--completion`）

- スクリプトは `internal/cli/completions/` に置き、`embed` で埋め込む。**決まった文で、補完の最中に bcuz を呼ばない**（`docs/design/decisions.md`）。候補を変えるときは、4つのスクリプトと `completion_test.go` の表、`docs/reference/cli.md` の表を同じ変更で直す
- bash のスクリプトは **bash 3.2（macOS の標準）でも動く書き方**にする。`mapfile`・`readarray`・連想配列を使わない（`compopt` のような 4 以降のものは、無くても動くように `2>/dev/null` で包む）。`*.bash` なので `qsoku shellcheck` の対象

## 実行

- コマンドは `exec.Command` で、**シェルを通さずに**起動する。`Stdin`・`Stdout`・`Stderr` には `os.Stdin`・`os.Stdout`・`os.Stderr` をそのまま渡す（パイプを挟まない。TTY を保つ）
- 見つからない（`exec.ErrNotFound`）は 127、起動できない（権限など）は 126。どちらも記録する
- 終了コード：`ExitError` の `ExitCode()`。Unix でシグナルで終わったときは `WaitStatus` から 128＋番号
- **シグナル**：起動の前に `signal.Notify` で受け、SIGINT・SIGQUIT は**受けて捨てる**（bcuz が先に終わらないため。転送しない）。SIGTERM・SIGHUP は子へ `Process.Signal` で転送する。子が終わったら `signal.Stop`
- Windows は `os.Interrupt` を受けて捨てるだけ。シグナルの処理は `_unix.go`・`_windows.go` に分け、`qsoku cross` を通す
- 時間は、起動の直前から `Wait` が返るまでを `time.Since` で測る（単調時計）。`ts` は起動の直前の `time.Now()`

## 記録

- 1行は、**項目の順を固定した構造体**から作る（`v`・`ts`・`why`・`cmd`・`dir`・`exit`・`ms`）。`map` は使わない（順が変わる）
- `json.Marshal` は `<`・`>`・`&` を `<` などにする。**`json.Encoder` の `SetEscapeHTML(false)`** を使い、理由やコマンドの文字をそのまま残す
- `ts` は `time.Now()` を **`.UTC()` にして** `2006-01-02T15:04:05.000Z07:00` で書く（ミリ秒、末尾は `Z`）。表示（`-H`）と日付の絞り込みだけが、見る側のタイムゾーン（`time.Local`）に直す。読む側は、古い記録のオフセット付きも受け付ける
- 追記は `os.OpenFile(…, O_APPEND|O_CREATE|O_WRONLY, 0o644)` に、改行まで含めた1行を**1回の `Write`** で書く
- `dir` は `filepath.Rel(root, cwd)` を `filepath.ToSlash` したもの。`os.Getwd()` をそのまま使う
- `.bcuz/` の探索は、`filepath.Dir` を繰り返し、`Dir(x) == x` になったら止める（Windows のドライブのルートでも止まる）。見つけるのは**ディレクトリ**の `.bcuz`（同名のファイルは無視する）
- **記録の失敗で終了コードを変えない。** 警告だけ

## 履歴

- 読み込みは `bufio.Reader` で行ごとに。**改行で終わっていない最後の行は飛ばす**（警告しない）。`bufio.Scanner` は既定で1行 64KB の上限があるので使わない
- 日付の比較は、`ts` を `time.Local` に直した日付で行う。テストでは `time.Local` を差し替える
- 時間の表示（`123ms`・`8.3s`・`2m05s`・`1h02m`）と引数の引用は、`docs/reference/cli.md` の決まりそのまま。境界（999・1000・59999・60000・3599999・3600000）をテストで押さえる

## テスト

- 表駆動。`t.TempDir()` の中で行い、リポジトリ内のファイルを書き換えない
- 通しテストの子のコマンドには、テストのバイナリ自身を使う（`.claude/rules/testing.md`「bcuz での動作確認」）
- わざとロジックを壊してテストが落ちることを確かめる（`.claude/rules/testing.md`「壊して確かめる」）
