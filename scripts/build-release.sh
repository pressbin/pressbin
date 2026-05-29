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

  echo "Building pressbin $OS/$ARCH..."
  OUT="dist/pressbin-$OS-$ARCH"
  [ "$OS" = "windows" ] && OUT="$OUT.exe"
  CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH \
    go build -ldflags="$LDFLAGS" -o "$OUT" .

  echo "Building pressbin-sync $OS/$ARCH..."
  OUT_SYNC="dist/pressbin-sync-$OS-$ARCH"
  [ "$OS" = "windows" ] && OUT_SYNC="$OUT_SYNC.exe"
  CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH \
    go build -ldflags="$LDFLAGS" -o "$OUT_SYNC" ./cmd/pressbin-sync
done

echo "Done. Binaries in ./dist/"
