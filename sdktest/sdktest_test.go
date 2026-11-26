package sdktest_test

import (
	"errors"
	"strings"
	"testing"

	sdk "github.com/mmoboy/hero-armory-sdk"
	"github.com/mmoboy/hero-armory-sdk/sdktest"
)

func TestOfflineSemanticsAndCurrency(t *testing.T) {
	h := sdktest.Install(t)
	if rc := sdk.GiveItemToHero("ghost", "mat_x", 1); rc != sdk.CodeHeroOffline {
		t.Fatalf("give to offline hero: rc=%d", rc)
	}
	h.SetOnline(sdk.HeroSnapshot{ID: "h1", Name: "Ayla", Gold: 50})
	if rc := sdk.GiveItemToHero("h1", "mat_x", 3); rc != 0 || len(h.GivenItems) != 1 {
		t.Fatalf("give to online hero: rc=%d given=%v", rc, h.GivenItems)
	}
	if err := sdk.GiveGold("h1", 25); err != nil {
		t.Fatal(err)
	}
	if err := sdk.GiveGold("h1", -100); !errors.Is(err, sdk.ErrNotEnoughGold) {
		t.Fatalf("overcharge: %v", err)
	}
	if err := sdk.GiveXP("h1", 0); !errors.Is(err, sdk.ErrInvalidArg) {
		t.Fatalf("zero xp: %v", err)
	}
	if err := sdk.GiveXP("ghost", 10); !errors.Is(err, sdk.ErrHeroOffline) {
		t.Fatalf("xp offline: %v", err)
	}
	hero, found, err := sdk.GetHero("h1")
	if err != nil || !found || hero.Gold != 75 {
		t.Fatalf("GetHero: %+v found=%v err=%v", hero, found, err)
	}
	if _, found, _ := sdk.GetHero("ghost"); found {
		t.Fatal("offline hero should not be found")
	}
}

func TestSpawnRemovalChatAndKV(t *testing.T) {
	h := sdktest.Install(t)
	sdk.SpawnMonster(sdk.MonsterSpawnSpec{ID: "rift_1", ArchetypeID: "wraith", PackSize: 2})
	if err := sdk.RemoveMonsterSpawn("rift_1"); err != nil || len(h.Spawns) != 0 {
		t.Fatalf("remove: %v spawns=%v", err, h.Spawns)
	}
	if err := sdk.RemoveMonsterSpawn("rift_1"); !errors.Is(err, sdk.ErrNotFound) {
		t.Fatalf("remove twice: %v", err)
	}

	h.SetOnline(sdk.HeroSnapshot{ID: "h1"})
	if err := sdk.SendChatToHero("h1", "Rift open \u2014 go!"); err != nil {
		t.Fatal(err)
	}
	if got := h.ChatsFor("h1"); len(got) != 1 || strings.ContainsAny(got[0], "\u2014") {
		t.Fatalf("chat not sanitized: %q", got)
	}
	if err := sdk.BroadcastChat("   "); !errors.Is(err, sdk.ErrInvalidArg) {
		t.Fatalf("empty chat: %v", err)
	}

	if err := sdk.KVSetJSON("hero/h1", map[string]int{"renown": 5}); err != nil {
		t.Fatal(err)
	}
	var v map[string]int
	if found, err := sdk.KVGetJSON("hero/h1", &v); !found || err != nil || v["renown"] != 5 {
		t.Fatalf("kv roundtrip: %v %v %v", v, found, err)
	}
	keys, _ := sdk.KVList("hero/", 0)
	if len(keys) != 1 {
		t.Fatalf("kv list: %v", keys)
	}
	h.Fail["server_kv_set"] = sdk.CodeQuotaOrBackend
	if err := sdk.KVSetString("x", "y"); !errors.Is(err, sdk.ErrQuotaOrBackend) {
		t.Fatalf("forced failure: %v", err)
	}
}

