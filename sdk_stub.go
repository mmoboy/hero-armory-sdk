//go:build !wasip1

package sdk

import "encoding/json"

// TestHost receives every hostcall made by plugin code in native (non-wasm)
// builds, so plugins can be unit-tested with `go test`. name is the host
// function name without the "host_" prefix (for example "server_spawn_monster")
// and payload is the JSON the real host would receive. A negative code is
// returned to the plugin as the hostcall result; out carries data for calls
// that return bytes (KV/Pref get and list, server_get_hero), where code -1
// means "not found".
//
// Without an installed TestHost every hostcall succeeds and does nothing.
// Most plugins should use the ready-made recorder in package sdk/sdktest.
type TestHost interface {
	HostCall(name string, payload []byte) (code int32, out []byte)
}

var testHost TestHost

// SetTestHost installs h for native builds and returns a func restoring the
// previous host. Not available in wasip1 builds.
func SetTestHost(h TestHost) (restore func()) {
	prev := testHost
	testHost = h
	return func() { testHost = prev }
}

// RegisteredServer returns the plugin passed to RegisterServer (native builds only).
func RegisteredServer() ServerPlugin { return activeServer }

// RegisteredClient returns the plugin passed to RegisterClient (native builds only).
func RegisteredClient() ClientPlugin { return activeClient }

