//go:build !wasip1

// Package sdktest is an in-memory Hero Armory host for unit-testing plugins
// with plain `go test` (no realm server, no client, no wasm).
//
// It records every hostcall (overlays, HUD actions, notices, packets, chat,
// spawns, items, gold/XP), keeps KV and Prefs in memory with the real host
// limits, and tracks which heroes are online so offline semantics match the
// real host (GiveItemToHero, GiveGold, GetHero, ...).
//
//	func TestMyPlugin(t *testing.T) {
//		h := sdktest.Install(t)
//		h.SetOnline(sdk.HeroSnapshot{ID: "h1", Name: "Ayla"})
//		p := newServer()
//		p.Init()
//		p.OnPlayerJoin(h.Hero("h1"))
//		if len(h.Notices) == 0 { t.Fatal("expected a welcome notice") }
//	}
package sdktest

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"

	sdk "github.com/mmoboy/hero-armory-sdk"
)

// ClientPacket is one SendToServer call.
type ClientPacket struct {
	Channel uint16
	Payload []byte
}

// Host is the recording test host. All fields are safe to read after the
// plugin call returns; use Reset to clear recordings between steps.
type Host struct {
	mu sync.Mutex

	// Calls lists every hostcall name in order.
	Calls []string
	Logs  []string

	// Client role.
	Overlay       *sdk.UIPanelSpec            // most recently mounted overlay (nil when none open)
	Overlays      map[string]*sdk.UIPanelSpec // all open named overlays by panel ID
	OverlayCount  int                         // number of SetOverlay calls
	Patches       []sdk.UIOverlayPatch
	HUD           []sdk.HUDActionSpec
	ClientPackets []ClientPacket
	CodexQueries  []string
	Audio         []string
	Meshes        map[string]sdk.CustomMeshSpec
	Prefs         map[string]string

	// Server role.
	Notices          []sdk.HostNoticeCmd
	Packets          []sdk.HostServerPacketCmd
	Chats            []sdk.HostChatCmd
	Monsters         map[string]sdk.MonsterSpec
	Spawns           map[string]sdk.MonsterSpawnSpec
	Items            map[string]sdk.ItemSpec
	Recipes          map[string]sdk.RecipeSpec
	NPCs             map[string]sdk.NPCSpec
	Rooms            map[[2]int]sdk.RoomSpec
	GatherNodes      map[string]sdk.GatherNodeSpec
	CodexConfig      *sdk.CodexConfigSpec
	CodexEntries     map[string]sdk.CodexEntrySpec
	PlayerSkills     map[string]sdk.PlayerSkillSpec
	Jobs             map[string]sdk.JobDefSpec
	StarterConfig    *sdk.StarterConfigSpec
	SpawnRoom        *sdk.SpawnRoomSpec
	HeroLoadouts     []sdk.SetHeroLoadoutCmd
	GivenItems       []sdk.GiveItemCmd
	TakenItems       []sdk.TakeItemCmd
	Teleports        []sdk.TeleportHeroCmd
	DynamicProps     map[string]sdk.SpawnPropCmd
	RoomAtmospheres  map[[2]int]sdk.RoomAtmosphereSpec
	MapConfig        *sdk.MapConfigSpec
	MapRegions       map[string]sdk.MapRegionSpec
	MapRoutes        map[string]sdk.MapRouteSpec
	MapMarkers       map[string]sdk.MapMarkerSpec
	Actors           map[string]sdk.ActorSpec
	VFXLog           []sdk.VFXSpec
	FloatingTexts    []sdk.FloatingTextSpec
	GroundMarkers    map[string]sdk.GroundMarkerSpec
	HeroHotbars      map[string][]sdk.HotbarSlotSpec
	HeroPoses        map[string]sdk.SetHeroPoseCmd
	HeroControlLocks map[string]bool
	KV               map[string][]byte

	// Online heroes, by ID. Gold/XP/Level/Room/Pos are updated by GiveGold/GiveXP/TeleportHero.
	Heroes        map[string]sdk.HeroSnapshot
	HeroInventory map[string]map[string]int
	HeroInBattle  map[string]bool

	// Fail forces a hostcall (by name, e.g. "server_kv_set") to return a code.
	Fail map[string]int32
}

// New returns an empty host. Most tests should use Install.
func New() *Host {
	h := &Host{}
	h.Reset()
	h.Heroes = map[string]sdk.HeroSnapshot{}
	h.HeroInventory = map[string]map[string]int{}
	h.HeroInBattle = map[string]bool{}
	h.KV = map[string][]byte{}
	h.Prefs = map[string]string{}
	h.Fail = map[string]int32{}
	return h
}

// Install creates a host, routes all SDK hostcalls to it for the duration of
// the test, and restores the previous host on cleanup.
func Install(t testing.TB) *Host {
	t.Helper()
	h := New()
	restore := sdk.SetTestHost(h)
	t.Cleanup(restore)
	return h
}

