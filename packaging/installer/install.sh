#!/usr/bin/env sh
# Andromeda shell installer (Volume 14). Downloads the release archive for the host platform
# from GitHub Releases, verifies its checksum, and installs the binary. POSIX sh; macOS/Linux.
#
#   curl -sSfL https://raw.githubusercontent.com/datamaia/andromeda/main/packaging/installer/install.sh | sh
#
# Environment:
#   ANDROMEDA_VERSION   release tag to install (default: latest)
#   ANDROMEDA_PREFIX    install prefix (default: /usr/local/bin, or ~/.local/bin if unwritable)
set -eu

REPO="datamaia/andromeda"
VERSION="${ANDROMEDA_VERSION:-latest}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) echo "unsupported architecture: $arch" >&2; exit 1 ;;
esac
case "$os" in
  darwin|linux) ;;
  *) echo "unsupported OS: $os (Windows is a later phase)" >&2; exit 1 ;;
esac

if [ "$VERSION" = "latest" ]; then
  VERSION=$(curl -sSfL "https://api.github.com/repos/$REPO/releases/latest" \
    | grep '"tag_name"' | head -1 | cut -d'"' -f4)
fi
[ -n "$VERSION" ] || { echo "could not determine version" >&2; exit 1; }
num="${VERSION#v}"

base="https://github.com/$REPO/releases/download/$VERSION"
archive="andromeda_${num}_${os}_${arch}.tar.gz"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "Downloading $archive ..."
curl -sSfL "$base/$archive" -o "$tmp/$archive"
curl -sSfL "$base/checksums.txt" -o "$tmp/checksums.txt"

echo "Verifying checksum ..."
( cd "$tmp" && grep " $archive\$" checksums.txt | { command -v sha256sum >/dev/null 2>&1 \
    && sha256sum -c - || shasum -a 256 -c -; } )

tar -xzf "$tmp/$archive" -C "$tmp"

prefix="${ANDROMEDA_PREFIX:-/usr/local/bin}"
if [ ! -w "$(dirname "$prefix")" ] && [ ! -w "$prefix" ]; then
  prefix="$HOME/.local/bin"
fi
mkdir -p "$prefix"
install -m 0755 "$tmp/andromeda" "$prefix/andromeda"
echo "Installed andromeda $VERSION to $prefix/andromeda"
"$prefix/andromeda" version || true
