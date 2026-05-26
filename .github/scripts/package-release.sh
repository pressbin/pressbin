#!/usr/bin/env bash
# Build a platform release folder: binary, config.yml.example, README, GitHub sync workflow.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

VERSION="${1:?version required (e.g. v1.0.0)}"
GOOS="${2:?GOOS required}"
GOARCH="${3:?GOARCH required}"
DIST_ROOT="${4:-$ROOT/dist}"

BUNDLE_NAME="pressbin-${VERSION}-${GOOS}-${GOARCH}"
STAGING="$DIST_ROOT/$BUNDLE_NAME"
rm -rf "$STAGING"
mkdir -p "$STAGING/.github/workflows" "$STAGING/.github/scripts"

BINARY="$STAGING/pressbin"
if [ "$GOOS" = "windows" ]; then
  BINARY="$STAGING/pressbin.exe"
fi

LDFLAGS="-s -w -X main.Version=$VERSION"
echo "Building pressbin $VERSION ($GOOS/$GOARCH)..."
CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" \
  go build -ldflags="$LDFLAGS" -o "$BINARY" .

cp config.yml.example release/README.md "$STAGING/"
cp templates/consumer/.github/workflows/sync.yml "$STAGING/.github/workflows/"
cp templates/consumer/.github/scripts/push.py templates/consumer/.github/scripts/delete.py "$STAGING/.github/scripts/"

ARCHIVE="$DIST_ROOT/${BUNDLE_NAME}.tar.gz"
mkdir -p "$DIST_ROOT"
tar -czf "$ARCHIVE" -C "$DIST_ROOT" "$BUNDLE_NAME"
echo "Wrote $ARCHIVE"
