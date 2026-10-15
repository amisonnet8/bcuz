# bcuzの使い方

**Bashで実行するコマンドは、必ず`bcuz`を通す。** `bcuz "理由" コマンド [引数...]`と打つと、コマンドがそのまま実行され、「いつ・なぜ・何を・どうなったか」が`.bcuz/history.jsonl`に1行追記される。コマンドの出力・終了コードは、そのまま素通しされる（`bcuz`が出すのは、自身のエラーだけ）。

```
bcuz "リファクタ後のlint" qsoku lint
bcuz "ts の UTC 化の確認" qsoku unit
bcuz "失敗の原因を探す" git log --oneline -5
```

## 強制されている

`.claude/settings.json`の`PreToolUse`フック（`.claude/hooks/require-bcuz.sh`）が、Bashツールのコマンドを実行前に調べる。次は**実行されずに断られる**（終了コード2、理由と書き方が返る）。

| 断られるもの | 直し方 |
|---|---|
| `bcuz`で始まらない（`qsoku check`、`cd x && …`） | `bcuz "理由" qsoku check` |
| 後ろに別のコマンドを続ける（`&&`・`\|\|`・`;`・`&`・改行） | 1コマンドずつ`bcuz`で実行する。まとめて1つの作業なら`bcuz "理由" sh -c 'a && b'` |
| コマンド置換（`$(…)`・`` `…` ``） | 先に別の`bcuz`で値を得るか、`sh -c`の中に書く |
| `bcuz`の知らないオプション（`bcuz -x …`） | 理由が`-`で始まるなら`bcuz -- "-x の確認" …` |

- パイプ（`|`）とリダイレクト（`> file`・`2>&1`）は通す（`bcuz "確認" qsoku check 2>&1 \| tail -20`）。パイプの後ろは記録されない（出力の整形用）
- 環境変数を足すときは`env`で：`bcuz "リリースのビルド" env VERSION=v0.1.0 qsoku dist`
- `bcuz`が`PATH`に無いとき（最初の`go install`など）は、フックは通す（作る前に詰まらないため）
- これは「うっかり」を防ぐ仕組みで、安全の境界ではない。`sh -c`で包めば何でも通る。回避のために使わない

## 理由の書き方

- **「何をするか」でなく「なぜするか」**。コマンドは`cmd`に残るので、理由で繰り返さない
  - 悪い：`bcuz "lint実行" qsoku lint`
  - 良い：`bcuz "srwr の edit を直したあとの確認" qsoku lint`
- 1行・短く。あとで`bcuz -H`を読む人（未来の自分や人間）に通じる言葉で。空白だけは不可（終了コード125）
- 日本語でよい（`mtqg`の記録と同じ）

## bcuz を通さなくてよいもの

Bashツールを通らないものは、フックの対象外で、そのまま使う。

- ファイルの変更：srwrの`look`・`edit`・`new`（理由`why`が付く。`.claude/rules/srwr-usage.md`）
- 読む・探す：組み込みの`Read`・`Grep`
- 記録：mtqgのMCPツール（`.claude/rules/mtqg-usage.md`）。Bashから`mtqg`を呼ぶなら、`bcuz "記録" mtqg …`
- フック（`mtqg hook`・`srwr hook`・`build.sh`）。Claude Codeが直接動かす

## 履歴を見る

```
bcuz -H                                  新しい順に全部
bcuz -H --date 2026-10-14                その日に始めた実行だけ
bcuz -H --since 2026-10-01 --until 2026-10-10
```

- 時刻は、UTCで持ち、表示だけ見る側のタイムゾーンに直す（`.claude/rules/datetime.md`と同じ考え方）
- 「このテストは、何のために、いつ回したか」が分かる。**作業の区切りで、人間への報告の根拠に使う**

## 終了コード

`bcuz`自身の終了コードは、コマンドのもの。次だけが`bcuz`のもの。

| 終了コード | 意味 |
|---|---|
| 125 | `bcuz`自身のエラー（理由が空、コマンドが無い、引数の誤り）。何も実行されていない |
| 126 | コマンドは見つかったが起動できない（実行権限が無い、など）。記録される |
| 127 | コマンドが見つからない（綴り違い）。記録される |

## 記録の置き場

- `.bcuz/`は、リポジトリ直下に置く。`.devcontainer/postCreate.sh`が`bcuz --init`で作る（すでにあれば何もしない）。無ければコマンドは実行されるが記録されず、`bcuz: not recorded: no .bcuz/ found`が出る
- **`.bcuz/history.jsonl`はコミットしない**（`.gitignore`の`/.*`で対象外）。実行のたびに増え、マージのたびに衝突するため。チームで共有したいときだけ、`.gitignore`に`!/.bcuz/`を足す（`merge=union`にするなら`.gitattributes`に`.bcuz/history.jsonl merge=union`）
- `.bcuz/`の中のファイルを直接編集しない
- `.srwrignore`で、`.bcuz/`はテープに記録しない（`bcuz`が書くもの）

## サンドボックス

Bashツールは`.claude/settings.json`のサンドボックスの中で動く。`.bcuz/`はリポジトリの中なので、書き込み先の追加は要らない。`bcuz`はシェルを通さずにコマンドを起動し、標準入出力をそのまま渡すので、サンドボックスの挙動は変わらない。
