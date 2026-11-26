package vanilla

import (
	"strings"
	"testing"

	sdk "github.com/mmoboy/hero-armory-sdk"
	"github.com/mmoboy/hero-armory-sdk/sdktest"
	"github.com/mmoboy/hero-armory-sdk/tactical"
)

func TestSanctuaryStarterClassesAndConfig(t *testing.T) {
	classes := SanctuaryStarterClasses()
	if len(classes) != 4 {
		t.Fatalf("expected 4 sanctuary starter classes, got %d", len(classes))
	}
	seen := map[string]bool{}
	for _, cls := range classes {
		if !sdk.IsValidContentID(cls.ID) {
			t.Fatalf("invalid starter class id %q", cls.ID)
		}
		if seen[cls.ID] {
			t.Fatalf("duplicate starter class id %q", cls.ID)
		}
		seen[cls.ID] = true
		if cls.Name == "" || cls.Description == "" || len(cls.Equipped) == 0 {
			t.Fatalf("starter class %q missing fields: %+v", cls.ID, cls)
		}
		if cls.BonusHP == 0 && cls.BonusAP == 0 && cls.BonusMP == 0 {
			t.Fatalf("expected stat bonuses on starter class %q", cls.ID)
		}
	}

	cfg := SanctuaryStarterConfig()
	if cfg.BaseHP <= 0 || cfg.BaseAP <= 0 || cfg.StarterGold <= 0 || len(cfg.StarterEquipped) == 0 {
		t.Fatalf("expected populated SanctuaryStarterConfig, got %+v", cfg)
	}
}

func TestRegisterAllAndCommands(t *testing.T) {
	h := sdktest.Install(t)
	RegisterAll()

	if len(h.Rooms) != 25 {
		t.Fatalf("expected 25 Sanctuary rooms, got %d", len(h.Rooms))
	}
	for coord, rm := range h.Rooms {
		if rm.Atmosphere == nil || rm.MapIcon == "" {
			t.Fatalf("room %v missing enriched Atmosphere or MapIcon", coord)
		}
		if len(rm.Liquids) > sdk.MaxRoomLiquids || len(rm.Lights) > sdk.MaxRoomLights || len(rm.Emitters) > sdk.MaxRoomEmitters {
			t.Fatalf("room %v exceeded API 4 limits", coord)
		}
		if err := sdk.SetRoomAtmosphere(coord[0], coord[1], *rm.Atmosphere); err != nil {
			t.Fatalf("room %v atmosphere invalid: %v", coord, err)
		}
		if err := sdk.ClearRoomAtmosphere(coord[0], coord[1]); err != nil {
			t.Fatalf("room %v clear atmosphere: %v", coord, err)
		}
	}
	if len(h.NPCs) != 7 {
		t.Fatalf("expected 7 Sanctuary NPCs, got %d", len(h.NPCs))
	}
	if len(h.PlayerSkills) != 3 {
		t.Fatalf("expected 3 Sanctuary player skills, got %d", len(h.PlayerSkills))
	}
	if len(h.Jobs) != 5 {
		t.Fatalf("expected 5 Sanctuary jobs, got %d", len(h.Jobs))
	}
	if len(h.CodexEntries) != 3 || h.CodexConfig == nil {
		t.Fatalf("expected Sanctuary codex config and 3 entries, got cfg=%v entries=%d", h.CodexConfig, len(h.CodexEntries))
	}

	// Test /help, /where, and /sanctuary commands.
	h.SetOnline(sdk.HeroSnapshot{ID: "h1", Name: "Ayla", RoomX: 2, RoomY: -2, Level: 5, Gold: 1200})
	if !HandleChatCommand(sdk.ChatEvent{HeroID: "h1", Text: "/help"}) {
		t.Fatal("expected /help to be handled")
	}
	if !HandleChatCommand(sdk.ChatEvent{HeroID: "h1", Text: "/where"}) {
		t.Fatal("expected /where to be handled")
	}
	if HandleChatCommand(sdk.ChatEvent{HeroID: "h1", Text: "hello world"}) {
		t.Fatal("normal chat should not be consumed")
	}

	h.SetInBattle("h1", true)
	if !HandleChatCommand(sdk.ChatEvent{HeroID: "h1", Text: "/sanctuary"}) {
		t.Fatal("expected /sanctuary to be handled even in battle")
	}
	if h.Hero("h1").RoomX != 2 || h.Hero("h1").RoomY != -2 {
		t.Fatal("hero in battle should not teleport")
	}

	h.SetInBattle("h1", false)
	if !HandleChatCommand(sdk.ChatEvent{HeroID: "h1", Text: "/sanctuary"}) {
		t.Fatal("expected /sanctuary to be handled")
	}
	if got := h.Hero("h1"); got.RoomX != 0 || got.RoomY != 0 {
		t.Fatalf("expected hero at (0,0) after /sanctuary, got (%d,%d)", got.RoomX, got.RoomY)
	}
	chats := strings.Join(h.ChatsFor("h1"), "\n")
	if !strings.Contains(chats, "Sanctuary commands:") || !strings.Contains(chats, "Lv.5, 1200g") {
		t.Fatalf("unexpected chat output: %q", chats)
	}
}

