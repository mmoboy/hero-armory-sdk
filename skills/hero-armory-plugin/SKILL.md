---
name: hero-armory-plugin
description: End-to-end Hero Armory plugin builder using hero-armory-sdk (Go compiled to wasip1/wasm). Use whenever a user asks to create, scaffold, modify, compile, pack, or deploy a Hero Armory plugin (server_plugin.wasm, client_plugin.gtpack, manifest.json, custom rooms, NPCs, starter config, skills, jobs, monsters, items, gather nodes, recipes, Codex entries, procedural 3D meshes, HUD actions, UI overlays, or plugin KV/Pref storage).
---

# Hero Armory Plugin Development Skill (`hero-armory-sdk` v1)

Use this skill whenever the user asks to create, scaffold, modify, debug, compile, or deploy a plugin for **Hero Armory**.

---

## 0. End-to-End Execution Checklist (Do All Steps Automatically)

When the user asks you to create a plugin (for example: `"Create a Hero Armory plugin that..."`), execute **every step end-to-end** without stopping early:

1. **Locate `hero-armory-sdk`**: For standalone GitHub repositories, use `go get github.com/mmoboy/hero-armory-sdk@latest` (and `.github/workflows/release.yml` via `github.com/mmoboy/hero-armory-plugin-example`). For local offline SDK checkouts, use a temporary `replace` directive or a `go.work` file pointing at your local `hero-armory-sdk` clone.
2. **Scaffold the plugin directory**: Create `go.mod`, `manifest.json`, `icon.png` (root square PNG icon, max **256x256** px and <= 128 KiB, referenced as `"icon": "icon.png"` in `manifest.json`), `proto/messages.go`, `server/main.go`, `client/main.go`, `assets/` (if needed), `screenshots/` (optional showcase `.png`/`.jpg` images for verified GitHub releases), and `build.sh` (`chmod +x`).
3. **Provide root `icon.png` + safe HUD icons**: Every plugin must include a root `icon.png` (valid PNG, max 256x256 px, <= 128 KiB) copied into `$STAGE/icon.png` before packing `.gtpack` so it renders in the Plugins Library and Host Server configuration screens. For in-game `HUDActionSpec.Icon`, use a built-in atlas icon name (such as `"star"`, `"shield"`, `"sword"`, `"book"`, `"compass"`).
4. **Implement `proto/`, `server/`, and `client/`** following all rules in Sections 1-5 below.
5. **Compile and pack**: Run `./build.sh` to produce `dist/server_plugin.wasm` and `dist/client_plugin.gtpack`, fixing any compiler or packer errors before finishing.
6. **Verify**: Confirm both `dist/server_plugin.wasm` and `dist/client_plugin.gtpack` exist and report the exact `cmd/pack` SHA-256 output.
7. **Deploy instructions**: If the user gave a `<realm-data-dir>`, copy both files into `<realm-data-dir>/plugins/`. Otherwise print the exact copy/flag commands from Section 6.

---

## 1. Non-Negotiable Rules

1. **Strict ASCII only**: Hero Armory uses a custom Latin-only TrueType glyph atlas. Never put emojis or non-ASCII Unicode (arrows, bullets, curly quotes, em-dashes) in UI labels, item/monster names, notices, Codex text, Go source, or `manifest.json`.
2. **Go 1.24+ pure Go to `wasip1/wasm`**:
   - Compile BOTH `server` and `client` with:
     `GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -buildmode=c-shared -trimpath -buildvcs=false -ldflags="-s -w" ...`
   - Never omit `-buildmode=c-shared` (host ABI exports depend on it).
   - Register the plugin in `func init()` via `sdk.RegisterServer(...)` or `sdk.RegisterClient(...)`, and keep `func main() {}` empty.
