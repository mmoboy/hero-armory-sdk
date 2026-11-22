package sdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var (
	ErrWrongRole      = errors.New("plugin storage API not allowed for this VM role")
	ErrInvalidArg     = errors.New("invalid plugin storage key, prefix, or value")
	ErrQuotaOrBackend = errors.New("plugin storage quota/rate limit exceeded or backend error")
	ErrMemoryABI      = errors.New("plugin storage ABI memory error")
)

func storageCodeToErr(code int32) error {
	switch code {
	case 0:
		return nil
	case -2:
		return ErrWrongRole
	case -3:
		return ErrMemoryABI
	case -4:
		return ErrInvalidArg
	case -5:
		return ErrQuotaOrBackend
	default:
		return fmt.Errorf("plugin storage error code %d", code)
	}
}

// ClientPlugin defines callbacks invoked by the Hero Armory native client on the
// main UI thread while connected to the realm that supplied client_plugin.gtpack.
type ClientPlugin interface {
	Init()
	OnEnterWorld(self HeroSnapshot)
	OnTick(dtMs uint32)
	OnPacket(channel uint16, data []byte)
	OnUIEvent(ev UIClickEvent)
}

// ServerPlugin defines callbacks invoked inside the Realm Server Wasm sandbox.
type ServerPlugin interface {
	Init()
	OnTick(dtMs uint32)
	OnPlayerJoin(hero HeroSnapshot)
	OnPlayerLeave(heroID string)
	OnRoomEnter(ev RoomEvent)
	OnChat(ev ChatEvent)
	OnPacket(senderHeroID string, channel uint16, data []byte)
}

// BaseClientPlugin provides default no-op implementations of ClientPlugin.
type BaseClientPlugin struct{}

func (BaseClientPlugin) Init()                     {}
func (BaseClientPlugin) OnEnterWorld(HeroSnapshot) {}
func (BaseClientPlugin) OnTick(uint32)             {}
func (BaseClientPlugin) OnPacket(uint16, []byte) {}
func (BaseClientPlugin) OnUIEvent(UIClickEvent)  {}

// CharacterCreateClientPlugin is optionally implemented by ClientPlugin to mount
// custom character-creation UI or drive 3D pedestal previews before spawning.
type CharacterCreateClientPlugin interface {
	OnEnterCharacterCreate(ctx CharacterCreateContext)
}

// CharacterCreatePlugin is optionally implemented by ServerPlugin to customize
// character creation rules, starter classes, validation, and starting state.
type CharacterCreatePlugin interface {
	OnBeginCharacterCreate(ev BeginCharacterCreateEvent) (CharacterCreateConfig, bool)
	OnFinalizeCharacterCreate(ev FinalizeCharacterCreateEvent) (CharacterCreateResult, error)
}

// BattleObserverPlugin is optionally implemented by ServerPlugin implementations
// that want callbacks whenever a tactical battle concludes.
type BattleObserverPlugin interface {
	OnBattleEnd(ev BattleEndEvent)
}

// WorldObserverPlugin is optionally implemented by ServerPlugin implementations
// that want callbacks when heroes harvest gatherables or craft recipes.
type WorldObserverPlugin interface {
	OnGather(ev GatherEvent)
	OnCraft(ev CraftEvent)
}

// NPCPlugin is optionally implemented by ServerPlugin implementations
// that handle dynamic NPC greetings, choices, blessings, or custom shops.
type NPCPlugin interface {
	OnNPCInteract(ev NPCInteractEvent) (NPCDialogResult, bool)
}

// InteractPlugin is optionally implemented by ServerPlugin implementations
// that handle room interactables (shrines, chests, altars, portals, etc.).
type InteractPlugin interface {
	OnInteract(ev InteractEvent) bool
}

// BaseServerPlugin provides default no-op implementations of ServerPlugin.
type BaseServerPlugin struct{}

