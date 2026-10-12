# qsokuの使い方

ビルド・テスト・lintなどの決まった作業を、リポジトリ直下の`qsokufile`に**名前を付けて**置き、`qsoku <名前>`で呼ぶツール。Makefileの代わり。手元（devcontainer）とCIで、同じ名前・同じコマンドで動かすために使う。**`qsokufile`に定義した作業は、`go`・`golangci-lint`などを直接打たず、`qsoku <名前>`で実行する。**

## このプロジェクトの名前

| 名前 | 内容 | いつ使うか |
|---|---|---|
| `qsoku build` | `go build ./...` | 編集後のビルド確認（Goファイルの編集後は、`.claude/hooks/build.sh`が自動で呼ぶ） |
| `qsoku fmt` | `golangci-lint fmt` | 整形を書き直したいとき |
| `qsoku vet` | `go vet ./...` | |
| `qsoku lint` | `golangci-lint run` | |
| `qsoku unit` | `go test ./...` | |
| `qsoku cross` | `GOOS=windows`・`darwin`で`vet`と`lint` | OS ごとに分けたコード（プロセス・シグナル）を触ったとき（`check`に含む） |
| `qsoku check` | `vet`・`lint`・`unit`・`cross`をまとめて実行 | **作業の区切りで必ず通す**（`.claude/rules/testing.md`） |
| `qsoku bin` | `go build -o bin/bcuz ./cmd/bcuz` | 手で`bcuz`を動かして確かめるとき |
| `qsoku try` | `bin`をビルドし、一時ディレクトリの作業場に`bcuz`を置いたシェルを開く（`dev/try.sh`）。閉じると記録を表示する | **人間が**本物の端末で、Ctrl-C・色・`vi`を確かめるとき（`dev/roadmap.md`の「R2の確認」） |
| `qsoku race` | `CGO_ENABLED=1 go test -race` | 並行性・排他に関わる変更 |
| `qsoku trivy` | 既知の脆弱性・ライセンスの検査 | 依存を足したとき |
| `qsoku shellcheck` | 追跡中の`*.sh`・`*.bash`をShellCheckにかける | シェルスクリプトを足したとき |
| `qsoku install-check [版]` | 一時ディレクトリに`go install`して、`--version`・`--init`・実行・`-H`を通す（`dev/install-check.sh`）。版なしはこのチェックアウト、`v0.1.0`・`latest`は公開済みのもの | 配布に関わる変更のあと、タグを公開したあと（CIの`check`ジョブも動かす） |
| `qsoku dist` | `dist/`に6つのバイナリと`SHA256SUMS`を作る（`dev/dist.sh`）。`VERSION=v0.1.0 qsoku dist`で、版の一致も確かめる | リリースのビルド（`release.yml`が動かす）、手元の試し |

一覧は、今ある`qsokufile`の内容そのもの。`qsoku .list`（名前とコマンド）か`qsoku .names`（名前だけ）で確かめる。**引数なしで`qsoku`とだけ打っても、使い方と定義済みの名前が出る。**

## 呼び方

```
qsoku <名前> [引数...]
```

- 引数は、項目のコマンドの中で`"$@"`・`$1`として受けられる。**このプロジェクトの`qsokufile`で引数を受けるのは`install-check`だけ**（ほかは、書いていない引数は捨てられる）。受けたいときは、項目を`unit: (cd //; go test "$@")`のように書き換える。そうすれば`qsoku unit -run TestFoo ./internal/...`が使える
- **qsoku自身のオプションは無い。** `--help`・`--version`は**名前として調べられる**ので、「定義されていない」というエラーになる。使い方は`qsoku .help`、バージョンは`qsoku .version`
- コマンドは、どのOSでも、また呼び出し元がどのシェルでも、常に`sh`で実行される
- 標準入出力・標準エラー出力はそのまま素通し。終了コードは、コマンドのものがそのまま返る

### 終了コード

| コード | 意味 |
|---|---|
| 0 | 成功 |
| 1 | qsokuが頼まれたことをできなかった（`qsokufile`が無い・解析できない、`sh`を起動できない、など） |
| 2 | コマンドラインの誤り（**定義されていない名前**、管理用コマンドの引数の数の誤り） |
| それ以外 | 項目のコマンドが実行された後は、そのコマンドの終了コード（qsokuが差し替えることは無い） |

## 管理用コマンド

`.`で始まる。`qsokufile`の名前は`.`で始められないので、名前とぶつからない。