// Reset clears recordings (calls, notices, packets, overlays, registries) but
// keeps KV, Prefs, online heroes, hero inventories, battle flags, and forced failures.
func (h *Host) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Calls, h.Logs = nil, nil
	h.Overlay, h.OverlayCount, h.Patches, h.HUD = nil, 0, nil, nil
	h.Overlays = map[string]*sdk.UIPanelSpec{}
	h.ClientPackets, h.CodexQueries, h.Audio = nil, nil, nil
	h.Notices, h.Packets, h.Chats = nil, nil, nil
	h.GivenItems, h.TakenItems, h.Teleports, h.HeroLoadouts = nil, nil, nil, nil
	h.CodexConfig, h.StarterConfig, h.SpawnRoom = nil, nil, nil
	h.Meshes = map[string]sdk.CustomMeshSpec{}
	h.Monsters = map[string]sdk.MonsterSpec{}
	h.Spawns = map[string]sdk.MonsterSpawnSpec{}
	h.Items = map[string]sdk.ItemSpec{}
	h.Recipes = map[string]sdk.RecipeSpec{}
	h.NPCs = map[string]sdk.NPCSpec{}
	h.Rooms = map[[2]int]sdk.RoomSpec{}
	h.GatherNodes = map[string]sdk.GatherNodeSpec{}
	h.CodexEntries = map[string]sdk.CodexEntrySpec{}
	h.PlayerSkills = map[string]sdk.PlayerSkillSpec{}
	h.Jobs = map[string]sdk.JobDefSpec{}
	h.DynamicProps = map[string]sdk.SpawnPropCmd{}
	h.RoomAtmospheres = map[[2]int]sdk.RoomAtmosphereSpec{}
	h.MapConfig = nil
	h.MapRegions = map[string]sdk.MapRegionSpec{}
	h.MapRoutes = map[string]sdk.MapRouteSpec{}
	h.MapMarkers = map[string]sdk.MapMarkerSpec{}
	h.Actors = map[string]sdk.ActorSpec{}
	h.VFXLog = nil
	h.FloatingTexts = nil
	h.GroundMarkers = map[string]sdk.GroundMarkerSpec{}
	h.HeroHotbars = map[string][]sdk.HotbarSlotSpec{}
	h.HeroPoses = map[string]sdk.SetHeroPoseCmd{}
	h.HeroControlLocks = map[string]bool{}
}

// FireActorClick dispatches ev to p if p implements sdk.CombatInputPlugin.
func (h *Host) FireActorClick(p sdk.ServerPlugin, ev sdk.ActorClickEvent) {
	if cb, ok := p.(sdk.CombatInputPlugin); ok && cb != nil {
		cb.OnActorClick(ev)
	}
}

// FireGroundClick dispatches ev to p if p implements sdk.CombatInputPlugin.
func (h *Host) FireGroundClick(p sdk.ServerPlugin, ev sdk.GroundClickEvent) {
	if cb, ok := p.(sdk.CombatInputPlugin); ok && cb != nil {
		cb.OnGroundClick(ev)
	}
}

// FireHotbarPress dispatches ev to p if p implements sdk.CombatInputPlugin.
func (h *Host) FireHotbarPress(p sdk.ServerPlugin, ev sdk.HotbarPressEvent) {
	if cb, ok := p.(sdk.CombatInputPlugin); ok && cb != nil {
		cb.OnHotbarPress(ev)
	}
}

// FireMapRoomClick dispatches ev to p if p implements sdk.MapInputPlugin.
func (h *Host) FireMapRoomClick(p sdk.ServerPlugin, ev sdk.MapRoomClickEvent) {
	if cb, ok := p.(sdk.MapInputPlugin); ok && cb != nil {
		cb.OnMapRoomClick(ev)
	}
}

// FireMapMarkerClick dispatches ev to p if p implements sdk.MapInputPlugin.
func (h *Host) FireMapMarkerClick(p sdk.ServerPlugin, ev sdk.MapMarkerClickEvent) {
	if cb, ok := p.(sdk.MapInputPlugin); ok && cb != nil {
		cb.OnMapMarkerClick(ev)
	}
}

// SetOnline marks hero as online (replacing any previous snapshot).
func (h *Host) SetOnline(hero sdk.HeroSnapshot) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if hero.Level == 0 {
		hero.Level = 1
	}
	h.Heroes[hero.ID] = hero
}

// SetOffline removes a hero from the online set.
func (h *Host) SetOffline(heroID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.Heroes, heroID)
}

// SetHeroItem sets the bag quantity of itemID for heroID (qty <= 0 removes it).
func (h *Host) SetHeroItem(heroID, itemID string, qty int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.HeroInventory[heroID] == nil {
		if qty <= 0 {
			return
		}
		h.HeroInventory[heroID] = map[string]int{}
	}
	if qty <= 0 {
		delete(h.HeroInventory[heroID], itemID)
		return
	}
	h.HeroInventory[heroID][itemID] = qty
}

// HeroItemCount returns the recorded bag quantity of itemID for heroID.
func (h *Host) HeroItemCount(heroID, itemID string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.HeroInventory[heroID][itemID]
}

// SetInBattle sets whether heroID is currently in a tactical battle.
func (h *Host) SetInBattle(heroID string, inBattle bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !inBattle {
		delete(h.HeroInBattle, heroID)
		return
	}
	h.HeroInBattle[heroID] = true
}

