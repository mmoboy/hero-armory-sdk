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
GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -buildmode=c-shared -trimpath -buildvcs=false -ldflags="-s -w" -o dist/server_plugin.wasm ./server
GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -buildmode=c-shared -trimpath -buildvcs=false -ldflags="-s -w" -o dist/stage/client.wasm ./client
cp manifest.json dist/stage/manifest.json
cp icon.png dist/stage/icon.png  # root plugin icon (PNG, max 256x256 px, <= 128 KiB)
cp -r assets dist/stage/assets   # if present
go run github.com/mmoboy/hero-armory-sdk/cmd/pack@v1 -dir dist/stage -out dist/client_plugin.gtpack
```


## Showcase Screenshots (Source-Verified Releases)

Add up to 6 `.png` or `.jpg`/`.jpeg` screenshots (max 4 MiB each) under a top-level
`screenshots/` directory in your public GitHub plugin repository (see
`examples/realm_bounty/screenshots/` and `hero-armory-plugin-example/screenshots/`).

- `screenshots/` is intentionally **excluded** from `client_plugin.gtpack` so realm join downloads stay fast.
- When publishing a release tag at `/plugins`, the Hero Armory registry verifies your Sigstore build attestation commit SHA and indexes `screenshots/` directly from that immutable commit on GitHub for display on the website and in the desktop client.

## Realm-Authoritative Character Creation & Vanilla Presets

Hero Armory separates client look presets from Realm-authoritative character creation rules:

- **Vanilla (`CharCreateModeVanilla`)**: Default realms render the built-in 3D Hero customizer and client-local **Saved Looks** presets (purely cosmetic body/face/eye/hair settings + preset name, reusable across Vanilla realms).
- **Vanilla Extended (`CharCreateModeVanillaExtended`)**: Keep the built-in customizer and Saved Looks bar while offering Realm starter classes (such as `vanilla.SanctuaryStarterClasses()`), custom validation, starter gear/bag/gold/base stats, and `{hero_id}` KV initialization via `sdk.CharacterCreatePlugin`.
- **Custom Plugin Screen (`CharCreateModePluginCustom`)**: Replace the creation controls via `sdk.CharacterCreateClientPlugin.OnEnterCharacterCreate(ctx)` using `sdk.SetOverlay`, drive the live 3D Sanctuary pedestal via `sdk.SetCharacterCreatePreview`, and finalize via `sdk.SubmitCharacterCreate`.

## API Level 6

`sdk.APILevel` is 6 (additive on top of `SDKVersion = 1`):
- **Unified Action Bar (`HotbarSlotSpec`)**: additive `Group` (`"spell"|"item"|"utility"|"end_turn"`), `KeyHint`, `Badge`, `Charges`, `AccentColor`, `Selected` rendered directly inside the client's unified bottom action bar.
- **Flexible World Map SDK (`RoleServer`)**:
  - `SetMapConfig(MapConfigSpec)` (custom title, subtitle, bg/grid palette, coordinate labels, fast-travel toggle).
  - `SetMapRegion(MapRegionSpec)` / `RemoveMapRegion(id)` (up to `MaxMapRegions = 16` shaded territory regions with danger level & labels).
  - `SetMapRoute(MapRouteSpec)` / `RemoveMapRoute(id)` (up to `MaxMapRoutes = 32` route connectors: `"road"`, `"portal"`, `"leyline"`, `"dashed"`).
  - Enriched `MapMarkerSpec` (`Category`, `Tooltip`, `Radius`, `Clickable`).
  - Optional `MapInputPlugin` (`OnMapRoomClick`, `OnMapMarkerClick`).
- **Tactical Battle HUD**: `tactical.Engine` streams `BattleHUDSnapshot` on `tactical.DefaultHUDChannel` (`5050`); client plugins mount/clear the top-center `tactical_hud` overlay in 1 line via `tactical.HandleHUDPacket(channel, data)`.

## API Level 5

`sdk.APILevel` is 5 (additive on top of `SDKVersion = 1`):
- **Generic Combat & Encounter Primitives (`RoleServer`)**:
  - Dynamic 3D actors with smooth interpolation, HP/shield bars, status badges, and click hit-testing: `SpawnActor`, `UpdateActor` (`ActorPatch` + `PtrFloat32`/`PtrInt`/`PtrString`/`PtrBool`), `RemoveActor` (max `MaxDynamicActors = 128`, `MaxDynamicActorsPerRoom = 32`).
  - One-shot 3D VFX & rising floating text: `PlayVFX` (`slash`, `bolt`, `nova`, `heal`, `shield`, `buff`, `debuff`, `ring`, `beam`), `ShowFloatingText`.
  - Room-wide or hero-private 3D ground markers/telegraphs: `SetGroundMarker` (`circle`, `ring`, `tile`, `line`), `RemoveGroundMarker`, `ClearGroundMarkers` (max `MaxGroundMarkers = 64` per room).
  - Hero combat HUD & control: `SetHeroHotbar` (slots 1..10, slot 10 = key `0`, empty slice hides), `SetHeroPose`, `SetHeroControlLock`.
  - Optional `CombatInputPlugin` callback interface: `OnActorClick`, `OnGroundClick`, `OnHotbarPress`.
  - Declarative opt-out: `PluginManifest.DisableLegacyCombat` (`"disable_legacy_combat": true`) bypasses legacy overworld monster combat so the plugin owns 100% of combat rules.
- **Turnkey Tactical Combat Package (`github.com/mmoboy/hero-armory-sdk/tactical`)**:
  - Ready-to-use turn-based tactical combat engine built purely on API Level 5 primitives.
  - 5 difficulty presets (`TierStory`, `TierNormal`, `TierVeteran`, `TierHeroic`, `TierMythic`) + `MobBalance` knobs (`HPMult`, `DamageMult`, `ShieldPct`, `APBonus`, `MPBonus`, `CritChance`, `CritMult`, `ArmorPct`, `RewardMult`, `AIProfile` `"basic"|"tactical"|"ruthless"`).
  - 8 default player spells (`heavy_strike`, `arcane_bolt`, `starlight_nova`, `ember_lance`, `frost_bind`, `aegis_ward`, `mending_light`, `gale_rush`) + slot 10 `End Turn`, status effects (`burn`, `poison`, `chill`, `stun`, `aegis`), and 8 distinct monster templates across combat roles.

## API Level 4

`sdk.APILevel` is 4 (additive on top of `SDKVersion = 1`):
- **Per-room atmosphere & live overrides**: `RoomSpec.Atmosphere` (`RoomAtmosphereSpec`: sky zenith/horizon, fog color/density, directional sun, ambient tint, weather, extra lights/emitters) plus `SetRoomAtmosphere` / `ClearRoomAtmosphere` for live event atmosphere shifts.
- **Animated liquid planes, local point lights & 3D particle emitters**: `RoomSpec.Liquids` (`water`, `lava`, `aether`, `acid`, `mist`), `RoomSpec.Lights` (pulsing local point lights), and `RoomSpec.Emitters` (`embers`, `fireflies`, `mist`, `sparkles`, `ash`, `snow`, `rain`, `portal_swirl`, `void`).
- **Walk-in trigger zones & custom-mesh interactables**: `RoomSpec.Zones` (`circle` / `rect` with optional glowing ground ring) + optional `ZoneObserverPlugin` (`OnZoneEnter`, `OnZoneExit`), and `ModelID`/`ModelPath`/`Scale`/`Yaw` on `RoomInteractableSpec`.
- **World-map styling & dynamic markers**: `RoomSpec.MapColor` / `RoomSpec.MapIcon` and `SetMapMarker` / `RemoveMapMarker` for pulsing live-event beacons.

Earlier API Level 3 & 2 additions remain unchanged: `GetHeroInventory`, `GetHeroItemCount`, `TakeItemFromHero`, `TeleportHero`, multiple named overlays (`ClearOverlayID`, `MaxOverlays = 4`, `"top_center"`/`"bottom_center"` anchors), `SpawnProp`/`RemoveProp`, `RemoveMonsterSpawn`, `GiveGold`/`GiveXP`, `GetHero`, plugin chat, `PatchOverlay`, `ValidateOverlay`, and `sdktest`. All changes are additive; v1-v3 plugins run unchanged.

## Local Sandbox Mode (One-Command Edit-Test Loop)

Launch any local plugin directory (containing `manifest.json`) directly in an offline loopback sandbox realm with zero login or lobby clicks:

```bash
hero-armory --sandbox ./my-plugin
hero-armory --sandbox ./my-plugin --sandbox-room 1,0
hero-armory --sandbox ./my-plugin --sandbox-no-build
hero-armory --sandbox-join --sandbox-hero "Tester2"
```

- `--sandbox <plugin-dir>`: opens a dedicated **Sandbox Boot Screen** (no login form), compiles `server_plugin.wasm` and `client_plugin.gtpack`, starts a local loopback (`127.0.0.1`) sandbox realm server, and connects straight into the 3D world with an offline dev hero.
- `--sandbox-no-build`: skips rebuilding `dist/` and launches directly with existing `dist/server_plugin.wasm` + `dist/client_plugin.gtpack`.
- `--sandbox-room <x,y>`: optional starting room coordinates (for example `1,0`).
- `--sandbox-join`: connects a 2nd (or 3rd) client straight into the currently running local loopback sandbox server for local multiplayer plugin testing without spawning a second server or rebuilding.
- `--sandbox-hero <name>`: optional dev hero display name (defaults to an auto-generated unique name when using `--sandbox-join` so multiple clients never collide).

While connected to a local sandbox realm (from `--sandbox`, `--sandbox-join`, or Plugin Studio's **Build & Run Sandbox**), an in-game **Sandbox Dev Bar** appears inside the world viewport:
- **Reload (`F5`)**: (host client) rebuilds WASM + `.gtpack`, restarts the loopback realm, and reconnects back into your current room (showing build errors in the log drawer if compilation fails).
- **Room Jump**: pick or cycle through any registered room in the realm to jump there in one click.
- **Logs**: toggleable live drawer displaying server and client plugin `sdk.Log(...)` output and build/WASM diagnostics in real time.

## Testing Plugins

```go
import "github.com/mmoboy/hero-armory-sdk/sdktest"

func TestWelcome(t *testing.T) {
	h := sdktest.Install(t)
	h.SetOnline(sdk.HeroSnapshot{ID: "h1", Name: "Ayla"})
	p := newServer()
	p.Init()
	p.OnPlayerJoin(h.Hero("h1"))
	if len(h.NoticesFor("h1")) == 0 {
		t.Fatal("expected a welcome notice")
	}
}
```

## Examples

- `examples/minimal_realm`: Bare-bones empty-shell world example (`server_plugin.wasm` only).
- `examples/standard_realm`: Full classic Sanctuary world rebuilt via `hero-armory-sdk/vanilla`.
- `examples/realm_bounty`: HUD actions, floating board overlay, Realm KV/Prefs, custom 3D mesh & boss monster.
- `examples/custom_char_create`: Realm-authoritative starter classes, live 3D creation preview, custom validation, and submit flow.
