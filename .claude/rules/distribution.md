# 配布方法

配布物は `bcuz` のバイナリ1つ（`dev/roadmap.md` の R5）。

- **`go install github.com/amisonnet8/bcuz/cmd/bcuz@latest`** で入れられるようにする（最初の配布方法）
- GitHub Releases に、Linux・macOS・Windows（amd64・arm64）のビルド済みバイナリを置く。**タグとリリースの作成は人間が行う**（`gh release create` は deny）
- **cgo は使わない**（`CGO_ENABLED=0`）。シグナルなど OS に依存する処理は、ビルドタグで OS ごとに分け、cgo なしで書く
- **外部依存を足さない。** JSON・プロセスの起動・シグナル・日付はすべて標準ライブラリで足りる。足すのは重要な判断（`.claude/rules/working-with-human.md` 1章）
- **バージョン**：`bcuz --version` は `runtime/debug.ReadBuildInfo()` のモジュールのバージョンを出す（`go install …@vX.Y.Z` で入れたときにタグが入る）。取れないとき（手元のビルド）は `(devel)`。出力の形は `docs/reference/cli.md`
- ビルド済みバイナリはリポジトリにコミットしない（`/bin/`・`/dist/` は `.gitignore` 済み）
- 記録の形式（`v`）を上げたら、古い記録も読めることを確かめる（`docs/reference/history.md` の互換性）
