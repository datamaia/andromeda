#!/usr/bin/env bash
# Andromeda installer.
#
#   curl -fsSL https://raw.githubusercontent.com/datamaia/andromeda/main/scripts/install.sh | bash
#
# Downloads the release binary for your OS/architecture from GitHub Releases and installs the
# `andromeda` executable onto your PATH. Configuration via environment variables:
#
#   ANDROMEDA_VERSION      version tag to install (default: latest)
#   ANDROMEDA_INSTALL_DIR  install directory (default: /usr/local/bin, else ~/.local/bin)
#   GITHUB_TOKEN           token for a PRIVATE repo's releases (otherwise anonymous)
#
set -euo pipefail

REPO="datamaia/andromeda"
BIN="andromeda"

info()  { printf '\033[38;5;99m▸\033[0m %s\n' "$*"; }
err()   { printf '\033[38;5;203m✗ %s\033[0m\n' "$*" >&2; exit 1; }

# --- detect platform -------------------------------------------------------
os="$(uname -s)"; arch="$(uname -m)"
case "$os" in
  Darwin) os="darwin" ;;
  Linux)  os="linux" ;;
  *) err "unsupported OS: $os (macOS and Linux are supported; Windows install is separate)";;
esac
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) err "unsupported architecture: $arch";;
esac
info "platform: ${os}/${arch}"

# --- auth header for private repos ----------------------------------------
auth_args=()
if [ -n "${GITHUB_TOKEN:-}" ]; then
  auth_args=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
fi

api="https://api.github.com/repos/${REPO}/releases"

# --- resolve version -------------------------------------------------------
version="${ANDROMEDA_VERSION:-latest}"
if [ "$version" = "latest" ]; then
  info "resolving latest release…"
  version="$(curl -fsSL "${auth_args[@]}" "${api}/latest" \
    | grep -m1 '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')" \
    || err "could not resolve the latest release (is a release published? is the repo private?)"
  [ -n "$version" ] || err "no releases found for ${REPO}"
fi
info "version: ${version}"

# --- download + extract ----------------------------------------------------
ver_nov="${version#v}"
asset="${BIN}_${ver_nov}_${os}_${arch}.tar.gz"
url="https://github.com/${REPO}/releases/download/${version}/${asset}"

tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
info "downloading ${asset}…"
curl -fsSL "${auth_args[@]}" -o "${tmp}/${asset}" "$url" \
  || err "download failed: ${url}"
tar -xzf "${tmp}/${asset}" -C "$tmp" || err "extraction failed"
[ -f "${tmp}/${BIN}" ] || err "archive did not contain the ${BIN} binary"

# --- install onto PATH -----------------------------------------------------
dir="${ANDROMEDA_INSTALL_DIR:-}"
if [ -z "$dir" ]; then
  if [ -w /usr/local/bin ] 2>/dev/null; then dir="/usr/local/bin"; else dir="${HOME}/.local/bin"; fi
fi
mkdir -p "$dir"
install -m 0755 "${tmp}/${BIN}" "${dir}/${BIN}" 2>/dev/null \
  || { chmod 0755 "${tmp}/${BIN}"; mv "${tmp}/${BIN}" "${dir}/${BIN}"; }
info "installed ${BIN} → ${dir}/${BIN}"

# --- PATH hint -------------------------------------------------------------
case ":${PATH}:" in
  *":${dir}:"*) ;;
  *) info "add ${dir} to your PATH, e.g.:  echo 'export PATH=\"${dir}:\$PATH\"' >> ~/.zshrc" ;;
esac

if command -v "$BIN" >/dev/null 2>&1; then
  info "done: $("$BIN" version 2>/dev/null || echo "$BIN installed")"
else
  info "done. Run: ${dir}/${BIN} version"
fi