func (BaseServerPlugin) Init()                                       {}
func (BaseServerPlugin) OnTick(uint32)                               {}
func (BaseServerPlugin) OnPlayerJoin(HeroSnapshot)                   {}
func (BaseServerPlugin) OnPlayerLeave(string)                        {}
func (BaseServerPlugin) OnRoomEnter(RoomEvent)                       {}
func (BaseServerPlugin) OnChat(ChatEvent)                            {}
func (BaseServerPlugin) OnPacket(string, uint16, []byte)             {}
func (BaseServerPlugin) OnBattleEnd(BattleEndEvent)                  {}
func (BaseServerPlugin) OnGather(GatherEvent)                        {}
func (BaseServerPlugin) OnCraft(CraftEvent)                          {}
func (BaseServerPlugin) OnNPCInteract(NPCInteractEvent) (NPCDialogResult, bool) {
	return NPCDialogResult{}, false
}
func (BaseServerPlugin) OnInteract(InteractEvent) bool { return false }

var (
	activeClient ClientPlugin
	activeServer ServerPlugin
)

// RegisterClient registers the client-side plugin implementation from func main/init.
func RegisterClient(p ClientPlugin) {
	activeClient = p
}

// RegisterServer registers the server-side plugin implementation from func main/init.
func RegisterServer(p ServerPlugin) {
	activeServer = p
}

// Log emits a sandboxed diagnostic log line tagged with the plugin ID.
func Log(msg string) {
	abiLog(0, msg)
}

// SendToServer sends an isolated OpPluginPacket frame from Client to Realm Server.
func SendToServer(channel uint16, payload []byte) int32 {
	return abiSendToServer(channel, payload)
}

// SendJSONToServer marshals v as JSON and sends it on channel to the Realm Server.
func SendJSONToServer(channel uint16, v any) int32 {
	raw, err := json.Marshal(v)
	if err != nil {
		return -1
	}
	return SendToServer(channel, raw)
}

// SetHUDActions registers compact icon+badge action buttons inside "HUD Actions".
func SetHUDActions(actions []HUDActionSpec) int32 {
	if len(actions) > MaxHUDActions {
		actions = actions[:MaxHUDActions]
	}
	raw, err := json.Marshal(actions)
	if err != nil {
		return -1
	}
	return abiSetHUDActions(raw)
}

// SetOverlay mounts or updates a declarative UI panel inside WorldModeState.
func SetOverlay(panel UIPanelSpec) int32 {
	raw, err := json.Marshal(panel)
	if err != nil {
		return -1
	}
	return abiSetOverlay(raw)
}

// ClearOverlay hides and removes the plugin's custom World HUD overlay.
func ClearOverlay() int32 {
	return abiSetOverlay(nil)
}

// OpenCodex opens the native Bestiary & Resource Codex panel, optionally pre-filtered to query.
func OpenCodex(query string) int32 {
	return abiOpenCodex(query)
}

// PlayAudio plays a bundled .wav or .ogg asset from inside the .gtpack archive.
func PlayAudio(assetPath string, volume float32) int32 {
	if volume < 0 {
		volume = 0
	}
	if volume > 1 {
		volume = 1
	}
	return abiPlayAudio(assetPath, uint32(volume*1000))
}

