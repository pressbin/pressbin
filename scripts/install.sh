#!/usr/bin/env bash
# Install Pressbin to ~/.pressbin (no root). Downloads from github.com/pressbin/pressbin releases.
set -euo pipefail

REPO="pressbin/pressbin"
INSTALL_DIR="${PRESSBIN_HOME:-$HOME/.pressbin}"
SITE_URL="${PRESSBIN_SITE_URL:-}"
SITE_TITLE="${PRESSBIN_SITE_TITLE:-My Blog}"

usage() {
  cat <<EOF
Usage: install.sh --site-url URL [options]

  --site-url URL     Public blog URL (required, or set PRESSBIN_SITE_URL)
  --site-title TITLE Site title (default: My Blog)
  --home DIR         Install directory (default: ~/.pressbin)

Environment:
  PRESSBIN_HOME      Install directory
  PRESSBIN_SITE_URL  Public site URL

Example:
  curl -fsSL https://raw.githubusercontent.com/pressbin/pressbin/main/scripts/install.sh | bash -s -- --site-url https://blog.example.com
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
    --site-url) SITE_URL="$2"; shift 2 ;;
    --site-title) SITE_TITLE="$2"; shift 2 ;;
    --home) INSTALL_DIR="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage; exit 1 ;;
  esac
done

if [ -z "$SITE_URL" ]; then
  echo "error: --site-url or PRESSBIN_SITE_URL is required" >&2
  usage
  exit 1
fi

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "unsupported architecture: $ARCH" >&2; exit 1 ;;
esac
case "$OS" in
  linux|darwin) ;;
  *) echo "unsupported OS: $OS" >&2; exit 1 ;;
esac

ASSET="pressbin-${OS}-${ARCH}"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "Fetching latest ${REPO} release..."
TAG="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -1)"
if [ -z "$TAG" ]; then
  echo "error: could not resolve latest release tag" >&2
  exit 1
fi
BASE="https://github.com/${REPO}/releases/download/${TAG}"

echo "Downloading ${ASSET} (${TAG})..."
curl -fsSL "${BASE}/${ASSET}" -o "${TMP}/${ASSET}"
curl -fsSL "${BASE}/checksums.txt" -o "${TMP}/checksums.txt"

(
  cd "$TMP"
  if command -v sha256sum >/dev/null 2>&1; then
    grep " ${ASSET}\$" checksums.txt | sha256sum -c -
  else
    grep " ${ASSET}\$" checksums.txt | shasum -a 256 -c -
  fi
)

mkdir -p "${INSTALL_DIR}/bin"
install -m 755 "${TMP}/${ASSET}" "${INSTALL_DIR}/bin/pressbin"

echo "Running setup..."
"${INSTALL_DIR}/bin/pressbin" setup \
  --home "${INSTALL_DIR}" \
  --site-url "${SITE_URL}" \
  --site-title "${SITE_TITLE}"

echo ""
echo "Add Pressbin to your PATH:"
echo "  export PATH=\"${INSTALL_DIR}/bin:\$PATH\""
