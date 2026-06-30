#!/usr/bin/env bash
# Download all vendored front-end assets.
# Run this once after cloning.
#
# Requirements: curl, tar (standard on Linux/macOS)
set -euo pipefail

KATEX_VERSION="0.16.11"
HTMX_VERSION="2.0.4"
MARKED_VERSION="14"
BULMA_VERSION="1.0.4"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
VENDOR="$ROOT/static/vendor"

mkdir -p "$VENDOR/katex/fonts" "$VENDOR/fonts"

echo "→ Downloading HTMX $HTMX_VERSION..."
curl -sL -o "$VENDOR/htmx.min.js" \
  "https://unpkg.com/htmx.org@${HTMX_VERSION}/dist/htmx.min.js"

echo "→ Downloading marked $MARKED_VERSION..."
curl -sL -o "$VENDOR/marked.min.js" \
  "https://cdn.jsdelivr.net/npm/marked@${MARKED_VERSION}/marked.min.js"

echo "→ Downloading Bulma $BULMA_VERSION..."
curl -sL -o "$VENDOR/bulma.min.css" \
  "https://cdn.jsdelivr.net/npm/bulma@${BULMA_VERSION}/css/bulma.min.css"

echo "→ Downloading KaTeX $KATEX_VERSION..."
TMP=$(mktemp -d)
curl -sL "https://registry.npmjs.org/katex/-/katex-${KATEX_VERSION}.tgz" \
  | tar -xz -C "$TMP"
cp "$TMP/package/dist/katex.min.css"              "$VENDOR/katex/"
cp "$TMP/package/dist/katex.min.js"               "$VENDOR/katex/"
cp "$TMP/package/dist/contrib/auto-render.min.js" "$VENDOR/katex/"
cp "$TMP/package/dist/fonts/"*.woff2              "$VENDOR/katex/fonts/"
rm -rf "$TMP"

echo "→ Downloading Inter font (latin + extended subsets)..."
declare -A INTER_FILES=(
  [inter-cyrillic-ext]=UcC73FwrK3iLTeHuS_nVMrMxCp50SjIa2JL7SUc
  [inter-cyrillic]=UcC73FwrK3iLTeHuS_nVMrMxCp50SjIa0ZL7SUc
  [inter-greek-ext]=UcC73FwrK3iLTeHuS_nVMrMxCp50SjIa2ZL7SUc
  [inter-greek]=UcC73FwrK3iLTeHuS_nVMrMxCp50SjIa1pL7SUc
  [inter-vietnamese]=UcC73FwrK3iLTeHuS_nVMrMxCp50SjIa2pL7SUc
  [inter-latin-ext]=UcC73FwrK3iLTeHuS_nVMrMxCp50SjIa25L7SUc
  [inter-latin]=UcC73FwrK3iLTeHuS_nVMrMxCp50SjIa1ZL7
)
for name in "${!INTER_FILES[@]}"; do
  hash="${INTER_FILES[$name]}"
  curl -sL -o "$VENDOR/fonts/${name}.woff2" \
    "https://fonts.gstatic.com/s/inter/v20/${hash}.woff2" &
done
wait

echo "✓ All assets ready."
