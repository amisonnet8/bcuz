# bcuz の文書

**bcuz** は、理由を付けてコマンドを実行し、「いつ・なぜ・何を・どうなったか」を JSONL に残すラッパー。

```
bcuz "xxテスト実施" make test
```

## 構成

| 場所 | 中身 | 読む人 |
|---|---|---|
| [reference/](reference/) | **正本。** 守る約束（コマンドの使い方、記録の形式） | 使う人・作る人 |
| [design/](design/) | なぜそう決めたか、やらないこと、制限と追加の候補 | 作る人 |
| [examples/](examples/) | 実際の入力と出力の例 | 使う人・作る人（テストの答え） |

### reference/

- [cli.md](reference/cli.md)：コマンドの使い方（実行、`--init`、`-H`/`--history`、`-h`/`--help`、`--version`）、`.bcuz/` の探し方、終了コード、メッセージ
- [history.md](reference/history.md)：記録の形式（`history.jsonl` の1行）

### design/

- [overview.md](design/overview.md)：目的、使う場面、やらないこと
- [decisions.md](design/decisions.md)：判断とその理由
- [limitations.md](design/limitations.md)：制限と、追加の候補

### examples/

- [README.md](examples/README.md)：例の一覧

## 文書の決まり

- `reference/` が正本。実装と食い違ったら、`reference/` に合わせる。仕様を変えるときは、先に `reference/` を直す
- `examples/` の出力は、実装の出力と一字一句同じにする（時刻・時間など毎回変わる値は、例の中で断る）
