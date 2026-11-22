# custom_char_create (Sanctuary Origins)

Reference `hero-armory-sdk` plugin demonstrating **Realm-authoritative character
creation** (`sdk.CharacterCreatePlugin` + `sdk.CharacterCreateClientPlugin`).

## What It Demonstrates

1. **Starter Classes via `vanilla.SanctuaryStarterClasses()`** (`server/main.go`):
   Exposes Kokiri Guardian, Gobball Warchief, Wakfu Chronomancer, and Twilight
   Sentinel origins with distinct gear, HP/AP/MP bonuses, starter bag stacks,
   and gold awards.
2. **Authoritative Name & Rule Validation** (`OnFinalizeCharacterCreate`):
   Rejects reserved names and unknown class IDs on the Realm Server before
   persisting any hero row, returning clear validation errors back to the client
   creation screen without disconnecting.
3. **Custom Client Creation Overlay + Live 3D Preview** (`client/main.go`):
   - `OnEnterCharacterCreate(ctx)` mounts a declarative panel via `sdk.SetOverlay`.
   - `sdk.SetCharacterCreatePreview(...)` updates the live 3D Sanctuary pedestal
     (class starter equipment, hair/iris appearance, and camera zoom).
   - `sdk.SubmitCharacterCreate(...)` sends the hero name, selected `ClassID`,
     backstory trait `CustomData`, and `EnterImmediately: true` to spawn
     straight into World Mode on approval.
4. **Vanilla Extended Alternative**:
   Changing `creationMode` in `server/main.go` to `sdk.CharCreateModeVanillaExtended`
   preserves the built-in Vanilla face/body/eye sliders and client-local Saved Looks
   presets while presenting realm starter class cards cleanly in native UI.

## Build

```bash
./build.sh
```

Outputs `dist/server_plugin.wasm` and `dist/client_plugin.gtpack`.
