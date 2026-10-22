# テスト方針

## 開発環境の前提

開発者の手元環境は**Linux（devcontainer）のみ**。Windows・macOSでの動作確認は手元では行わず、GitHub Actionsのホステッドランナーに委ねる（`.github/workflows/ci.yml`の3OSマトリクス）。

## 実装後の動作確認

実装したら、ビルド確認に加えて実際の動作確認を行うこと。ロジック上正しそうに見えても、動かして初めて見つかる不具合はある。

- `qsoku check`（vet・lint・単体テスト、他 OS 向けの型検査と lint（`cross`）。整形の差異は`lint`が報告する）は作業の区切りで必ず通す
- 並行性・排他制御に関わる変更は`qsoku race`も通す
- 依存を足したら`qsoku trivy`（既知の脆弱性・ライセンス）を通す
- シェルスクリプトを足したら`qsoku shellcheck`を通す（`git add`してから走らせる。追跡中のものだけが対象）
- `.claude/hooks/require-bcuz.sh`を変えたら`qsoku hook`（判定の表）を通す。判定を変えたら、表も同じ変更の中で直す

## bcuz での動作確認

- **答えは `docs/`。** `docs/reference/` の書式・メッセージ・終了コードと、`docs/examples/` の入出力に、一字一句合わせる。食い違えば実装を直す（`docs/` が誤りだと思うときは、`.claude/rules/working-with-human.md` 1章の「重要な判断」として聞く）
- **`docs/` の英語と日本語（`_ja.md`）は、`internal/docs/translations_test.go` が対で確かめる**（揃っていること、見出し・表・箇条書き・コードブロックの構造が同じこと、先頭の切り替え、リンクが同じ言語を指すこと）。`examples/` の照合と `cli.md` の使い方の照合は、両方の言語で回る。例を足すときは、英語を正本に書き、日本語を対にする
- **`examples/` は `internal/docs` のテストが照合する。** `history-filter.md` の各ブロックは、`examples/history.jsonl` を履歴の表示に通した結果と、タイムゾーンを固定して（Asia/Tokyo・UTC）比べる。`json` のブロックは、記録の形式（7項目、この順）に合うかを確かめる。例を書き換えて合わせない
- **`basic.md`・`errors.md`・`shell.md`（とその `_ja.md`）と、`README.md`・`README_ja.md` の最初の例（README の使い方の行も）**は、`internal/docs/transcript_unix_test.go` が、書き起こしを本物の bcuz に通す（Unix のみ）。`make`・`git` などは、テストのバイナリ自身への偽のコマンド（`PATH` はその置き場だけ。`mkae` は無いので 127）で、文書に書かれた出力を出し、書かれた終了コードで終わる。確かめるのは bcuz の出力・メッセージ・終了コード・記録で、`ts`・`ms`（履歴の日時・時間の欄）は比べない。`json` ブロックは直前の実行の記録と、`$` の無い履歴の行は、その記録の表示と比べる
- **実行の通しテスト**は、`go build` した本物の `bcuz` を `t.TempDir()` の作業場で動かす。**子のコマンドには、テストのバイナリ自身を使う**（環境変数で「指定の終了コードで終わる」「眠る」「シグナルを待つ」などに切り替える）。`sh`・`sleep` などの外部コマンドは、Windows に無いので使わない
- **シグナル**（SIGINT で bcuz が先に終わらない、SIGTERM・SIGHUP の転送、130 などの終了コード、記録が残ること）は、Unix のビルドタグのテストで押さえる。Windows は Ctrl-C で bcuz が終わらないことだけ
- **TTY の素通し**は、Linux で `script -qec '<bcuz …>' /dev/null` を使い、子が端末を見ていること（`isatty`）を確かめる。`script` が無ければ `t.Skip`
- **補完スクリプト**は、`internal/cli/completion_test.go` が本物のシェル（bash・bash-completion つきの bash・zsh・fish・pwsh）で動かし、`docs/reference/cli.md` の表どおりに出るかを確かめる。入っていないシェルは飛ばすので、通ったことはそのシェルで確かめたことにならない（開発環境は4つとも入れる。GitHub のランナーには pwsh があり、macOS には zsh がある）。zsh は端末が無いと補完の仕組みを起こせないので、`compadd`・`_normal` を代役にして呼ぶ。**代役は本物と同じく、打っている語の頭で絞る**（絞らないと、正しいスクリプトでも落ちる）
  - **シェルを新しく入れたら、`go test -count=1` でキャッシュを使わずに回す。** テストのキャッシュは外のシェルの有無を見ないので、入れる前の「飛ばした」結果がそのまま通ったように見える。飛ばされていないことは `-v` で確かめる
  - **PowerShell のスクリプトは、標準入力（`pwsh -Command -`）では1行ずつ対話のように読まれる。** 複数行の文は空行で終わる：ブロックの中に空行があるとそこで切れ、ブロックの後に空行が無いと次の行が同じ文に飲み込まれる（どちらも何も出ず、誤りも出ない）。補完のスクリプトのブロックには空行を入れず、テストではブロックの後に空行を置く
  - **PowerShell の補完に渡る `$commandAst` の文字列は、末尾に打った空白を含まない。** カーソルの位置まで切り出すときは、足りない分を空白で埋める
