*[English](cli.md) | **日本語***

# コマンドの使い方

## 書式

```
bcuz [--] <理由> <コマンド> [引数...]    コマンドを実行して記録する
bcuz -H [絞り込み]                     履歴を表示する（--history と同じ）
bcuz --init                          カレントディレクトリに .bcuz/ を作る
bcuz -h                              使い方を表示する（--help と同じ）
bcuz --version                       バージョンを表示する
```

### 引数の読み方

- **オプションとして読むのは、第1引数だけ。** 第1引数が `-H`・`--history`・`--init`・`-h`・`--help`・`--version` のどれかなら、その機能になる
- 第1引数が `--` なら、それを読み捨て、第2引数を理由として実行する
- 第1引数がそれ以外で `-` から始まるなら、エラー（125）。理由を `-` で始めたいときは `--` を前に付ける
- それ以外は、第1引数を理由、残りをコマンドと引数として実行する。**コマンドの引数は一切解釈しない**（`bcuz "確認" ls -h` の `-h` は `ls` に渡る）

## 実行

```
bcuz "xxテスト実施" make test
```

1. 理由が空（空白だけを含む）なら、エラー（125）。何も実行しない
2. コマンドが無ければ、エラー（125）
3. `.bcuz/` を探す（下の「`.bcuz/` の探し方」）。見つからなくても止めない
4. コマンドを `PATH` から探して起動する。**シェルを通さない**（`&&`・`|`・`*` などは、bcuz を呼んだシェルが先に解釈する）
5. コマンドが終わるのを待つ
6. `.bcuz/` が見つかっていれば、1行追記する（[history_ja.md](history_ja.md)）
7. コマンドの終了コードで終わる

### 素通しするもの

- 標準入力・標準出力・標準エラー出力は、bcuz が受けたものを**そのまま**コマンドに渡す（パイプを挟まない）。TTY もそのまま。bcuz は出力を読まない・残さない
- 環境変数・カレントディレクトリは、そのまま引き継ぐ
- **Ctrl-C**（SIGINT）・SIGQUIT：端末からはコマンドにも届く。bcuz 自身はこれで終わらず、コマンドの終了を待って記録する（転送はしない。二重に届くため）
- **SIGTERM・SIGHUP**：bcuz に届いたら、コマンドに転送し、終了を待って記録する
- Windows では、Ctrl-C（Ctrl-Break）で bcuz 自身は終わらず、コマンドの終了を待つ

### 終了コード

| 終了コード | 意味 | 記録 |
|---|---|---|
| コマンドのもの | コマンドが終わった | する |
| 128＋シグナル番号 | コマンドがシグナルで終わった（例：Ctrl-C なら 130） | する |
| 125 | bcuz 自身のエラー（引数の誤り、理由が空、コマンドが無い） | しない |
| 126 | コマンドは見つかったが起動できない（実行権限が無い、など） | する |
| 127 | コマンドが見つからない | する |

- 126・127 も「理由を付けて実行しようとした」ことなので記録する（`ms` は 0）
- **記録に失敗しても、終了コードはコマンドのものを返す**（警告だけ出す）

### `.bcuz/` が見つからないとき

コマンドは実行する。終わったあと、標準エラー出力に次を出す。終了コードはコマンドのもの。

```
bcuz: not recorded: no .bcuz/ found (run 'bcuz --init')
```

## `.bcuz/` の探し方

- カレントディレクトリから親へ順にたどり、**最初に見つかった `.bcuz/` ディレクトリ**を使う。ファイルシステムのルートまで無ければ「見つからない」
- 記録は、その中の `history.jsonl` に追記する。`.bcuz/` はあるが `history.jsonl` が無いときは、作る
- git などのリポジトリの境界では止まらない
- 履歴（`-H`）も同じ探し方をする

## `--init`

```
bcuz --init
```

- カレントディレクトリに `.bcuz/` と空の `.bcuz/history.jsonl` を作る
- すでにカレントディレクトリに `.bcuz/` があれば、何も変えずに知らせて、成功（0）で終わる
- 親に `.bcuz/` があっても作る（そのことを知らせる）。以後、このディレクトリより下は、新しい `.bcuz/` に記録される
- `.gitignore` などは作らない。`.bcuz/` を git に入れるかは利用者が決める

出力（標準出力）：

| 場合 | 出力 |
|---|---|
| 作った | `bcuz: created .bcuz/` |
| 作った（親にもある） | `bcuz: created .bcuz/ (shadows /home/me/work/.bcuz/)` |
| すでにある | `bcuz: .bcuz/ already exists` |

## `-H`・`--history`

```
bcuz -H                                      全部
bcuz -H --date 2026-10-12                    その日だけ
bcuz -H --since 2026-10-01                   その日から
bcuz -H --until 2026-10-10                   その日まで
bcuz -H --since 2026-10-01 --until 2026-10-10
```

### 絞り込み

| オプション | 意味 |
|---|---|
| `--date YYYY-MM-DD` | その日に**始めた**実行だけ |
| `--since YYYY-MM-DD` | その日以降（その日を含む） |
| `--until YYYY-MM-DD` | その日まで（その日を含む） |

