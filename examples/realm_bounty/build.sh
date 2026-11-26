#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SDK_DIR="$(cd "$DIR/../.." && pwd)"
STAGE="$DIR/dist/stage"

rm -rf "$STAGE"
mkdir -p "$STAGE/assets"

echo "[realm_bounty] Building server_plugin.wasm..."
(
  cd "$DIR"
  GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -buildmode=c-shared -trimpath -buildvcs=false -ldflags="-s -w" -o "$DIR/dist/server_plugin.wasm" ./server
)

echo "[realm_bounty] Building client.wasm..."
(
  cd "$DIR"
  GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -buildmode=c-shared -trimpath -buildvcs=false -ldflags="-s -w" -o "$STAGE/client.wasm" ./client
)
cp "$DIR/manifest.json" "$STAGE/manifest.json"
if [[ -f "$DIR/icon.png" ]]; then
  cp "$DIR/icon.png" "$STAGE/icon.png"
fi
cp -r "$DIR/assets/." "$STAGE/assets/"

echo "[realm_bounty] Packing client_plugin.gtpack..."
(
  cd "$SDK_DIR"
  go run ./cmd/pack -dir "$STAGE" -out "$DIR/dist/client_plugin.gtpack"
)
rm -rf "$STAGE"

echo "[realm_bounty] Done:"
ls -lh "$DIR/dist/server_plugin.wasm" "$DIR/dist/client_plugin.gtpack"
