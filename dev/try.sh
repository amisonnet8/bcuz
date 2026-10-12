#!/usr/bin/env bash
# Open a shell in a new temporary workspace, with the built bcuz on PATH, to
# feel bcuz in a real terminal (dev/roadmap.md, "R2 の確認"). Run it with
# `qsoku try`, which builds bin/bcuz first. Nothing outside the temporary
# directory is touched, and it is removed when the shell is closed.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

mkdir -p "$work/bin" "$work/proj"
cp "$root/bin/bcuz" "$work/bin/bcuz"
cd "$work/proj"
bcuz_bin="$work/bin/bcuz"
"$bcuz_bin" --init

cat <<'EOF'

== bcuz を手で試す（シェルを開きます。終わったら exit か Ctrl-D）

 1. bcuz "色" ls --color=auto
      -> ls の出力に色が付く（bcuz を挟んでも端末のまま）
 2. bcuz "止める" sleep 30      （そのまま Ctrl-C を押す）
      -> すぐ戻り、続けて echo $? で 130 が出る
 3. bcuz "編集" vi x.txt        （適当に打って :q! で終わる）
      -> vi が普通に使える（画面が崩れない）

シェルを閉じると、記録（history.jsonl）を表示します。

EOF

PATH="$work/bin:$PATH" PS1='bcuz-try$ ' bash --norc -i || true

echo
echo "== history.jsonl（1回の実行が1行。exit が 0・130 などになっているはず）"
cat "$work/proj/.bcuz/history.jsonl"
