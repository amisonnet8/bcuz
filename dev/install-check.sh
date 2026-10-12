#!/usr/bin/env bash
# Install bcuz with `go install` into a temporary directory and use it in a
# temporary workspace (dev/roadmap.md, R5). Run it with `qsoku install-check`.
#
#   qsoku install-check            install this checkout (./cmd/bcuz)
#   qsoku install-check v0.1.0     install the published module at that version
#   qsoku install-check latest     the same, at @latest
#
# Nothing outside the temporary directory is touched.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
ver="${1:-}"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# A native Go on Windows needs a Windows path for GOBIN.
native="$tmp"
if command -v cygpath >/dev/null 2>&1; then
  native="$(cygpath -m "$tmp")"
fi

export GOBIN="$native/bin"
export CGO_ENABLED=0

if [ -z "$ver" ]; then
  (cd "$root" && go install ./cmd/bcuz)
else
  # Outside the checkout, so that go.mod here cannot decide the version.
  (cd "$tmp" && go install "github.com/amisonnet8/bcuz/cmd/bcuz@$ver")
fi
bcuz="$tmp/bin/bcuz"

fail() {
  echo "install-check: $*" >&2
  exit 1
}

# --version: a tag when the module was installed at a tag, (devel) for a
# build from a checkout.
got="$("$bcuz" --version)"
case "$ver" in
  "" | latest)
    if ! printf '%s\n' "$got" | grep -Eq '^bcuz (\(devel\)|v[0-9]+\.[0-9]+\.[0-9]+.*)$'; then
      fail "--version printed '$got'"
    fi
    ;;
  *)
    [ "$got" = "bcuz $ver" ] || fail "--version printed '$got', want 'bcuz $ver'"
    ;;
esac

work="$tmp/work"
mkdir "$work"
cd "$work"

got="$("$bcuz" --init)"
[ "$got" = "bcuz: created .bcuz/" ] || fail "--init printed '$got'"

"$bcuz" "install check" go version >/dev/null || fail "running a command failed"

got="$("$bcuz" -H)"
case "$got" in
  *"  install check  \$ go version") ;;
  *) fail "-H printed '$got'" ;;
esac

echo "install-check: ok ($("$bcuz" --version))"
