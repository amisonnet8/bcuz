# CLAUDE.md

このファイルは、このプロジェクトで作業する際、Claude Codeが常に踏まえるべき前提を示す。

## このプロジェクトについて

**bcuz** は、理由を付けてコマンドを実行し、「いつ・なぜ・何を・どうなったか」を `.bcuz/history.jsonl` に残すラッパー（`bcuz "xxテスト実施" make test`）。Go の単一バイナリ、外部依存なし、Linux・macOS・Windows。

- **仕様は `docs/` で確定している（正本）。`docs/` は英語が正本（`X.md`）で、日本語は翻訳（`X_ja.md`）。** 使い方・出力・メッセージ・終了コードは `docs/reference/cli.md`、記録の形式は `docs/reference/history.md`、入出力の例は `docs/examples/`。実装はこれに一字一句合わせる
- 判断の理由は `docs/design/decisions.md`、やらないことは `docs/design/overview.md`、追加の候補は `docs/design/limitations.md`（候補は、人間が決めるまで作らない）

### 最初に読むもの

1. **`dev/roadmap.md`** — 段階（R0〜R5）と完了条件。**この順に進める**。今どこかは `mtqg context`
2. **`docs/README.md`** → `docs/reference/cli.md`・`history.md` → `docs/examples/`
3. **`.claude/rules/working-with-human.md`** — 人間とのやり取り（下の3原則の詳細）

### 人間とのやり取りの3原則（必ず守る）

1. **小さなつまずき**（サンドボックスで書けない、など）は、原因と「してほしいこと」を5行以内で頼む。人間は「やる／やらない」だけを判断する。「どうしますか」と丸投げしない。些細な判断は聞かずに決めて記録する
2. **人間がする確認**は、準備・起動・取得を自動にし、人間は見て OK／NG を付けるだけにする。1コマンドで、何度でも同じ手順。確認を頼む前に、その自動化を作る
3. **利用者の目に触れる見え方**（出力・メッセージ・使い方・記録の形式）を `docs/` から変える・足すときは、変更前後の例を並べた短い資料で了承をもらってから実装する

### 進め方の要点

- **動くものを小さく作る。** 段階ごとに完了条件を実際に動かして確かめる。先回りして作らない
- 進捗は mtqg の todo。**最初のセッションで、`dev/roadmap.md` の段階を `mtqg t add` で立てる**（段階ごとに1件。中のタスクは着手前に足す）
- 仕様と実装がずれたら `.claude/rules/documentation.md`

## 守ること（設計の芯）

実装の細部より優先する。これを崩す変更は、実装する前に必ず相談すること。

- **コマンドは素通し。** シェルを通さずに起動し、標準入出力・標準エラー出力は保持も加工もしない。終了コードは子のものをそのまま返す（bcuz 自身のエラーは 125・126・127 のみ）
- **記録は追記だけ。** 書き換えない・消さない。1行は項目の順を固定した構造体から作り、1回の `Write` で追記する
- **記録の失敗で、コマンドの終了コードを変えない。** `.bcuz/` が無くても実行する（警告だけ）
- **日時は UTC で持ち、表示のときだけ見る側のタイムゾーンに直す。** 記録の項目（`v`・`ts`・`why`・`cmd`・`dir`・`exit`・`ms`）と、`docs/reference/` の出力・メッセージは変えない（変えるときは人間の了承が先）
- **役割外の機能を足さない。** 権限の昇格・出力の保存・設定ファイル・色・`--json` などは作らない（候補は `docs/design/limitations.md`。人間が決めるまで作らない）
- **外部依存ゼロ・cgo なし。** 標準ライブラリだけで書く

## ドキュメントの言語

- **`docs/` と `README.md` は英語が正本**（`X.md`）。日本語は翻訳で、同じ場所に `X_ja.md`。同じ変更の中で両方を直す
- `docs/` 以外の文書（`dev/`・`.claude/`・`CLAUDE.md`）は**日本語**
- コード中のコメントと、bcuz が出すメッセージは**英語**
- 詳しくは `.claude/rules/documentation.md`

## 参照すべきファイル

作業を始める前に、以下を確認すること。