// RegisterCustomMesh uploads a pure-Go procedural 3D model under spec.ID on the client (RoleClient only).
// Registered models can be referenced by monsters, gatherables, room props, and items via ModelID.
func RegisterCustomMesh(spec CustomMeshSpec) error {
	if !IsValidContentID(spec.ID) || len(spec.Parts) == 0 || len(spec.Parts) > MaxCustomMeshParts {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	if len(raw) > MaxCustomMeshBytes {
		return ErrInvalidArg
	}
	return storageCodeToErr(abiRegisterCustomMesh(raw))
}

// SendToHero delivers an OpPluginPacket frame from server plugin to one connected hero.
func SendToHero(heroID string, channel uint16, payload []byte) int32 {
	raw, err := json.Marshal(HostServerPacketCmd{
		TargetHeroID: heroID,
		Channel:      channel,
		Payload:      payload,
	})
	if err != nil {
		return -1
	}
	return abiServerSendPacket(raw)
}

// SendJSONToHero marshals v and sends it on channel to one hero.
func SendJSONToHero(heroID string, channel uint16, v any) int32 {
	raw, err := json.Marshal(v)
	if err != nil {
		return -1
	}
	return SendToHero(heroID, channel, raw)
}

// BroadcastPacket delivers an OpPluginPacket frame from server plugin to all online heroes.
func BroadcastPacket(channel uint16, payload []byte) int32 {
	raw, err := json.Marshal(HostServerPacketCmd{
		BroadcastAll: true,
		Channel:      channel,
		Payload:      payload,
	})
	if err != nil {
		return -1
	}
	return abiServerSendPacket(raw)
}

// BroadcastJSON marshals v and broadcasts it to all online heroes on the realm.
func BroadcastJSON(channel uint16, v any) int32 {
	raw, err := json.Marshal(v)
	if err != nil {
		return -1
	}
	return BroadcastPacket(channel, raw)
}

// BroadcastRoomPacket delivers an OpPluginPacket frame to all heroes in (roomX, roomY).
func BroadcastRoomPacket(roomX, roomY int, channel uint16, payload []byte) int32 {
	raw, err := json.Marshal(HostServerPacketCmd{
		BroadcastRm: true,
		RoomX:       roomX,
		RoomY:       roomY,
		Channel:     channel,
		Payload:     payload,
	})
	if err != nil {
		return -1
	}
	return abiServerSendPacket(raw)
}

// SendNoticeToHero displays a SystemNotice banner on one hero's screen.
func SendNoticeToHero(heroID, severity, title, message string) int32 {
	raw, err := json.Marshal(HostNoticeCmd{
		TargetHeroID: heroID,
		Severity:     severity,
		Title:        title,
		Message:      message,
	})
	if err != nil {
		return -1
	}
	return abiServerSendNotice(raw)
}

// BroadcastNotice displays a SystemNotice banner to all heroes connected to the realm.
func BroadcastNotice(severity, title, message string) int32 {
	raw, err := json.Marshal(HostNoticeCmd{
		BroadcastAll: true,
		Severity:     severity,
		Title:        title,
		Message:      message,
	})
	if err != nil {
		return -1
	}
	return abiServerSendNotice(raw)
}

// RegisterMonster defines a custom monster archetype usable by SpawnMonster on the realm server.
func RegisterMonster(spec MonsterSpec) int32 {
	raw, err := json.Marshal(spec)
	if err != nil {
		return -1
	}
	return abiServerRegisterMonster(raw)
}

// SpawnMonster places or updates a monster pack in a realm room on the realm server.
func SpawnMonster(spec MonsterSpawnSpec) int32 {
	raw, err := json.Marshal(spec)
	if err != nil {
		return -1
	}
	return abiServerSpawnMonster(raw)
}

// RegisterItem registers a custom equipment piece, material resource, or consumable.
func RegisterItem(spec ItemSpec) int32 {
	raw, err := json.Marshal(spec)
	if err != nil {
		return -1
	}
	return abiServerRegisterItem(raw)
}

// SpawnGatherNode places or updates a gatherable resource node in a realm room.
func SpawnGatherNode(spec GatherNodeSpec) int32 {
	raw, err := json.Marshal(spec)
	if err != nil {
		return -1
	}
	return abiServerSpawnGatherNode(raw)
}

// RegisterRecipe adds a custom profession crafting recipe to the realm.
func RegisterRecipe(spec RecipeSpec) int32 {
	raw, err := json.Marshal(spec)
	if err != nil {
		return -1
	}
	return abiServerRegisterRecipe(raw)
}

// GiveItemToHero grants quantity of itemID into an online hero's bag.
func GiveItemToHero(heroID, itemID string, quantity int) int32 {
	raw, err := json.Marshal(GiveItemCmd{
		TargetHeroID: heroID,
		ItemID:       itemID,
		Quantity:     quantity,
	})
	if err != nil {
		return -1
	}
	return abiServerGiveItem(raw)
}

// ConfigureCodex customizes the title, banner, spoiler rules, and custom tab of the Codex.
func ConfigureCodex(spec CodexConfigSpec) int32 {
	raw, err := json.Marshal(spec)
	if err != nil {
		return -1
	}
	return abiServerConfigureCodex(raw)
}

// RegisterCodexEntry adds or updates a custom database/guide entry in the realm Codex.
func RegisterCodexEntry(spec CodexEntrySpec) int32 {
	raw, err := json.Marshal(spec)
	if err != nil {
		return -1
	}
	return abiServerRegisterCodexEntry(raw)
}

// RegisterRoom defines or replaces a room chunk in the realm map (RoleServer only).
func RegisterRoom(spec RoomSpec) int32 {
	raw, err := json.Marshal(spec)
	if err != nil {
		return -1
	}
	return abiServerRegisterRoom(raw)
}

// RemoveRoom removes a room coordinate from the realm map (RoleServer only).
func RemoveRoom(x, y int) int32 {
	raw, err := json.Marshal(RoomCoordCmd{X: x, Y: y})
	if err != nil {
		return -1
	}
	return abiServerRemoveRoom(raw)
}

// SetSpawnRoom sets the realm's initial spawn room coordinate and local position (RoleServer only).
func SetSpawnRoom(x, y int, pos [3]float32) int32 {
	raw, err := json.Marshal(SpawnRoomSpec{X: x, Y: y, Pos: pos})
	if err != nil {
		return -1
	}
	return abiServerSetSpawnRoom(raw)
}

// RegisterNPC defines or replaces a realm NPC with outfit, appearance, dialog options, and shop flag (RoleServer only).
func RegisterNPC(spec NPCSpec) int32 {
	raw, err := json.Marshal(spec)
	if err != nil {
		return -1
	}
	return abiServerRegisterNPC(raw)
}

// SetHeroLoadout updates an online hero's equipped items or pose from server plugin logic (RoleServer only).
func SetHeroLoadout(cmd SetHeroLoadoutCmd) int32 {
	raw, err := json.Marshal(cmd)
	if err != nil {
		return -1
	}
	return abiServerSetHeroLoadout(raw)
}

// SetStarterConfig sets starter spawn, gold, equipped loadout, starter bag items, and base hero stats (RoleServer only).
func SetStarterConfig(spec StarterConfigSpec) int32 {
	raw, err := json.Marshal(spec)
	if err != nil {
		return -1
	}
	return abiServerSetStarterConfig(raw)
}

// RegisterPlayerSkill defines or replaces a player tactical battle skill for this realm (RoleServer only).
func RegisterPlayerSkill(spec PlayerSkillSpec) int32 {
	raw, err := json.Marshal(spec)
	if err != nil {
		return -1
	}
	return abiServerRegisterPlayerSkill(raw)
}

// RegisterJob defines or customizes metadata for one gathering/crafting profession (RoleServer only).
func RegisterJob(spec JobDefSpec) int32 {
	raw, err := json.Marshal(spec)
	if err != nil {
		return -1
	}
	return abiServerRegisterJob(raw)
}

// KVSet persists a binary or UTF-8 value under key in the realm server's plugin KV store (RoleServer only).
func KVSet(key string, value []byte) error {
	if !IsValidStorageKey(key, MaxKVKeyLen) || len(value) > MaxKVValueBytes {
		return ErrInvalidArg
	}
	return storageCodeToErr(abiServerKVSet(key, value))
}

// KVSetString persists a string value under key in the realm server's plugin KV store (RoleServer only).
func KVSetString(key, value string) error {
	return KVSet(key, []byte(value))
}

// KVSetJSON marshals v as JSON and persists it under key in the realm server's plugin KV store (RoleServer only).
func KVSetJSON(key string, v any) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return KVSet(key, raw)
}

