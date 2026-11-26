#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SDK_DIR="$(cd "$DIR/../.." && pwd)"
STAGE="$DIR/dist/stage"

rm -rf "$STAGE"
mkdir -p "$STAGE/assets"

echo "[custom_char_create] Building server_plugin.wasm..."
(
  cd "$DIR"
  GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -buildmode=c-shared -trimpath -buildvcs=false -ldflags="-s -w" -o "$DIR/dist/server_plugin.wasm" ./server
)

echo "[custom_char_create] Building client.wasm..."
(
  cd "$DIR"
  GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -buildmode=c-shared -trimpath -buildvcs=false -ldflags="-s -w" -o "$STAGE/client.wasm" ./client
)
cp "$DIR/manifest.json" "$STAGE/manifest.json"
if [ -d "$DIR/assets" ]; then
  cp -r "$DIR/assets/." "$STAGE/assets/"
fi

echo "[custom_char_create] Packing client_plugin.gtpack..."
(
  cd "$SDK_DIR"
  go run ./cmd/pack -dir "$STAGE" -out "$DIR/dist/client_plugin.gtpack"
)
rm -rf "$STAGE"

echo "[custom_char_create] Done:"
ls -lh "$DIR/dist/server_plugin.wasm" "$DIR/dist/client_plugin.gtpack"