- **タイムゾーン**に頼るテストは、`time.Local` を差し替えるか `TZ` を明示する。CI の各 OS のタイムゾーンに頼らない
- `.bcuz/` の探索は、`t.TempDir()` の中に入れ子の作業場を作り、`t.Chdir` してから確かめる。リポジトリ直下の `.bcuz/` は本物の記録（コミットする）なので、テストはそこを探索・書き込みしない（一時ディレクトリから親をたどっても、リポジトリには届かない）

## Goのテスト

- **期待値は、テストを書く前に決める。** 推測で埋めて、落ちてから直す、をしない
  - **計算で決まる値**（ハッシュ・表示幅・日時の書式など）は、実装とは別の道具（`sha256sum`など）で求めて書く。実装の出力を写すだけでは、実装の誤りを固定してしまう
  - **動かして決まる値**（走査・照合の結果など）は、先に実装を動かして出力を読み、1件ずつ正しいと確かめてから貼る
  - **件数より中身で比べる**（名前の一覧など）。件数は数え間違えても気づきにくく、落ちたときに何が違うか分からない
  - テストが落ちたときは、実装と期待値のどちらが誤りかを先に確かめる
  - 「〇〇と同じ書式・動き」と仕様にあるものは、その本家を道具として使って期待値を求める（gitignore の書式なら`git check-ignore`。パスに末尾の`/`を付けて聞くと、ディレクトリそのものの判定が本来と変わることがある。付けずに聞く）
- **表の形の行（行末のコメント、map の値）を手で桁揃えしない。** gofmt が揃え方を決めるので、`qsoku lint`が`File is not properly formatted (gofmt)`で落ちる。書いたら`qsoku fmt`で整形してから`qsoku check`する（整形ツールが書いた変更は、テープに理由なしで載るが、それでよい）
- **gitを使うテストは、開発者の設定の影響を受けないようにする。** `HOME`と`GIT_CONFIG_GLOBAL`を一時ディレクトリに向け、`user.name`・`user.email`はテストの中で決める（`t.Setenv`）
- **Goのソースに、見えない文字（ゼロ幅、U+2028・U+2029、結合文字、文字の向きを変える制御文字、そのほかの制御文字）をそのまま書かない。** 書き込みの途中で本物の文字に置き換わると、見た目では気づけない不具合になる
  - **UTF-8 のバイトのエスケープで書く**（U+2028 なら`\xe2\x80\xa8`）。何の文字かは、コメントかテストの名前に残す。バイトのエスケープは文字列リテラルの中でだけ使える（rune リテラルには使えない）
  - **バックスラッシュと`u`で始まる、コードポイントのエスケープは使わない。** srwrの`edit`・`new`の引数はJSONで渡るので、そう書くと引数の段階で本物の文字になって届く（U+2028・U+2029・U+0301・U+200Dで実測）
  - 書いたあとに`grep -P '[\x{2028}\x{2029}\x{200D}\x{0301}]'`で、本物の文字が無いことを確かめる。`qsoku lint`（staticcheckのST1018）が見つけるのは書式文字だけ

## 壊して確かめる（mutation-check）

通るだけのテストは、効いているとは限らない。実装を1か所ずつ壊して、テストが落ちるかを見る。手順・道具（`mutants/`）・仕様の変異の一覧（`spec.txt`）は、Skill`mutation-check`（`.claude/skills/mutation-check/`）にある。生き残った変異は、握りつぶさずに読み、テストを足すか、同じ意味だと理由を書いて残す。

