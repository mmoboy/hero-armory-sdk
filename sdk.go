package sdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
)

var (
	ErrWrongRole      = errors.New("hostcall not allowed for this VM role")
	ErrInvalidArg     = errors.New("invalid hostcall argument")
	ErrQuotaOrBackend = errors.New("hostcall quota/rate limit exceeded or backend error")
	ErrMemoryABI      = errors.New("hostcall ABI memory error")
	ErrHeroOffline    = errors.New("target hero is not online")
	ErrNotEnoughGold  = errors.New("target hero does not have enough gold")
	ErrNotFound       = errors.New("target not found")
	ErrNotEnoughItems = errors.New("target hero does not have enough items")
	ErrInBattle       = errors.New("target hero is currently in battle")
)

// Hostcall result codes shared by the host and the SDK (API level 2+ calls,
// KV/Pref storage, and GiveItemToHero).
const (
	CodeOK             int32 = 0
	CodeWrongRole      int32 = -2
	CodeMemoryABI      int32 = -3
	CodeInvalidArg     int32 = -4
	CodeQuotaOrBackend int32 = -5
	CodeHeroOffline    int32 = -6
	CodeNotEnoughGold  int32 = -7
	CodeNotFound       int32 = -8
	CodeNotEnoughItems int32 = -9
	CodeInBattle       int32 = -10
)

func storageCodeToErr(code int32) error {
	return CodeToErr(code)
}