func TestOverlayValidationAndPatch(t *testing.T) {
	h := sdktest.Install(t)
	big := sdk.ColNode("root", 4)
	for i := 0; i < sdk.MaxUINodes; i++ {
		big.Children = append(big.Children, sdk.LabelNode("", "x", 10, [4]float32{}, false))
	}
	panel := sdk.UIPanelSpec{ID: "p", Visible: true, Root: &big}
	if err := sdk.ValidateOverlay(panel); err == nil {
		t.Fatal("129 nodes should fail validation")
	}
	if rc := sdk.SetOverlay(panel); rc != sdk.CodeInvalidArg || h.OverlayCount != 0 {
		t.Fatalf("oversized overlay must not reach the host: rc=%d", rc)
	}

	root := sdk.ColNode("root", 4,
		sdk.LabelNode("timer", "5:00", 10, [4]float32{}, false),
		sdk.ProgressNode("bar", 0.1, 10, [4]float32{}),
	)
	if rc := sdk.SetOverlay(sdk.UIPanelSpec{ID: "hub", Visible: true, Root: &root}); rc != 0 {
		t.Fatalf("SetOverlay rc=%d", rc)
	}
	hidden := true
	err := sdk.PatchOverlay("hub",
		sdk.UINodePatch{ID: "timer", Text: sdk.PatchText("4:59"), Hidden: &hidden},
		sdk.UINodePatch{ID: "bar", Progress: sdk.PatchProgress(0.5)})
	if err != nil {
		t.Fatal(err)
	}
	if h.FindNode("timer").Text != "4:59" || !h.FindNode("timer").Hidden || h.FindNode("bar").Progress != 0.5 {
		t.Fatalf("patch not applied: %s", h.OverlayText())
	}
	if err := sdk.PatchOverlay("other", sdk.UINodePatch{ID: "timer"}); !errors.Is(err, sdk.ErrNotFound) {
		t.Fatalf("patch wrong panel: %v", err)
	}
}

func TestHUDBadgeTruncation(t *testing.T) {
	h := sdktest.Install(t)
	sdk.SetHUDActions([]sdk.HUDActionSpec{{ID: "a", Icon: "star", Badge: "RIFT 4:12"}})
	if h.HUD[0].Badge != "RIFT 4" {
		t.Fatalf("badge should be cut to %d chars like the client: %q", sdk.MaxHUDBadgeLen, h.HUD[0].Badge)
	}
}

func TestInventoryInspectionAndTakeItem(t *testing.T) {
	h := sdktest.Install(t)
	if _, err := sdk.GetHeroInventory("ghost"); !errors.Is(err, sdk.ErrHeroOffline) {
		t.Fatalf("offline inventory: %v", err)
	}
	if err := sdk.TakeItemFromHero("ghost", "plug_rift_shard", 1); !errors.Is(err, sdk.ErrHeroOffline) {
		t.Fatalf("offline take: %v", err)
	}

	h.SetOnline(sdk.HeroSnapshot{ID: "h1", Name: "Ayla"})
	if rc := sdk.GiveItemToHero("h1", "plug_rift_shard", 5); rc != 0 {
		t.Fatalf("give shard rc=%d", rc)
	}
	h.SetHeroItem("h1", "plug_rift_core", 2)

	cnt, err := sdk.GetHeroItemCount("h1", "plug_rift_shard")
	if err != nil || cnt != 5 {
		t.Fatalf("shard count=%d err=%v", cnt, err)
	}
	inv, err := sdk.GetHeroInventory("h1")
	if err != nil || len(inv) != 2 || inv[0].ItemID != "plug_rift_core" || inv[1].ItemID != "plug_rift_shard" {
		t.Fatalf("inventory=%+v err=%v", inv, err)
	}
	if err := sdk.TakeItemFromHero("h1", "plug_rift_shard", 6); !errors.Is(err, sdk.ErrNotEnoughItems) {
		t.Fatalf("over-take: %v", err)
	}
	if h.HeroItemCount("h1", "plug_rift_shard") != 5 {
		t.Fatalf("failed take must not mutate bag: %d", h.HeroItemCount("h1", "plug_rift_shard"))
	}
	if err := sdk.TakeItemFromHero("h1", "plug_rift_shard", 3); err != nil {
		t.Fatalf("take 3: %v", err)
	}
	if cnt, err := sdk.GetHeroItemCount("h1", "plug_rift_shard"); err != nil || cnt != 2 {
		t.Fatalf("remaining=%d err=%v", cnt, err)
	}
	if err := sdk.TakeItemFromHero("h1", "BAD ID", 1); !errors.Is(err, sdk.ErrInvalidArg) {
		t.Fatalf("bad item id: %v", err)
	}
	if err := sdk.TakeItemFromHero("h1", "plug_rift_shard", 0); !errors.Is(err, sdk.ErrInvalidArg) {
		t.Fatalf("zero qty: %v", err)
	}
}

