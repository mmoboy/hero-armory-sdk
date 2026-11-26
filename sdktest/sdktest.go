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
	Overlay       *sdk.UIPanelSpec // current overlay (nil when cleared)
	OverlayCount  int              // number of SetOverlay calls
	Patches       []sdk.UIOverlayPatch
	HUD           []sdk.HUDActionSpec
	ClientPackets []ClientPacket
	CodexQueries  []string
	Audio         []string
	Meshes        map[string]sdk.CustomMeshSpec
	Prefs         map[string]string

	// Server role.
	Notices      []sdk.HostNoticeCmd
	Packets      []sdk.HostServerPacketCmd
	Chats        []sdk.HostChatCmd
	Monsters     map[string]sdk.MonsterSpec
	Spawns       map[string]sdk.MonsterSpawnSpec
	Items        map[string]sdk.ItemSpec
	Recipes      map[string]sdk.RecipeSpec
	NPCs         map[string]sdk.NPCSpec
	Rooms        map[[2]int]sdk.RoomSpec
	GatherNodes  map[string]sdk.GatherNodeSpec
	CodexEntries map[string]sdk.CodexEntrySpec
	GivenItems   []sdk.GiveItemCmd
	KV           map[string][]byte

	// Online heroes, by ID. Gold/XP/Level are updated by GiveGold/GiveXP.
	Heroes map[string]sdk.HeroSnapshot

	// Fail forces a hostcall (by name, e.g. "server_kv_set") to return a code.
	Fail map[string]int32
}

// New returns an empty host. Most tests should use Install.
func New() *Host {
	h := &Host{}
	h.Reset()
	h.Heroes = map[string]sdk.HeroSnapshot{}
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
// keeps KV, Prefs, online heroes, and forced failures.
func (h *Host) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Calls, h.Logs = nil, nil
	h.Overlay, h.OverlayCount, h.Patches, h.HUD = nil, 0, nil, nil
	h.ClientPackets, h.CodexQueries, h.Audio = nil, nil, nil
	h.Notices, h.Packets, h.Chats, h.GivenItems = nil, nil, nil, nil
	h.Meshes = map[string]sdk.CustomMeshSpec{}
	h.Monsters = map[string]sdk.MonsterSpec{}
	h.Spawns = map[string]sdk.MonsterSpawnSpec{}
	h.Items = map[string]sdk.ItemSpec{}
	h.Recipes = map[string]sdk.RecipeSpec{}
	h.NPCs = map[string]sdk.NPCSpec{}
	h.Rooms = map[[2]int]sdk.RoomSpec{}
	h.GatherNodes = map[string]sdk.GatherNodeSpec{}
	h.CodexEntries = map[string]sdk.CodexEntrySpec{}
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
			return 0, nil
		}
		p, ok := decode[sdk.UIPanelSpec](payload)
		if !ok || sdk.ValidateOverlay(p) != nil {
			return invalid, nil
		}
		h.Overlay = &p
	case "ui_patch_overlay":
		p, ok := decode[sdk.UIOverlayPatch](payload)
		if !ok {
			return invalid, nil
		}
		if h.Overlay == nil || h.Overlay.ID != p.PanelID || h.Overlay.Root == nil {
			return sdk.CodeNotFound, nil
		}
		for _, pt := range p.Patches {
			n := findNode(h.Overlay.Root, pt.ID)
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
