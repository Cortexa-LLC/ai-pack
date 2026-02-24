#!/usr/bin/env bash
set -euo pipefail

VERSION=${1:-0.8.2}
PLATFORM=${2:-$(go env GOOS)-$(go env GOARCH)}
DEST="lib/kuzu/$PLATFORM"

case "$PLATFORM" in
  linux-amd64)  KUZU_ARCH="linux-x86_64"  ;;
  linux-arm64)  KUZU_ARCH="linux-aarch64" ;;
  darwin-arm64) KUZU_ARCH="macos-arm64"   ;;
  darwin-amd64) KUZU_ARCH="macos-x86_64"  ;;
  *) echo "Unsupported platform: $PLATFORM"; exit 1 ;;
esac

URL="https://github.com/kuzudb/kuzu/releases/download/v${VERSION}/kuzu_${KUZU_ARCH}.tar.gz"

echo "Downloading Kuzu $VERSION for $PLATFORM..."
mkdir -p "$DEST"
curl -sL "$URL" | tar -xz -C "$DEST" --strip-components=1
echo "✓ Kuzu $VERSION → $DEST"