// OverlayByID returns the open overlay panel matching panelID, or nil.
func (h *Host) OverlayByID(panelID string) *sdk.UIPanelSpec {
	h.mu.Lock()
	defer h.mu.Unlock()
	if panelID == "" {
		panelID = "default"
	}
	return h.Overlays[panelID]
}

// FindNodeIn returns the node with nodeID inside the open overlay panelID, or nil.
func (h *Host) FindNodeIn(panelID, nodeID string) *sdk.UINodeSpec {
	h.mu.Lock()
	defer h.mu.Unlock()
	if panelID == "" {
		panelID = "default"
	}
	ov := h.Overlays[panelID]
	if ov == nil || ov.Root == nil {
		return nil
	}
	return findNode(ov.Root, nodeID)
}

// Hero returns the online snapshot for id (zero value if offline).
func (h *Host) Hero(id string) sdk.HeroSnapshot {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.Heroes[id]
}

// NoticesFor returns the notices sent to heroID (including broadcasts).
func (h *Host) NoticesFor(heroID string) []sdk.HostNoticeCmd {
	h.mu.Lock()
	defer h.mu.Unlock()
	var out []sdk.HostNoticeCmd
	for _, n := range h.Notices {
		if n.BroadcastAll || n.TargetHeroID == heroID {
			out = append(out, n)
		}
	}
	return out
}

// ChatsFor returns the plugin chat lines heroID received (including broadcasts).
func (h *Host) ChatsFor(heroID string) []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	var out []string
	for _, c := range h.Chats {
		if c.BroadcastAll || c.TargetHeroID == heroID {
			out = append(out, c.Text)
		}
	}
	return out
}

// PacketsFor decodes every JSON packet on channel delivered to heroID
// (direct sends and broadcasts) as T.
func PacketsFor[T any](h *Host, heroID string, channel uint16) []T {
	h.mu.Lock()
	defer h.mu.Unlock()
	var out []T
	for _, p := range h.Packets {
		if p.Channel != channel || !(p.BroadcastAll || p.TargetHeroID == heroID) {
			continue
		}
		var v T
		if json.Unmarshal(p.Payload, &v) == nil {
			out = append(out, v)
		}
	}
	return out
}

// FindNode returns the node with id in the current overlay, or nil.
func (h *Host) FindNode(id string) *sdk.UINodeSpec {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.Overlay == nil || h.Overlay.Root == nil {
		return nil
	}
	return findNode(h.Overlay.Root, id)
}

// OverlayText concatenates every label/button text in the current overlay,
// which is handy for strings.Contains assertions.
func (h *Host) OverlayText() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.Overlay == nil || h.Overlay.Root == nil {
		return ""
	}
	var b strings.Builder
	var walk func(n *sdk.UINodeSpec)
	walk = func(n *sdk.UINodeSpec) {
		if n.Text != "" {
			b.WriteString(n.Text)
			b.WriteByte('\n')
		}
		for i := range n.Children {
			walk(&n.Children[i])
		}
	}
	walk(h.Overlay.Root)
	return b.String()
}

func findNode(n *sdk.UINodeSpec, id string) *sdk.UINodeSpec {
	if n.ID == id {
		return n
	}
	for i := range n.Children {
		if f := findNode(&n.Children[i], id); f != nil {
			return f
		}
	}
	return nil
}

func decode[T any](raw []byte) (T, bool) {
	var v T
	return v, json.Unmarshal(raw, &v) == nil
}

func kvBytes(m map[string][]byte) int {
	n := 0
	for k, v := range m {
		n += len(k) + len(v)
	}
	return n
}