func TestTeleportHero(t *testing.T) {
	h := sdktest.Install(t)
	sdk.RegisterRoom(sdk.RoomSpec{X: 0, Y: 0, Name: "Courtyard", Biome: "sanctuary", SpawnPos: [3]float32{0, 0, 2.0}})
	sdk.RegisterRoom(sdk.RoomSpec{X: 1, Y: -1, Name: "Rift Site", Biome: "twilight", SpawnPos: [3]float32{1, 0, -2}})

	if err := sdk.TeleportHero("ghost", 1, -1, [3]float32{}, false); !errors.Is(err, sdk.ErrHeroOffline) {
		t.Fatalf("offline teleport: %v", err)
	}
	h.SetOnline(sdk.HeroSnapshot{ID: "h1", Name: "Ayla", RoomX: 0, RoomY: 0})
	if err := sdk.TeleportHero("h1", 9, 9, [3]float32{}, false); !errors.Is(err, sdk.ErrNotFound) {
		t.Fatalf("missing room teleport: %v", err)
	}
	h.SetInBattle("h1", true)
	if err := sdk.TeleportHero("h1", 1, -1, [3]float32{}, false); !errors.Is(err, sdk.ErrInBattle) {
		t.Fatalf("in-battle teleport: %v", err)
	}
	h.SetInBattle("h1", false)
	if err := sdk.TeleportHero("h1", 1, -1, [3]float32{}, false); err != nil {
		t.Fatalf("teleport default spawn: %v", err)
	}
	got := h.Hero("h1")
	if got.RoomX != 1 || got.RoomY != -1 || got.PosX != 1 || got.PosZ != -2 {
		t.Fatalf("unexpected hero pos after default teleport: %+v", got)
	}
	if err := sdk.TeleportHero("h1", 0, 0, [3]float32{3.5, 0, -4.5}, true); err != nil {
		t.Fatalf("teleport custom pos: %v", err)
	}
	got = h.Hero("h1")
	if got.RoomX != 0 || got.RoomY != 0 || got.PosX != 3.5 || got.PosZ != -4.5 {
		t.Fatalf("unexpected hero pos after custom teleport: %+v", got)
	}
}

func TestMultiNamedOverlaysAndClearID(t *testing.T) {
	h := sdktest.Install(t)
	rootA := sdk.ColNode("root_a", 4, sdk.LabelNode("lbl_a", "Hub", 12, [4]float32{}, false))
	rootB := sdk.ColNode("root_b", 4, sdk.LabelNode("lbl_b", "Rift 04:00", 12, [4]float32{}, false))
	if rc := sdk.SetOverlay(sdk.UIPanelSpec{ID: "guild_hub", Visible: true, Anchor: "center", Root: &rootA}); rc != 0 {
		t.Fatalf("mount guild_hub rc=%d", rc)
	}
	if rc := sdk.SetOverlay(sdk.UIPanelSpec{ID: "rift_bar", Visible: true, Anchor: "top_center", Root: &rootB}); rc != 0 {
		t.Fatalf("mount rift_bar rc=%d", rc)
	}
	if len(h.Overlays) != 2 {
		t.Fatalf("expected 2 concurrent overlays, got %d", len(h.Overlays))
	}
	if err := sdk.PatchOverlay("rift_bar", sdk.UINodePatch{ID: "lbl_b", Text: sdk.PatchText("Rift 03:59")}); err != nil {
		t.Fatalf("patch rift_bar: %v", err)
	}
	if h.FindNodeIn("rift_bar", "lbl_b").Text != "Rift 03:59" {
		t.Fatalf("rift_bar patch not applied")
	}
	if err := sdk.ClearOverlayID("guild_hub"); err != nil {
		t.Fatalf("clear guild_hub: %v", err)
	}
	if h.OverlayByID("guild_hub") != nil || h.OverlayByID("rift_bar") == nil {
		t.Fatalf("closing guild_hub must preserve rift_bar")
	}
	if err := sdk.ClearOverlayID("guild_hub"); !errors.Is(err, sdk.ErrNotFound) {
		t.Fatalf("clear missing overlay: %v", err)
	}
	// Quota check: fill up to MaxOverlays (4), 5th distinct ID must fail with CodeQuotaOrBackend.
	for _, id := range []string{"p2", "p3", "p4"} {
		r := sdk.ColNode("r", 2)
		if rc := sdk.SetOverlay(sdk.UIPanelSpec{ID: id, Visible: true, Root: &r}); rc != 0 {
			t.Fatalf("mount %s rc=%d", id, rc)
		}
	}
	r5 := sdk.ColNode("r5", 2)
	if rc := sdk.SetOverlay(sdk.UIPanelSpec{ID: "p5", Visible: true, Root: &r5}); rc != sdk.CodeQuotaOrBackend {
		t.Fatalf("5th overlay should hit quota, got rc=%d", rc)
	}
	sdk.ClearOverlay()
	if len(h.Overlays) != 0 || h.Overlay != nil {
		t.Fatalf("ClearOverlay should clear all overlays")
	}
}

