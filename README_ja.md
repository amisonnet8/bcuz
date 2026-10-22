*[English](README.md) | **日本語***

# bcuz

<p align="center">
  <img src="docs/images/banner.svg" alt="bcuz: 理由を付けてコマンドを実行する" width="720">
</p>

<p align="center">
  <b>理由を付けてコマンドを実行し、「いつ・なぜ・何を・どうなったか」を残します。</b>
</p>

<p align="center">
  <a href="https://github.com/amisonnet8/bcuz/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/amisonnet8/bcuz/actions/workflows/ci.yml/badge.svg"></a>
  <a href="https://github.com/amisonnet8/bcuz/releases"><img alt="Release" src="https://img.shields.io/github/v/release/amisonnet8/bcuz"></a>
  <a href="LICENSE"><img alt="License: MIT" src="https://img.shields.io/github/license/amisonnet8/bcuz"></a>
  <img alt="Go" src="https://img.shields.io/github/go-mod/go-version/amisonnet8/bcuz">
  <img alt="Platforms" src="https://img.shields.io/badge/platform-linux%20%7C%20macos%20%7C%20windows-informational">
</p>

<p align="center">
  <a href="#ひと目で">ひと目で</a> ·
  <a href="#特徴">特徴</a> ·
  <a href="#入れ方">入れ方</a> ·
  <a href="#使い方">使い方</a> ·
  <a href="#終了コード">終了コード</a> ·
  <a href="#文書">文書</a>
</p>

## ひと目で

```console
$ bcuz --init
bcuz: created .bcuz/
$ bcuz "テストを実行" make test
$ bcuz -H
2026-10-12 06:35:12    0    8.3s  .  テストを実行  $ make test
```

コマンドの前に、理由を1つ付けるだけです。あとで履歴を見れば、**なぜ**打ったのかが分かります。

## 特徴

- 📝 **理由が一緒に残る**：`bcuz "理由" コマンド`。`sudo` と同じく、前に付けるだけ
- 🔍 **あとで分かる**：いつ・なぜ・何を・どうなったか（終了コード・時間）を、1回の実行につき1行の JSON で
- 🧩 **何も変えない**：出力・TTY・終了コードはそのまま。シェルも通さない
- 📦 **単一バイナリ・依存なし**：Go 製。Linux・macOS・Windows
- 🤖 **AI エージェントにも**：エージェントが何のために何を打ったかを、`bcuz -H` で読める
- 🗂️ **ただのファイル**：追記だけの `.bcuz/history.jsonl`。`grep`・`jq`・git でそのまま扱える

## 入れ方

```bash
go install github.com/amisonnet8/bcuz/cmd/bcuz@latest
```

ビルド済みのバイナリ（Linux・macOS・Windows の amd64・arm64）は、[Releases](https://github.com/amisonnet8/bcuz/releases) にあります。

シェルの補完（`bash`・`zsh`・`fish`・`powershell`。[`--completion`](docs/reference/cli_ja.md#--completion)）：

```bash
bcuz --completion bash > ~/.local/share/bash-completion/completions/bcuz
```

## 使い方

```bash
bcuz --init                           # ここに .bcuz/ を作る（最初の1回）
bcuz "テストを実行" make test         # 実行して記録する
bcuz -- "-v を付けて確認" go test -v  # 理由が - で始まるとき
bcuz -H                               # 履歴を新しい順に見る
bcuz -H --date 2026-10-12             # その日だけ（--since・--until も）
gzip -dc old.jsonl.gz | bcuz -H -     # 移した記録を標準入力から見る
```

```text
bcuz [--] <理由> <コマンド> [引数...]   実行して記録
bcuz -H, --history [-] [--date D | --since D] [--until D]
bcuz --init                         カレントディレクトリに .bcuz/ を作る
bcuz --completion <shell>           シェルの補完スクリプトを出す
bcuz -h, --help
bcuz --version
```

<details>
<summary>記録の形（<code>.bcuz/history.jsonl</code> の1行）</summary>

```json
{"v":1,"ts":"2026-10-11T21:35:12.345Z","why":"テストを実行","cmd":["make","test"],"dir":".","exit":0,"ms":8312}
```

</details>

## 知っておくこと

- 記録は、カレントディレクトリから親へたどって最初の `.bcuz/history.jsonl` に追記されます
- シェルは通しません。`&&` やパイプは、呼んだシェルが解釈します（まとめて記録するなら `sh -c '…'`）
- `.bcuz/` が無くてもコマンドは実行されます（警告だけ）

## 終了コード

| 終了コード | 意味 |
|---|---|
| コマンドのもの | そのまま返る |
| 128＋番号 | シグナルで終わった（Ctrl-C なら 130） |
| 125 | bcuz 自身のエラー（記録しない） |
| 126 | 見つかったが起動できない（記録する） |
| 127 | 見つからない（記録する） |

## 文書

- [コマンドの使い方](docs/reference/cli_ja.md)：出力・メッセージ・終了コード
- [記録の形式](docs/reference/history_ja.md)：`history.jsonl` の1行
- [例](docs/examples/README_ja.md)：実際の入力と出力
- [設計](docs/design/overview_ja.md)：目的、判断の理由、制限
- [対話形式のガイド](https://notebook.google.com/notebook/6e5c260f-9493-489a-994f-d3c6b710e24d)：質問しながら bcuz を調べられる（Gemini Notebook 製）

## ライセンス

[MIT](LICENSE)