// KVCall and PrefCall are the payloads of the storage hostcalls in native builds.
type KVCall struct {
	Key    string `json:"key,omitempty"`
	Value  []byte `json:"value,omitempty"`
	Prefix string `json:"prefix,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

func call(name string, payload []byte) int32 {
	if testHost == nil {
		return 0
	}
	code, _ := testHost.HostCall(name, payload)
	return code
}

func callOut(name string, payload []byte) ([]byte, bool, int32) {
	if testHost == nil {
		return nil, false, 0
	}
	code, out := testHost.HostCall(name, payload)
	if code == -1 {
		return nil, false, 0
	}
	if code < 0 {
		return nil, false, code
	}
	return out, true, 0
}

func kvPayload(c KVCall) []byte {
	raw, _ := json.Marshal(c)
	return raw
}

func listOut(name string, c KVCall) ([]byte, int32) {
	raw, ok, code := callOut(name, kvPayload(c))
	if code < 0 {
		return nil, code
	}
	if !ok || raw == nil {
		return []byte("[]"), 0
	}
	return raw, 0
}

func abiLog(_ uint32, msg string) { call("log", []byte(msg)) }

func abiSendToServer(channel uint16, payload []byte) int32 {
	raw, _ := json.Marshal(struct {
		Channel uint16 `json:"channel"`
		Payload []byte `json:"payload"`
	}{channel, payload})
	return call("send_packet", raw)
}

func abiServerSendPacket(raw []byte) int32          { return call("server_send_packet", raw) }
func abiServerSendNotice(raw []byte) int32          { return call("server_send_notice", raw) }
func abiServerRegisterMonster(raw []byte) int32     { return call("server_register_monster", raw) }
func abiServerSpawnMonster(raw []byte) int32        { return call("server_spawn_monster", raw) }
func abiServerRegisterItem(raw []byte) int32        { return call("server_register_item", raw) }
func abiServerSpawnGatherNode(raw []byte) int32     { return call("server_spawn_gather_node", raw) }
func abiServerRegisterRecipe(raw []byte) int32      { return call("server_register_recipe", raw) }
func abiServerGiveItem(raw []byte) int32            { return call("server_give_item", raw) }
func abiServerConfigureCodex(raw []byte) int32      { return call("server_configure_codex", raw) }
func abiServerRegisterCodexEntry(raw []byte) int32  { return call("server_register_codex_entry", raw) }
func abiSetHUDActions(raw []byte) int32             { return call("ui_set_hud_actions", raw) }
func abiSetOverlay(raw []byte) int32                { return call("ui_set_overlay", raw) }
func abiOpenCodex(query string) int32               { return call("ui_open_codex", []byte(query)) }
func abiRegisterCustomMesh(raw []byte) int32        { return call("ui_register_mesh", raw) }
func abiServerRegisterRoom(raw []byte) int32        { return call("server_register_room", raw) }
func abiServerRemoveRoom(raw []byte) int32          { return call("server_remove_room", raw) }
func abiServerSetSpawnRoom(raw []byte) int32        { return call("server_set_spawn_room", raw) }
func abiServerRegisterNPC(raw []byte) int32         { return call("server_register_npc", raw) }
func abiServerSetHeroLoadout(raw []byte) int32      { return call("server_set_hero_loadout", raw) }
func abiServerSetStarterConfig(raw []byte) int32    { return call("server_set_starter_config", raw) }
func abiServerRegisterPlayerSkill(raw []byte) int32 { return call("server_register_player_skill", raw) }
func abiServerRegisterJob(raw []byte) int32         { return call("server_register_job", raw) }
func abiServerSetNPCResult(raw []byte) int32        { return call("server_set_npc_result", raw) }
func abiSetCreatePreview(raw []byte) int32          { return call("ui_set_create_preview", raw) }
func abiSubmitCharCreate(raw []byte) int32          { return call("ui_submit_char_create", raw) }
func abiServerSetCharCreateConfig(raw []byte) int32 {
	return call("server_set_char_create_config", raw)
}
func abiServerSetCharCreateResult(raw []byte) int32 {
	return call("server_set_char_create_result", raw)
}
func abiServerRemoveMonsterSpawn(raw []byte) int32 { return call("server_remove_monster_spawn", raw) }
func abiServerGiveCurrency(raw []byte) int32       { return call("server_give_currency", raw) }
func abiServerSendChat(raw []byte) int32           { return call("server_send_chat", raw) }
func abiPatchOverlay(raw []byte) int32             { return call("ui_patch_overlay", raw) }
func abiServerTakeItem(raw []byte) int32           { return call("server_take_item", raw) }
func abiServerTeleportHero(raw []byte) int32       { return call("server_teleport_hero", raw) }
func abiClearOverlayID(raw []byte) int32           { return call("ui_clear_overlay_id", raw) }
func abiServerSpawnProp(raw []byte) int32          { return call("server_spawn_prop", raw) }
func abiServerRemoveProp(raw []byte) int32         { return call("server_remove_prop", raw) }
func abiServerSetRoomAtmosphere(raw []byte) int32  { return call("server_set_room_atmosphere", raw) }
func abiServerSetMapMarker(raw []byte) int32       { return call("server_set_map_marker", raw) }
func abiServerRemoveMapMarker(raw []byte) int32    { return call("server_remove_map_marker", raw) }
func abiServerSpawnActor(raw []byte) int32         { return call("server_spawn_actor", raw) }
func abiServerUpdateActor(raw []byte) int32        { return call("server_update_actor", raw) }
func abiServerRemoveActor(raw []byte) int32        { return call("server_remove_actor", raw) }
func abiServerPlayVFX(raw []byte) int32            { return call("server_play_vfx", raw) }
func abiServerShowFloatingText(raw []byte) int32   { return call("server_show_floating_text", raw) }
func abiServerSetGroundMarker(raw []byte) int32    { return call("server_set_ground_marker", raw) }
func abiServerRemoveGroundMarker(raw []byte) int32 { return call("server_remove_ground_marker", raw) }
func abiServerClearGroundMarkers(raw []byte) int32 { return call("server_clear_ground_markers", raw) }
func abiServerSetHeroHotbar(raw []byte) int32      { return call("server_set_hero_hotbar", raw) }
func abiServerSetHeroPose(raw []byte) int32        { return call("server_set_hero_pose", raw) }
func abiServerSetHeroControlLock(raw []byte) int32 { return call("server_set_hero_control_lock", raw) }
func abiServerSetMapConfig(raw []byte) int32       { return call("server_set_map_config", raw) }
func abiServerSetMapRegion(raw []byte) int32       { return call("server_set_map_region", raw) }
func abiServerRemoveMapRegion(raw []byte) int32    { return call("server_remove_map_region", raw) }
func abiServerSetMapRoute(raw []byte) int32        { return call("server_set_map_route", raw) }
func abiServerRemoveMapRoute(raw []byte) int32     { return call("server_remove_map_route", raw) }

func abiServerGetHero(req []byte) ([]byte, bool, int32) { return callOut("server_get_hero", req) }
func abiServerGetInventory(req []byte) ([]byte, int32) {
	if testHost == nil {
		return []byte("[]"), 0
	}
	code, out := testHost.HostCall("server_get_inventory", req)
	if code < 0 {
		return nil, code
	}
	if out == nil {
		return []byte("[]"), 0
	}
	return out, 0
}

func abiPlayAudio(assetPath string, volMilli uint32) int32 {
	raw, _ := json.Marshal(struct {
		Path     string `json:"path"`
		VolMilli uint32 `json:"vol_milli"`
	}{assetPath, volMilli})
	return call("play_audio", raw)
}

func abiServerKVSet(key string, val []byte) int32 {
	return call("server_kv_set", kvPayload(KVCall{Key: key, Value: val}))
}
func abiServerKVGet(key string) ([]byte, bool, int32) {
	return callOut("server_kv_get", kvPayload(KVCall{Key: key}))
}
func abiServerKVDelete(key string) int32 {
	return call("server_kv_delete", kvPayload(KVCall{Key: key}))
}
func abiServerKVList(prefix string, limit int) ([]byte, int32) {
	return listOut("server_kv_list", KVCall{Prefix: prefix, Limit: limit})
}
func abiPrefSet(key, val string) int32 {
	return call("pref_set", kvPayload(KVCall{Key: key, Value: []byte(val)}))
}
func abiPrefGet(key string) ([]byte, bool, int32) {
	return callOut("pref_get", kvPayload(KVCall{Key: key}))
}
func abiPrefDelete(key string) int32 {
	return call("pref_delete", kvPayload(KVCall{Key: key}))
}
func abiPrefList(prefix string, limit int) ([]byte, int32) {
	return listOut("pref_list", KVCall{Prefix: prefix, Limit: limit})
}