3. **Strict VM Role Separation**:
   - `RoleServer` (`server_plugin.wasm`): Runs on the Realm Server. Can call `RegisterRoom`, `RemoveRoom`, `SetSpawnRoom`, `RegisterNPC`, `SetHeroLoadout`, `SetStarterConfig`, `RegisterPlayerSkill`, `RegisterJob`, `RegisterMonster`, `SpawnMonster`, `RemoveMonsterSpawn`, `RegisterItem`, `SpawnGatherNode`, `RegisterRecipe`, `GiveItemToHero`, `GiveGold`, `GiveXP`, `GetHero`, `ConfigureCodex`, `RegisterCodexEntry`, `SendToHero`, `SendJSONToHero`, `BroadcastPacket`, `BroadcastJSON`, `BroadcastRoomPacket`, `SendNoticeToHero`, `BroadcastNotice`, `SendChatToHero`, `BroadcastChat`, and `KV*`, plus optional lifecycle hooks `OnBeginCharacterCreate` and `OnFinalizeCharacterCreate`.
   - `RoleClient` (`client_plugin.gtpack` -> `client.wasm`): Runs on the native game client. Can call `SendToServer`, `SendJSONToServer`, `SetHUDActions`, `SetOverlay`, `PatchOverlay`, `ClearOverlay`, `SetCharacterCreatePreview`, `SubmitCharacterCreate`, `OpenCodex`, `PlayAudio`, `RegisterCustomMesh`, and `Pref*`, plus optional `OnEnterCharacterCreate`.
   - `sdk.Log(msg)` works in both roles.
   - Calling a hostcall from the wrong VM role returns `sdk.ErrWrongRole` / negative error code.
4. **Content IDs & Naming Convention**:
   - Content IDs (`ItemSpec.ID`, `MonsterSpec.ID`, `GatherNodeSpec.ID`, `RecipeSpec.ID`, `NPCSpec.ID`, `PlayerSkillSpec.ID`, `JobDefSpec.ID`, `CodexEntrySpec.ID`) must satisfy `IsValidContentID`: `2-48` lowercase ASCII chars `[a-z0-9_-]`.
   - Prefixing custom add-on items with `plug_` (e.g. `plug_astral_shard`) is recommended when extending `hero-armory-sdk/vanilla` so custom items never collide with vanilla item IDs.
5. **Plugin IDs & Storage Keys**:
   - `manifest.json` `id` and `CustomMeshSpec.ID`: 3-40 lowercase ASCII chars `[a-z0-9_-]` (`2-48` for `IsValidContentID`).
   - KV and Pref keys: allowed chars `[a-zA-Z0-9._:,/-]`, no `..`, no leading/trailing `/`.

---

## 2. Project Layout & Templates

```text
<plugin-dir>/
+-- go.mod
+-- manifest.json
+-- icon.png                # required plugin icon at project root (PNG, max 256x256 px, <= 128 KiB, packed into .gtpack)
+-- build.sh
+-- assets/                 # optional runtime files packed into client_plugin.gtpack: .png, .wav, .ogg, .glb, .json
+-- screenshots/            # optional showcase .png/.jpg images displayed from verified GitHub commits (not packed into .gtpack)
+-- proto/                  # shared uint16 channel IDs & JSON packet structs
|   +-- messages.go
+-- server/
|   +-- main.go             # embeds sdk.BaseServerPlugin
+-- client/
    +-- main.go             # embeds sdk.BaseClientPlugin
```

### `go.mod`

For a standalone public GitHub plugin repository (required for Source-Verified GitHub Actions releases):

```go
module github.com/<owner>/<my-plugin>

go 1.24

require github.com/mmoboy/hero-armory-sdk v1.0.0
```

When developing offline against a local `hero-armory-sdk` checkout (not committed to a Source-Verified GitHub release repo), temporarily point at your local SDK directory:

```go
replace github.com/mmoboy/hero-armory-sdk => ../../sdk/hero-armory-sdk
```

### `manifest.json`

```json
{
  "sdk_version": 1,
  "id": "my_plugin",
  "name": "My Realm Plugin",
  "version": "1.0.0",
  "author": "Author Name",
  "description": "Short ASCII description (max 240 chars).",
  "entry_wasm": "client.wasm",
  "icon": "icon.png",
  "permissions": ["hud_actions", "overlay_ui", "packets", "audio"],
  "required": false
}
```