- **作業ツリーではなく、写しで壊す。** `$TMPDIR`に写しを作り（書き方は Skill の手順1）、その中で変異を入れる。作業ツリーをBashで書き換えると、srwrのテープに理由のない`external`が残り、戻し忘れの心配もある。足したテストは、写しへコピーしてから同じ変異をもう一度入れ、落ちることを確かめる
- **パッケージ自身のテストで生き残ったものは、全テスト（`go test ./...`。e2eがあれば、それも含む）でもう一度回す。** CLIの出力の変異は、e2eが捕まえることが多い
- **見落としやすい形**：整える処理（書き出しの整形など）を通して比べるテストは、手前の誤り（読み込みで空白を残す、など）を隠す。整える処理がもう一度直すため。読み込んだ結果そのものを比べる。期待値が偶然一致することもある（境目ちょうどの例を足す）
- **意味が同じなので残した変異**：（プロジェクトで、残したものと理由を足す。例：エラーを返すときに添える値、ありえない入力の境界）

## Trivy：既知の脆弱性とライセンス

- `trivy fs`で、依存モジュールの既知の脆弱性（CVE）とライセンスを検査する（`qsoku trivy`）
- 何を失敗とみなすか（深刻度、禁止するライセンスの種類）は`trivy.yaml`で固定する
- `go get`で依存を足したら、必ずTrivyを通す
- Trivyは脆弱性データベースの取得に通信が要る（devcontainerのBashサンドボックスを使う場合は下記「Bashサンドボックスの落とし穴」参照）

## ShellCheck：シェルスクリプト

- シェルスクリプトは最小限にする。込み入った処理はGoで書く
- `qsoku shellcheck`は`git ls-files '*.sh' '*.bash'`で追跡中のものを列挙してShellCheckにかける
- シェルスクリプト中のコメントを、行頭が小文字の`# shellcheck`で始まる文にしてはいけない（ShellCheckがインラインディレクティブとして誤解釈し、SC1072/SC1073でパースエラーになる）。「ShellCheck」のように大文字を混ぜて書く
- **`postCreate.sh`の中で`. /etc/os-release`をsourceしない。** ShellCheckのSC1091（追えない形で指定されたファイル）が情報の水準で報告され、終了コードが0でなくなる。ディストリビューションの版は`lsb_release -rs`で取る
- `# shellcheck disable=…`は、`case`の枝の前には置けない（SC1124 のパースエラー）。`case`の文の前に置くと、全部の枝に効く

## シェル向けのもの：補完のスクリプト・シェルとの連携

- devcontainer のシェルは、`postCreate.sh`の`shells`で決める。**このリポジトリは4つ**（`(bash zsh fish pwsh)`。`bcuz --completion`のスクリプトを本物のシェルで確かめるため。テンプレートの初めは`(bash)`だけ）
- 補完のスクリプトなど、自分のシェル向けのものを配るプロジェクトは、**本物のシェルで動かして確かめる**。`shells`を`(bash zsh fish pwsh)`にし、`devcontainer.json`の`features`に`"ghcr.io/devcontainers/features/powershell:2.0.3": {}`を足して、コンテナを作り直す（bash・zsh はベースイメージ、fish は`postCreate.sh`の apt、pwsh は feature）。`pwsh`を挙げて feature が無いと、`postCreate.sh`が止まる
- テストは、入っていないシェルを飛ばす（CI のランナーにあるのは bash と`pwsh`）。飛んだときは、コンテナを作り直したかを確かめる。bcuz の補完のテストの組み立ては、上の「bcuz での動作確認」の「補完スクリプト」

## `-race`の運用

- devcontainerは`CGO_ENABLED=0`。`-race`はcgoを要するため、`qsoku race`（`CGO_ENABLED=1 go test -race -count=1 ./...`）として`qsoku check`とは別にする
- CIでは`race`を別ジョブにし、`ubuntu-latest`・`macos-latest`に限る（`windows-latest`には標準でCコンパイラがない）

## Bashサンドボックスの落とし穴

> devcontainer内でClaude Codeが実行するBashコマンドは、既定でOSレベルのサンドボックス（Linux bubblewrap）にかかる。`.claude/settings.json`の`sandbox`でファイルシステムの書き込み先とネットワーク接続先を許可リストで絞っている。