| コマンド | 内容 |
|---|---|
| `qsoku .list` | 定義されている名前とコマンドを、ファイルの順に出す |
| `qsoku .names` | 名前だけを出す（補完が使う。決して失敗しない） |
| `qsoku .add <名前> <コマンド>` | 項目を足す。**すでにある名前なら、その行をその場で置き換える**（コメントや並び順はそのまま） |
| `qsoku .rm <名前>` | 項目を消す。無い名前はエラー |
| `qsoku .edit` | `$EDITOR`で`qsokufile`を開く |
| `qsoku .where` | `qsokufile`のあるディレクトリを出す |
| `qsoku .init` | カレントディレクトリに空の`qsokufile`を作る |
| `qsoku .version` | qsokuのバージョン |
| `qsoku .shell <shell>` | シェル連携のコードを出す（下記） |
| `qsoku .help` | 使い方と、定義済みの名前 |

## `qsokufile`の書き方

```
# 行頭が#の行はコメント。
build: (cd //; go build ./...)
unit: (cd //; go test "$@")
check: qsoku vet && qsoku lint && qsoku unit
```

- **1行が1項目**：`名前: コマンド`。最初の`:`が区切り。**行の継続は無い**（1項目は必ず1行）
- コメントは**行頭が`#`の行だけ**。行の途中・末尾の`#`はqsokuには特別な意味が無く、`sh`がシェルのコメントとして読む
- 名前に使えるのは、英数字（ASCII）・`.`・`_`・`-`。**先頭を`.`にはできない**。大文字・小文字は区別される
- **同じ名前を2回定義するとエラー**（両方の行番号が出る）。何も実行されない
- 別の項目を呼ぶときは、`qsoku <名前>`と書く（`check`の例）
- **項目の中で`exit`を呼ばない。** qsokuが同じスクリプトの末尾に後始末の行を足すので、`exit`があるとそこで終わって届かない。成否は、項目の最後のコマンドの終了コードに任せる

### `//`：`qsokufile`のあるディレクトリ

`//`は、使われている`qsokufile`のあるディレクトリ（環境変数`QSOKU_ROOT`）に置き換わる。**どのディレクトリから`qsoku`を呼んでも、リポジトリ直下で動かすために**、項目は`(cd //; ...)`の形で書く。

- 置き換わるのは、**引用符の外で、単語の先頭**にあるとき（コマンドの先頭、または空白・`;`・`&`・`|`・`(`の直後）
- 置き換わらない：`http://example.com`（`:`の後）、`"//"`・`'//'`（引用符の中）、コメントの中
- **括弧で囲む**（`(cd //; ...)`）と、`cd`が子プロセスの中に閉じるので、呼び出し元のシェルの居場所は動かない。括弧なしの`cd //src && ...`は、呼び出し元のシェルもそこへ動く（シェル連携が居場所を持ち帰るため）
- 引用符の中で場所を使いたいときは、`$QSOKU_ROOT`を直接書く

### 探し方

`qsokufile`は、カレントディレクトリから**親へ順にたどって最初に見つかったもの**が使われる。gitリポジトリの境界では止まらない。サブディレクトリからでも同じ`qsokufile`が使える。見つからなければ、エラー（終了コード1）で`qsoku .init`を案内する。

## 名前を足す・変える

- 新しい作業は、`qsoku .add <名前> '<コマンド>'`で足すか、`qsokufile`を直接編集する。**足したら、上の「このプロジェクトの名前」の表と、`.claude/rules/testing.md`の該当箇所も同じ変更の中で直す**
- CIで走らせたい作業は、`.github/workflows/ci.yml`の`run: qsoku <名前>`に書く（手元とCIでコマンドが食い違わない）
- 例：e2eテストを足す

  ```
  test: (cd //; go test -tags e2e -count=1 ./e2e/...)
  ```

## シェル連携

`.devcontainer/postCreate.sh`が、`~/.bashrc`に次の1行を足す。

```
eval "$(qsoku .shell bash)"
```

これで、次の2つが入る。

- **`qsoku`というシェル関数**：項目が`cd`した居場所を、呼び出し元のシェルへ持ち帰る。関数の中は`command qsoku`で本物のバイナリを呼ぶ
- **TAB補完**：`qsoku `の後でTABを押すと、その場で`qsoku .names`を呼んで、今ある`qsokufile`の名前を出す。管理用コマンド（`.add`など）は、シェルのスクリプトに固定で入っている

## OSごとの注意

- **Windows**：`qsokufile`のコマンドは、Git for WindowsのMSYS `sh.exe`で動く。`Git\bin`（`C:\Program Files\Git\bin`）を`PATH`に足すこと。CIの`run:`は`shell: bash`を明示する（`.claude/rules/testing.md`「GitHub Actions CIの落とし穴」）
- **Git Bash（Windows）**：引数が`//`で始まると、MSYSが先に書き換えてしまい、qsokuに届くのは`/`1つになる。`MSYS_NO_PATHCONV=1`で止められる
