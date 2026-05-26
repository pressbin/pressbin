#!/usr/bin/env bash
# Build cross-platform release binaries into ./dist/
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

VERSION="${1:?Usage: ./scripts/build-release.sh v1.0.0}"

LDFLAGS="-s -w -X main.Version=$VERSION"
TARGETS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
)

mkdir -p dist

for TARGET in "${TARGETS[@]}"; do
  OS="${TARGET%/*}"
  ARCH="${TARGET#*/}"
  OUTPUT="dist/pressbin-$OS-$ARCH"

  [ "$OS" = "windows" ] && OUTPUT="$OUTPUT.exe"

  echo "Building $OS/$ARCH..."
  CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH \
    go build -ldflags="$LDFLAGS" -o "$OUTPUT" .
done

echo "Done. Binaries in ./dist/"