func TestDynamicProps(t *testing.T) {
	h := sdktest.Install(t)
	sdk.RegisterRoom(sdk.RoomSpec{X: 0, Y: 0, Name: "Courtyard", Biome: "sanctuary"})
	sdk.RegisterRoom(sdk.RoomSpec{X: 1, Y: 0, Name: "Forge", Biome: "astral_forge"})

	obelisk := sdk.RoomPropSpec{ID: "rift_obelisk_1", ModelID: "rift_obelisk_mesh", X: 0, Y: 0, Z: 0, Scale: 1.5}
	if err := sdk.SpawnProp(9, 9, obelisk); !errors.Is(err, sdk.ErrNotFound) {
		t.Fatalf("spawn prop in missing room: %v", err)
	}
	if err := sdk.SpawnProp(0, 0, obelisk); err != nil {
		t.Fatalf("spawn prop: %v", err)
	}
	if p, ok := h.DynamicProps["rift_obelisk_1"]; !ok || p.RoomX != 0 || p.Prop.ModelID != "rift_obelisk_mesh" {
		t.Fatalf("prop not recorded: %+v", h.DynamicProps)
	}
	// Upsert moves prop to another room cleanly.
	if err := sdk.SpawnProp(1, 0, obelisk); err != nil {
		t.Fatalf("move prop: %v", err)
	}
	if h.DynamicProps["rift_obelisk_1"].RoomX != 1 {
		t.Fatalf("expected moved prop in room (1,0)")
	}
	if err := sdk.RemoveProp("rift_obelisk_1"); err != nil {
		t.Fatalf("remove prop: %v", err)
	}
	if err := sdk.RemoveProp("rift_obelisk_1"); !errors.Is(err, sdk.ErrNotFound) {
		t.Fatalf("remove prop twice: %v", err)
	}
}

