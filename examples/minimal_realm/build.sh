#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
mkdir -p "$DIR/dist"

cd "$DIR"

echo "[minimal_realm] Building server_plugin.wasm..."
GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -buildmode=c-shared -trimpath -buildvcs=false -ldflags="-s -w" -o "$DIR/dist/server_plugin.wasm" ./server

echo "[minimal_realm] Done (server-only plugin, no client_plugin.gtpack needed):"
ls -lh "$DIR/dist/server_plugin.wasm"
