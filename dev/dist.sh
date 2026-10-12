#!/usr/bin/env bash
# Build the binaries to publish into dist/ (.claude/rules/distribution.md).
# Run it with `qsoku dist`.
#
#   VERSION=v0.1.0 qsoku dist    the release build: the binary must say v0.1.0
#   qsoku dist                   a trial build, named "dev"
#
# The version comes from the build information, which Go fills from the git
# tag when the checkout is exactly at the tag and has no changes.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

version="${VERSION:-}"
label="${version:-dev}"

rm -rf dist
mkdir dist

targets="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64"
for target in $targets; do
  os="${target%/*}"
  arch="${target#*/}"
  ext=""
  if [ "$os" = windows ]; then
    ext=".exe"
  fi
  out="dist/bcuz-$label-$os-$arch$ext"
  echo "== $out"
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath -o "$out" ./cmd/bcuz
done

# The binary for this machine must say the version that is being released.
if [ -n "$version" ]; then
  host_os="$(go env GOOS)"
  host_arch="$(go env GOARCH)"
  ext=""
  if [ "$host_os" = windows ]; then
    ext=".exe"
  fi
  got="$("dist/bcuz-$label-$host_os-$host_arch$ext" --version)"
  if [ "$got" != "bcuz $version" ]; then
    echo "dist: the binary says '$got', want 'bcuz $version'." >&2
    echo "dist: is the checkout exactly at the tag, without changes?" >&2
    exit 1
  fi
fi

(
  cd dist
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum bcuz-* >SHA256SUMS
  else
    shasum -a 256 bcuz-* >SHA256SUMS
  fi
)
echo "== dist/SHA256SUMS"
cat dist/SHA256SUMS
