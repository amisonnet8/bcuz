# 配布方法

配布物は `bcuz` のバイナリ1つ（`dev/roadmap.md` の R5）。

- **`go install github.com/amisonnet8/bcuz/cmd/bcuz@latest`** で入れられるようにする（最初の配布方法）
- GitHub Releases に、Linux・macOS・Windows（amd64・arm64）のビルド済みバイナリを置く。**タグとリリースの作成は人間が行う**（`gh release create` は deny）
- **cgo は使わない**（`CGO_ENABLED=0`）。シグナルなど OS に依存する処理は、ビルドタグで OS ごとに分け、cgo なしで書く
- **外部依存を足さない。** JSON・プロセスの起動・シグナル・日付はすべて標準ライブラリで足りる。足すのは重要な判断（`.claude/rules/working-with-human.md` 1章）
- **バージョン**：`bcuz --version` は `runtime/debug.ReadBuildInfo()` のモジュールのバージョンを出す（`go install …@vX.Y.Z` で入れたときにタグが入る）。取れないとき（手元のビルド）は `(devel)`。出力の形は `docs/reference/cli.md`
- **`release.yml` が動かすもの**：`qsoku unit` を通してから `VERSION=<タグ> qsoku dist` で配布物を作り、6つのバイナリと `SHA256SUMS` を下書きの Release に添付する（同じタグの Release がすでにあれば、付け直す）。人間と Claude Code の役割は、下の「リリースの流れ」
- **配布物**（`qsoku dist`、`dev/dist.sh`）：`bcuz-<版>-<os>-<arch>`（Windows は `.exe`）を、linux・darwin・windows × amd64・arm64 の6つ。`CGO_ENABLED=0`・`-trimpath`。`VERSION` を付けたときは、この環境向けのバイナリの `--version` がその版と一致しなければ失敗する（タグの上で、変更の無い状態でビルドしていない印）
- **`go install` の確認**（`qsoku install-check`、`dev/install-check.sh`）：一時ディレクトリに入れ、`--version`・`--init`・実行・`-H` を通す。引数なしはこのチェックアウト、`v0.1.0`・`latest` は公開済みのモジュール。3 OS の CI（`check` ジョブ）が引数なしを動かす。**タグを公開したあとは、`qsoku install-check <タグ>` で確かめる**（サンドボックスでは、モジュールのキャッシュ `/go/pkg/mod` に書けず動かないことがある。そのときは人間のターミナルで）
- ビルド済みバイナリはリポジトリにコミットしない（`/bin/`・`/dist/` は `.gitignore` 済み）
- 記録の形式（`v`）を上げたら、古い記録も読めることを確かめる（`docs/reference/history.md` の互換性）

## リリースの流れ

ビルド済みのバイナリなどを GitHub Releases で配るときは、次の流れにする。

1. **人間が**タグ（`v0.1.0`の形）を打って push する
2. タグで動く workflow（`.github/workflows/release.yml`）が、テストを通してから配布物を作り、**下書きの Release**（`--draft`）に添付する
3. **人間が**下書きの中身を見て、公開する

- **Claude Codeは、タグも Release も作らない。** `gh release create`は`.claude/settings.json`で拒否（deny）されている。Release を作るのは workflow の中だけ
- bcuz の workflow が何を動かすかは、上の「`release.yml` が動かすもの」「配布物」
- ビルド済みのバイナリは、リポジトリにコミットしない（`.claude/rules/directory-structure.md`「配布物」）
