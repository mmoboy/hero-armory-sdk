#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SDK_DIR="$(cd "$DIR/../.." && pwd)"
STAGE="$DIR/dist/stage"

rm -rf "$STAGE"
mkdir -p "$STAGE/assets"

echo "[standard_realm] Building server_plugin.wasm..."
(
  cd "$DIR"
  GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -buildmode=c-shared -trimpath -buildvcs=false -ldflags="-s -w" -o "$DIR/dist/server_plugin.wasm" ./server
)

echo "[standard_realm] Building client.wasm..."
(
  cd "$DIR"
  GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -buildmode=c-shared -trimpath -buildvcs=false -ldflags="-s -w" -o "$STAGE/client.wasm" ./client
)
cp "$DIR/manifest.json" "$STAGE/manifest.json"
if [ -d "$DIR/assets" ]; then
  cp -r "$DIR/assets/." "$STAGE/assets/"
fi

echo "[standard_realm] Packing client_plugin.gtpack..."
(
  cd "$SDK_DIR"
  go run ./cmd/pack -dir "$STAGE" -out "$DIR/dist/client_plugin.gtpack"
)
rm -rf "$STAGE"

echo "[standard_realm] Done:"
ls -lh "$DIR/dist/server_plugin.wasm" "$DIR/dist/client_plugin.gtpack"