### `build.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SDK_DIR="$(cd "$DIR/../../sdk/hero-armory-sdk" && pwd)"
STAGE="$DIR/dist/stage"

rm -rf "$STAGE"
mkdir -p "$STAGE/assets"

GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build \
  -buildmode=c-shared -trimpath -buildvcs=false -ldflags="-s -w" \
  -o "$DIR/dist/server_plugin.wasm" ./server

GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build \
  -buildmode=c-shared -trimpath -buildvcs=false -ldflags="-s -w" \
  -o "$STAGE/client.wasm" ./client

cp "$DIR/manifest.json" "$STAGE/manifest.json"
if [[ -f "$DIR/icon.png" ]]; then
  cp "$DIR/icon.png" "$STAGE/icon.png"
fi
if [ -d "$DIR/assets" ] && [ -n "$(ls -A "$DIR/assets" 2>/dev/null)" ]; then
  cp -r "$DIR/assets/." "$STAGE/assets/"
fi

(
  cd "$SDK_DIR"
  go run ./cmd/pack -dir "$STAGE" -out "$DIR/dist/client_plugin.gtpack"
)
rm -rf "$STAGE"
ls -lh "$DIR/dist/server_plugin.wasm" "$DIR/dist/client_plugin.gtpack"
```

---

## 3. Server Plugin API (`RoleServer`)

`hero-armory-server` starts as an empty shell (loading the embedded `standard_realm` bundle only when no external plugin is configured). Inside `Init()`, a server plugin can either define a custom world from scratch (`examples/minimal_realm`) or import `"github.com/mmoboy/hero-armory-sdk/vanilla"` and call `vanilla.RegisterAll()` + `vanilla.HandleNPCInteract(ev)` (`examples/standard_realm`) before adding custom content.

Embed `sdk.BaseServerPlugin` and implement only the hooks needed:

- `Init()` - Register rooms, NPCs, starter config, skills, jobs, monsters, items, gather nodes, recipes, Codex config/entries, and load persisted state from KV.
- `OnTick(dtMs uint32)`
- `OnPlayerJoin(hero sdk.HeroSnapshot)` / `OnPlayerLeave(heroID string)`
- `OnRoomEnter(ev sdk.RoomEvent)`
- `OnChat(ev sdk.ChatEvent)`
- `OnPacket(senderHeroID string, channel uint16, data []byte)`
- Optional `sdk.BattleObserverPlugin`: `OnBattleEnd(ev sdk.BattleEndEvent)`
- Optional `sdk.WorldObserverPlugin`: `OnGather(ev sdk.GatherEvent)` and `OnCraft(ev sdk.CraftEvent)`
- Optional `sdk.NPCPlugin`: `OnNPCInteract(ev sdk.NPCInteractEvent) (sdk.NPCDialogResult, bool)`
- Optional `sdk.InteractPlugin`: `OnInteract(ev sdk.InteractEvent) bool`
- Optional `sdk.ChatCommandPlugin` (API level 2): `OnChatCommand(ev sdk.ChatEvent) bool` runs BEFORE a local/server/commerce/party message is broadcast. Return true to consume it: nobody else sees it, the sender sees only their own line, and `OnChat` is not called. Reply with `sdk.SendChatToHero` or `sdk.SendNoticeToHero`.

`BattleEndEvent.MonsterID` is the `MonsterSpawnSpec.ID` of the engaged pack. `time.Now()` works inside the sandbox and returns the host wall clock.

### Server Hostcalls

- **Networking & Notices**:
  - `sdk.SendToHero(heroID, channel, payload)` / `sdk.SendJSONToHero(heroID, channel, v)`
  - `sdk.BroadcastPacket(channel, payload)` / `sdk.BroadcastJSON(channel, v)`
  - `sdk.BroadcastRoomPacket(roomX, roomY, channel, payload)`
  - `sdk.SendNoticeToHero(heroID, severity, title, message)` / `sdk.BroadcastNotice(severity, title, message)` (`severity`: `"info"`, `"warn"`, `"alert"`)
- **World Map, Rooms & NPCs**:
  - `sdk.RegisterRoom(sdk.RoomSpec{...})` (`TerrainOps`: `"rect"`, `"slope"`, `"void"`; `Obstacles`, `Boxes`, `Portals`, `Houses`, `Interactables`, `Props`, `SafeZone`, `HousingZone`, `LightingPreset`)
  - `sdk.RemoveRoom(x, y)` / `sdk.SetSpawnRoom(x, y, pos)`
  - `sdk.RegisterNPC(sdk.NPCSpec{...})` (`Outfit`, `Greeting`, up to 12 `Options`, `ShopEnabled`)
- **Starter Config, Loadouts, Skills & Jobs**:
  - `sdk.SetStarterConfig(sdk.StarterConfigSpec{...})` (`SpawnRoomX/Y`, `SpawnPos`, `StarterGold`, `StarterEquipped`, `StarterBag`, `IncludeAllEquip`, `IncludePortalStyles`, `BaseHP`, `BaseAP`, `BaseMP`, `DisableHousing`)
  - `sdk.SetHeroLoadout(sdk.SetHeroLoadoutCmd{...})`
  - `sdk.RegisterPlayerSkill(sdk.PlayerSkillSpec{...})` (`ScalingStat`: `"strength"`, `"wisdom"`, `"agility"`)
  - `sdk.RegisterJob(sdk.JobDefSpec{...})`
- **Content & Progression**:
  - `sdk.RegisterMonster(sdk.MonsterSpec{...})` (`AIStyle`: `""`, `"brawler"`, `"caster"`, `"kiter"`, `"support"`; up to 6 `Spells`; `TargetKind`: `"enemy"`, `"self"`, `"ally"`; `EffectKind`: `"damage"`, `"heal"`, `"lifesteal"`, `"push"`, `"pull"`; `VFXKind`: `"slash"`, `"bolt"`, `"nova"`, `"ring"`, `"pillar"`; optional `ModelPath` `.glb` or `ModelID`)
  - `sdk.SpawnMonster(sdk.MonsterSpawnSpec{...})` (`RespawnSec`: `0` = 30s default, `<0` = one-shot)
  - `sdk.RegisterItem(sdk.ItemSpec{...})` (`ID`: `2-48` chars `[a-z0-9_-]`; `Kind`: `"equipment"`, `"material"`, `"consumable"`, `"portal_style"`; `Slot`: `"hat"`, `"amulet"`, `"chest"`, `"belt"`, `"gloves"`, `"ring"`, `"boots"`, `"weapon"`; `Rarity`: `"Common"`, `"Uncommon"`, `"Rare"`, `"Heroic"`, `"Mythic"`, `"Relic"`, `"Royal"`; optional `VisualPreset`, `ModelPath`, `ModelID`, `Price`, `SellPrice`, `ShopHidden`, `ShopCategory`, `SetID`)
  - `sdk.SpawnGatherNode(sdk.GatherNodeSpec{...})` (`Job`: `"bucheron"`, `"mineur"`, `"herborist"`, `"fisherman"`, `"craftsman"`)
  - `sdk.RegisterRecipe(sdk.RecipeSpec{...})`
  - `sdk.GiveItemToHero(heroID, itemID, quantity)` (1..99; returns `sdk.CodeHeroOffline` (-6) for offline heroes, items are NOT queued)
  - `sdk.GiveGold(heroID, amount)` (negative = charge, `sdk.ErrNotEnoughGold` if the balance is too low) / `sdk.GiveXP(heroID, amount)` (applies level-ups) -- API level 2, no host caps
  - `sdk.RemoveMonsterSpawn(spawnID)` -- API level 2; an idle pack vanishes at once, a pack in battle when the battle ends
  - `sdk.GetHero(heroID) (HeroSnapshot, found, err)` -- API level 2; `HeroSnapshot` now carries `Level`, `XP`, `Gold`, `ClassID`
  - `sdk.SendChatToHero(heroID, text)` / `sdk.BroadcastChat(text)` -- API level 2; shown on the system channel tagged `[<plugin name>]`, max 240 ASCII chars
- **Codex**:
  - `sdk.ConfigureCodex(sdk.CodexConfigSpec{...})`
  - `sdk.RegisterCodexEntry(sdk.CodexEntrySpec{...})` (`Category`: `"bestiary"`, `"resources"`, `"items"`, `"recipes"`, `"custom"`)
- **Persistent Realm KV Storage**:
  - `sdk.KVSet(key, bytes)` / `sdk.KVSetString(key, str)` / `sdk.KVSetJSON(key, v)`
  - `sdk.KVGet(key)` / `sdk.KVGetString(key)` / `sdk.KVGetJSON(key, &dst)`
  - `sdk.KVDelete(key)` / `sdk.KVList(prefix, limit)`

---

## 4. Client Plugin API (`RoleClient`)

Embed `sdk.BaseClientPlugin` and implement:

- `Init()` - Register procedural 3D meshes (`sdk.RegisterCustomMesh`) and load saved UI prefs (`sdk.PrefGetString`).
- `OnEnterWorld(self sdk.HeroSnapshot)` - Register initial HUD actions and optionally open overlay.
- `OnTick(dtMs uint32)`
- `OnPacket(channel uint16, data []byte)` - Decode server snapshots and refresh HUD/overlay.
- `OnUIEvent(ev sdk.UIClickEvent)` - Handle HUD button clicks (`ev.ControlID`), input submit (`ev.Value`), and form values (`ev.Inputs`).

### Client Hostcalls & Declarative UI Builders

- `sdk.SendToServer(channel, payload)` / `sdk.SendJSONToServer(channel, v)`
- `sdk.SetHUDActions([]sdk.HUDActionSpec{...})` (max 6 buttons; `Icon` = atlas icon name like `"star"` or bundle `.png` path; `Badge` max 6 chars, longer badges are cut)
- `sdk.SetOverlay(sdk.UIPanelSpec{...})` / `sdk.ClearOverlay()`
  - `Anchor`: `"top_left"`, `"top_right"`, `"bottom_left"`, `"bottom_right"`, `"center"`
  - Node kinds: `"panel"`, `"col"`, `"row"`, `"label"`, `"button"`, `"input"`, `"icon"`, `"image"`, `"progress"`, `"scroll"`
  - Button `Variant`: `"default"`, `"active"` (alias `"primary"`), `"emerald"` (alias `"success"`), `"danger"`
  - `SetOverlay` rejects panels over the limits locally (-4); `sdk.ValidateOverlay(panel)` says why.
  - `sdk.PatchOverlay(panelID, sdk.UINodePatch{ID, Text, Value, Progress, Color, Hidden})` (API level 2) updates nodes by `ID` in place, keeping focus, typed input, and scroll position. Use it for timers and counters; use `Value: sdk.PatchText("")` to clear an input after submit. `UINodeSpec.Hidden` renders a node hidden.
  - Helper constructors: `sdk.ColNode`, `sdk.RowNode`, `sdk.CardNode`, `sdk.LabelNode`, `sdk.ButtonNode`, `sdk.InputNode`, `sdk.ProgressNode`, `sdk.ImageNode`
- `sdk.OpenCodex(query)`
- `sdk.PlayAudio(assetPath, volume)` (`.wav` or `.ogg` inside `.gtpack`, volume `0.0..1.0`)
- `sdk.RegisterCustomMesh(sdk.CustomMeshSpec{ID, Parts})` (`AnimMode`: `""`, `"spin_y"`, `"bob"`, `"orbit"`; `ActiveOnly`: hide gatherable part on cooldown). Triangles are front-facing when wound **counter-clockwise seen from outside**; when `Normals` are omitted they are computed from that winding, so clockwise triangles light as if inside-out. Set `DoubleSided` only for thin surfaces.
- **Local Client Prefs** (scoped per realm + plugin):
  - `sdk.PrefSet(key, val)` / `sdk.PrefSetJSON(key, v)`
  - `sdk.PrefGet(key)` / `sdk.PrefGetString(key, def)` / `sdk.PrefGetJSON(key, &dst)`
  - `sdk.PrefDelete(key)` / `sdk.PrefList(prefix, limit)`

---

## 4b. API Levels, Error Codes, and Testing

- `sdk.APILevel` = 2. Level-2 calls: `RemoveMonsterSpawn`, `GiveGold`, `GiveXP`, `GetHero`, `SendChatToHero`, `BroadcastChat`, `ChatCommandPlugin`, `PatchOverlay`. A plugin calling one on an older realm/client fails to load with an error naming the missing host function.
- Result codes (`sdk.CodeToErr`): `-2` wrong role (`ErrWrongRole`), `-3` ABI memory, `-4` invalid argument (`ErrInvalidArg`), `-5` quota/backend, `-6` hero offline (`ErrHeroOffline`), `-7` not enough gold (`ErrNotEnoughGold`), `-8` not found (`ErrNotFound`).
- **Unit tests with `go test`**: import `github.com/mmoboy/hero-armory-sdk/sdktest` and call `h := sdktest.Install(t)`. It records overlays, HUD actions, notices, packets, chat, spawns, and given items, keeps KV/Prefs in memory with the real limits, and models online/offline heroes (`h.SetOnline(sdk.HeroSnapshot{...})`). Helpers: `h.NoticesFor`, `h.ChatsFor`, `sdktest.PacketsFor[T]`, `h.FindNode`, `h.OverlayText`, `h.Fail["server_kv_set"] = sdk.CodeQuotaOrBackend`.

## 5. Hard Limits (`manifest.go` / Host VM)

- `.gtpack` archive: `<= 16 MiB` compressed, `<= 32 MiB` extracted, `<= 256` files.
- Allowed bundle extensions: `.wasm`, `.json`, `.png`, `.wav`, `.ogg`, `.glb`.
- Packet payload: `<= 64 KiB` (`MaxPacketBytes`).
- UI Overlay: `<= 128 KiB` JSON, `<= 128` nodes, `<= 8` depth, `<= 6` HUD actions.
- Server KV: key `<= 128` bytes, value `<= 64 KiB` everywhere. Key count, total size, and write rate depend on the realm's hosting mode: **local and dedicated realms are unlimited** by default; **cloud realms** default to `<= 2,000` keys, `<= 4 MiB` total, and `<= 50` KVSet/KVDelete calls per second (over-limit calls return `ErrQuotaOrBackend`, -5). Realm operators can change each limit (`-plugin-kv-max-keys`, `-plugin-kv-max-bytes`, `-plugin-kv-writes-per-sec`, or the matching `HERO_ARMORY_PLUGIN_KV_*` env vars), so design for the cloud defaults if your plugin should run anywhere. `sdktest` enforces the cloud defaults.
- Hero grants (`GiveGold`, `GiveXP`, `GiveItemToHero`) are never limited; they apply immediately. On cloud realms the hero's save to the database is batched (default every 2 s, and when the hero leaves; `-plugin-grant-save-interval-ms`).
- Client Prefs: `<= 128` keys, `<= 128 KiB` total, key `<= 96` bytes, value `<= 8 KiB`.
- Custom 3D Meshes & GLBs: `<= 32` meshes, `<= 16` parts/mesh, `<= 16,384` verts, `<= 32,768` tris, `<= 1 MiB` JSON; `.glb` assets `<= 50,000` verts and `<= 8` textures.
- Rooms & NPCs: `RoomSpec <= 64 KiB` JSON (`MaxRoomSpecBytes`), `<= 256` terrain ops, `<= 128` props/obstacles, `<= 16` portals/houses, `<= 32` interactables, `<= 12` NPC options (`MaxNPCOptions`).
- Realm custom registries: `<= 128` monsters, `<= 512` monster spawns, `<= 512` items, `<= 256` recipes.

---

## 6. Deploying to a Hero Armory Realm

Option A -- Drop into the realm data directory (auto-discovered on startup):

```bash
mkdir -p <realm-data-dir>/plugins
cp dist/server_plugin.wasm   <realm-data-dir>/plugins/server_plugin.wasm
cp dist/client_plugin.gtpack <realm-data-dir>/plugins/client_plugin.gtpack
```

Option B -- Explicit CLI flags on headless `hero-armory-server`:

```bash
go run ./apps/hero-armory/cmd/server \
  -server-plugin ./dist/server_plugin.wasm \
  -client-plugin ./dist/client_plugin.gtpack