- **Trivyの脆弱性DBの取得先は`ghcr.io`ではなく`mirror.gcr.io`（`mirror.gcr.io/aquasec/trivy-db:2`）。** 名前から`ghcr.io`（GitHub Container Registry）を許可すればよいと思い込むと、`qsoku trivy`が`sandbox_violations`で失敗する
- **Trivyの脆弱性DBは`~/.cache/trivy`に書き込む。** サンドボックスの`filesystem.allowWrite`にこれが無いと、ネットワークを許可してもダウンロード後の書き込みで`read-only file system`になる
- **`allowWrite`に書いたパスは、ディレクトリが既に存在しないと効かない。** 新しいコンテナには`~/.cache/golangci-lint`・`~/.cache/trivy`が無く、`~/.cache`自体は読み取り専用なので、`qsoku lint`・`qsoku check`は`failed to initialize build cache ... read-only file system`で**失敗**し、`qsoku trivy`もDBを保存できず終了コード1になる。`postCreate.sh`が`.claude/settings.json`の`allowWrite`を`jq`で読み、`~/`と絶対パス（`/go`など。`/dev`・`/proc`・`/sys`と相対パスは除く。作れなければ警告だけ）を`mkdir -p`して先に作る（`allowWrite`を変えても`postCreate.sh`は直さなくてよい）。**すでにあるコンテナを直すには、サンドボックスの外で該当のディレクトリを`mkdir -p`する**
- **Linuxでは`bubblewrap`（`bwrap`）と`socat`が無いと、`sandbox.enabled: true`でも黙って無効のままになる。** エラーも出ないので、サンドボックスが効いているつもりで効いていない状態になる。`postCreate.sh`が、無いものだけ`apt`で入れる（ベースイメージが持っていれば何もしない）
- **golangci-lintは、Goより古いと標準ライブラリを読めず、`typecheck`で落ちる**（`could not load export data … export data version 5 is greater than maximum supported version 4`）。golangci-lintは、自分が抱えるx/toolsの版までのexport dataしか読めない。コンテナのGoが新しいパッチ・版に上がると、固定した版（ビルド済みの`install.sh`の版も）が突然落ちる。**同じ版をコンテナのGoで作り直しても直らない**（実測）。そのため、版を固定せず`go install …/golangci-lint/v2/cmd/golangci-lint@latest`で入れる（`.devcontainer/postCreate.sh`、CIの`version: latest`）。Goが上がった直後で、対応するlintがまだ無いときは落ちたままになる。`qsoku lint`が、このエラーを見つけたら更新のしかたを案内する（`qsokufile`）。直し方は、`go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`を動かす（`proxy.golang.org`と`storage.googleapis.com`への接続が要る）。固定しない代わりに、新しいlintの指摘で赤くなることがある（そのときは設定か指摘を直す）
- **golangci-lintのキャッシュは`~/.cache/golangci-lint`に書き込む。** `filesystem.allowWrite`にこれが無いと、`qsoku lint`・`qsoku check`が`Failed to persist facts to cache ... read-only file system`の警告を出す。結果（`0 issues`）と終了コードには影響せず、キャッシュが効かなくなるだけ
- **`check.trivy.dev`への接続はTrivyのバージョン確認機能で、許可リストに無くても`qsoku trivy`の結果・終了コードには影響しない。** `sandbox_violations`として警告は出るが実害はないので、許可リストに足すかは任意
- **許可リストに無いものに当たったら、自分で回り道を探し続けない。** `.claude/rules/working-with-human.md` 2章の形で、何をどこに足すかを短く頼む
- **`git fetch`・`git pull`と`gh pr create`には`github.com`・`api.github.com`への許可が要る**
- **`sudo`は、サンドボックスの中では承認しても失敗する。** サンドボックスのuidの付け替え（特権の無いユーザー名前空間）で、本来rootが持つファイルがnobodyの持ち物に見え、`sudo`自身の所有者の確認に落ちる。`sudo`が要るときは、人間のターミナルで動かすよう頼む（`.claude/rules/working-with-human.md` 2章）
- **`srwr init`は、書き換えるものがあるときは、サンドボックスの中では`.mcp.json`・`.claude/settings.json`を置き換えられない（書き換えが無ければ動く。そもそもテンプレートから作ったプロジェクトでは動かさなくてよい）**（一時ファイルを`rename`する先が書き込み保護で、`device or resource busy`。`.srwr/`の鍵だけが先にできる）。人間のターミナルで動かす（`.claude/rules/srwr-usage.md`「落とし穴」）。`srwr view`は`~/.cache/srwr`に書くので、同じく人間のターミナルで動かす
- **サンドボックスの書き込み保護が、作業ディレクトリ直下に`.bashrc`・`.gitconfig`などのダミーファイルを出現させることがある。** `git status`に大量の未追跡ファイルとして見えて驚くが、実害はない（未追跡のままなのでコミットには影響しない）。これはサンドボックス内のBashからだけ見える見かけ上のファイル（`/dev/null`のbind mount）で、サンドボックスの外の人間側には存在しない。`.gitignore`はルート直下のドットファイルを許可リスト方式にしてあり、`git add -A`で混ざらない
- **道具のリリースのタグが打たれた直後にコンテナを作り直すと、前の版が入ることがある**（srwr の v0.3.0 で起きた）。Go のモジュールプロキシは`@latest`の答えをしばらく覚えているので、`postCreate.sh`の`go install …@latest`が前の版を入れ、その版に無い補完も警告を出して飛ばす。作り直したら、新しく出た道具の`--version`を見る。古ければ`go install github.com/amisonnet8/<道具>/cmd/<道具>@<新しい版>`で入れ直し（`proxy.golang.org`・`sum.golang.org`・`storage.googleapis.com`への接続が要る）、補完は人間のターミナルで入れ直す（書き先がサンドボックスの外）

