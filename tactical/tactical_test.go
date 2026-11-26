package tactical_test

import (
	"strings"
	"testing"

	sdk "github.com/mmoboy/hero-armory-sdk"
	"github.com/mmoboy/hero-armory-sdk/sdktest"
	"github.com/mmoboy/hero-armory-sdk/tactical"
)

func TestDifficultyPresetsAndScaling(t *testing.T) {
	tiers := []tactical.DifficultyTier{
		tactical.TierStory,
		tactical.TierNormal,
		tactical.TierVeteran,
		tactical.TierHeroic,
		tactical.TierMythic,
	}
	tmpl, ok := tactical.MonsterByID("ironhide_bruiser")
	if !ok {
		t.Fatal("expected ironhide_bruiser in DefaultMonsters")
	}
	prevHP := 0
	prevDmg := 0
	for _, tier := range tiers {
		parsed, ok := tactical.ParseDifficulty(strings.ToUpper(string(tier)))
		if !ok || parsed != tier {
			t.Fatalf("ParseDifficulty(%q) = (%q, %v)", tier, parsed, ok)
		}
		bal := tactical.PresetBalance(tier)
		st := tactical.ScaleMob(tmpl, 5, bal)
		if st.MaxHP <= prevHP || st.Damage <= prevDmg {
			t.Fatalf("tier %s did not scale monotonically: %+v (prev HP=%d DMG=%d)", tier, st, prevHP, prevDmg)
		}
		prevHP = st.MaxHP
		prevDmg = st.Damage
	}
	if _, ok := tactical.ParseDifficulty("impossible"); ok {
		t.Fatal("unknown difficulty should fail ParseDifficulty")
	}
}

func TestDefaultRosterAndSpellbook(t *testing.T) {
	spells := tactical.DefaultSpells()
	if len(spells) != 8 {
		t.Fatalf("expected 8 default spells, got %d", len(spells))
	}
	for i, sp := range spells {
		if sp.Slot != i+1 || !sdk.IsValidContentID(sp.ID) || sp.Tooltip == "" {
			t.Fatalf("invalid spell %+v", sp)
		}
	}
	mobs := tactical.DefaultMonsters()
	if len(mobs) != 8 {
		t.Fatalf("expected 8 default monsters, got %d", len(mobs))
	}
	for _, m := range mobs {
		if !sdk.IsValidContentID(m.ID) || m.BaseHP <= 0 || m.BaseDamage <= 0 || len(m.Spells) == 0 {
			t.Fatalf("invalid monster template %+v", m)
		}
	}
}