```

Option C -- Environment variables:

```bash
export HERO_ARMORY_SERVER_PLUGIN=./dist/server_plugin.wasm
export HERO_ARMORY_CLIENT_PLUGIN=./dist/client_plugin.gtpack
```

Option D -- Bundled dev launcher (for workspace examples):

```bash
./apps/hero-armory/scripts/run_realm.sh <standard_realm|realm_bounty|minimal_realm> -id realm-1
```

Option E -- Publish to the public Plugin Library (`https://kmmo.enlumis.com/plugins`, Recommended):

Add `.github/workflows/release.yml` in your public GitHub plugin repo (no `replace` directives in `go.mod`):

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

Push a matching tag (for example `v1.0.0` for `"version": "1.0.0"` in `manifest.json`), wait for GitHub Actions to build and attest `server_plugin.wasm` + `client_plugin.gtpack` for free, then submit the repo URL + tag at `https://kmmo.enlumis.com/plugins` for a **Source-Verified** catalog entry selectable directly inside the Hero Armory desktop client.

For working references, inspect `sdk/hero-armory-plugin-example` (`github.com/mmoboy/hero-armory-plugin-example`) and `sdk/hero-armory-sdk/vanilla` alongside `sdk/hero-armory-sdk/examples/{realm_bounty,custom_char_create,standard_realm,minimal_realm}`.

