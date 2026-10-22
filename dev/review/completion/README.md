# シェル補完（`bcuz --completion <shell>`）

## 何を足すか、なぜか

- `bcuz --completion <shell>` で、シェルの補完スクリプトを標準出力に出す。シェルは `bash`・`zsh`・`fish`・`powershell`（mtqg・crube と同じ4つ）
- mtqg・crube は `completion <shell>` だが、bcuz は第1引数だけをオプションとして読み、それ以外は理由になる。`completion` を予約すると「理由 completion でコマンドを動かす」が変わるので、`--completion` にする（q abfc6efbca、了承済み）
- 補完で一番効くのは、**理由の後ろのコマンドを、そのコマンド自身の補完に任せる**こと（`bcuz "テスト" git che<TAB>` → `checkout`）。bcuz の前に付けただけで、いつもの補完が消えないようにする

## 変更前後の例

変更前：

```
$ bcuz --completion bash
bcuz: unknown option: --completion (use '--' before a reason starting with '-')
$ echo $?
125
```

変更後（スクリプトが出る。入れ方は文書に書く）：

```
$ bcuz --completion bash > ~/.local/share/bash-completion/completions/bcuz
$ bcuz --completion zsh > "${fpath[1]}/_bcuz"
$ bcuz --completion fish > ~/.config/fish/completions/bcuz.fish
$ bcuz --completion powershell >> $PROFILE
```

誤り（125。今の `--date` のメッセージと同じ形）：

```
$ bcuz --completion
bcuz: missing shell after --completion (bash, zsh, fish, powershell)
$ bcuz --completion tcsh
bcuz: unknown shell: tcsh (bash, zsh, fish, powershell)
$ bcuz --completion bash zsh
bcuz: unknown argument: zsh
```

TAB で出るもの：

| 打ったところ | 出るもの |
|---|---|
| `bcuz -<TAB>` | `-H` `--history` `--init` `-h` `--help` `--version` `--completion` `--` |
| `bcuz <TAB>`（理由） | 何も出さない（理由は自由な文） |
| `bcuz --completion <TAB>` | `bash` `zsh` `fish` `powershell` |
| `bcuz -H <TAB>` | `-` `--date` `--since` `--until`（行にまだ無いもの） |
| `bcuz -H --date <TAB>` | 何も出さない（日付） |
| `bcuz "理由" <TAB>` | コマンドの名前（シェルのいつもの補完） |
| `bcuz "理由" git <TAB>` | `git` 自身の補完に任せる |
| `bcuz -- "-x の確認" <TAB>` | コマンドの名前（`--` の後は1つずれる） |

`-h` の使い方の行（1行足す）：

```
  bcuz --init
  bcuz --completion <shell>
  bcuz -h, --help
```

## 新しく決めること（推奨案）

1. メッセージは上の3つ。シェルが無い・知らないシェルは、選べる4つを括弧で添える（mtqg と同じく、何を書けばよいかが分かるように）
2. **スクリプトは決まった文で、bcuz に候補を聞きに来ない**（mtqg は TAB のたびに `mtqg candidates` を呼ぶが、bcuz のオプションは少なく変わらない。聞きに来る隠しオプションを足すと、第1引数で読む言葉が増える）。4つのスクリプトが同じものを出すことは、テストで本物のシェルに打たせて確かめる（入っているシェルだけ。今のコンテナは bash・zsh）
3. コマンドの補完への任せ方は、各シェルの決まった仕組み：bash は `_command_offset`（bash-completion が無ければ、コマンドの名前とファイル）、zsh は `_normal`、fish は `complete -C`（実装で `__fish_complete_subcommand` から変えた。こちらは `-` で始まる語・`=` を含む語を飛ばし、理由を数え違える）、PowerShell は `CompleteInput`
4. `-H` の後の `-` と絞り込みは、行にすでにあるものを出さない
5. 開発環境（`.devcontainer/`）は mtqg に揃える：fish（apt）と PowerShell（feature）を入れ、bcuz・mtqg・crube の補完を4つのシェルすべてに入れる。zsh・fish・pwsh には qsoku のシェル連携も入れる
6. **`postCreate.sh` は bcuz を `@latest` で入れるので、次の版を出すまで `--completion` が無い。** 無いときは警告を出して飛ばす（コンテナの作成は止めない）。次の版のあとは、そのまま入る

## 文書に足すもの

- `docs/reference/cli.md`：書式と使い方の行、節「`--completion`」（入れ方と、TAB で出るものの表）、エラーの表に2行、「引数の読み方」のオプションの一覧に `--completion`
- `docs/design/decisions.md`：`completion` でなく `--completion` にした理由、スクリプトが聞きに来ない理由
- `README.md`：入れ方の後ろに補完の1行
- （どれも英語が正本で、`_ja.md` も同じ変更で直す）

## 了承してほしいこと

1. 上の使い方・メッセージ・TAB で出るもので、`docs/` を直して実装してよいか
2. 決めること 2（決まった文のスクリプト）と 6（次の版までは補完を入れずに警告）でよいか
