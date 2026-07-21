#!/usr/bin/env bash
# Authenticode-sign a Windows binary during `goreleaser release` (invoked by the binary_signs
# block in .goreleaser.yaml). This is credential-gated and a strict no-op unless a signing cert is
# provisioned — mirroring the macOS `notarize` block — so releases keep succeeding with an unsigned
# binary until then. A signed andromeda.exe barely trips Microsoft Defender ASR / SmartScreen (see
# the README "Windows Defender / SmartScreen" section for why the unsigned build is flagged).
#
# To activate, set these in the release environment (GitHub Actions secrets) and install
# osslsigncode on the runner (`apt-get install -y osslsigncode`):
#   WINDOWS_SIGN_CERT      base64-encoded PKCS#12 (.pfx) certificate
#   WINDOWS_SIGN_PASSWORD  its password
# For Azure Trusted Signing, swap the osslsigncode call below for `signtool` + the Azure dlib.
set -euo pipefail

artifact="${1:-}"
[ -n "$artifact" ] || { echo "sign-windows: no artifact given" >&2; exit 1; }

# Only Windows binaries carry an Authenticode signature; skip macOS/Linux binaries.
case "$artifact" in
  *.exe) ;;
  *) exit 0 ;;
esac

# No cert provisioned -> no-op: leave the binary unsigned and let the release proceed (as today).
if [ -z "${WINDOWS_SIGN_CERT:-}" ]; then
  echo "sign-windows: WINDOWS_SIGN_CERT not set — leaving $(basename "$artifact") unsigned (no-op)."
  exit 0
fi

command -v osslsigncode >/dev/null 2>&1 || {
  echo "sign-windows: osslsigncode not found on PATH — install it on the runner to sign." >&2
  exit 1
}

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
pfx="$tmp/cert.pfx"
printf '%s' "$WINDOWS_SIGN_CERT" | base64 -d > "$pfx"

echo "sign-windows: signing $(basename "$artifact") with osslsigncode (RFC3161 timestamp)…"
osslsigncode sign \
  -pkcs12 "$pfx" -pass "${WINDOWS_SIGN_PASSWORD:-}" \
  -n "Andromeda CLI" -i "https://andromedacli.com" \
  -ts "http://timestamp.digicert.com" \
  -h sha256 \
  -in "$artifact" -out "$artifact.signed"
mv -f "$artifact.signed" "$artifact"
echo "sign-windows: signed $(basename "$artifact")."