// KVGet retrieves raw bytes for key from the realm server's plugin KV store (RoleServer only).
func KVGet(key string) ([]byte, bool, error) {
	if !IsValidStorageKey(key, MaxKVKeyLen) {
		return nil, false, ErrInvalidArg
	}
	raw, found, code := abiServerKVGet(key)
	if code < 0 {
		return nil, false, storageCodeToErr(code)
	}
	if !found {
		return nil, false, nil
	}
	return raw, true, nil
}

// KVGetString retrieves a string value for key from the realm server's plugin KV store (RoleServer only).
func KVGetString(key string) (string, bool, error) {
	raw, found, err := KVGet(key)
	if err != nil || !found {
		return "", found, err
	}
	return string(raw), true, nil
}

// KVGetJSON retrieves and unmarshals a JSON value into dst from the realm server's plugin KV store (RoleServer only).
func KVGetJSON(key string, dst any) (bool, error) {
	raw, found, err := KVGet(key)
	if err != nil || !found {
		return found, err
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return false, err
	}
	return true, nil
}

// KVDelete removes key from the realm server's plugin KV store (RoleServer only).
func KVDelete(key string) error {
	if !IsValidStorageKey(key, MaxKVKeyLen) {
		return ErrInvalidArg
	}
	return storageCodeToErr(abiServerKVDelete(key))
}

