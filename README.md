# hero-armory-sdk

Go WebAssembly (`wasip1/wasm`) SDK for authoring Hero Armory realm and client
plugins (`server_plugin.wasm` + `client_plugin.gtpack`).

## Quick Start

Use the standalone template repository as a starting point:
- Template repository: `https://github.com/mmoboy/hero-armory-plugin-example`
- Web Portal & Plugin Library: `https://kmmo.enlumis.com/plugins`

Or install directly in a new Go module:

```bash
go get github.com/mmoboy/hero-armory-sdk@latest
```

Plugin layout expected by the official GitHub Actions release workflow:

```text
my-plugin/
  go.mod
  manifest.json
  server/main.go
  client/main.go
  assets/               # optional PNG icons, GLB models, WAV/OGG audio
  .github/workflows/
    release.yml
```

## Source-Verified GitHub Releases (Free for Public Repos)

Add `.github/workflows/release.yml` to your public plugin repository:

```yaml
name: Release Plugin

on:
  push:
    tags:
      - "v*"

jobs:
  release:
    permissions:
      id-token: write
      attestations: write
      contents: write
    uses: mmoboy/hero-armory-sdk/.github/workflows/build-plugin.yml@v1
```

When you push a git tag (for example `v1.0.0`, matching `"version": "1.0.0"`
in `manifest.json`), GitHub Actions builds `server_plugin.wasm` and
`client_plugin.gtpack`, generates cryptographic Sigstore build provenance
attestations, and attaches all release files to the GitHub Release. You can
then submit the repository URL and tag on `https://kmmo.enlumis.com/plugins`
to publish it to the community catalog with a **Source-Verified** badge.

## Local Development Build

```bash
GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -buildmode=c-shared -ldflags="-s -w" -o dist/server_plugin.wasm ./server
GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -buildmode=c-shared -ldflags="-s -w" -o dist/stage/client.wasm ./client
cp manifest.json dist/stage/manifest.json
cp -r assets dist/stage/assets   # if present
go run github.com/mmoboy/hero-armory-sdk/cmd/pack@v1 -dir dist/stage -out dist/client_plugin.gtpack
```


## Showcase Screenshots (Source-Verified Releases)

Add up to 6 `.png` or `.jpg`/`.jpeg` screenshots (max 4 MiB each) under a top-level
`screenshots/` directory in your public GitHub plugin repository (see
`examples/realm_bounty/screenshots/` and `hero-armory-plugin-example/screenshots/`).

- `screenshots/` is intentionally **excluded** from `client_plugin.gtpack` so realm join downloads stay fast.
- When publishing a release tag at `/plugins`, the Hero Armory registry verifies your Sigstore build attestation commit SHA and indexes `screenshots/` directly from that immutable commit on GitHub for display on the website and in the desktop client.