- **`.claude/rules/bcuz-usage.md`** — bcuzの使い方。**Bashで実行するコマンドは、必ず`bcuz "理由" コマンド…`の形で打つ**（フックが、そうでないものを断る）。理由は「なぜ」を書く。**コマンドを打つ前に読む**
- **`.claude/rules/mtqg-usage.md`** — mtqgの使い方（記録者、種類の使い分け、こまめに記録する）
- **`.claude/rules/srwr-usage.md`** — srwrの使い方（**厳格モード**）。**`Edit`・`Write`は禁止されており、ファイルの変更は`look`／`edit`／`new`（理由`why`つき）で行う。作業のまとまりごとに`session`で題つきのテープを始め、コミットの前にテープを閉じる。ファイルを変える前に読む**
- **`.claude/rules/qsoku-usage.md`** — qsokuの使い方（`qsokufile`の名前の一覧、呼び方、書式、`//`、シェル連携）。**ビルド・テスト・lintは`go`などを直接打たず、ここの名前で実行する**
- **`.claude/rules/testing.md`** — テスト方針（開発環境の前提、動作確認の粒度、Trivy、ShellCheck、`-race`、Bashサンドボックス・CIの落とし穴）
- **`.claude/rules/directory-structure.md`** — ディレクトリ構成と、ファイルの置き場所の判断基準。**新しいファイルを追加するときに読む**
- **`.claude/rules/datetime.md`** — 日時の扱い（データは UTC で持ち、表示だけ利用者のタイムゾーンに直す。瞬間で比べる、時計を差し替える）。**日時を保存・表示・比較するコードを書くときに読む**
- **`.claude/rules/working-with-human.md`** — 人間とのやり取り（頼み方の3種類、小さなつまずきの頼み方、確認の自動化、見え方を変えるときの了承）。**人間に何かを頼む前に読む**
- **`.claude/rules/naming.md`** — 命名規則（製品名、用語の日本語と英語の対応、JSON のフィールド、メッセージ、Go の名前）
- **`.claude/rules/distribution.md`** — 配布方法（`go install`、GitHub Releases、cgo なし、バージョン）
- **`.claude/rules/documentation.md`** — ドキュメントの書き方（文書の役割、`docs/` が正本、仕様と実装がずれたとき、言語）
- **`.claude/rules/go-code.md`** — Go の bcuz 固有の決まり（引数の読み方、実行とシグナル、記録の書き方、履歴の読み方）。**コードを書く前に読む**

これらのルールファイルは、実装中に得た細かい気づき・教訓をClaude Code自身が育てていくものである。新しく気づいたルール・踏んだ落とし穴があれば、該当するファイルに追記すること。どのファイルにも当てはまらない新しい種類の気づきであれば、新しいルールファイルを作ってよい（作ったらこの一覧にも足すこと）。

## 開発の進め方

- **Bashのコマンドは、すべて`bcuz "理由" コマンド…`で実行する**（`bcuz "lintの確認" qsoku lint`）。理由は「何をするか」でなく「なぜするか」。`bcuz`で始まらない・`&&`や`;`でつなぐコマンドは、`.claude/settings.json`の`PreToolUse`フック（`.claude/hooks/require-bcuz.sh`）が実行前に断る。実行の記録は`.bcuz/history.jsonl`に残り、`bcuz -H`で見られる（`.claude/rules/bcuz-usage.md`）
- **ファイルの変更は、srwrの道具（`look`で見て`edit`、新規は`new`）で行い、すべてに理由（`why`）を添える。** 組み込みの`Edit`・`Write`は`.claude/settings.json`で禁止されている（厳格モード）。Bashでの書き換え（`sed -i`・リダイレクト）もしない
- **作業のまとまり（todo・bug・人間の依頼・作業の種類の変わり目）ごとに、最初に`session`で題つきのテープを始める。** 1つの作業＝1本のテープ、コミットも同じ区切り
- **コミットの前に、テープを閉じる**（`srwr tapes check`のあと`srwr tapes new`）。書き込み中の生のテープは`.gitignore`で対象外で、閉じて圧縮した`.srwr/tapes/*.tape.jsonl.gz`だけがgitに入る（`.claude/rules/srwr-usage.md`）
- 実装後は`.claude/rules/testing.md`に従い、ビルド確認に加えて実際の動作確認（`bcuz "…" qsoku check`、必要に応じて`qsoku race`・`qsoku trivy`・`qsoku shellcheck`）を行うこと。判断に迷ったら実行する側に倒すこと
- 依存ライブラリを足したら、Trivyでライセンス・脆弱性を確認すること（`qsoku trivy`）。bcuz は外部依存ゼロが方針なので、足すこと自体が重要な判断（人間に聞く）
- 人間に確認を頼むときは、`.claude/rules/working-with-human.md` 3章の形で、1コマンドと見るところだけを渡すこと

## gitの扱い（Claude Codeの作業として）