func TestFullTacticalBattleLifecycleAndAllSpells(t *testing.T) {
	h := sdktest.Install(t)
	sdk.RegisterRoom(sdk.RoomSpec{X: 0, Y: 0, Name: "Starlight Nexus", Biome: "astral_forge"})
	h.SetOnline(sdk.HeroSnapshot{ID: "h1", Name: "Ayla", RoomX: 0, RoomY: 0, PosX: 0, PosZ: 0, Level: 5})

	var ended []tactical.BattleResult
	eng := tactical.NewEngine(tactical.Config{
		DefaultDifficulty: tactical.TierNormal,
		TurnStepMs:        500,
		Seed:              42,
		ResolveLoot: func(res tactical.BattleResult) []tactical.LootDrop {
			return []tactical.LootDrop{{ItemID: "mat_wakfu_crystal", Name: "Wakfu Crystal", Qty: 2}}
		},
		OnBattleEnd: func(res tactical.BattleResult) {
			ended = append(ended, res)
		},
	})

	if err := eng.AddEncounter(tactical.EncounterSpec{
		ID: "enc_sentinel", RoomX: 0, RoomY: 0, PosX: 0, PosZ: -6,
		TemplateID: "frost_sentinel", Level: 3, PackSize: 1, RespawnSec: 5.0,
	}); err != nil {
		t.Fatalf("AddEncounter: %v", err)
	}
	if _, ok := h.Actors["enc_sentinel"]; !ok {
		t.Fatal("expected overworld encounter actor to spawn")
	}

	// Test /difficulty chat command updates idle encounter HP.
	baseHP := h.Actors["enc_sentinel"].MaxHP
	if !eng.HandleDifficultyCommand(sdk.ChatEvent{HeroID: "h1", Text: "/difficulty veteran"}) {
		t.Fatal("expected /difficulty command to be consumed")
	}
	if eng.RealmDifficulty() != tactical.TierVeteran || h.Actors["enc_sentinel"].MaxHP <= baseHP {
		t.Fatalf("expected veteran difficulty to boost idle actor HP (base=%d now=%d)", baseHP, h.Actors["enc_sentinel"].MaxHP)
	}
	// Right-click inspect before battle.
	eng.OnActorClick(sdk.ActorClickEvent{HeroID: "h1", RoomX: 0, RoomY: 0, ActorID: "enc_sentinel", Button: "right"})
	if len(h.ChatsFor("h1")) == 0 {
		t.Fatal("expected right-click inspect chat output")
	}

	// Left-click starts tactical battle.
	eng.OnActorClick(sdk.ActorClickEvent{HeroID: "h1", RoomX: 0, RoomY: 0, ActorID: "enc_sentinel", Button: "left"})
	if !eng.InBattle("h1") || !h.HeroControlLocks["h1"] {
		t.Fatal("expected hero in battle with movement locked")
	}
	if len(h.HeroHotbars["h1"]) != 9 {
		t.Fatalf("expected 8 spells + End Turn (9 slots), got %d", len(h.HeroHotbars["h1"]))
	}
	if sl := h.HeroHotbars["h1"][0]; sl.Group != "spell" || sl.KeyHint != "1" || sl.Badge != "2 AP" || !sl.Selected {
		t.Fatalf("expected unified action bar metadata on slot 1, got %+v", sl)
	}
	if endSl := h.HeroHotbars["h1"][8]; endSl.Group != "end_turn" || endSl.KeyHint != "SPC" || endSl.Label != "END" || endSl.Badge != "END" {
		t.Fatalf("expected end_turn metadata on slot 10, got %+v", endSl)
	}
	// Verify movement area circle, single spell cast-range circle, and red target ring markers.
	if m, ok := h.GroundMarkers["tac_mp_h1"]; !ok || m.Shape != "circle" {
		t.Fatalf("expected movement area circle marker, got %+v", h.GroundMarkers["tac_mp_h1"])
	}
	if m, ok := h.GroundMarkers["tac_rng_h1"]; !ok || m.Shape != "circle" {
		t.Fatalf("expected cast-range circle marker, got %+v", h.GroundMarkers["tac_rng_h1"])
	}
	if m, ok := h.GroundMarkers["tac_tgt_h1"]; !ok || m.Shape != "ring" {
		t.Fatalf("expected target ring marker under selected enemy, got %+v", h.GroundMarkers["tac_tgt_h1"])
	}
	// Verify client HUD packet mounts both tactical_hud (with End Turn button) and tactical_timeline overlays cleanly.
	hudPkts := sdktest.PacketsFor[tactical.BattleHUDSnapshot](h, "h1", tactical.DefaultHUDChannel)
	if len(hudPkts) == 0 || !hudPkts[len(hudPkts)-1].Active || len(hudPkts[len(hudPkts)-1].Timeline) != 2 {
		t.Fatalf("expected active BattleHUDSnapshot with 2 timeline entries, got %+v", hudPkts)
	}
	if !hudPkts[len(hudPkts)-1].Timeline[0].Active || hudPkts[len(hudPkts)-1].Timeline[0].Team != "ally" {
		t.Fatalf("expected hero active in timeline entry 0: %+v", hudPkts[len(hudPkts)-1].Timeline[0])
	}
	rawActive := h.Packets[len(h.Packets)-1].Payload
	if !tactical.HandleHUDPacket(tactical.DefaultHUDChannel, rawActive) ||
		h.OverlayByID(tactical.BattleHUDOverlayID) == nil ||
		h.OverlayByID(tactical.BattleTimelineOverlayID) == nil ||
		h.FindNodeIn(tactical.BattleHUDOverlayID, tactical.EndTurnButtonID) == nil {
		t.Fatal("expected HandleHUDPacket to mount tactical_hud (with End Turn button) and tactical_timeline overlays")
	}
	if len(h.GroundMarkers) == 0 {
		t.Fatal("expected movement/range ground markers")
	}

	// Cast Ember Lance (slot 4) -> applies Burn status.
	eng.OnHotbarPress(sdk.HotbarPressEvent{HeroID: "h1", Slot: 4, ActionID: "ember_lance"})
	eng.OnActorClick(sdk.ActorClickEvent{HeroID: "h1", RoomX: 0, RoomY: 0, ActorID: "enc_sentinel", Button: "left"})
	foundBurn := false
	for _, st := range h.Actors["enc_sentinel"].Statuses {
		if st.ID == "burn" {
			foundBurn = true
		}
	}
	if !foundBurn {
		t.Fatalf("expected burn status on sentinel after ember_lance: %+v", h.Actors["enc_sentinel"].Statuses)
	}

	// Cast Aegis Ward (slot 6, self-cast shield).
	eng.OnHotbarPress(sdk.HotbarPressEvent{HeroID: "h1", Slot: 6, ActionID: "aegis_ward"})

	// Ground click moves hero within MP.
	eng.OnGroundClick(sdk.GroundClickEvent{HeroID: "h1", RoomX: 0, RoomY: 0, PosX: 0, PosZ: -2.0, Button: "left"})
	if h.Hero("h1").PosZ != -2.0 {
		t.Fatalf("expected hero moved to z=-2.0, got %+v", h.Hero("h1"))
	}

	// End Turn via clickable tactical_hud End Turn button -> OnPacket -> OnTick executes enemy AI turn + burn tick.
	if !tactical.HandleUIEvent(sdk.UIClickEvent{ControlID: tactical.EndTurnButtonID}) || len(h.ClientPackets) == 0 {
		t.Fatal("expected HandleUIEvent to send End Turn packet to server")
	}
	endPkt := h.ClientPackets[len(h.ClientPackets)-1]
	if !eng.OnPacket("h1", endPkt.Channel, endPkt.Payload) {
		t.Fatal("expected Engine.OnPacket to handle End Turn action")
	}
	eng.OnTick(500)
	eng.OnTick(500)

	// Next player turn: test Frost Bind (slot 5), Mending Light (slot 7), Gale Rush dash (slot 8).
	eng.OnHotbarPress(sdk.HotbarPressEvent{HeroID: "h1", Slot: 5, ActionID: "frost_bind"})
	eng.OnActorClick(sdk.ActorClickEvent{HeroID: "h1", RoomX: 0, RoomY: 0, ActorID: "enc_sentinel", Button: "left"})
	eng.OnHotbarPress(sdk.HotbarPressEvent{HeroID: "h1", Slot: 7, ActionID: "mending_light"})
	eng.OnHotbarPress(sdk.HotbarPressEvent{HeroID: "h1", Slot: 10, ActionID: tactical.EndTurnActionID})
	eng.OnTick(500)
	eng.OnTick(500)

	eng.OnHotbarPress(sdk.HotbarPressEvent{HeroID: "h1", Slot: 8, ActionID: "gale_rush"})
	eng.OnGroundClick(sdk.GroundClickEvent{HeroID: "h1", RoomX: 0, RoomY: 0, PosX: 1.0, PosZ: -4.0, Button: "left"})

	// Finish off sentinel across turns using full 6 AP (2x Arcane Bolt or Ember Lance + Arcane Bolt).
	for round := 0; round < 12 && eng.InBattle("h1"); round++ {
		eng.OnHotbarPress(sdk.HotbarPressEvent{HeroID: "h1", Slot: 2, ActionID: "arcane_bolt"})
		eng.OnActorClick(sdk.ActorClickEvent{HeroID: "h1", RoomX: 0, RoomY: 0, ActorID: "enc_sentinel", Button: "left"})
		if !eng.InBattle("h1") {
			break
		}
		eng.OnHotbarPress(sdk.HotbarPressEvent{HeroID: "h1", Slot: 2, ActionID: "arcane_bolt"})
		eng.OnActorClick(sdk.ActorClickEvent{HeroID: "h1", RoomX: 0, RoomY: 0, ActorID: "enc_sentinel", Button: "left"})
		if !eng.InBattle("h1") {
			break
		}
		eng.OnHotbarPress(sdk.HotbarPressEvent{HeroID: "h1", Slot: 10, ActionID: tactical.EndTurnActionID})
		eng.OnTick(500)
		eng.OnTick(500)
	}

	if eng.InBattle("h1") {
		t.Fatal("expected battle to conclude within 12 turns")
	}
	if len(ended) != 1 || !ended[0].Won || ended[0].GoldAwarded <= 0 || ended[0].XPAwarded <= 0 {
		t.Fatalf("unexpected BattleResult: %+v", ended)
	}
	if h.HeroControlLocks["h1"] {
		t.Fatal("movement lock should be cleared after battle")
	}
	if len(h.HeroHotbars["h1"]) != 0 {
		t.Fatal("hotbar should be cleared after battle")
	}
	if len(h.GroundMarkers) != 0 {
		t.Fatal("ground markers should be cleared after battle")
	}
	if h.Hero("h1").Gold <= 0 || h.Hero("h1").XP <= 0 {
		t.Fatalf("expected gold & XP rewards granted: %+v", h.Hero("h1"))
	}
	// Verify end-of-battle summary packet clears battle HUD/timeline and mounts Victory Loot Summary overlay.
	lastPkt := h.Packets[len(h.Packets)-1]
	if !tactical.HandleHUDPacket(lastPkt.Channel, lastPkt.Payload) ||
		h.OverlayByID(tactical.BattleHUDOverlayID) != nil ||
		h.OverlayByID(tactical.BattleTimelineOverlayID) != nil ||
		h.OverlayByID(tactical.BattleSummaryOverlayID) == nil {
		t.Fatal("expected Victory summary packet to clear battle HUD/timeline and mount tactical_summary overlay")
	}
	if !strings.Contains(h.OverlayText(), "VICTORY!") || !strings.Contains(h.OverlayText(), "Wakfu Crystal x2") {
		t.Fatalf("expected Victory & loot text in summary overlay, got: %s", h.OverlayText())
	}
	if !tactical.HandleUIEvent(sdk.UIClickEvent{ControlID: tactical.SummaryCloseButtonID}) ||
		h.OverlayByID(tactical.BattleSummaryOverlayID) != nil {
		t.Fatal("expected Continue button to close tactical_summary overlay")
	}

	// Respawn timer restores overworld encounter actor after 5 seconds.
	eng.OnTick(5100)
	if _, ok := h.Actors["enc_sentinel"]; !ok {
		t.Fatal("expected encounter actor to respawn after 5s")
	}
}