func TestRoomAtmosphereAndMapMarkers(t *testing.T) {
	h := sdktest.Install(t)
	sdk.RegisterRoom(sdk.RoomSpec{
		X: 0, Y: 0, Name: "Observatory", Biome: "astral_forge",
		MapColor: [3]float32{0.2, 0.7, 0.95},
		MapIcon:  "star",
		Atmosphere: &sdk.RoomAtmosphereSpec{
			SkyZenith:  [3]float32{0.05, 0.08, 0.22},
			SkyHorizon: [3]float32{0.18, 0.34, 0.56},
			FogDensity: 0.015,
			Weather:    "starfall",
		},
		Liquids: []sdk.RoomLiquidSpec{
			{ID: "pool", Kind: "aether", MinX: -8, MaxX: 8, MinZ: -8, MaxZ: 8, Y: -0.2, Opacity: 0.75},
		},
		Lights: []sdk.RoomLightSpec{
			{ID: "prism", X: 0, Y: 3, Z: 0, Color: [3]float32{0.4, 0.9, 1.0}, Intensity: 1.8, Range: 12},
		},
		Emitters: []sdk.RoomEmitterSpec{
			{ID: "sparks", Kind: "sparkles", X: 0, Y: 1.5, Z: 0, Rate: 12},
		},
		Zones: []sdk.RoomTriggerZoneSpec{
			{ID: "rune_pad", Shape: "circle", X: 0, Z: 0, Radius: 3.0, Label: "Starlight Pad"},
		},
	})

	if err := sdk.SetRoomAtmosphere(9, 9, sdk.RoomAtmosphereSpec{Weather: "void"}); !errors.Is(err, sdk.ErrNotFound) {
		t.Fatalf("atmosphere missing room: %v", err)
	}
	if err := sdk.SetRoomAtmosphere(0, 0, sdk.RoomAtmosphereSpec{Weather: "hurricane"}); !errors.Is(err, sdk.ErrInvalidArg) {
		t.Fatalf("invalid weather should fail: %v", err)
	}
	riftAtmo := sdk.RoomAtmosphereSpec{
		SkyZenith:        [3]float32{0.12, 0.02, 0.22},
		FogColor:         [3]float32{0.24, 0.06, 0.38},
		FogDensity:       0.03,
		Weather:          "void",
		WeatherColor:     [4]float32{0.78, 0.28, 1.0, 1.0},
		WeatherIntensity: 0.85,
		Lights:           []sdk.RoomLightSpec{{ID: "rift_glow", X: 0, Y: 3, Z: 0, Color: [3]float32{0.8, 0.3, 1.0}, Intensity: 2.0, Range: 14}},
		Emitters:         []sdk.RoomEmitterSpec{{ID: "rift_swirl", Kind: "portal_swirl", X: 0, Y: 1.5, Z: 0, Rate: 20}},
	}
	if err := sdk.SetRoomAtmosphere(0, 0, riftAtmo); err != nil {
		t.Fatalf("set room atmosphere: %v", err)
	}
	if got, ok := h.RoomAtmospheres[[2]int{0, 0}]; !ok || got.Weather != "void" || len(got.Lights) != 1 || len(got.Emitters) != 1 {
		t.Fatalf("atmosphere not recorded: %+v", got)
	}
	if err := sdk.ClearRoomAtmosphere(0, 0); err != nil {
		t.Fatalf("clear room atmosphere: %v", err)
	}
	if _, ok := h.RoomAtmospheres[[2]int{0, 0}]; ok {
		t.Fatal("atmosphere override should be cleared")
	}

	if err := sdk.SetMapMarker(sdk.MapMarkerSpec{ID: "rift_1", RoomX: 9, RoomY: 9, Label: "Rift"}); !errors.Is(err, sdk.ErrNotFound) {
		t.Fatalf("marker missing room: %v", err)
	}
	if err := sdk.SetMapMarker(sdk.MapMarkerSpec{ID: "rift_1", RoomX: 0, RoomY: 0, Label: "Rift Incursion", Icon: "star", Color: [4]float32{0.8, 0.3, 1, 1}, Pulse: true}); err != nil {
		t.Fatalf("set map marker: %v", err)
	}
	if m, ok := h.MapMarkers["rift_1"]; !ok || !m.Pulse || m.Label != "Rift Incursion" {
		t.Fatalf("marker not recorded: %+v", h.MapMarkers)
	}
	if err := sdk.RemoveMapMarker("rift_1"); err != nil {
		t.Fatalf("remove map marker: %v", err)
	}
	if err := sdk.RemoveMapMarker("rift_1"); !errors.Is(err, sdk.ErrNotFound) {
		t.Fatalf("remove map marker twice: %v", err)
	}
}

type combatInputSpy struct {
	sdk.BaseServerPlugin
	actorClicks  []sdk.ActorClickEvent
	groundClicks []sdk.GroundClickEvent
	hotbarPress  []sdk.HotbarPressEvent
}

