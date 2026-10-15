#!/usr/bin/env bash
# PreToolUse hook for the Bash tool: every command must be run through bcuz
# (`bcuz "<reason>" <command> [args...]`), so that when / why / what / result
# land in .bcuz/history.jsonl (.claude/rules/bcuz-usage.md).
#
# Allowed:  bcuz "<reason>" cmd ...   bcuz -- "-x reason" cmd ...
#           bcuz -H [...]  bcuz --init  bcuz -h  bcuz --version
# Refused:  anything that does not start with bcuz, and anything that chains
#           more commands behind the first one (&&  ||  ;  &  newline, command
#           substitution), because those would run without a record.
# Pipes and redirects are allowed (output shaping: `bcuz "x" qsoku check 2>&1 | tail`).
# Not installed yet (bootstrap, e.g. the first `go install`): allowed.
#
# This is a guard against forgetting, not a security boundary: `bcuz "r" sh -c
# '...'` is allowed on purpose (one logical step, one record).
# The single quotes below are intentional: they are literal characters to match.
# shellcheck disable=SC1003,SC2016
set -euo pipefail

# Remove quoted text, so that operators inside quotes are not mistaken for
# shell syntax. Backslash escapes outside quotes drop the escaped character.
strip_quotes() {
  local s=$1 out="" q="" c i
  for ((i = 0; i < ${#s}; i++)); do
    c=${s:i:1}
    if [ -n "$q" ]; then
      if [ "$q" = '"' ] && [ "$c" = '\' ]; then
        i=$((i + 1))
        continue
      fi
      if [ "$c" = "$q" ]; then q=""; fi
      continue
    fi
    case $c in
      "'" | '"') q=$c ;;
      '\') i=$((i + 1)) ;;
      *) out+=$c ;;
    esac
  done
  printf '%s' "$out"
}

# Prints the reason to refuse, or nothing if the command is fine.
check() {
  local cmd=$1 bare
  cmd="${cmd#"${cmd%%[![:space:]]*}"}" # trim leading whitespace

  if [[ ! $cmd =~ ^bcuz([[:space:]]|$) ]]; then
    echo "does not start with bcuz"
    return
  fi
  # The first argument is read as an option only by bcuz itself; here we only
  # let through its own options, a reason, or "--".
  if [[ $cmd =~ ^bcuz[[:space:]]+- ]] &&
    ! [[ $cmd =~ ^bcuz[[:space:]]+(-H|--history|--init|-h|--help|--version|--)([[:space:]]|$) ]]; then
    echo "unknown bcuz option (put '--' before a reason that starts with '-')"
    return
  fi
  if [[ $cmd == *'$('* || $cmd == *'`'* || $cmd == *'<('* || $cmd == *'>('* ]]; then
    echo "contains command or process substitution (it would run without a record)"
    return
  fi

  bare=$(strip_quotes "$cmd")
  # 2>&1, >&2, &> are redirects, not control operators.
  bare=$(sed -E 's/[0-9]*>&[0-9-]*//g; s/&>>?//g' <<<"$bare")
  if [[ $bare == *$'\n'* ]]; then
    echo "contains a newline (one command per call; put a script in a file)"
    return
  fi
  if [[ $bare == *'&'* || $bare == *';'* || $bare == *'||'* ]]; then
    echo "chains commands (&& || ; &): run each one with its own bcuz, or wrap one step in: bcuz \"reason\" sh -c '...'"
    return
  fi
}

input=$(cat)
command=$(jq -r '.tool_input.command // empty' <<<"$input")
if [ -z "$command" ]; then
  exit 0
fi
if ! command -v bcuz >/dev/null 2>&1; then
  exit 0
fi

reason=$(check "$command")
if [ -z "$reason" ]; then
  exit 0
fi

cat >&2 <<EOF
Refused: this Bash command is not run through bcuz ($reason).
Run it as:  bcuz "<why you are running it>" <command> [args...]
Example:    bcuz "lint after the refactor" qsoku lint
See .claude/rules/bcuz-usage.md.
EOF
exit 2