// KVList returns sorted keys matching prefix (up to limit, max MaxKVListLimit) from the server plugin KV store.
func KVList(prefix string, limit int) ([]string, error) {
	if !IsValidStoragePrefix(prefix, MaxKVKeyLen) {
		return nil, ErrInvalidArg
	}
	if limit <= 0 || limit > MaxKVListLimit {
		limit = MaxKVListLimit
	}
	raw, code := abiServerKVList(prefix, limit)
	if code < 0 {
		return nil, storageCodeToErr(code)
	}
	if len(raw) == 0 {
		return nil, nil
	}
	var keys []string
	if err := json.Unmarshal(raw, &keys); err != nil {
		return nil, err
	}
	return keys, nil
}

// PrefSet persists a local UI preference scoped to this (realm, plugin) on the client (RoleClient only).
func PrefSet(key, value string) error {
	if !IsValidStorageKey(key, MaxPrefKeyLen) || len(value) > MaxPrefValueBytes || strings.IndexByte(value, 0) >= 0 || !utf8.ValidString(value) {
		return ErrInvalidArg
	}
	return storageCodeToErr(abiPrefSet(key, value))
}

// PrefSetJSON marshals v as JSON and persists it in the client plugin preference store (RoleClient only).
func PrefSetJSON(key string, v any) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return PrefSet(key, string(raw))
}

// PrefGet retrieves a local UI preference scoped to this (realm, plugin) on the client (RoleClient only).
func PrefGet(key string) (string, bool, error) {
	if !IsValidStorageKey(key, MaxPrefKeyLen) {
		return "", false, ErrInvalidArg
	}
	raw, found, code := abiPrefGet(key)
	if code < 0 {
		return "", false, storageCodeToErr(code)
	}
	if !found {
		return "", false, nil
	}
	return string(raw), true, nil
}

// PrefGetString returns the saved preference string for key, or defaultVal if unset or invalid.
func PrefGetString(key, defaultVal string) string {
	val, found, err := PrefGet(key)
	if err != nil || !found {
		return defaultVal
	}
	return val
}

// PrefGetJSON retrieves and unmarshals a JSON preference into dst (RoleClient only).
func PrefGetJSON(key string, dst any) (bool, error) {
	val, found, err := PrefGet(key)
	if err != nil || !found {
		return found, err
	}
	if err := json.Unmarshal([]byte(val), dst); err != nil {
		return false, err
	}
	return true, nil
}

// PrefDelete deletes one client plugin preference key (RoleClient only).
func PrefDelete(key string) error {
	if !IsValidStorageKey(key, MaxPrefKeyLen) {
		return ErrInvalidArg
	}
	return storageCodeToErr(abiPrefDelete(key))
}

// PrefList returns sorted client preference keys matching prefix for this (realm, plugin).
func PrefList(prefix string, limit int) ([]string, error) {
	if !IsValidStoragePrefix(prefix, MaxPrefKeyLen) {
		return nil, ErrInvalidArg
	}
	if limit <= 0 || limit > MaxPrefListLimit {
		limit = MaxPrefListLimit
	}
	raw, code := abiPrefList(prefix, limit)
	if code < 0 {
		return nil, storageCodeToErr(code)
	}
	if len(raw) == 0 {
		return nil, nil
	}
	var keys []string
	if err := json.Unmarshal(raw, &keys); err != nil {
		return nil, err
	}
	return keys, nil
}

// SetCharacterCreatePreview updates the live 3D Sanctuary pedestal gear,
// appearance, pose, or camera zoom during ScreenCharacterCreate (RoleClient only).
func SetCharacterCreatePreview(spec CharacterCreatePreviewSpec) error {
	raw, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	if len(raw) > 32768 {
		return ErrInvalidArg
	}
	return storageCodeToErr(abiSetCreatePreview(raw))
}

// SubmitCharacterCreate submits hero creation to the active Realm from a custom
// RoleClient character-creation overlay (RoleClient only).
func SubmitCharacterCreate(sub CharacterCreateSubmission) error {
	if len(sub.CustomData) > 16*1024 {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(sub)
	if err != nil {
		return err
	}
	if len(raw) > 32768 {
		return ErrInvalidArg
	}
	return storageCodeToErr(abiSubmitCharCreate(raw))
}