func TestSanctuaryTacticalEncountersAndLoot(t *testing.T) {
	h := sdktest.Install(t)
	RegisterRooms()
	h.SetOnline(sdk.HeroSnapshot{ID: "h1", Name: "Ayla", RoomX: 0, RoomY: 0, PosX: 9, PosZ: 10, Level: 5})

	eng := tactical.NewEngine(tactical.Config{
		DefaultDifficulty: tactical.TierNormal,
		TurnStepMs:        400,
		Seed:              11,
		ResolveLoot:       AwardTacticalLoot,
	})
	RegisterTacticalEncounters(eng)

	if len(h.Actors) != 49 {
		t.Fatalf("expected 49 tactical encounter actors across 25 Sanctuary rooms, got %d", len(h.Actors))
	}

	// Engage Courtyard Wild Gobball and defeat it.
	eng.OnActorClick(sdk.ActorClickEvent{HeroID: "h1", RoomX: 0, RoomY: 0, ActorID: "tac_2_2_0", Button: "left"})
	if !eng.InBattle("h1") {
		t.Fatal("expected tactical battle against Courtyard gobball")
	}
	for round := 0; round < 6 && eng.InBattle("h1"); round++ {
		eng.OnHotbarPress(sdk.HotbarPressEvent{HeroID: "h1", Slot: 2, ActionID: "arcane_bolt"})
		eng.OnActorClick(sdk.ActorClickEvent{HeroID: "h1", RoomX: 0, RoomY: 0, ActorID: "tac_2_2_0", Button: "left"})
		if !eng.InBattle("h1") {
			break
		}
		eng.OnHotbarPress(sdk.HotbarPressEvent{HeroID: "h1", Slot: 2, ActionID: "arcane_bolt"})
		eng.OnActorClick(sdk.ActorClickEvent{HeroID: "h1", RoomX: 0, RoomY: 0, ActorID: "tac_2_2_0", Button: "left"})
		if !eng.InBattle("h1") {
			break
		}
		eng.OnHotbarPress(sdk.HotbarPressEvent{HeroID: "h1", Slot: 10, ActionID: tactical.EndTurnActionID})
		eng.OnTick(400)
		eng.OnTick(400)
	}
	if eng.InBattle("h1") {
		t.Fatal("expected Courtyard gobball defeated")
	}
	if h.HeroItemCount("h1", "mat_gobball_wool") < 2 {
		t.Fatalf("expected gobball loot drop, got %d", h.HeroItemCount("h1", "mat_gobball_wool"))
	}
	lastPkt := h.Packets[len(h.Packets)-1]
	if !tactical.HandleHUDPacket(lastPkt.Channel, lastPkt.Payload) || h.OverlayByID(tactical.BattleSummaryOverlayID) == nil {
		t.Fatal("expected Sanctuary tactical victory to mount tactical_summary loot overlay")
	}
	if !strings.Contains(h.OverlayText(), "Royal Gobball Fleece x2") {
		t.Fatalf("expected resolved item name in summary overlay, got: %s", h.OverlayText())
	}
	if !tactical.HandleUIEvent(sdk.UIClickEvent{ControlID: tactical.SummaryCloseButtonID}) || h.OverlayByID(tactical.BattleSummaryOverlayID) != nil {
		t.Fatal("expected Continue button to dismiss tactical_summary")
	}
}