// CodeToErr maps a hostcall result code to an SDK error (nil for >= 0).
func CodeToErr(code int32) error {
	if code >= 0 {
		return nil
	}
	switch code {
	case CodeWrongRole:
		return ErrWrongRole
	case CodeMemoryABI:
		return ErrMemoryABI
	case CodeInvalidArg:
		return ErrInvalidArg
	case CodeQuotaOrBackend:
		return ErrQuotaOrBackend
	case CodeHeroOffline:
		return ErrHeroOffline
	case CodeNotEnoughGold:
		return ErrNotEnoughGold
	case CodeNotFound:
		return ErrNotFound
	case CodeNotEnoughItems:
		return ErrNotEnoughItems
	case CodeInBattle:
		return ErrInBattle
	default:
		return fmt.Errorf("hostcall error code %d", code)
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
func (BaseClientPlugin) OnPacket(uint16, []byte)   {}
func (BaseClientPlugin) OnUIEvent(UIClickEvent)    {}

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

// ChatCommandPlugin is optionally implemented by ServerPlugin implementations
// that handle chat commands (API level 2+). The host calls OnChatCommand
// BEFORE a local/server/commerce/party message is broadcast; returning true
// consumes the message: nobody else sees it and only the sender gets their
// own line echoed back. Reply with SendChatToHero or SendNoticeToHero.
// Messages that are not consumed are broadcast and then reach OnChat as usual.
type ChatCommandPlugin interface {
	OnChatCommand(ev ChatEvent) bool
}

// InteractPlugin is optionally implemented by ServerPlugin implementations
// that handle room interactables (shrines, chests, altars, portals, etc.).
type InteractPlugin interface {
	OnInteract(ev InteractEvent) bool
}

// ZoneObserverPlugin is optionally implemented by ServerPlugin implementations
// that want callbacks when heroes walk into or out of a RoomTriggerZoneSpec
// (API level 4+).
type ZoneObserverPlugin interface {
	OnZoneEnter(ev ZoneEvent)
	OnZoneExit(ev ZoneEvent)
}

// CombatInputPlugin is optionally implemented by ServerPlugin implementations
// that drive combat or interactive encounters using API level 5 primitives.
type CombatInputPlugin interface {
	OnActorClick(ev ActorClickEvent)
	OnGroundClick(ev GroundClickEvent)
	OnHotbarPress(ev HotbarPressEvent)
}

// MapInputPlugin is optionally implemented by ServerPlugin implementations
// that handle interactive World Map room or marker clicks (API level 6+).
type MapInputPlugin interface {
	OnMapRoomClick(ev MapRoomClickEvent)
	OnMapMarkerClick(ev MapMarkerClickEvent)
}

// BaseServerPlugin provides default no-op implementations of ServerPlugin.
type BaseServerPlugin struct{}

func (BaseServerPlugin) Init()                           {}
func (BaseServerPlugin) OnTick(uint32)                   {}
func (BaseServerPlugin) OnPlayerJoin(HeroSnapshot)       {}
func (BaseServerPlugin) OnPlayerLeave(string)            {}
func (BaseServerPlugin) OnRoomEnter(RoomEvent)           {}
func (BaseServerPlugin) OnChat(ChatEvent)                {}
func (BaseServerPlugin) OnPacket(string, uint16, []byte) {}
func (BaseServerPlugin) OnBattleEnd(BattleEndEvent)      {}
func (BaseServerPlugin) OnGather(GatherEvent)            {}
func (BaseServerPlugin) OnCraft(CraftEvent)              {}
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

// PatchOverlay updates individual nodes of the currently shown overlay panel
// (matched by UINodeSpec.ID) without rebuilding it, keeping input focus,
// typed text, and scroll position (RoleClient only, API level 2+). Patches
// are also applied to the stored panel, so a later full re-render keeps them.
// Returns ErrNotFound if panelID is not the active overlay.
func PatchOverlay(panelID string, patches ...UINodePatch) error {
	if len(patches) == 0 {
		return nil
	}
	if len(patches) > MaxUINodes {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(UIOverlayPatch{PanelID: panelID, Patches: patches})
	if err != nil {
		return err
	}
	if len(raw) > MaxUISpecBytes {
		return ErrInvalidArg
	}
	return CodeToErr(abiPatchOverlay(raw))
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
// Panels over the overlay limits are rejected locally (-4); call
// ValidateOverlay to get the reason.
func SetOverlay(panel UIPanelSpec) int32 {
	if ValidateOverlay(panel) != nil {
		return CodeInvalidArg
	}
	raw, err := json.Marshal(panel)
	if err != nil {
		return -1
	}
	return abiSetOverlay(raw)
}

// ClearOverlay hides and removes all custom World HUD overlays for this plugin.
func ClearOverlay() int32 {
	return abiSetOverlay(nil)
}

// ClearOverlayID hides and removes a single named overlay panel by ID
// (RoleClient only, API level 3+). Returns ErrNotFound if panelID is not open.
func ClearOverlayID(panelID string) error {
	if panelID == "" {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(ClearOverlayCmd{PanelID: panelID})
	if err != nil {
		return err
	}
	return CodeToErr(abiClearOverlayID(raw))
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

// GiveItemToHero grants quantity (1..99) of itemID into an online hero's bag.
// Returns 0 on success, CodeHeroOffline (-6) if the hero is not online (the
// items are NOT queued), or another negative code on invalid input.
// Hosts before API level 2 returned 0 for offline heroes.
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

// RemoveMonsterSpawn removes a plugin-placed monster pack by MonsterSpawnSpec.ID
// (RoleServer only, API level 2+). An idle pack disappears immediately; a pack
// that is in battle is removed when that battle ends, whatever the result.
// Returns ErrNotFound for unknown IDs.
func RemoveMonsterSpawn(id string) error {
	if !IsValidContentID(id) {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(RemoveMonsterSpawnCmd{ID: id})
	if err != nil {
		return err
	}
	return CodeToErr(abiServerRemoveMonsterSpawn(raw))
}

// GiveGold adds amount gold to an online hero (RoleServer only, API level 2+).
// A negative amount charges the hero and fails with ErrNotEnoughGold if the
// balance is too low. Returns ErrHeroOffline if the hero is not online.
func GiveGold(heroID string, amount int64) error {
	return giveCurrency(GiveCurrencyCmd{TargetHeroID: heroID, Gold: amount})
}

// GiveXP grants amount (> 0) hero XP to an online hero, applying level-ups
// (RoleServer only, API level 2+). Returns ErrHeroOffline if not online.
func GiveXP(heroID string, amount int64) error {
	if amount <= 0 {
		return ErrInvalidArg
	}
	return giveCurrency(GiveCurrencyCmd{TargetHeroID: heroID, XP: amount})
}

func giveCurrency(cmd GiveCurrencyCmd) error {
	if cmd.TargetHeroID == "" || cmd.XP < 0 || (cmd.Gold == 0 && cmd.XP == 0) {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	return CodeToErr(abiServerGiveCurrency(raw))
}

// GetHero returns the current snapshot (room, position, level, gold, class)
// of an online hero (RoleServer only, API level 2+). found is false if the
// hero is not online.
func GetHero(heroID string) (hero HeroSnapshot, found bool, err error) {
	if heroID == "" {
		return HeroSnapshot{}, false, ErrInvalidArg
	}
	req, err := json.Marshal(HeroQueryCmd{HeroID: heroID})
	if err != nil {
		return HeroSnapshot{}, false, err
	}
	raw, ok, code := abiServerGetHero(req)
	if code < 0 {
		return HeroSnapshot{}, false, CodeToErr(code)
	}
	if !ok {
		return HeroSnapshot{}, false, nil
	}
	if err := json.Unmarshal(raw, &hero); err != nil {
		return HeroSnapshot{}, false, err
	}
	return hero, true, nil
}

// SendChatToHero shows a chat line from this plugin to one online hero, on
// the "system" channel tagged with the plugin name (RoleServer only, API
// level 2+). Text is ASCII-sanitized and capped at MaxChatTextLen.
func SendChatToHero(heroID, text string) error {
	if heroID == "" {
		return ErrInvalidArg
	}
	return sendChat(HostChatCmd{TargetHeroID: heroID, Text: text})
}

// BroadcastChat shows a chat line from this plugin to every online hero
// (RoleServer only, API level 2+).
func BroadcastChat(text string) error {
	return sendChat(HostChatCmd{BroadcastAll: true, Text: text})
}

func sendChat(cmd HostChatCmd) error {
	cmd.Text = SanitizeASCII(cmd.Text, MaxChatTextLen)
	if cmd.Text == "" {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	return CodeToErr(abiServerSendChat(raw))
}

func isFinite32(v float32) bool {
	f := float64(v)
	return !math.IsNaN(f) && !math.IsInf(f, 0)
}

// TakeItemFromHero atomically removes quantity (1..999) of itemID from an
// online hero's bag (not equipped slots) (RoleServer only, API level 3+).
// Returns ErrHeroOffline if not online, or ErrNotEnoughItems if the hero's
// bag holds fewer than quantity (without mutating the bag).
func TakeItemFromHero(heroID, itemID string, quantity int) error {
	if heroID == "" || !IsValidContentID(itemID) || quantity < 1 || quantity > 999 {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(TakeItemCmd{
		TargetHeroID: heroID,
		ItemID:       itemID,
		Quantity:     quantity,
	})
	if err != nil {
		return err
	}
	return CodeToErr(abiServerTakeItem(raw))
}

// GetHeroInventory returns the bag item stacks of an online hero aggregated by
// ItemID (RoleServer only, API level 3+). Returns ErrHeroOffline if offline.
func GetHeroInventory(heroID string) ([]ItemStackSpec, error) {
	if heroID == "" {
		return nil, ErrInvalidArg
	}
	req, err := json.Marshal(HeroQueryCmd{HeroID: heroID})
	if err != nil {
		return nil, err
	}
	raw, code := abiServerGetInventory(req)
	if code < 0 {
		return nil, CodeToErr(code)
	}
	if len(raw) == 0 {
		return nil, nil
	}
	var stacks []ItemStackSpec
	if err := json.Unmarshal(raw, &stacks); err != nil {
		return nil, err
	}
	return stacks, nil
}

// GetHeroItemCount returns the total quantity of itemID in an online hero's
// bag (RoleServer only, API level 3+). Returns ErrHeroOffline if offline.
func GetHeroItemCount(heroID, itemID string) (int, error) {
	if heroID == "" || !IsValidContentID(itemID) {
		return 0, ErrInvalidArg
	}
	stacks, err := GetHeroInventory(heroID)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, st := range stacks {
		if st.ItemID == itemID {
			total += st.Quantity
		}
	}
	return total, nil
}

// TeleportHero moves an online hero to room (roomX, roomY) (RoleServer only,
// API level 3+). When usePos is true, the hero lands at pos (clamped to room
// bounds); when false, the hero lands at the room's default spawn position.
// Returns ErrHeroOffline if offline, ErrInBattle if currently in a battle, or
// ErrNotFound if the destination room does not exist.
func TeleportHero(heroID string, roomX, roomY int, pos [3]float32, usePos bool) error {
	if heroID == "" {
		return ErrInvalidArg
	}
	if usePos && (!isFinite32(pos[0]) || !isFinite32(pos[1]) || !isFinite32(pos[2])) {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(TeleportHeroCmd{
		TargetHeroID: heroID,
		RoomX:        roomX,
		RoomY:        roomY,
		Pos:          pos,
		UsePos:       usePos,
	})
	if err != nil {
		return err
	}
	return CodeToErr(abiServerTeleportHero(raw))
}

// SpawnProp places or updates a dynamic 3D world prop in room (roomX, roomY)
// (RoleServer only, API level 3+). spec.ID must satisfy IsValidContentID.
// Returns ErrNotFound if the room does not exist, or ErrQuotaOrBackend if
// MaxDynamicProps or MaxDynamicPropsPerRoom would be exceeded.
func SpawnProp(roomX, roomY int, spec RoomPropSpec) error {
	if !IsValidContentID(spec.ID) {
		return ErrInvalidArg
	}
	if !isFinite32(spec.X) || !isFinite32(spec.Y) || !isFinite32(spec.Z) ||
		!isFinite32(spec.Scale) || !isFinite32(spec.Yaw) ||
		!isFinite32(spec.SolidRadius) || spec.SolidRadius < 0 {
		return ErrInvalidArg
	}
	if spec.ModelPath != "" {
		if _, err := CleanBundlePath(spec.ModelPath); err != nil {
			return ErrInvalidArg
		}
	}
	if spec.ModelID != "" && !IsValidContentID(spec.ModelID) {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(SpawnPropCmd{
		RoomX: roomX,
		RoomY: roomY,
		Prop:  spec,
	})
	if err != nil {
		return err
	}
	return CodeToErr(abiServerSpawnProp(raw))
}

// RemoveProp removes a plugin-spawned dynamic 3D world prop by ID
// (RoleServer only, API level 3+). Returns ErrNotFound if id is unknown.
func RemoveProp(id string) error {
	if !IsValidContentID(id) {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(RemovePropCmd{ID: id})
	if err != nil {
		return err
	}
	return CodeToErr(abiServerRemoveProp(raw))
}

func isFiniteVec3(v [3]float32) bool {
	return isFinite32(v[0]) && isFinite32(v[1]) && isFinite32(v[2])
}

func isFiniteVec4(v [4]float32) bool {
	return isFinite32(v[0]) && isFinite32(v[1]) && isFinite32(v[2]) && isFinite32(v[3])
}

func isValidWeather(w string) bool {
	switch w {
	case "", "none", "rain", "snow", "ash", "spores", "embers", "void", "starfall":
		return true
	default:
		return false
	}
}

func isValidEmitterKind(k string) bool {
	switch k {
	case "embers", "fireflies", "mist", "sparkles", "ash", "snow", "rain", "portal_swirl", "void":
		return true
	default:
		return false
	}
}

func validateRoomLight(l RoomLightSpec) bool {
	return isFinite32(l.X) && isFinite32(l.Y) && isFinite32(l.Z) &&
		isFiniteVec3(l.Color) &&
		isFinite32(l.Intensity) && l.Intensity >= 0 &&
		isFinite32(l.Range) && l.Range >= 0 &&
		isFinite32(l.PulseSpeed) && l.PulseSpeed >= 0
}

func validateRoomEmitter(e RoomEmitterSpec) bool {
	return isValidEmitterKind(e.Kind) &&
		isFinite32(e.X) && isFinite32(e.Y) && isFinite32(e.Z) &&
		isFiniteVec3(e.Spread) && isFiniteVec4(e.Color) &&
		isFinite32(e.Rate) && e.Rate >= 0 &&
		isFinite32(e.Size) && e.Size >= 0 &&
		isFinite32(e.Speed) && e.Speed >= 0 &&
		isFinite32(e.Lifetime) && e.Lifetime >= 0
}

func validateRoomAtmosphere(a RoomAtmosphereSpec) bool {
	if !isValidWeather(a.Weather) {
		return false
	}
	if !isFiniteVec3(a.SkyZenith) || !isFiniteVec3(a.SkyHorizon) ||
		!isFiniteVec3(a.FogColor) || !isFinite32(a.FogDensity) || a.FogDensity < 0 ||
		!isFiniteVec3(a.SunColor) || !isFinite32(a.SunIntensity) || a.SunIntensity < 0 ||
		!isFinite32(a.SunPitch) || !isFinite32(a.SunYaw) ||
		!isFiniteVec3(a.AmbientColor) || !isFiniteVec4(a.WeatherColor) ||
		!isFinite32(a.WeatherIntensity) || a.WeatherIntensity < 0 {
		return false
	}
	if len(a.Lights) > MaxRoomLights || len(a.Emitters) > MaxRoomEmitters {
		return false
	}
	for _, l := range a.Lights {
		if !validateRoomLight(l) {
			return false
		}
	}
	for _, e := range a.Emitters {
		if !validateRoomEmitter(e) {
			return false
		}
	}
	return true
}

// SetRoomAtmosphere dynamically overrides the sky, fog, lighting, weather, and
// optional extra point lights/emitters for room (roomX, roomY) (RoleServer
// only, API level 4+). Syncs live to clients in that room and to future
// entrants. Returns ErrNotFound if the room does not exist.
func SetRoomAtmosphere(roomX, roomY int, spec RoomAtmosphereSpec) error {
	if !validateRoomAtmosphere(spec) {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(SetRoomAtmosphereCmd{
		RoomX:      roomX,
		RoomY:      roomY,
		Atmosphere: &spec,
	})
	if err != nil {
		return err
	}
	return CodeToErr(abiServerSetRoomAtmosphere(raw))
}

// ClearRoomAtmosphere removes any live atmosphere override on (roomX, roomY),
// reverting the room to its base RoomSpec.Atmosphere (RoleServer only, API
// level 4+). Returns ErrNotFound if the room does not exist.
func ClearRoomAtmosphere(roomX, roomY int) error {
	raw, err := json.Marshal(SetRoomAtmosphereCmd{
		RoomX: roomX,
		RoomY: roomY,
	})
	if err != nil {
		return err
	}
	return CodeToErr(abiServerSetRoomAtmosphere(raw))
}

// SetMapMarker places or updates a dynamic marker on the world map and minimap
// (RoleServer only, API level 4+, extended in API level 6). spec.ID must
// satisfy IsValidContentID. Returns ErrNotFound if (RoomX, RoomY) does not
// exist, or ErrQuotaOrBackend if MaxMapMarkers would be exceeded.
func SetMapMarker(spec MapMarkerSpec) error {
	if !IsValidContentID(spec.ID) || !isFinite32(spec.PosX) || !isFinite32(spec.PosZ) ||
		!isFiniteVec4(spec.Color) || !isFinite32(spec.Radius) || spec.Radius < 0 {
		return ErrInvalidArg
	}
	spec.Label = SanitizeASCII(spec.Label, 48)
	spec.Icon = SanitizeASCII(spec.Icon, 48)
	spec.Category = SanitizeASCII(spec.Category, 24)
	spec.Tooltip = SanitizeASCII(spec.Tooltip, 160)
	raw, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	return CodeToErr(abiServerSetMapMarker(raw))
}

// RemoveMapMarker removes a dynamic world-map marker by ID (RoleServer only,
// API level 4+). Returns ErrNotFound if id is unknown.
func RemoveMapMarker(id string) error {
	if !IsValidContentID(id) {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(RemoveMapMarkerCmd{ID: id})
	if err != nil {
		return err
	}
	return CodeToErr(abiServerRemoveMapMarker(raw))
}

// Pointer helpers for ActorPatch and partial update structs.
func PtrFloat32(v float32) *float32 { return &v }
func PtrInt(v int) *int             { return &v }
func PtrString(v string) *string    { return &v }
func PtrBool(v bool) *bool          { return &v }

func isValidActorTeam(team string) bool {
	switch team {
	case "", "neutral", "ally", "enemy":
		return true
	default:
		return false
	}
}

func isValidVFXKind(kind string) bool {
	switch kind {
	case "slash", "bolt", "nova", "heal", "shield", "buff", "debuff", "ring", "beam":
		return true
	default:
		return false
	}
}

func isValidGroundMarkerShape(shape string) bool {
	switch shape {
	case "circle", "ring", "tile", "line":
		return true
	default:
		return false
	}
}

func sanitizeStatuses(in []ActorStatusSpec) ([]ActorStatusSpec, bool) {
	if len(in) > MaxActorStatuses {
		return nil, false
	}
	if len(in) == 0 {
		return in, true
	}
	out := make([]ActorStatusSpec, len(in))
	for i, s := range in {
		if !IsValidContentID(s.ID) || !isFiniteVec3(s.Color) || s.Stacks < 0 {
			return nil, false
		}
		s.Label = SanitizeASCII(s.Label, 24)
		out[i] = s
	}
	return out, true
}

// SpawnActor spawns or replaces a dynamic 3D actor in a room (RoleServer only,
// API level 5+). spec.ID must satisfy IsValidContentID. Returns ErrNotFound if
// the room does not exist, or ErrQuotaOrBackend if actor limits are exceeded.
func SpawnActor(spec ActorSpec) error {
	if !IsValidContentID(spec.ID) || !isValidActorTeam(spec.Team) ||
		!isFinite32(spec.PosX) || !isFinite32(spec.PosZ) || !isFinite32(spec.Yaw) ||
		!isFinite32(spec.Scale) || spec.Scale < 0 ||
		!isFiniteVec3(spec.BodyColor) || !isFiniteVec3(spec.TrimColor) ||
		spec.HP < 0 || spec.MaxHP < 0 || spec.Shield < 0 {
		return ErrInvalidArg
	}
	if spec.ModelPath != "" {
		if _, err := CleanBundlePath(spec.ModelPath); err != nil {
			return ErrInvalidArg
		}
	}
	if spec.ModelID != "" && !IsValidContentID(spec.ModelID) {
		return ErrInvalidArg
	}
	statuses, ok := sanitizeStatuses(spec.Statuses)
	if !ok {
		return ErrInvalidArg
	}
	spec.Statuses = statuses
	spec.Name = SanitizeASCII(spec.Name, 48)
	spec.Kind = SanitizeASCII(spec.Kind, 32)
	spec.Pose = SanitizeASCII(spec.Pose, 24)
	raw, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	return CodeToErr(abiServerSpawnActor(raw))
}

// UpdateActor partially updates an existing dynamic actor (RoleServer only,
// API level 5+). Returns ErrNotFound if patch.ID does not exist.
func UpdateActor(patch ActorPatch) error {
	if !IsValidContentID(patch.ID) {
		return ErrInvalidArg
	}
	if patch.PosX != nil && !isFinite32(*patch.PosX) {
		return ErrInvalidArg
	}
	if patch.PosZ != nil && !isFinite32(*patch.PosZ) {
		return ErrInvalidArg
	}
	if patch.Yaw != nil && !isFinite32(*patch.Yaw) {
		return ErrInvalidArg
	}
	if patch.HP != nil && *patch.HP < 0 {
		return ErrInvalidArg
	}
	if patch.MaxHP != nil && *patch.MaxHP < 0 {
		return ErrInvalidArg
	}
	if patch.Shield != nil && *patch.Shield < 0 {
		return ErrInvalidArg
	}
	if patch.Name != nil {
		s := SanitizeASCII(*patch.Name, 48)
		patch.Name = &s
	}
	if patch.Pose != nil {
		s := SanitizeASCII(*patch.Pose, 24)
		patch.Pose = &s
	}
	if patch.ReplaceStatuses || len(patch.Statuses) > 0 {
		statuses, ok := sanitizeStatuses(patch.Statuses)
		if !ok {
			return ErrInvalidArg
		}
		patch.Statuses = statuses
	}
	raw, err := json.Marshal(patch)
	if err != nil {
		return err
	}
	return CodeToErr(abiServerUpdateActor(raw))
}

// RemoveActor removes a dynamic actor by ID (RoleServer only, API level 5+).
// Returns ErrNotFound if id does not exist.
func RemoveActor(id string) error {
	if !IsValidContentID(id) {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(RemoveActorCmd{ID: id})
	if err != nil {
		return err
	}
	return CodeToErr(abiServerRemoveActor(raw))
}

// PlayVFX triggers a one-shot 3D visual effect in a room (RoleServer only,
// API level 5+). Returns ErrNotFound if the room does not exist.
func PlayVFX(spec VFXSpec) error {
	if !isValidVFXKind(spec.Kind) ||
		!isFinite32(spec.FromX) || !isFinite32(spec.FromY) || !isFinite32(spec.FromZ) ||
		!isFinite32(spec.ToX) || !isFinite32(spec.ToY) || !isFinite32(spec.ToZ) ||
		!isFinite32(spec.Radius) || spec.Radius < 0 ||
		!isFiniteVec4(spec.Color) ||
		!isFinite32(spec.DurationSec) || spec.DurationSec < 0 {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	return CodeToErr(abiServerPlayVFX(raw))
}

// ShowFloatingText spawns rising world-space text in a room (RoleServer only,
// API level 5+). Returns ErrNotFound if the room does not exist.
func ShowFloatingText(spec FloatingTextSpec) error {
	if !isFinite32(spec.PosX) || !isFinite32(spec.PosY) || !isFinite32(spec.PosZ) ||
		!isFiniteVec4(spec.Color) || !isFinite32(spec.Scale) || spec.Scale < 0 {
		return ErrInvalidArg
	}
	spec.Text = SanitizeASCII(spec.Text, 48)
	if spec.Text == "" {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	return CodeToErr(abiServerShowFloatingText(raw))
}

// SetGroundMarker places or updates a 3D ground telegraph or highlight
// (RoleServer only, API level 5+). Empty OwnerHeroID makes the marker visible
// to everyone in the room; non-empty restricts it to that hero.
func SetGroundMarker(spec GroundMarkerSpec) error {
	if !IsValidContentID(spec.ID) || !isValidGroundMarkerShape(spec.Shape) ||
		!isFinite32(spec.PosX) || !isFinite32(spec.PosZ) ||
		!isFinite32(spec.ToX) || !isFinite32(spec.ToZ) ||
		!isFinite32(spec.Radius) || spec.Radius < 0 ||
		!isFiniteVec4(spec.Color) {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	return CodeToErr(abiServerSetGroundMarker(raw))
}

// RemoveGroundMarker removes one ground marker by ID (RoleServer only,
// API level 5+). Returns ErrNotFound if id does not exist.
func RemoveGroundMarker(id string) error {
	if !IsValidContentID(id) {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(RemoveGroundMarkerCmd{ID: id})
	if err != nil {
		return err
	}
	return CodeToErr(abiServerRemoveGroundMarker(raw))
}

// ClearGroundMarkers removes ground markers in (roomX, roomY) (RoleServer
// only, API level 5+). If ownerHeroID is non-empty, only markers owned by that
// hero are removed; if empty, all markers in the room are cleared.
func ClearGroundMarkers(roomX, roomY int, ownerHeroID string) error {
	raw, err := json.Marshal(ClearGroundMarkersCmd{
		RoomX:       roomX,
		RoomY:       roomY,
		OwnerHeroID: ownerHeroID,
	})
	if err != nil {
		return err
	}
	return CodeToErr(abiServerClearGroundMarkers(raw))
}

// SetHeroHotbar configures one online hero's plugin action hotbar (slots 1..10,
// slot 10 maps to key 0) (RoleServer only, API level 5+). Passing an empty
// slice hides the plugin hotbar. Returns ErrHeroOffline if not online.
func SetHeroHotbar(heroID string, slots []HotbarSlotSpec) error {
	if heroID == "" || len(slots) > MaxHotbarSlots {
		return ErrInvalidArg
	}
	clean := make([]HotbarSlotSpec, len(slots))
	seen := make(map[int]bool, len(slots))
	for i, s := range slots {
		if s.Slot < 1 || s.Slot > MaxHotbarSlots || seen[s.Slot] || !IsValidContentID(s.ID) ||
			s.Cooldown < 0 || s.Charges < 0 || !isFiniteVec3(s.AccentColor) {
			return ErrInvalidArg
		}
		seen[s.Slot] = true
		s.Label = SanitizeASCII(s.Label, 24)
		s.IconName = SanitizeASCII(s.IconName, 48)
		s.CostText = SanitizeASCII(s.CostText, 16)
		s.Tooltip = SanitizeASCII(s.Tooltip, 160)
		s.Group = SanitizeASCII(s.Group, 16)
		s.KeyHint = SanitizeASCII(s.KeyHint, 8)
		s.Badge = SanitizeASCII(s.Badge, 12)
		clean[i] = s
	}
	raw, err := json.Marshal(SetHeroHotbarCmd{
		TargetHeroID: heroID,
		Slots:        clean,
	})
	if err != nil {
		return err
	}
	return CodeToErr(abiServerSetHeroHotbar(raw))
}

// SetHeroPose sets an online hero's animation pose and facing yaw (RoleServer
// only, API level 5+). Returns ErrHeroOffline if not online.
func SetHeroPose(heroID string, pose string, yaw float32) error {
	if heroID == "" || !isFinite32(yaw) {
		return ErrInvalidArg
	}
	pose = SanitizeASCII(pose, 24)
	if pose == "" {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(SetHeroPoseCmd{
		TargetHeroID: heroID,
		Pose:         pose,
		Yaw:          yaw,
	})
	if err != nil {
		return err
	}
	return CodeToErr(abiServerSetHeroPose(raw))
}

// SetHeroControlLock locks or unlocks free overworld movement for an online
// hero (RoleServer only, API level 5+). Returns ErrHeroOffline if not online.
func SetHeroControlLock(heroID string, lockMove bool) error {
	if heroID == "" {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(SetHeroControlLockCmd{
		TargetHeroID: heroID,
		LockMove:     lockMove,
	})
	if err != nil {
		return err
	}
	return CodeToErr(abiServerSetHeroControlLock(raw))
}

func isValidMapRouteStyle(style string) bool {
	switch style {
	case "", "road", "portal", "leyline", "dashed":
		return true
	default:
		return false
	}
}

// SetMapConfig customizes the realm's World Map title, subtitle, colors,
// coordinate display, and fast-travel hint (RoleServer only, API level 6+).
func SetMapConfig(spec MapConfigSpec) error {
	if !isFiniteVec4(spec.BgColor) || !isFiniteVec4(spec.GridColor) {
		return ErrInvalidArg
	}
	spec.Title = SanitizeASCII(spec.Title, 48)
	spec.Subtitle = SanitizeASCII(spec.Subtitle, 96)
	raw, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	return CodeToErr(abiServerSetMapConfig(raw))
}

// SetMapRegion places or updates a shaded territory region on the World Map
// (RoleServer only, API level 6+). Returns ErrQuotaOrBackend if MaxMapRegions
// would be exceeded.
func SetMapRegion(spec MapRegionSpec) error {
	if !IsValidContentID(spec.ID) || spec.MinRoomX > spec.MaxRoomX || spec.MinRoomY > spec.MaxRoomY ||
		spec.DangerLevel < 0 || !isFiniteVec4(spec.FillColor) || !isFiniteVec4(spec.BorderColor) ||
		!isFiniteVec3(spec.LabelColor) {
		return ErrInvalidArg
	}
	spec.Name = SanitizeASCII(spec.Name, 48)
	spec.Subtitle = SanitizeASCII(spec.Subtitle, 64)
	if spec.Name == "" {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	return CodeToErr(abiServerSetMapRegion(raw))
}

// RemoveMapRegion removes a World Map region by ID (RoleServer only, API level 6+).
func RemoveMapRegion(id string) error {
	if !IsValidContentID(id) {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(RemoveMapRegionCmd{ID: id})
	if err != nil {
		return err
	}
	return CodeToErr(abiServerRemoveMapRegion(raw))
}

// SetMapRoute draws or updates a route connector between two rooms on the World
// Map (RoleServer only, API level 6+). Returns ErrQuotaOrBackend if
// MaxMapRoutes would be exceeded.
func SetMapRoute(spec MapRouteSpec) error {
	if !IsValidContentID(spec.ID) || !isValidMapRouteStyle(spec.Style) || !isFiniteVec4(spec.Color) {
		return ErrInvalidArg
	}
	spec.Label = SanitizeASCII(spec.Label, 48)
	raw, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	return CodeToErr(abiServerSetMapRoute(raw))
}

// RemoveMapRoute removes a World Map route by ID (RoleServer only, API level 6+).
func RemoveMapRoute(id string) error {
	if !IsValidContentID(id) {
		return ErrInvalidArg
	}
	raw, err := json.Marshal(RemoveMapRouteCmd{ID: id})
	if err != nil {
		return err
	}
	return CodeToErr(abiServerRemoveMapRoute(raw))
}
