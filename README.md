# bcuz

理由を付けてコマンドを実行し、「いつ・なぜ・何を・どうなったか」を残すラッパー。

```
$ bcuz --init
bcuz: created .bcuz/
$ bcuz "xxテスト実施" make test
$ bcuz -H
2026-10-12 06:35:12    0    8.3s  .  xxテスト実施  $ make test
```

> **開発中です。** 仕様は [docs/](docs/README.md) で確定しています。

## 入れ方

```
go install github.com/amisonnet8/bcuz/cmd/bcuz@latest
```

## 使い方

```
bcuz [--] <理由> <コマンド> [引数...]   実行して記録
bcuz -H, --history [--date D | --since D] [--until D]
bcuz --init                         カレントディレクトリに .bcuz/ を作る
bcuz -h, --help
bcuz --version
```

- 記録は、カレントディレクトリから親へたどって最初の `.bcuz/history.jsonl` に、1回の実行＝1行の JSON で追記されます
- 終了コードはコマンドのものがそのまま返ります。`.bcuz/` が無くてもコマンドは実行されます（警告だけ）
- 詳しくは [docs/reference/cli.md](docs/reference/cli.md)、記録の形式は [docs/reference/history.md](docs/reference/history.md)、例は [docs/examples/](docs/examples/README.md)

## 開発

devcontainer で開き、`CLAUDE.md` と `dev/roadmap.md` を読む。ビルドとテストは `qsoku check`。

## ライセンス

MIT
