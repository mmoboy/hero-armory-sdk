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