- 日付は、**表示する側のタイムゾーン**の日付で比べる（`ts` を手元のタイムゾーンに直してから、その日付を見る）
- `--date` と `--since`・`--until` は一緒に使えない（125）
- 同じオプションを2回書くと、エラー（125）
- 日付の形が違う・存在しない日付（`2026-02-30`）は、エラー（125）
- `--since` が `--until` より後なら、何も出さずに成功（0）

### 表示

- 標準出力に、**新しい順**に1行1件
- 1件も無ければ、何も出さずに成功（0）

```
2026-10-12 06:35:12    0    8.3s  backend  xxテスト実施  $ make test
```

| 欄 | 内容 |
|---|---|
| 日時 | `ts` を手元のタイムゾーンに直した `YYYY-MM-DD HH:MM:SS` |
| 終了コード | 右寄せ3桁 |
| 時間 | 右寄せ6桁。下の「時間の書き方」 |
| 場所 | `dir`（直下は `.`） |
| 理由 | `why`。改行とタブは空白1つに置き換える |
| コマンド | `$ ` の後に、`cmd` をシェルで打てる形に直したもの（下） |

欄の区切りは空白2つ。場所と理由は幅をそろえない。

#### 時間の書き方

| `ms` | 表示 |
|---|---|
| 1000 未満 | `123ms` |
| 60000 未満 | `8.3s`・`35.0s`（常に小数1桁、切り捨て） |
| 3600000 未満 | `2m05s` |
| それ以上 | `1h02m` |

#### コマンドの書き方

- 引数を空白1つでつなぐ
- 英数字と `-_./=:,+@%` だけでできた引数は、そのまま。それ以外（空白・引用符・`$` などを含む、空文字列）は `'…'` で囲み、中の `'` は `'\''` にする
- 例：`["git","commit","-m","fix: a b"]` → `git commit -m 'fix: a b'`

### 読めない行

- 改行で終わっていない最後の行は、書き込みの途中として飛ばす（警告しない）
- JSON として読めない行・必須の項目が無い行は飛ばし、標準エラー出力に警告を出す。終了コードは 0 のまま

```
bcuz: warning: .bcuz/history.jsonl:12: skipped unreadable line
```

### `.bcuz/` が見つからないとき

```
bcuz: no .bcuz/ found (run 'bcuz --init')
```

標準エラー出力に出して、125 で終わる。

### 履歴を読めないとき

`history.jsonl` が無ければ、1件も無いものとして扱う（何も出さずに 0）。あるのに読めなければ、標準エラー出力に出して、125 で終わる。

```
bcuz: cannot read .bcuz/history.jsonl: permission denied
```

## `-h`・`--help`

使い方を標準出力に出して、0 で終わる。

```
bcuz - run a command with a reason, and record it

usage:
  bcuz [--] <reason> <command> [args...]
  bcuz -H, --history [--date D | --since D] [--until D]
  bcuz --init
  bcuz -h, --help
  bcuz --version

dates are YYYY-MM-DD. records go to the nearest .bcuz/history.jsonl.
```

## `--version`

```
bcuz v0.1.0
```

バージョンは、ビルド情報（`go install …@v0.1.0` で入るタグ）から取る。取れないとき（手元のビルド）は `bcuz (devel)`。

## エラーのメッセージ

すべて標準エラー出力に、`bcuz: ` で始まる1行で出す。

| 場合 | メッセージ | 終了コード |
|---|---|---|
| 引数が無い | `bcuz: missing reason and command (see 'bcuz -h')` | 125 |
| 理由が空 | `bcuz: reason is empty` | 125 |
| コマンドが無い | `bcuz: missing command after reason` | 125 |
| 知らないオプション | `bcuz: unknown option: -x (use '--' before a reason starting with '-')` | 125 |
| 日付の誤り | `bcuz: invalid date: 2026-02-30` | 125 |
| `--date` と期間の併用 | `bcuz: --date cannot be used with --since or --until` | 125 |
| オプションの重複 | `bcuz: duplicate option: --since` | 125 |
| `-H` の後の知らない引数 | `bcuz: unknown argument: --foo` | 125 |
| 日付が無い | `bcuz: missing date after --date` | 125 |
| `.bcuz/` が無い（実行） | `bcuz: not recorded: no .bcuz/ found (run 'bcuz --init')` | コマンドのもの |
| `.bcuz/` が無い（`-H`） | `bcuz: no .bcuz/ found (run 'bcuz --init')` | 125 |
| 履歴を読めない（`-H`） | `bcuz: cannot read .bcuz/history.jsonl: permission denied` | 125 |
| 読めない行（`-H`） | `bcuz: warning: .bcuz/history.jsonl:12: skipped unreadable line` | 0 |
| コマンドが見つからない | `bcuz: command not found: mkae` | 127 |
| 起動できない | `bcuz: cannot run ./build.sh: permission denied` | 126 |
| 記録できない | `bcuz: warning: not recorded: <理由>` | コマンドのもの |
| `--init` で作れない | `bcuz: cannot create .bcuz/: <理由>` | 125 |