## GitHub Actions CIの落とし穴

- **`windows-latest`で`qsoku`を`go install`しても、内部で呼ぶ`sh`がそのままPATHに乗るとは限らない。** `qsokufile`のコマンドは常に`sh`（Git Bash付属）で実行される。CIの`run:`ステップは`shell: bash`を明示し、Git Bashが通ったPATHでqsokuを実行する
- **`/dev/stderr`等のUnix固有のパスはWindows（Git Bash）で壊れる。** `tee /dev/stderr`などを使わず、標準のリダイレクトだけで書く
- **Windowsのcheckoutで改行がCRLFになると、`gofmt -l`が全ファイルを未整形と誤検知する。** ルートの`.gitattributes`（`* text=auto eol=lf`）で防ぐ
- **`uses: owner/repo@TAG`はタグ名と厳密に一致しないと失敗する。** `v`の有無を見落としやすい。書く前に実際のタグ名を確かめる
- **Windowsには実行ビットの概念がない。** `os.Chmod`後に実行ビットを確かめるテストは、`runtime.GOOS != "windows"`でガードする
- **実行権限の無いファイルで「起動できない（126）」を作るテストは、Windowsでは「見つからない（127）」になる。** 拡張子の無い名前（`script`・`build.sh`）は`PATHEXT`に合わず、コマンドとして探されないため。Windowsでは`.exe`の名前で、中身が実行形式でないファイルにすると、起動に失敗して126になる
- **出力に出るパスと比べるテストは、一時ディレクトリを`filepath.EvalSymlinks`で解決してから比べる。** 同じディレクトリでも、macOSは`/var`→`/private/var`のシンボリックリンク、Windowsは8.3の短縮名（`RUNNER~1`）と長い名前が食い違う。`t.TempDir()`の値のまま比べると、macOS・WindowsのCIだけで落ちる
- **Windowsでは、開いたままのファイルを置き換え（`rename`）られないことがあり、開いたハンドルのあるディレクトリは消せない。** 一時ファイル経由の置き換えや、失敗したときの後片付け（`os.RemoveAll`）の前に、先に`Close`する。Windowsのランナーでだけ落ちる
- **Windowsの`os.Symlink`は、リンクの宛先の`/`を`\`にして保存する。** `os.Readlink`は`releases\v2`を返す。宛先を見せる・比べるときは`filepath.ToSlash`で`/`にそろえる。テストの宛先は、区切りを含むもの（`src/auth`）にする。区切りの無い宛先（`src`）では、この食い違いを捕まえられない（crube の`show`の`link`で、CI の Windows でだけ落ちた）
- **OSで分かれる動き（区切りの`\`など）は、手元のLinuxのテストからも確かめる。** `filepath.Separator`のようなOSの値を、パッケージの変数（例：crube の`internal/core`の`osSeparator`）に入れて使い、テストでWindowsの値に差し替える（`t.Cleanup`で戻す。`t.Parallel`にしない）。CIのWindowsまで待たずに、手元で落ちる