- **pushは人間が行う。** Claude Codeはコミットまでで止めること。`git push`は`.claude/settings.json`で拒否（deny）されている
- 作業の区切りでこまめにコミットしてよい（消してしまったときに戻せるように）
- `git reset --hard`・`git clean`は、コミットしていない作業を消すので、使う前に必ず理由を説明すること
- **コミットの前に`git config user.email`／`user.name`を確かめる。** 正しい値は`65074207+amisonnet8@users.noreply.github.com`／`amisonnet8`（GitHubのプライバシー用noreplyアドレス）。違う値なら、`.git/config`の`[user]`（ローカルの上書き）を疑う。ローカルに設定が無いと、コンテナのglobalの設定（noreplyでないことがある）が使われる

## 権限・自動化について

`.claude/settings.json`（人間が管理する）により、次の設定になっている。

- 拒否（deny）：`git push`、`gh pr merge`、`gh release create`
- **`bcuz`を通さないBashコマンドの拒否**：`PreToolUse`フック（`.claude/hooks/require-bcuz.sh`）が、`bcuz`で始まらないコマンドと、後ろに別のコマンドをつなぐもの（`&&`・`||`・`;`・`&`・改行・コマンド置換）を、実行前に断る。直し方は断りのメッセージと`.claude/rules/bcuz-usage.md`に書いてある
- 実行前に確認（ask）：`git reset --hard`、`git clean`、`sudo`、`gh pr create`
- **Bashコマンドは既定でOSレベルのサンドボックス（Linux bubblewrap）内で実行される。** ファイルシステムの書き込み先とネットワーク接続先は許可リストで絞られている。この範囲内なら`curl`・`wget`・`go get`などは確認なしに実行できる。リストに無い宛先・書き込み先は`sandbox_violations`として拒否される。そのときは、`.claude/rules/working-with-human.md` 2章の形（原因と、`.claude/settings.json` のどこに何を足すか）で**短く頼む**（許可を広げるかは人間が判断する）。落とし穴は`.claude/rules/testing.md`「Bashサンドボックスの落とし穴」参照

確認を求められた場合、無理に実行しようとしない。`.claude/settings.json`を変更したいときは、変更案を差分で示すにとどめること（人間が貼るだけで済む形にする）。

**srwrの厳格モード**：`permissions.deny`に`Edit`・`Write`・`MultiEdit`・`NotebookEdit`があり、`.mcp.json`にsrwrのMCPサーバー、`PostToolUse`に`srwr hook`（読み取りと外部変更の記録）がある。テープ（`.srwr/tapes/`）はgitで共有する（鍵`.srwr/key`は入れない）。許可を広げる（緩いモードにする）のは人間の判断。詳細は`.claude/rules/srwr-usage.md`。

ビルドの自動フック（`PostToolUse`）が設定されている。srwrの`edit`・`new`で`.go`・`go.mod`・`go.sum`を変えると`.claude/hooks/build.sh`が`qsoku build`を実行し、失敗すると理由がClaude Codeに返る。複数ファイルにまたがる編集の途中は、まだ書いていないファイルを参照して一時的に失敗することがある。

## mtqgでの記録について

作業の中で生まれる「気づき・やること・質問と答え・不具合とそのやり取り・用語の合意・この先も効き続ける決まり事」は、コードにもコミットにも残らず、チャットの履歴やメモアプリに散って失われやすい。`mtqg`（`m`emo & rules・`t`odo・`q`a & bugs・`g`lossary）はそれを`.mtqg/journal.jsonl`に追記し、gitに乗せて共有する。使い方は`.claude/rules/mtqg-usage.md`。`.mtqg/`にはこのテンプレートの時点で決まり事（rule）を3件登録済み——`mtqg context`で確認できる。AIエージェントからは、フック（`.claude/settings.json`の`mtqg hook`）とMCPサーバー（`.mcp.json`の`mtqg mcp`）の両方で使える。

## ルール・スキルの提案

作業を進める中で、以下に気づいたら、都度こちらから提案すること（提案するだけで、勝手に作成・適用はしない。判断はこちらが行う）。

- **新しいルールにした方がよさそうな知見**: 同じ種類の判断や落とし穴に複数回遭遇した、既存の`.claude/rules/`のどれにも当てはまらない、といった場合。どのファイルに追記すべきか、あるいは新規ファイルが必要かも合わせて提案する
- **Skill化した方が効率的そうな作業**: 「同じ手順を3回目繰り返している」など、パターン化できそうな一連の作業に気づいた場合。まだ実装が薄い段階では無理に提案する必要はない。パターンが実際に繰り返されてから提案すること

提案は気づいたタイミングで随時行ってよく、まとめて報告するために貯めておく必要はない。

This repository records its development with mtqg: run `mtqg context` at the start of a session, and see `.mtqg/SCHEMA.md` for the data format.