func TestRuthlessAIAndAbortCleanup(t *testing.T) {
	h := sdktest.Install(t)
	sdk.RegisterRoom(sdk.RoomSpec{X: 0, Y: 0, Name: "Caldera", Biome: "cindervault"})
	h.SetOnline(sdk.HeroSnapshot{ID: "h1", Name: "Ayla", RoomX: 0, RoomY: 0, PosX: 0, PosZ: 0, Level: 5})

	eng := tactical.NewEngine(tactical.Config{DefaultDifficulty: tactical.TierMythic, TurnStepMs: 250, Seed: 7})
	if err := eng.AddEncounter(tactical.EncounterSpec{
		ID: "boss_overlord", RoomX: 0, RoomY: 0, PosX: 0, PosZ: -5,
		TemplateID: "astral_overlord", Level: 5, Tier: tactical.TierMythic,
	}); err != nil {
		t.Fatal(err)
	}
	eng.OnActorClick(sdk.ActorClickEvent{HeroID: "h1", RoomX: 0, RoomY: 0, ActorID: "boss_overlord", Button: "left"})
	if !eng.InBattle("h1") {
		t.Fatal("expected mythic battle active")
	}
	// Let ruthless boss take a turn.
	eng.OnHotbarPress(sdk.HotbarPressEvent{HeroID: "h1", Slot: 10, ActionID: tactical.EndTurnActionID})
	eng.OnTick(250)
	if len(h.VFXLog) == 0 {
		t.Fatal("expected ruthless boss to cast spells / play VFX")
	}

	// Aborting via player leave cleans up lock, hotbar, markers, and restores idle boss actor.
	eng.OnPlayerLeave("h1")
	if eng.InBattle("h1") || h.HeroControlLocks["h1"] || len(h.HeroHotbars["h1"]) != 0 {
		t.Fatal("expected clean battle teardown on OnPlayerLeave")
	}
	if _, ok := h.Actors["boss_overlord"]; !ok {
		t.Fatal("expected idle boss actor restored after aborted battle")
	}
}