func (s *combatInputSpy) OnActorClick(ev sdk.ActorClickEvent) {
	s.actorClicks = append(s.actorClicks, ev)
}
func (s *combatInputSpy) OnGroundClick(ev sdk.GroundClickEvent) {
	s.groundClicks = append(s.groundClicks, ev)
}
func (s *combatInputSpy) OnHotbarPress(ev sdk.HotbarPressEvent) {
	s.hotbarPress = append(s.hotbarPress, ev)
}

func TestCombatPrimitivesAPI5(t *testing.T) {
	h := sdktest.Install(t)
	sdk.RegisterRoom(sdk.RoomSpec{X: 0, Y: 0, Name: "Arena", Biome: "sanctuary"})
	h.SetOnline(sdk.HeroSnapshot{ID: "h1", Name: "Ayla", RoomX: 0, RoomY: 0})

	// Dynamic actors: spawn, patch, remove.
	if err := sdk.SpawnActor(sdk.ActorSpec{ID: "sentinel_1", RoomX: 9, RoomY: 9, Name: "Ghost"}); !errors.Is(err, sdk.ErrNotFound) {
		t.Fatalf("spawn actor missing room: %v", err)
	}
	if err := sdk.SpawnActor(sdk.ActorSpec{
		ID: "sentinel_1", RoomX: 0, RoomY: 0, PosX: 4, PosZ: -3, Name: "Astral Sentinel",
		Kind: "sentinel", Team: "enemy", HP: 120, MaxHP: 120, Shield: 20, Targetable: true,
		Statuses: []sdk.ActorStatusSpec{{ID: "ward", Label: "Ward", Color: [3]float32{0.3, 0.8, 1}, Stacks: 2}},
	}); err != nil {
		t.Fatalf("spawn actor: %v", err)
	}
	if err := sdk.UpdateActor(sdk.ActorPatch{
		ID: "sentinel_1", PosX: sdk.PtrFloat32(5.5), MoveTo: true, HP: sdk.PtrInt(85), Shield: sdk.PtrInt(0),
		Pose: sdk.PtrString("slash"), ReplaceStatuses: true,
	}); err != nil {
		t.Fatalf("update actor: %v", err)
	}
	got := h.Actors["sentinel_1"]
	if got.PosX != 5.5 || got.HP != 85 || got.Shield != 0 || got.Pose != "slash" || len(got.Statuses) != 0 {
		t.Fatalf("actor patch mismatch: %+v", got)
	}
	if err := sdk.RemoveActor("sentinel_1"); err != nil {
		t.Fatalf("remove actor: %v", err)
	}
	if err := sdk.RemoveActor("sentinel_1"); !errors.Is(err, sdk.ErrNotFound) {
		t.Fatalf("remove actor twice: %v", err)
	}

	// VFX and floating text.
	if err := sdk.PlayVFX(sdk.VFXSpec{RoomX: 0, RoomY: 0, Kind: "bolt", FromX: 0, FromZ: 0, ToX: 5, ToZ: -3, DurationSec: 0.4}); err != nil {
		t.Fatalf("play vfx: %v", err)
	}
	if err := sdk.PlayVFX(sdk.VFXSpec{RoomX: 0, RoomY: 0, Kind: "bogus"}); !errors.Is(err, sdk.ErrInvalidArg) {
		t.Fatalf("invalid vfx kind: %v", err)
	}
	if err := sdk.ShowFloatingText(sdk.FloatingTextSpec{RoomX: 0, RoomY: 0, PosX: 5, PosY: 2, PosZ: -3, Text: "-35!", Critical: true}); err != nil {
		t.Fatalf("show floating text: %v", err)
	}
	if len(h.VFXLog) != 1 || len(h.FloatingTexts) != 1 || !h.FloatingTexts[0].Critical {
		t.Fatalf("vfx/floating text not recorded")
	}

	// Ground markers: room-wide + owner-filtered clear.
	if err := sdk.SetGroundMarker(sdk.GroundMarkerSpec{ID: "gm_public", RoomX: 0, RoomY: 0, Shape: "ring", PosX: 0, PosZ: 0, Radius: 3}); err != nil {
		t.Fatalf("set public ground marker: %v", err)
	}
	if err := sdk.SetGroundMarker(sdk.GroundMarkerSpec{ID: "gm_priv", RoomX: 0, RoomY: 0, Shape: "circle", PosX: 2, PosZ: 2, Radius: 1.5, OwnerHeroID: "h1"}); err != nil {
		t.Fatalf("set private ground marker: %v", err)
	}
	if err := sdk.ClearGroundMarkers(0, 0, "h1"); err != nil {
		t.Fatalf("clear owner ground markers: %v", err)
	}
	if _, ok := h.GroundMarkers["gm_priv"]; ok {
		t.Fatal("owner ground marker should be cleared")
	}
	if _, ok := h.GroundMarkers["gm_public"]; !ok {
		t.Fatal("room-wide ground marker should remain after owner-scoped clear")
	}
	if err := sdk.RemoveGroundMarker("gm_public"); err != nil {
		t.Fatalf("remove ground marker: %v", err)
	}

	// Hero hotbar, pose, control lock.
	if err := sdk.SetHeroHotbar("ghost", []sdk.HotbarSlotSpec{{Slot: 1, ID: "slash", Label: "Slash"}}); !errors.Is(err, sdk.ErrHeroOffline) {
		t.Fatalf("offline hotbar: %v", err)
	}
	if err := sdk.SetHeroHotbar("h1", []sdk.HotbarSlotSpec{
		{Slot: 1, ID: "star_slash", Label: "Star Slash", CostText: "2 AP", Active: true},
		{Slot: 10, ID: "end_turn", Label: "End Turn"},
	}); err != nil {
		t.Fatalf("set hotbar: %v", err)
	}
	if len(h.HeroHotbars["h1"]) != 2 {
		t.Fatalf("hotbar slots=%d", len(h.HeroHotbars["h1"]))
	}
	if err := sdk.SetHeroHotbar("h1", nil); err != nil || len(h.HeroHotbars["h1"]) != 0 {
		t.Fatalf("clear hotbar: err=%v slots=%v", err, h.HeroHotbars["h1"])
	}
	if err := sdk.SetHeroPose("h1", "channel", 1.5); err != nil || h.HeroPoses["h1"].Pose != "channel" {
		t.Fatalf("set hero pose: %v", err)
	}
	if err := sdk.SetHeroControlLock("h1", true); err != nil || !h.HeroControlLocks["h1"] {
		t.Fatalf("set control lock: %v", err)
	}

	// CombatInputPlugin dispatch helpers.
	spy := &combatInputSpy{}
	h.FireActorClick(spy, sdk.ActorClickEvent{HeroID: "h1", ActorID: "sentinel_1", Button: "left"})
	h.FireGroundClick(spy, sdk.GroundClickEvent{HeroID: "h1", PosX: 3, PosZ: -1, Button: "left"})
	h.FireHotbarPress(spy, sdk.HotbarPressEvent{HeroID: "h1", Slot: 1, ActionID: "star_slash"})
	if len(spy.actorClicks) != 1 || len(spy.groundClicks) != 1 || len(spy.hotbarPress) != 1 {
		t.Fatalf("combat input dispatch failed: %+v", spy)
	}
}

