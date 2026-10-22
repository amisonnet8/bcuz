# ディレクトリ構成

新しいファイルを追加する際は、どこに置くかをここで確かめる。**構成を変えたら、この文書も同じ変更の中で直す。**

**このリポジトリの説明（`.crube/`）も保つ。** ファイル・ディレクトリを足す・動かす・消すときは、同じ変更の中で英語と日本語の説明を直し、コミットの前に両方の言語で`crube status --require all`を通す。やり方（MCPとCLI、`.crube/ignore`の保ち方）は`.claude/rules/crube-usage.md`

下の図は、段階（`dev/roadmap.md`）を進めた後の姿。`(R1)` などの印は、その段階で作るもの。

```
.
├── CLAUDE.md               ← プロジェクトルール（参照先の案内）
├── LICENSE                 （MIT）
├── README.md               ← 利用者向けの入口（英語が正本。日本語は README_ja.md）
├── qsokufile               ← ビルド・テストの近道（qsoku）
├── go.mod / go.sum         （module github.com/amisonnet8/bcuz）
├── .gitattributes          （`* text=auto eol=lf`）
├── .gitignore
├── .golangci.yaml          ← lintの設定
├── trivy.yaml              ← 脆弱性・ライセンス検査の設定
├── .goreleaser.yaml        ← 配布物の作り方（goreleaser。`release.yml`と`qsoku snapshot`が使う。`.claude/rules/distribution.md`）
├── .mcp.json               ← mtqg・srwrとcrube（英語の`crube`・日本語の`crube-ja`）のMCPサーバーの配線
├── .srwrignore             ← srwrがテープに記録しないファイル（`.mtqg/`・`.bcuz/`など）
├── .bcuz/                  ← bcuzの記録（`history.jsonl`）。`postCreate.sh`が`bcuz --init`で作る。**コミットする**（`merge=union`）
├── .crube/                 ← このリポジトリのファイルの説明（crube。`structure.txt`は英語、`structure_ja.txt`は日本語、`ignore`は`.gitignore`に合わせて手で保つ）。**コミットする**
├── .srwr/                  ← srwrの鍵・ロック・テープ（最初のhook・MCPの呼び出しでできる。**閉じた`tapes/*.gz`だけコミットする**）
├── cmd/bcuz/main.go        ← 引数を internal/cli に渡して、終了コードで終わるだけ（R0）
├── internal/
│   ├── cli/                ← 引数の振り分け、メッセージ、終了コード、各機能の入口（R0〜）。`completions/` に埋め込む補完スクリプト（`--completion`）
│   ├── record/             ← 記録の型、1行の読み書き、`.bcuz/` の探索、追記、`--init`（R1）
│   ├── run/                ← コマンドの起動、終了コード、シグナル（`_unix.go`・`_windows.go`）（R2）
│   ├── history/            ← 履歴の読み込み、日付の絞り込み、1行の表示（R3）
│   └── docs/               ← `docs/` の検査だけのテスト（リンク切れ、`examples/` との照合）（R0〜R4）
├── tools/licenses/         ← 配布物に入れる依存のライセンス文を集める道具（goreleaser が`go run`で動かす。実行ファイルには組み込まれない）
├── docs/                   ← 正本（英語。日本語は `_ja.md`）。reference/・design/・examples/・images/（README のヘッダー画像。`docs/README.md`）
├── dev/                    ← 開発のうちうち。roadmap.md、review/（了承をもらう資料）
├── .devcontainer/          ← 開発環境（devcontainer.json、postCreate.sh）
├── .claude/                ← rules/（ルール）、skills/（commit-with-tape：テープを閉じてコミットする手順。mutation-check：壊して確かめる手順と、変異の道具`mutants/`（別の Go モジュール）・仕様の変異の一覧`spec.txt`）、hooks/（build.sh、require-bcuz.sh とそのテスト、crube-untracked.sh：srwr の new のあとに未説明を知らせる）、settings.json（人間が管理する）
├── .github/workflows/      ← CI（`ci.yml`）と、タグで動くリリース（`release.yml`。`ci.yml`を呼んでから Release を作って公開する）
└── .mtqg/                  ← mtqgの記録
```

## 配置の判断基準

- **`cmd/bcuz/main.go` は薄く**：リリースのビルドが埋める`var version string`を`cli.ReleaseVersion`に渡し、`os.Exit(cli.Main(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))`するだけ（`.claude/rules/distribution.md`「版」）
- **引数の読み方・メッセージの文言・bcuz 自身の終了コード（125）**は `internal/cli`。ほかのパッケージはメッセージを組み立てず、エラーの種類を返す
- **記録の形式に関わるもの**（項目と順、`ts` の書き方、追記の仕方、`.bcuz/` の探し方）は `internal/record` に集める。ほかのパッケージが JSON を直接組み立てない
- **プロセスとシグナル**は `internal/run`。記録のことを知らない（終了コード・開始時刻・時間を返すだけ）
- **履歴の見せ方**（絞り込み、時間の書き方、引数の引用、改行の置き換え）は `internal/history`。`internal/record` で読んだものを受け取り、書き込まない
- **OS ごとの違い**は `internal/run` の `_unix.go`・`_windows.go` に閉じ込める。ほかに OS の分岐を書かない
- 迷ったら「記録の形式か」「プロセスか」「見せ方か」「引数とメッセージか」で分ける

## 各ファイル・ディレクトリの補足

- **`cmd/bcuz/`**: R0 で作る。`main.go`は引数を`internal/cli`に渡すだけにする
- **`tools/licenses/`**: 配布物に同梱する依存のライセンス文を集める（`.claude/rules/distribution.md`）。`go run ./tools/licenses <出力先> <main パッケージ>`。`TestRun`は`./cmd/bcuz`の一覧のテスト（依存が変わると直す）。ほかのテスト（`TestRunDependencies`・`TestRunNoLicense`・`TestCommand`）は、`replace`で作った仮のモジュールで道具そのものを確かめるもので、どのPJでも同じ（直さない）
- **`.claude/skills/mutation-check/mutants/`**: 変異の道具。`go.mod`を持つ別のモジュールなので、`qsoku check`（`./...`）の対象に入らない。変えたら、そのディレクトリで整形・vet・テストを通す（Skill の「気をつけること」）
- **`internal/`**: Goの仕組みとして、モジュールの外からimportできない。外へ公開するパッケージは`internal/`に置かない
- **`go.sum`**: コミットする（改ざん検知・再現可能なビルドのため）
- **`.mtqg/`**: コミットされる。中のファイルは直接編集せず、すべて`mtqg`のコマンド経由で書く（`.claude/rules/mtqg-usage.md`）。`.mtqg/.local/`はコミットされない
- **`.claude/hooks/require-bcuz.sh`**: Bashツールのコマンドを、`bcuz`を通っていなければ実行前に断る`PreToolUse`フック。判定の表は`require-bcuz.test.sh`（`bash .claude/hooks/require-bcuz.test.sh`）。判定を変えたら、テストの表も同じ変更の中で直す
- **`.mcp.json`**: `mtqg mcp`と`srwr mcp`（MCPサーバー）の配線。`mtqg init --agent claude-code`・`srwr init`が書くものと同じ。その後ろに、crubeのMCPサーバー（`crube mcp --allow-write`。`crube-ja`は`CRUBE_LANG=ja`）を手で足している（`.claude/rules/crube-usage.md`）
- **`.crube/`**: このリポジトリのファイルとディレクトリの説明（`.claude/rules/crube-usage.md`）。**コミットする**（`.gitignore`の`!/.crube/`）。テンプレートには、テンプレートのファイルの説明を両方の言語で入れてあり、コンテナを作ったあと（`.bcuz/`・`devcontainer-lock.json`・`.srwr/`ができたあと）は、そのまま`status --require all`が通る。閉じたテープ（`*.tape.jsonl.gz`）は、説明を付けず`.crube/ignore`で外す（パターンのエントリにすると、最初のテープを閉じるまで何にも当たらず`orphan`になり、無ければ新しいテープが`untracked`になるため）。`ignore`は`.gitignore`を手で写したもの（crube は`.gitignore`を読まない）
- **`.srwr/`**: 最初のhook・MCPの呼び出しで、srwrが作る（`srwr init`でも作られる。鍵`key`・`lock`・`active`・`tapes/`・`init-backup/`）。**コミットするのは閉じて圧縮したテープ`tapes/*.tape.jsonl.gz`だけ**。書き込み中の生のテープ`*.tape.jsonl`は`.gitignore`で対象外（閉じると`.gz`に置き換わり、生のままコミットすると「削除＋追加」の差分が出るため）。鍵は共有しない。テンプレートには置いていない（鍵をプロジェクトごとに別にするため）
- **`.srwrignore`**: テープに記録しないファイル（`.gitignore`と同じ書き方）。生成物や巨大なファイルはここに足す。`.env`・`*.pem`・`*.key`などはsrwrが元から外す
- **`.devcontainer/devcontainer-lock.json`**: コンテナのビルドで自動で作られる。**できたらコミットする**（この版のfeatureでテストなどが通った証明になる）。featureを新しくしたいときは、lockを消してビルドし直し、新しいlockをコミットする
- **VSCodeの設定**: `.devcontainer/devcontainer.json`の`customizations.vscode`に置く（拡張と設定を1か所にまとめる）。`.vscode/`は作らない
- **`.gitattributes`**: `* text=auto eol=lf`。Windowsランナーでの改行コード変換による誤検知を防ぐ（`.claude/rules/testing.md`）。`.mtqg/`の中の`.gitattributes`は`mtqg init`が置くもので、別物
- **`.gitignore`のドットファイル**: ルート直下のドットファイル・ドットディレクトリは**許可リスト方式**（`/.*`を無視し、追跡するものだけ`!`で戻す。`.claude/`の中は`/.claude/*`を無視して`settings.json`・`rules/`・`skills/`・`hooks/`だけ戻す）。サンドボックスが作業ディレクトリ直下に出すダミーファイルを`git add -A`で混ぜないため（`.claude/rules/testing.md`）。**追跡するドットファイルを新しく足すときは、`.gitignore`に`!`の行を足すこと**。足さないと`git add`が無視する
- **配布物（ビルド済みバイナリ）**: リポジトリにコミットしない。`.gitignore`には**ルート直下に限定して**書く（`/名前`）。名前だけを書くと、同名のソースディレクトリまで無視される
- **`docs/`**: 正本の文書（`.claude/rules/documentation.md`）。`docs/examples/history.jsonl`（英語。日本語は `history_ja.jsonl`）はテストの答えでもある。英語と日本語は対で、同じ構造にする
- **`dev/`**: 開発のうちうち。`docs/` から参照しない。`roadmap.md`・`review/`のほか、人間の確認を始める`try.sh`（`qsoku try`が呼ぶ）、`go install`の確認の`install-check.sh`（`qsoku install-check`が呼ぶ。`.claude/rules/distribution.md`）を置く。シェルスクリプトは`qsoku shellcheck`の対象
- **`.bcuz/`**: bcuzが`.bcuz/history.jsonl`に追記する、コマンドの実行の記録（`.claude/rules/bcuz-usage.md`）。**コミットする**（`.gitignore`の`!/.bcuz/`）。行を足すだけなので、`.gitattributes`の`merge=union`で、マージしても両側の行が残る。コミットの前に`git add .bcuz/history.jsonl`する（`git commit`は`bcuz`を通さない）。秘密をコマンドに書かない（記録に残る）。中のファイルは直接編集しない。`.srwrignore`でテープにも載せない。bcuz を動かすテストは `t.TempDir()` の中で行う
- **シェルスクリプト**: 最小限にする。追跡中のものは`qsoku shellcheck`の対象になる（`.claude/rules/testing.md`）