func listKeys[V any](m map[string]V, prefix string, limit, maxLimit int) []byte {
	if limit <= 0 || limit > maxLimit {
		limit = maxLimit
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	if len(keys) > limit {
		keys = keys[:limit]
	}
	raw, _ := json.Marshal(keys)
	return raw
}

// HostCall implements sdk.TestHost.
func (h *Host) HostCall(name string, payload []byte) (int32, []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Calls = append(h.Calls, name)
	if code, ok := h.Fail[name]; ok {
		return code, nil
	}
	const invalid = sdk.CodeInvalidArg

	switch name {
	case "log":
		h.Logs = append(h.Logs, string(payload))

	// ---- client ----
	case "send_packet":
		p, ok := decode[ClientPacket](payload)
		if !ok || len(p.Payload) > sdk.MaxPacketBytes {
			return invalid, nil
		}
		h.ClientPackets = append(h.ClientPackets, p)
	case "ui_set_hud_actions":
		a, ok := decode[[]sdk.HUDActionSpec](payload)
		if !ok {
			return invalid, nil
		}
		for i := range a {
			if len(a[i].Badge) > sdk.MaxHUDBadgeLen {
				a[i].Badge = a[i].Badge[:sdk.MaxHUDBadgeLen] // same truncation as the client
			}
		}
		h.HUD = a
	case "ui_set_overlay":
		h.OverlayCount++
		if len(payload) == 0 {
			h.Overlay = nil
			h.Overlays = map[string]*sdk.UIPanelSpec{}
			return 0, nil
		}
		p, ok := decode[sdk.UIPanelSpec](payload)
		if !ok || sdk.ValidateOverlay(p) != nil {
			return invalid, nil
		}
		id := p.ID
		if id == "" {
			id = "default"
		}
		if _, exists := h.Overlays[id]; !exists && len(h.Overlays) >= sdk.MaxOverlays {
			return sdk.CodeQuotaOrBackend, nil
		}
		h.Overlays[id] = &p
		h.Overlay = &p
	case "ui_clear_overlay_id":
		c, ok := decode[sdk.ClearOverlayCmd](payload)
		if !ok || c.PanelID == "" {
			return invalid, nil
		}
		if _, exists := h.Overlays[c.PanelID]; !exists {
			return sdk.CodeNotFound, nil
		}
		delete(h.Overlays, c.PanelID)
		if h.Overlay != nil {
			curID := h.Overlay.ID
			if curID == "" {
				curID = "default"
			}
			if curID == c.PanelID {
				h.Overlay = nil
				keys := make([]string, 0, len(h.Overlays))
				for k := range h.Overlays {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				if len(keys) > 0 {
					h.Overlay = h.Overlays[keys[0]]
				}
			}
		}
	case "ui_patch_overlay":
		p, ok := decode[sdk.UIOverlayPatch](payload)
		if !ok {
			return invalid, nil
		}
		pid := p.PanelID
		if pid == "" {
			pid = "default"
		}
		ov := h.Overlays[pid]
		if ov == nil || ov.Root == nil {
			return sdk.CodeNotFound, nil
		}
		for _, pt := range p.Patches {
			n := findNode(ov.Root, pt.ID)
			if n == nil {
				continue
			}
			if pt.Text != nil {
				n.Text = *pt.Text
			}
			if pt.Value != nil {
				n.Text = *pt.Value
			}
			if pt.Progress != nil {
				n.Progress = *pt.Progress
			}
			if pt.Color != nil {
				n.Color = *pt.Color
			}
			if pt.Hidden != nil {
				n.Hidden = *pt.Hidden
			}
		}
		h.Patches = append(h.Patches, p)
	case "ui_open_codex":
		h.CodexQueries = append(h.CodexQueries, string(payload))
	case "play_audio":
		h.Audio = append(h.Audio, string(payload))
	case "ui_register_mesh":
		m, ok := decode[sdk.CustomMeshSpec](payload)
		if !ok || !sdk.IsValidContentID(m.ID) {
			return invalid, nil
		}
		if _, exists := h.Meshes[m.ID]; !exists && len(h.Meshes) >= sdk.MaxCustomMeshes {
			return sdk.CodeQuotaOrBackend, nil
		}
		h.Meshes[m.ID] = m
	case "pref_set":
		c, _ := decode[sdk.KVCall](payload)
		if _, exists := h.Prefs[c.Key]; !exists && len(h.Prefs) >= sdk.MaxPrefsPerRealm {
			return sdk.CodeQuotaOrBackend, nil
		}
		h.Prefs[c.Key] = string(c.Value)
	case "pref_get":
		c, _ := decode[sdk.KVCall](payload)
		v, ok := h.Prefs[c.Key]
		if !ok {
			return -1, nil
		}
		return 0, []byte(v)
	case "pref_delete":
		c, _ := decode[sdk.KVCall](payload)
		delete(h.Prefs, c.Key)
	case "pref_list":
		c, _ := decode[sdk.KVCall](payload)
		return 0, listKeys(h.Prefs, c.Prefix, c.Limit, sdk.MaxPrefListLimit)

	// ---- server ----
	case "server_send_packet":
		p, ok := decode[sdk.HostServerPacketCmd](payload)
		if !ok || len(p.Payload) > sdk.MaxPacketBytes {
			return invalid, nil
		}
		h.Packets = append(h.Packets, p)
	case "server_send_notice":
		n, ok := decode[sdk.HostNoticeCmd](payload)
		if !ok || n.Message == "" {
			return invalid, nil
		}
		h.Notices = append(h.Notices, n)
	case "server_send_chat":
		c, ok := decode[sdk.HostChatCmd](payload)
		if !ok || c.Text == "" {
			return invalid, nil
		}
		if !c.BroadcastAll {
			if _, online := h.Heroes[c.TargetHeroID]; !online {
				return sdk.CodeHeroOffline, nil
			}
		}
		h.Chats = append(h.Chats, c)
	case "server_register_monster":
		m, ok := decode[sdk.MonsterSpec](payload)
		if !ok || !sdk.IsValidContentID(m.ID) {
			return invalid, nil
		}
		h.Monsters[m.ID] = m
	case "server_spawn_monster":
		s, ok := decode[sdk.MonsterSpawnSpec](payload)
		if !ok || !sdk.IsValidContentID(s.ID) || !sdk.IsValidContentID(s.ArchetypeID) {
			return invalid, nil
		}
		h.Spawns[s.ID] = s
	case "server_remove_monster_spawn":
		c, ok := decode[sdk.RemoveMonsterSpawnCmd](payload)
		if !ok {
			return invalid, nil
		}
		if _, exists := h.Spawns[c.ID]; !exists {
			return sdk.CodeNotFound, nil
		}
		delete(h.Spawns, c.ID)
	case "server_register_item":
		it, ok := decode[sdk.ItemSpec](payload)
		if !ok || !sdk.IsValidContentID(it.ID) {
			return invalid, nil
		}
		h.Items[it.ID] = it
	case "server_register_recipe":
		r, ok := decode[sdk.RecipeSpec](payload)
		if !ok || !sdk.IsValidContentID(r.ID) {
			return invalid, nil
		}
		h.Recipes[r.ID] = r
	case "server_register_npc":
		n, ok := decode[sdk.NPCSpec](payload)
		if !ok {
			return invalid, nil
		}
		h.NPCs[n.ID] = n
	case "server_register_room":
		r, ok := decode[sdk.RoomSpec](payload)
		if !ok || len(payload) > sdk.MaxRoomSpecBytes {
			return invalid, nil
		}
		if len(r.Liquids) > sdk.MaxRoomLiquids || len(r.Lights) > sdk.MaxRoomLights ||
			len(r.Emitters) > sdk.MaxRoomEmitters || len(r.Zones) > sdk.MaxRoomZones {
			return invalid, nil
		}
		h.Rooms[[2]int{r.X, r.Y}] = r
	case "server_remove_room":
		c, _ := decode[sdk.RoomCoordCmd](payload)
		delete(h.Rooms, [2]int{c.X, c.Y})
	case "server_spawn_gather_node":
		g, ok := decode[sdk.GatherNodeSpec](payload)
		if !ok {
			return invalid, nil
		}
		h.GatherNodes[g.ID] = g
	case "server_register_codex_entry":
		e, ok := decode[sdk.CodexEntrySpec](payload)
		if !ok {
			return invalid, nil
		}
		h.CodexEntries[e.ID] = e
	case "server_give_item":
		g, ok := decode[sdk.GiveItemCmd](payload)
		if !ok || g.Quantity < 1 || g.Quantity > 99 {
			return invalid, nil
		}
		if _, online := h.Heroes[g.TargetHeroID]; !online {
			return sdk.CodeHeroOffline, nil
		}
		h.GivenItems = append(h.GivenItems, g)
		if h.HeroInventory[g.TargetHeroID] == nil {
			h.HeroInventory[g.TargetHeroID] = map[string]int{}
		}
		h.HeroInventory[g.TargetHeroID][g.ItemID] += g.Quantity
	case "server_take_item":
		c, ok := decode[sdk.TakeItemCmd](payload)
		if !ok || c.TargetHeroID == "" || !sdk.IsValidContentID(c.ItemID) || c.Quantity < 1 || c.Quantity > 999 {
			return invalid, nil
		}
		if _, online := h.Heroes[c.TargetHeroID]; !online {
			return sdk.CodeHeroOffline, nil
		}
		inv := h.HeroInventory[c.TargetHeroID]
		if inv[c.ItemID] < c.Quantity {
			return sdk.CodeNotEnoughItems, nil
		}
		inv[c.ItemID] -= c.Quantity
		if inv[c.ItemID] <= 0 {
			delete(inv, c.ItemID)
		}
		h.TakenItems = append(h.TakenItems, c)
	case "server_get_inventory":
		q, ok := decode[sdk.HeroQueryCmd](payload)
		if !ok || q.HeroID == "" {
			return invalid, nil
		}
		if _, online := h.Heroes[q.HeroID]; !online {
			return sdk.CodeHeroOffline, nil
		}
		inv := h.HeroInventory[q.HeroID]
		ids := make([]string, 0, len(inv))
		for id, qty := range inv {
			if qty > 0 {
				ids = append(ids, id)
			}
		}
		sort.Strings(ids)
		stacks := make([]sdk.ItemStackSpec, 0, len(ids))
		for _, id := range ids {
			stacks = append(stacks, sdk.ItemStackSpec{ItemID: id, Quantity: inv[id]})
		}
		raw, _ := json.Marshal(stacks)
		return 0, raw
	case "server_teleport_hero":
		c, ok := decode[sdk.TeleportHeroCmd](payload)
		if !ok || c.TargetHeroID == "" {
			return invalid, nil
		}
		hero, online := h.Heroes[c.TargetHeroID]
		if !online {
			return sdk.CodeHeroOffline, nil
		}
		if h.HeroInBattle[c.TargetHeroID] {
			return sdk.CodeInBattle, nil
		}
		room, exists := h.Rooms[[2]int{c.RoomX, c.RoomY}]
		if !exists {
			return sdk.CodeNotFound, nil
		}
		hero.RoomX = c.RoomX
		hero.RoomY = c.RoomY
		if c.UsePos {
			hero.PosX = c.Pos[0]
			hero.PosZ = c.Pos[2]
		} else {
			hero.PosX = room.SpawnPos[0]
			hero.PosZ = room.SpawnPos[2]
		}
		h.Heroes[c.TargetHeroID] = hero
		h.Teleports = append(h.Teleports, c)
	case "server_spawn_prop":
		c, ok := decode[sdk.SpawnPropCmd](payload)
		if !ok || !sdk.IsValidContentID(c.Prop.ID) {
			return invalid, nil
		}
		if _, exists := h.Rooms[[2]int{c.RoomX, c.RoomY}]; !exists {
			return sdk.CodeNotFound, nil
		}
		prev, existed := h.DynamicProps[c.Prop.ID]
		if !existed && len(h.DynamicProps) >= sdk.MaxDynamicProps {
			return sdk.CodeQuotaOrBackend, nil
		}
		if !existed || prev.RoomX != c.RoomX || prev.RoomY != c.RoomY {
			roomCount := 0
			for _, p := range h.DynamicProps {
				if p.RoomX == c.RoomX && p.RoomY == c.RoomY {
					roomCount++
				}
			}
			if roomCount >= sdk.MaxDynamicPropsPerRoom {
				return sdk.CodeQuotaOrBackend, nil
			}
		}
		h.DynamicProps[c.Prop.ID] = c
	case "server_remove_prop":
		c, ok := decode[sdk.RemovePropCmd](payload)
		if !ok || !sdk.IsValidContentID(c.ID) {
			return invalid, nil
		}
		if _, exists := h.DynamicProps[c.ID]; !exists {
			return sdk.CodeNotFound, nil
		}
		delete(h.DynamicProps, c.ID)
	case "server_set_room_atmosphere":
		c, ok := decode[sdk.SetRoomAtmosphereCmd](payload)
		if !ok {
			return invalid, nil
		}
		coord := [2]int{c.RoomX, c.RoomY}
		if _, exists := h.Rooms[coord]; !exists {
			return sdk.CodeNotFound, nil
		}
		if c.Atmosphere == nil {
			delete(h.RoomAtmospheres, coord)
		} else {
			if len(c.Atmosphere.Lights) > sdk.MaxRoomLights || len(c.Atmosphere.Emitters) > sdk.MaxRoomEmitters {
				return invalid, nil
			}
			h.RoomAtmospheres[coord] = *c.Atmosphere
		}
	case "server_set_map_marker":
		m, ok := decode[sdk.MapMarkerSpec](payload)
		if !ok || !sdk.IsValidContentID(m.ID) {
			return invalid, nil
		}
		if _, exists := h.Rooms[[2]int{m.RoomX, m.RoomY}]; !exists {
			return sdk.CodeNotFound, nil
		}
		if _, exists := h.MapMarkers[m.ID]; !exists && len(h.MapMarkers) >= sdk.MaxMapMarkers {
			return sdk.CodeQuotaOrBackend, nil
		}
		h.MapMarkers[m.ID] = m
	case "server_remove_map_marker":
		c, ok := decode[sdk.RemoveMapMarkerCmd](payload)
		if !ok || !sdk.IsValidContentID(c.ID) {
			return invalid, nil
		}
		if _, exists := h.MapMarkers[c.ID]; !exists {
			return sdk.CodeNotFound, nil
		}
		delete(h.MapMarkers, c.ID)
	case "server_spawn_actor":
		a, ok := decode[sdk.ActorSpec](payload)
		if !ok || !sdk.IsValidContentID(a.ID) {
			return invalid, nil
		}
		if _, exists := h.Rooms[[2]int{a.RoomX, a.RoomY}]; !exists {
			return sdk.CodeNotFound, nil
		}
		prev, existed := h.Actors[a.ID]
		if !existed && len(h.Actors) >= sdk.MaxDynamicActors {
			return sdk.CodeQuotaOrBackend, nil
		}
		if !existed || prev.RoomX != a.RoomX || prev.RoomY != a.RoomY {
			roomCount := 0
			for _, cur := range h.Actors {
				if cur.RoomX == a.RoomX && cur.RoomY == a.RoomY {
					roomCount++
				}
			}
			if roomCount >= sdk.MaxDynamicActorsPerRoom {
				return sdk.CodeQuotaOrBackend, nil
			}
		}
		h.Actors[a.ID] = a
	case "server_update_actor":
		p, ok := decode[sdk.ActorPatch](payload)
		if !ok || !sdk.IsValidContentID(p.ID) {
			return invalid, nil
		}
		cur, exists := h.Actors[p.ID]
		if !exists {
			return sdk.CodeNotFound, nil
		}
		if p.PosX != nil {
			cur.PosX = *p.PosX
		}
		if p.PosZ != nil {
			cur.PosZ = *p.PosZ
		}
		if p.Yaw != nil {
			cur.Yaw = *p.Yaw
		}
		if p.Name != nil {
			cur.Name = *p.Name
		}
		if p.HP != nil {
			cur.HP = *p.HP
		}
		if p.MaxHP != nil {
			cur.MaxHP = *p.MaxHP
		}
		if p.Shield != nil {
			cur.Shield = *p.Shield
		}
		if p.Pose != nil {
			cur.Pose = *p.Pose
		}
		if p.ReplaceStatuses || len(p.Statuses) > 0 {
			cur.Statuses = append([]sdk.ActorStatusSpec(nil), p.Statuses...)
		}
		if p.Targetable != nil {
			cur.Targetable = *p.Targetable
		}
		h.Actors[p.ID] = cur
	case "server_remove_actor":
		c, ok := decode[sdk.RemoveActorCmd](payload)
		if !ok || !sdk.IsValidContentID(c.ID) {
			return invalid, nil
		}
		if _, exists := h.Actors[c.ID]; !exists {
			return sdk.CodeNotFound, nil
		}
		delete(h.Actors, c.ID)
	case "server_play_vfx":
		v, ok := decode[sdk.VFXSpec](payload)
		if !ok {
			return invalid, nil
		}
		if _, exists := h.Rooms[[2]int{v.RoomX, v.RoomY}]; !exists {
			return sdk.CodeNotFound, nil
		}
		h.VFXLog = append(h.VFXLog, v)
	case "server_show_floating_text":
		f, ok := decode[sdk.FloatingTextSpec](payload)
		if !ok || f.Text == "" {
			return invalid, nil
		}
		if _, exists := h.Rooms[[2]int{f.RoomX, f.RoomY}]; !exists {
			return sdk.CodeNotFound, nil
		}
		h.FloatingTexts = append(h.FloatingTexts, f)
	case "server_set_ground_marker":
		m, ok := decode[sdk.GroundMarkerSpec](payload)
		if !ok || !sdk.IsValidContentID(m.ID) {
			return invalid, nil
		}
		if _, exists := h.Rooms[[2]int{m.RoomX, m.RoomY}]; !exists {
			return sdk.CodeNotFound, nil
		}
		prev, existed := h.GroundMarkers[m.ID]
		if !existed || prev.RoomX != m.RoomX || prev.RoomY != m.RoomY {
			roomCount := 0
			for _, cur := range h.GroundMarkers {
				if cur.RoomX == m.RoomX && cur.RoomY == m.RoomY {
					roomCount++
				}
			}
			if roomCount >= sdk.MaxGroundMarkers {
				return sdk.CodeQuotaOrBackend, nil
			}
		}
		h.GroundMarkers[m.ID] = m
	case "server_remove_ground_marker":
		c, ok := decode[sdk.RemoveGroundMarkerCmd](payload)
		if !ok || !sdk.IsValidContentID(c.ID) {
			return invalid, nil
		}
		if _, exists := h.GroundMarkers[c.ID]; !exists {
			return sdk.CodeNotFound, nil
		}
		delete(h.GroundMarkers, c.ID)
	case "server_clear_ground_markers":
		c, ok := decode[sdk.ClearGroundMarkersCmd](payload)
		if !ok {
			return invalid, nil
		}
		if _, exists := h.Rooms[[2]int{c.RoomX, c.RoomY}]; !exists {
			return sdk.CodeNotFound, nil
		}
		for id, cur := range h.GroundMarkers {
			if cur.RoomX != c.RoomX || cur.RoomY != c.RoomY {
				continue
			}
			if c.OwnerHeroID == "" || cur.OwnerHeroID == c.OwnerHeroID {
				delete(h.GroundMarkers, id)
			}
		}
	case "server_set_hero_hotbar":
		c, ok := decode[sdk.SetHeroHotbarCmd](payload)
		if !ok || c.TargetHeroID == "" || len(c.Slots) > sdk.MaxHotbarSlots {
			return invalid, nil
		}
		if _, online := h.Heroes[c.TargetHeroID]; !online {
			return sdk.CodeHeroOffline, nil
		}
		if len(c.Slots) == 0 {
			delete(h.HeroHotbars, c.TargetHeroID)
		} else {
			h.HeroHotbars[c.TargetHeroID] = append([]sdk.HotbarSlotSpec(nil), c.Slots...)
		}
	case "server_set_hero_pose":
		c, ok := decode[sdk.SetHeroPoseCmd](payload)
		if !ok || c.TargetHeroID == "" || c.Pose == "" {
			return invalid, nil
		}
		if _, online := h.Heroes[c.TargetHeroID]; !online {
			return sdk.CodeHeroOffline, nil
		}
		h.HeroPoses[c.TargetHeroID] = c
	case "server_set_hero_control_lock":
		c, ok := decode[sdk.SetHeroControlLockCmd](payload)
		if !ok || c.TargetHeroID == "" {
			return invalid, nil
		}
		if _, online := h.Heroes[c.TargetHeroID]; !online {
			return sdk.CodeHeroOffline, nil
		}
		h.HeroControlLocks[c.TargetHeroID] = c.LockMove
	case "server_set_map_config":
		c, ok := decode[sdk.MapConfigSpec](payload)
		if !ok {
			return invalid, nil
		}
		h.MapConfig = &c
	case "server_set_map_region":
		r, ok := decode[sdk.MapRegionSpec](payload)
		if !ok || !sdk.IsValidContentID(r.ID) || r.MinRoomX > r.MaxRoomX || r.MinRoomY > r.MaxRoomY {
			return invalid, nil
		}
		if _, exists := h.MapRegions[r.ID]; !exists && len(h.MapRegions) >= sdk.MaxMapRegions {
			return sdk.CodeQuotaOrBackend, nil
		}
		h.MapRegions[r.ID] = r
	case "server_remove_map_region":
		c, ok := decode[sdk.RemoveMapRegionCmd](payload)
		if !ok || !sdk.IsValidContentID(c.ID) {
			return invalid, nil
		}
		if _, exists := h.MapRegions[c.ID]; !exists {
			return sdk.CodeNotFound, nil
		}
		delete(h.MapRegions, c.ID)
	case "server_set_map_route":
		r, ok := decode[sdk.MapRouteSpec](payload)
		if !ok || !sdk.IsValidContentID(r.ID) {
			return invalid, nil
		}
		if _, exists := h.MapRoutes[r.ID]; !exists && len(h.MapRoutes) >= sdk.MaxMapRoutes {
			return sdk.CodeQuotaOrBackend, nil
		}
		h.MapRoutes[r.ID] = r
	case "server_remove_map_route":
		c, ok := decode[sdk.RemoveMapRouteCmd](payload)
		if !ok || !sdk.IsValidContentID(c.ID) {
			return invalid, nil
		}
		if _, exists := h.MapRoutes[c.ID]; !exists {
			return sdk.CodeNotFound, nil
		}
		delete(h.MapRoutes, c.ID)
	case "server_configure_codex":
		c, ok := decode[sdk.CodexConfigSpec](payload)
		if !ok {
			return invalid, nil
		}
		h.CodexConfig = &c
	case "server_register_player_skill":
		s, ok := decode[sdk.PlayerSkillSpec](payload)
		if !ok {
			return invalid, nil
		}
		h.PlayerSkills[s.ID] = s
	case "server_register_job":
		j, ok := decode[sdk.JobDefSpec](payload)
		if !ok {
			return invalid, nil
		}
		h.Jobs[j.ID] = j
	case "server_set_starter_config":
		s, ok := decode[sdk.StarterConfigSpec](payload)
		if !ok {
			return invalid, nil
		}
		h.StarterConfig = &s
	case "server_set_spawn_room":
		s, ok := decode[sdk.SpawnRoomSpec](payload)
		if !ok {
			return invalid, nil
		}
		h.SpawnRoom = &s
	case "server_set_hero_loadout":
		c, ok := decode[sdk.SetHeroLoadoutCmd](payload)
		if !ok {
			return invalid, nil
		}
		h.HeroLoadouts = append(h.HeroLoadouts, c)
	case "server_give_currency":
		g, ok := decode[sdk.GiveCurrencyCmd](payload)
		if !ok || g.XP < 0 {
			return invalid, nil
		}
		hero, online := h.Heroes[g.TargetHeroID]
		if !online {
			return sdk.CodeHeroOffline, nil
		}
		if hero.Gold+g.Gold < 0 {
			return sdk.CodeNotEnoughGold, nil
		}
		hero.Gold += g.Gold
		hero.XP += g.XP // level-ups are not simulated
		h.Heroes[g.TargetHeroID] = hero
	case "server_get_hero":
		q, _ := decode[sdk.HeroQueryCmd](payload)
		hero, online := h.Heroes[q.HeroID]
		if !online {
			return -1, nil
		}
		raw, _ := json.Marshal(hero)
		return 0, raw
	case "server_kv_set":
		c, _ := decode[sdk.KVCall](payload)
		if !sdk.IsValidStorageKey(c.Key, sdk.MaxKVKeyLen) || len(c.Value) > sdk.MaxKVValueBytes {
			return invalid, nil
		}
		old, exists := h.KV[c.Key]
		if !exists && len(h.KV) >= sdk.MaxKVKeysPerRealm {
			return sdk.CodeQuotaOrBackend, nil
		}
		if kvBytes(h.KV)-len(old)+len(c.Value)+boolInt(!exists)*len(c.Key) > sdk.MaxKVTotalBytes {
			return sdk.CodeQuotaOrBackend, nil
		}
		h.KV[c.Key] = append([]byte(nil), c.Value...)
	case "server_kv_get":
		c, _ := decode[sdk.KVCall](payload)
		v, ok := h.KV[c.Key]
		if !ok {
			return -1, nil
		}
		return 0, append([]byte(nil), v...)
	case "server_kv_delete":
		c, _ := decode[sdk.KVCall](payload)
		delete(h.KV, c.Key)
	case "server_kv_list":
		c, _ := decode[sdk.KVCall](payload)
		return 0, listKeys(h.KV, c.Prefix, c.Limit, sdk.MaxKVListLimit)

	default:
		// Accepted without recording: configure_codex, starter config, skills,
		// jobs, loadouts, NPC results, character creation, spawn room.
	}
	return 0, nil
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// String summarizes the recorded state, for test failure messages.
func (h *Host) String() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return fmt.Sprintf("sdktest.Host{calls=%d notices=%d packets=%d chats=%d spawns=%d kv=%d online=%d}",
		len(h.Calls), len(h.Notices), len(h.Packets), len(h.Chats), len(h.Spawns), len(h.KV), len(h.Heroes))
}
