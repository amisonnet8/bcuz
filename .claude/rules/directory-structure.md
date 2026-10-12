# ディレクトリ構成

新しいファイルを追加する際は、どこに置くかをここで確かめる。**構成を変えたら、この文書も同じ変更の中で直す。**

下の図は、段階（`dev/roadmap.md`）を進めた後の姿。`(R1)` などの印は、その段階で作るもの。

```
.
├── CLAUDE.md               ← プロジェクトルール（参照先の案内）
├── LICENSE                 （MIT）
├── README.md               ← 利用者向けの入口
├── qsokufile               ← ビルド・テストの近道（qsoku）
├── go.mod / go.sum         （module github.com/amisonnet8/bcuz）
├── .gitattributes          （`* text=auto eol=lf`）
├── .gitignore
├── .golangci.yaml          ← lintの設定
├── trivy.yaml              ← 脆弱性・ライセンス検査の設定
├── .mcp.json               ← mtqgとsrwrのMCPサーバーの配線
├── .srwrignore             ← srwrがテープに記録しないファイル（`.mtqg/`など）
├── .srwr/                  ← srwrの鍵・ロック・テープ（最初のhook・MCPの呼び出しでできる。**閉じた`tapes/*.gz`だけコミットする**）
├── cmd/bcuz/main.go        ← 引数を internal/cli に渡して、終了コードで終わるだけ（R0）
├── internal/
│   ├── cli/                ← 引数の振り分け、メッセージ、終了コード、各機能の入口（R0〜）
│   ├── record/             ← 記録の型、1行の読み書き、`.bcuz/` の探索、追記、`--init`（R1）
│   ├── run/                ← コマンドの起動、終了コード、シグナル（`_unix.go`・`_windows.go`）（R2）
│   ├── history/            ← 履歴の読み込み、日付の絞り込み、1行の表示（R3）
│   └── docs/               ← `docs/` の検査だけのテスト（リンク切れ、`examples/` との照合）（R0〜R4）
├── docs/                   ← 正本。reference/・design/・examples/（`docs/README.md`）
├── dev/                    ← 開発のうちうち。roadmap.md、review/（了承をもらう資料）
├── .devcontainer/          ← 開発環境（devcontainer.json、postCreate.sh）
├── .claude/                ← rules/（ルール）、hooks/、settings.json（人間が管理する）
├── .github/workflows/      ← CI（ci.yml）、タグで下書きの Release を作る（release.yml）
└── .mtqg/                  ← mtqgの記録
```

## 配置の判断基準

- **`cmd/bcuz/main.go` は薄く**：`os.Exit(cli.Main(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))` だけ
- **引数の読み方・メッセージの文言・bcuz 自身の終了コード（125）**は `internal/cli`。ほかのパッケージはメッセージを組み立てず、エラーの種類を返す
- **記録の形式に関わるもの**（項目と順、`ts` の書き方、追記の仕方、`.bcuz/` の探し方）は `internal/record` に集める。ほかのパッケージが JSON を直接組み立てない
- **プロセスとシグナル**は `internal/run`。記録のことを知らない（終了コード・開始時刻・時間を返すだけ）
- **履歴の見せ方**（絞り込み、時間の書き方、引数の引用、改行の置き換え）は `internal/history`。`internal/record` で読んだものを受け取り、書き込まない
- **OS ごとの違い**は `internal/run` の `_unix.go`・`_windows.go` に閉じ込める。ほかに OS の分岐を書かない
- 迷ったら「記録の形式か」「プロセスか」「見せ方か」「引数とメッセージか」で分ける

## 各ファイル・ディレクトリの補足

- **`cmd/bcuz/`**: R0 で作る。`main.go`は引数を`internal/cli`に渡すだけにする
- **`internal/`**: Goの仕組みとして、モジュールの外からimportできない。外へ公開するパッケージは`internal/`に置かない
- **`go.sum`**: コミットする（改ざん検知・再現可能なビルドのため）
- **`.mtqg/`**: コミットされる。中のファイルは直接編集せず、すべて`mtqg`のコマンド経由で書く（`.claude/rules/mtqg-usage.md`）。`.mtqg/.local/`はコミットされない
- **`.mcp.json`**: `mtqg mcp`と`srwr mcp`（MCPサーバー）の配線。`mtqg init --agent claude-code`・`srwr init`が書くものと同じ
- **`.srwr/`**: 最初のhook・MCPの呼び出しで、srwrが作る（`srwr init`でも作られる。鍵`key`・`lock`・`active`・`tapes/`・`init-backup/`）。**コミットするのは閉じて圧縮したテープ`tapes/*.tape.jsonl.gz`だけ**。書き込み中の生のテープ`*.tape.jsonl`は`.gitignore`で対象外（閉じると`.gz`に置き換わり、生のままコミットすると「削除＋追加」の差分が出るため）。鍵は共有しない。テンプレートには置いていない（鍵をプロジェクトごとに別にするため）
- **`.srwrignore`**: テープに記録しないファイル（`.gitignore`と同じ書き方）。生成物や巨大なファイルはここに足す。`.env`・`*.pem`・`*.key`などはsrwrが元から外す
- **VSCodeの設定**: `.devcontainer/devcontainer.json`の`customizations.vscode`に置く（拡張と設定を1か所にまとめる）。`.vscode/`は作らない
- **`.gitattributes`**: `* text=auto eol=lf`。Windowsランナーでの改行コード変換による誤検知を防ぐ（`.claude/rules/testing.md`）。`.mtqg/`の中の`.gitattributes`は`mtqg init`が置くもので、別物
- **`.gitignore`のドットファイル**: ルート直下のドットファイル・ドットディレクトリは**許可リスト方式**（`/.*`を無視し、追跡するものだけ`!`で戻す。`.claude/`の中は`/.claude/*`を無視して`settings.json`・`rules/`・`skills/`・`hooks/`だけ戻す）。サンドボックスが作業ディレクトリ直下に出すダミーファイルを`git add -A`で混ぜないため（`.claude/rules/testing.md`）。**追跡するドットファイルを新しく足すときは、`.gitignore`に`!`の行を足すこと**。足さないと`git add`が無視する
- **配布物（ビルド済みバイナリ）**: リポジトリにコミットしない。`.gitignore`には**ルート直下に限定して**書く（`/名前`）。名前だけを書くと、同名のソースディレクトリまで無視される
- **`docs/`**: 正本の文書（`.claude/rules/documentation.md`）。`docs/examples/history.jsonl` はテストの答えでもある
- **`dev/`**: 開発のうちうち。`docs/` から参照しない。`roadmap.md`・`review/`のほか、人間の確認を始める`try.sh`（`qsoku try`が呼ぶ）、配布の`install-check.sh`・`dist.sh`（`qsoku install-check`・`qsoku dist`が呼ぶ。`.claude/rules/distribution.md`）を置く。シェルスクリプトは`qsoku shellcheck`の対象
- **`.bcuz/`**: この開発リポジトリの直下には作らない（`.gitignore` の `/.*` で無視される）。bcuz を動かすテストは `t.TempDir()` の中で行う
- **シェルスクリプト**: 最小限にする。追跡中のものは`qsoku shellcheck`の対象になる（`.claude/rules/testing.md`）