### Realm-Authoritative Character Creation

- **Server hook**: Implement optional `sdk.CharacterCreatePlugin` (`OnBeginCharacterCreate`, `OnFinalizeCharacterCreate`). Do NOT add these methods on `BaseServerPlugin`; only realms customizing character creation implement them.
  - `CharacterCreateConfig.Mode`: `sdk.CharCreateModeVanilla` (`"vanilla"`, default UI + portable client Saved Looks presets), `sdk.CharCreateModeVanillaExtended` (`"vanilla_extended"`, keeps built-in sliders/presets while presenting Realm `Classes []StarterClassSpec`, e.g. `vanilla.SanctuaryStarterClasses()`), or `sdk.CharCreateModePluginCustom` (`"plugin_custom"`, delegates creation UI to `client.wasm`).
  - `CharacterCreateResult`: returns `RejectMessage` (or Go `error` from `OnFinalizeCharacterCreate`) to keep the player on `ScreenCharacterCreate` with the server error, or overrides `HeroName`, `AppearanceOverride`/`UseAppearanceOverride`, `Equipped`/`OverrideEquipped`, `StarterBag`/`OverrideStarterBag`, `StarterGold`/`OverrideStarterGold`, `BaseHP`, `BaseAP`, `BaseMP`, `SpawnRoomX`/`SpawnRoomY`/`SpawnPos`/`OverrideSpawn`, `CustomKV` (`{hero_id}` placeholder is substituted with the created hero's ID), and `WelcomeNotice`.
- **Client hook (`plugin_custom` or overlay helpers)**: Implement optional `sdk.CharacterCreateClientPlugin` (`OnEnterCharacterCreate(ctx CharacterCreateContext)`).
  - Use `sdk.SetOverlay(UIPanelSpec{...})` to render custom class/trait/name controls on `ScreenCharacterCreate`.
  - Call `sdk.SetCharacterCreatePreview(CharacterCreatePreviewSpec{...})` to drive the live 3D Sanctuary pedestal (`Equipped`/`ApplyEquipped`, `Appearance`/`UseAppearance`, `Pose`, `CameraFocus` `"body"` or `"face"`).
  - Call `sdk.SubmitCharacterCreate(CharacterCreateSubmission{Name, ClassID, CustomData, EnterImmediately: true})` when the player submits.
