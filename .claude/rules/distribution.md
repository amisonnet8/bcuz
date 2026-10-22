# 配布方法

配布物は `bcuz` のバイナリ1つ（`dev/roadmap.md` の R5）。

- **`go install github.com/amisonnet8/bcuz/cmd/bcuz@latest`** で入れられるようにする（最初の配布方法）
- GitHub Releases に、Linux・macOS・Windows（amd64・arm64）のビルド済みのアーカイブを置く（下の「配布物の作り方」）
- **外部依存を足さない。** JSON・プロセスの起動・シグナル・日付はすべて標準ライブラリで足りる。足すのは重要な判断（`.claude/rules/working-with-human.md` 1章）
- 出力の形（`bcuz --version`）は `docs/reference/cli.md`
- **`go install` の確認**（`qsoku install-check`、`dev/install-check.sh`）：一時ディレクトリに入れ、`--version`・`--init`・実行・`-H` を通す。引数なしはこのチェックアウト、`v0.1.0`・`latest` は公開済みのモジュール。3 OS の CI（`check` ジョブ）が引数なしを動かす。**タグを公開したあとは、`qsoku install-check <タグ>` で確かめる**（サンドボックスでは、モジュールのキャッシュ `/go/pkg/mod` に書けず動かないことがある。そのときは人間のターミナルで）
- 記録の形式（`v`）を上げたら、古い記録も読めることを確かめる（`docs/reference/history.md` の互換性）

## 配布物の作り方（goreleaser）

`.goreleaser.yaml`・`.github/workflows/release.yml`・`tools/licenses`で作る（テンプレートと同じ形。見本は crube（https://github.com/amisonnet8/crube ））。

- **cgo を使わない**（`CGO_ENABLED=0`）。どの OS でも、ほかのライブラリなしに動く。シグナルなど OS に依存する処理は、ビルドタグで OS ごとに分け、cgo なしで書く
- 配布物は linux・darwin・windows の amd64・arm64。linux・darwin は`.tar.gz`、windows は`.zip`。名前は`bcuz_<版>_<OS>_<アーキテクチャ>`（版は`v`なし）。`checksums.txt`を添える。アーカイブには、実行ファイル・`LICENSE`・`README.md`・`README_ja.md`・`licenses/`（依存のライセンス文）を入れる
- **版**：goreleaser が`-X main.version=<タグ>`（`v0.1.0`の形）を埋める。`main`パッケージに`var version string`を置き、空のとき（`go install`で入れたもの）は`runtime/debug.ReadBuildInfo`のモジュールの版を出すと、どちらでも同じ`v0.1.0`になる（crube の`cli.BuildVersion`）。どちらも取れないとき（手元のビルド）は `(devel)`
- **依存のライセンス文**を配布物に同梱する。`tools/licenses`が、3つの OS で実行ファイルに組み込まれるモジュールを`go list -deps`で求め、そのライセンスのファイルと Go の標準ライブラリの`LICENSE`を`bin/licenses/<モジュールのパス>/`に写す（goreleaser の`before`の手順）。bcuz は外部依存が無いので、今は Go の`LICENSE`だけ
  - 依存を足す・外すと、`tools/licenses`の`TestRun`の一覧が落ちる。新しいモジュールのライセンスを確かめてから一覧を直す（Trivy の`qsoku trivy`も通す）

## リリースの流れ

ビルド済みのバイナリなどを GitHub Releases で配るときは、次の流れにする。どのプロジェクトも同じ。

1. **人間が**タグ（`v0.1.0`の形）を打って push する。人間がするのはこれだけ
2. タグで動く workflow（`.github/workflows/release.yml`）が、検査を通してから配布物を作り、Release に添付して**公開する**（本文は goreleaser が前のタグからのコミットで作る）

- **Claude Codeは、タグも Release も作らない。** `gh release create`は`.claude/settings.json`で拒否（deny）されている。Release を作るのは workflow の中だけ
- `release.yml`は、まず`ci.yml`をそのまま呼んで（`workflow_call`）普段の CI の検査をすべて通し、そのあと goreleaser（`.goreleaser.yaml`。`goreleaser-action`の`version: "~> v2"`で、v2 の最新）で配布物を作る。書き込みの権限（`contents: write`）は、Release を作るジョブだけに与える
- goreleaser はいったん**下書き**に添付し、すべて添付し終えてから、最後の手順（`gh release edit --draft=false`）で公開する。途中で落ちると下書きだけが残り、公開された Release に配布物が欠けることはない。同じタグでもう一度動かすと、その下書きを置き換える（`replace_existing_draft`）。落ちたときは、原因を直したコミットに新しいタグを打つか、同じコミットのままなら workflow を再実行する
- **公開済みの Release があるタグでは、何も変えずに失敗する**（最初の手順）。出し直すときは、新しいタグを打つ
- **タグを打つ前に、手元で`qsoku snapshot`を動かして中身を見る**。タグなしで`dist/`に同じ配布物を作る（版は前のタグのパッチを1つ上げた`0.3.1-snapshot`の形、バイナリの`--version`は前のタグ`v0.3.0`のまま。タグが1つも無ければ`0.0.1-snapshot`・`v0.0.0`）。goreleaser は`.devcontainer/postCreate.sh`が入れる。コンテナを作り直す前に要るときは、`GOTOOLCHAIN=auto go install github.com/goreleaser/goreleaser/v2@latest`（goreleaser がコンテナより新しい Go を求めるとき、その Go を取ってきてビルドする。Bashサンドボックスの中では、`proxy.golang.org`と`storage.googleapis.com`への接続が要る）。`.goreleaser.yaml`を変えたら`qsoku snapshot`で確かめる
- ビルド済みのバイナリは、リポジトリにコミットしない（`/bin/`・`/dist/` は `.gitignore` 済み。`.claude/rules/directory-structure.md`「配布物」）