type mapInputSpy struct {
	sdk.BaseServerPlugin
	roomClicks   []sdk.MapRoomClickEvent
	markerClicks []sdk.MapMarkerClickEvent
}

func (s *mapInputSpy) OnMapRoomClick(ev sdk.MapRoomClickEvent) {
	s.roomClicks = append(s.roomClicks, ev)
}
func (s *mapInputSpy) OnMapMarkerClick(ev sdk.MapMarkerClickEvent) {
	s.markerClicks = append(s.markerClicks, ev)
}

func TestFlexibleWorldMapAndUnifiedHotbarAPI6(t *testing.T) {
	h := sdktest.Install(t)
	sdk.RegisterRoom(sdk.RoomSpec{X: 0, Y: 0, Name: "Hub", Biome: "sanctuary"})
	sdk.RegisterRoom(sdk.RoomSpec{X: 1, Y: 0, Name: "Spire", Biome: "astral_forge"})
	h.SetOnline(sdk.HeroSnapshot{ID: "h1", Name: "Ayla"})

	if err := sdk.SetMapConfig(sdk.MapConfigSpec{
		Title: "Cartographer Atlas", Subtitle: "Click rooms to fast-travel",
		ShowCoords: true, AllowFastTravel: true,
	}); err != nil {
		t.Fatalf("SetMapConfig: %v", err)
	}
	if h.MapConfig == nil || !h.MapConfig.AllowFastTravel || h.MapConfig.Title != "Cartographer Atlas" {
		t.Fatalf("MapConfig mismatch: %+v", h.MapConfig)
	}

	if err := sdk.SetMapRegion(sdk.MapRegionSpec{
		ID: "reg_north", Name: "Astral Reach", DangerLevel: 3,
		MinRoomX: 0, MaxRoomX: 1, MinRoomY: 0, MaxRoomY: 1,
		FillColor: [4]float32{0.2, 0.5, 0.9, 0.25},
	}); err != nil {
		t.Fatalf("SetMapRegion: %v", err)
	}
	if err := sdk.SetMapRegion(sdk.MapRegionSpec{ID: "bad_bounds", Name: "Bad", MinRoomX: 2, MaxRoomX: 0}); !errors.Is(err, sdk.ErrInvalidArg) {
		t.Fatalf("invalid region bounds: %v", err)
	}
	if err := sdk.SetMapRoute(sdk.MapRouteSpec{
		ID: "rt_ley", FromX: 0, FromY: 0, ToX: 1, ToY: 0, Style: "leyline", Label: "Star Leyline",
	}); err != nil {
		t.Fatalf("SetMapRoute: %v", err)
	}
	if err := sdk.SetMapRoute(sdk.MapRouteSpec{ID: "rt_bad", Style: "warp"}); !errors.Is(err, sdk.ErrInvalidArg) {
		t.Fatalf("invalid route style: %v", err)
	}
	if err := sdk.SetMapMarker(sdk.MapMarkerSpec{
		ID: "mk_boss", RoomX: 1, RoomY: 0, Label: "Boss", Category: "boss", Tooltip: "Danger!", Radius: 2.5, Clickable: true,
	}); err != nil {
		t.Fatalf("SetMapMarker API6: %v", err)
	}
	if m := h.MapMarkers["mk_boss"]; m.Category != "boss" || !m.Clickable || m.Radius != 2.5 {
		t.Fatalf("enriched marker mismatch: %+v", m)
	}

	if err := sdk.SetHeroHotbar("h1", []sdk.HotbarSlotSpec{{
		Slot: 1, ID: "heavy_strike", Label: "Strike", Group: "spell", KeyHint: "1",
		Badge: "2 AP", Charges: 3, AccentColor: [3]float32{1, 0.8, 0.3}, Selected: true,
	}}); err != nil {
		t.Fatalf("SetHeroHotbar API6: %v", err)
	}
	if sl := h.HeroHotbars["h1"][0]; sl.Group != "spell" || sl.KeyHint != "1" || sl.Badge != "2 AP" || !sl.Selected || sl.Charges != 3 {
		t.Fatalf("unified hotbar slot mismatch: %+v", sl)
	}

	spy := &mapInputSpy{}
	h.FireMapRoomClick(spy, sdk.MapRoomClickEvent{HeroID: "h1", RoomX: 1, RoomY: 0, Button: "left"})
	h.FireMapMarkerClick(spy, sdk.MapMarkerClickEvent{HeroID: "h1", MarkerID: "mk_boss", RoomX: 1, RoomY: 0, Button: "right"})
	if len(spy.roomClicks) != 1 || len(spy.markerClicks) != 1 {
		t.Fatalf("map input dispatch failed: %+v", spy)
	}

	if err := sdk.RemoveMapRegion("reg_north"); err != nil {
		t.Fatalf("RemoveMapRegion: %v", err)
	}
	if err := sdk.RemoveMapRegion("reg_north"); !errors.Is(err, sdk.ErrNotFound) {
		t.Fatalf("RemoveMapRegion twice: %v", err)
	}
	if err := sdk.RemoveMapRoute("rt_ley"); err != nil {
		t.Fatalf("RemoveMapRoute: %v", err)
	}
	if err := sdk.RemoveMapRoute("rt_ley"); !errors.Is(err, sdk.ErrNotFound) {
		t.Fatalf("RemoveMapRoute twice: %v", err)
	}
}
