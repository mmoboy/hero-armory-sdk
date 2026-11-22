package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	sdk "github.com/mmoboy/hero-armory-sdk"
	"github.com/mmoboy/hero-armory-sdk/vanilla"
	"realm_bounty/proto"
)

type challengeTemplate struct {
	Kind        string
	Title       string
	Description string
	Target      int
}

const (
	bountyMonsterArchID = "bounty_harbinger"
	bountyHubSpawnID    = "bounty_stalker_hub"
)

var challenges = []challengeTemplate{
	{
		Kind:        "explore",
		Title:       "Chunk Explorer",
		Description: "Explore 3 different rooms around the realm.",
		Target:      3,
	},
	{
		Kind:        "chat",
		Title:       "Town Crier",
		Description: "Say hello twice in public realm or local chat.",
		Target:      2,
	},
	{
		Kind:        "cheer",
		Title:       "Realm Rally",
		Description: "Send 2 Realm Cheers from the Bounty Board.",
		Target:      2,
	},
	{
		Kind:        "hunt",
		Title:       "Slay Gilded Stalker",
		Description: "Defeat the Gilded Stalker in combat or click Summon Stalker nearby.",
		Target:      1,
	},
	{
		Kind:        "gather",
		Title:       "Prospect Gilded Vein",
		Description: "Mine the Gilded Ore Vein in hub (0,0) twice to gather Gilded Ore.",
		Target:      2,
	},
}

type heroBountyState struct {
	HeroID         string
	HeroName       string
	OriginID       string `json:"origin_id,omitempty"`
	OriginName     string `json:"origin_name,omitempty"`
	RoomX          int
	RoomY          int
	PosX           float32
	PosZ           float32
	Online         bool
	ChallengeIdx   int
	Progress       int
	Completed      bool
	CompletedCount int
	VisitedRooms   map[string]bool
	LastEvent      string
}

type bountyOriginDef struct {
	Class          sdk.StarterClassSpec
	StartChallenge int
	StarterGold    int64
	StarterBag     []sdk.ItemStackSpec
}

var bountyOrigins = []bountyOriginDef{
	{
		Class: sdk.StarterClassSpec{
			ID:          "bounty_pathfinder",
			Name:        "Bounty Pathfinder",
			Description: "Swift scout equipped for Chunk Explorer runs (+10 HP, +1 MP).",
			Icon:        "ha_spell_strike",
			Equipped:    vanilla.KokiriGear(),
			BonusHP:     10,
			BonusMP:     1,
		},
		StartChallenge: 0,
		StarterGold:    220,
		StarterBag: []sdk.ItemStackSpec{
			{ItemID: "cons_wakfu_potion", Quantity: 3},
			{ItemID: "plug_bounty_ore", Quantity: 1},
		},
	},
	{
		Class: sdk.StarterClassSpec{
			ID:          "bounty_prospector",
			Name:        "Ore Prospector",
			Description: "Gilded Vein miner carrying Stalker Drafts (+20 HP, +1 AP).",
			Icon:        "ha_job_mineur",
			Equipped:    vanilla.GobballGear(),
			BonusHP:     20,
			BonusAP:     1,
		},
		StartChallenge: 4,
		StarterGold:    260,
		StarterBag: []sdk.ItemStackSpec{
			{ItemID: "plug_bounty_ore", Quantity: 2},
			{ItemID: "plug_bounty_elixir", Quantity: 1},
			{ItemID: "cons_wakfu_potion", Quantity: 2},
		},
	},
	{
		Class: sdk.StarterClassSpec{
			ID:          "bounty_stalker",
			Name:        "Stalker Hunter",
			Description: "Relic hunter wielding the Gilded Halberd for boss hunts (+25 HP, +1 AP).",
			Icon:        "ha_spell_nova",
			Equipped: map[string]string{
				"hat":    "hat_twilight_crown",
				"chest":  "chest_twilight_hauberk",
				"boots":  "boots_twilight_greaves",
				"weapon": "plug_stalker_halberd",
			},
			BonusHP: 25,
			BonusAP: 1,
		},
		StartChallenge: 3,
		StarterGold:    300,
		StarterBag: []sdk.ItemStackSpec{
			{ItemID: "cons_red_heart_elixir", Quantity: 2},
			{ItemID: "plug_bounty_elixir", Quantity: 1},
		},
	},
}

func bountyStarterClasses() []sdk.StarterClassSpec {
	out := make([]sdk.StarterClassSpec, 0, len(bountyOrigins))
	for _, o := range bountyOrigins {
		out = append(out, o.Class)
	}
	return out
}

type bountyServer struct {
	sdk.BaseServerPlugin
	heroes      map[string]*heroBountyState
	realmCheers int
}

type persistedGlobalState struct {
	RealmCheers int `json:"realm_cheers"`
}

func newBountyServer() *bountyServer {
	return &bountyServer{
		heroes: make(map[string]*heroBountyState),
	}
}

func (s *bountyServer) loadPersistedState() {
	var global persistedGlobalState
	if ok, err := sdk.KVGetJSON("global", &global); err == nil && ok {
		if global.RealmCheers > 0 {
			s.realmCheers = global.RealmCheers
		}
	}
	keys, err := sdk.KVList("hero:", sdk.MaxKVListLimit)
	if err != nil {
		return
	}
	restored := 0
	for _, key := range keys {
		var st heroBountyState
		if ok, err := sdk.KVGetJSON(key, &st); err == nil && ok {
			if st.HeroID == "" {
				st.HeroID = strings.TrimPrefix(key, "hero:")
			}
			if st.HeroID == "" {
				continue
			}
			st.Online = false
			if st.VisitedRooms == nil {
				st.VisitedRooms = make(map[string]bool)
			}
			if st.OriginName == "" {
				st.OriginName = "Freelance"
			}
			s.heroes[st.HeroID] = &st
			restored++
		}
	}
	if restored > 0 || s.realmCheers > 0 {
		sdk.Log(fmt.Sprintf("Restored %d hero bounties and %d realm cheers from Realm KV", restored, s.realmCheers))
	}
}

func (s *bountyServer) saveGlobalState() {
	_ = sdk.KVSetJSON("global", persistedGlobalState{RealmCheers: s.realmCheers})
}

func (s *bountyServer) saveHeroState(st *heroBountyState) {
	if st == nil || st.HeroID == "" {
		return
	}
	cp := *st
	cp.Online = false
	_ = sdk.KVSetJSON("hero:"+cp.HeroID, cp)
}

func (s *bountyServer) Init() {
	sdk.Log("Realm Bounty Board server plugin initialized")
	vanilla.RegisterAll()

	for _, rm := range vanilla.Rooms() {
		if rm.X == 0 && rm.Y == 2 {
			rm.Portals = append(rm.Portals, sdk.RoomPortalSpec{
				Edge: "north", TargetX: 0, TargetY: 3, Label: "Gilded Vault",
			})
			rm.TerrainOps = append(rm.TerrainOps, sdk.RoomTerrainOpSpec{
				Op: "slope", Surface: "ramp", MinX: -6, MaxX: 6, MinZ: -30, MaxZ: -20, RiseDir: "south", Elev: 0, ElevEnd: 0.80,
			})
			sdk.RegisterRoom(rm)
			break
		}
	}
	sdk.RegisterRoom(sdk.RoomSpec{
		X: 0, Y: 3, Name: "Gilded Vault", Biome: "astral_forge", LightingPreset: "astral", HalfSize: 22.0,
		FloorColor:  [3]float32{0.18, 0.14, 0.10},
		WallColor:   [3]float32{0.24, 0.18, 0.10},
		AccentColor: [3]float32{0.96, 0.78, 0.24},
		SpawnPos:    [3]float32{0, 0, 4.0},
		TerrainOps: []sdk.RoomTerrainOpSpec{
			{Op: "rect", Surface: "dais", MinX: -16, MaxX: 16, MinZ: -16, MaxZ: 16, Elev: 0.80},
			{Op: "slope", Surface: "ramp", MinX: -6, MaxX: 6, MinZ: 16, MaxZ: 22, RiseDir: "north", Elev: 0, ElevEnd: 0.80},
		},
		Portals: []sdk.RoomPortalSpec{
			{Edge: "south", TargetX: 0, TargetY: 2, Label: "Sylvan Watchtower"},
		},
		Interactables: []sdk.RoomInteractableSpec{
			{
				ID: "gilded_vault_altar", Kind: "shrine", Label: "Gilded Vault Altar",
				Description: "An ornate auric plinth humming with compressed Wakfu energy.",
				RewardItem:  "plug_bounty_ore", X: 0, Z: -9.5, Radius: 1.25,
			},
		},
		Props: []sdk.RoomPropSpec{
			{
				ID: "gilded_statue_prop", Kind: "custom", X: -6.5, Z: -8.0, Scale: 1.25, Yaw: 0.45,
				ModelPath: "assets/gilded_stalker.glb", SolidRadius: 1.15,
			},
			{
				ID: "gilded_ore_trophy", Kind: "custom", X: 6.5, Z: -8.0, Scale: 1.85, Yaw: -0.45,
				ModelID: "gilded_ore_item", SolidRadius: 0.9,
			},
		},
	})

	// Register custom plugin items BEFORE recipes/monsters that reference them.
	sdk.RegisterItem(sdk.ItemSpec{
		ID:           "plug_bounty_ore",
		Name:         "Gilded Ore",
		Kind:         "material",
		Description:  "Shining metallic ore veined with raw Wakfu traces.",
		Rarity:       "Rare",
		MaxStack:     99,
		VisualPreset: "mat_wakfu_crystal",
		ModelID:      "gilded_ore_item",
		ModelScale:   1.0,
		PrimaryCol:   [4]float32{0.96, 0.78, 0.22, 1.0},
		AccentCol:    [4]float32{0.68, 0.24, 0.14, 1.0},
		GlowCol:      [4]float32{1.0, 0.88, 0.38, 1.0},
	})
	sdk.RegisterItem(sdk.ItemSpec{
		ID:           "plug_bounty_elixir",
		Name:         "Stalker Draft",
		Kind:         "consumable",
		Description:  "Bright golden tonic distilled from Gilded Ore.",
		Rarity:       "Rare",
		MaxStack:     99,
		VisualPreset: "cons_red_heart_elixir",
		PrimaryCol:   [4]float32{0.96, 0.74, 0.22, 1.0},
		AccentCol:    [4]float32{0.78, 0.18, 0.18, 1.0},
		GlowCol:      [4]float32{1.0, 0.84, 0.34, 1.0},
	})
	sdk.RegisterItem(sdk.ItemSpec{
		ID:           "plug_stalker_crown",
		Name:         "Stalker Halo",
		Kind:         "equipment",
		Slot:         "hat",
		Rarity:       "Relic",
		Description:  "Crown hammered from Gilded Ore and Wakfu shards.",
		MaxStack:     1,
		VisualPreset: "hat_twilight_crown",
		PrimaryCol:   [4]float32{0.94, 0.74, 0.20, 1.0},
		AccentCol:    [4]float32{0.72, 0.18, 0.16, 1.0},
		GlowCol:      [4]float32{1.0, 0.88, 0.40, 1.0},
		Vitality:     35,
		Strength:     20,
		Wisdom:       20,
		Agility:      25,
	})
	sdk.RegisterItem(sdk.ItemSpec{
		ID:           "plug_stalker_halberd",
		Name:         "Gilded Halberd",
		Kind:         "equipment",
		Slot:         "weapon",
		Rarity:       "Mythic",
		Description:  "Heavy hunter polearm forged from Gilded Ore and Forged Ingots.",
		MaxStack:     1,
		VisualPreset: "weapon_guardian_halberd",
		PrimaryCol:   [4]float32{0.96, 0.78, 0.24, 1.0},
		AccentCol:    [4]float32{0.68, 0.16, 0.14, 1.0},
		GlowCol:      [4]float32{1.0, 0.90, 0.44, 1.0},
		Vitality:     20,
		Strength:     40,
		Wisdom:       15,
		Agility:      30,
	})

	sdk.SpawnGatherNode(sdk.GatherNodeSpec{
		ID:           "bounty_gilded_vein",
		Job:          "mineur",
		Name:         "Gilded Ore Vein",
		RoomX:        0,
		RoomY:        0,
		PosX:         -3.5,
		PosZ:         -9.0,
		RewardItemID: "plug_bounty_ore",
		MinLevel:     1,
		GatherSec:    1.5,
		XP:           25,
		RespawnSec:   3,
		VisualColor:  [4]float32{0.96, 0.78, 0.22, 1.0},
		ModelID:      "gilded_vein_mesh",
		ModelScale:   1.0,
	})

	sdk.RegisterRecipe(sdk.RecipeSpec{
		ID:           "plug_craft_elixir",
		Job:          "mineur",
		Name:         "Brew Stalker Draft",
		Description:  "Refine 2 Gilded Ore into a potent golden Stalker Draft.",
		OutputItemID: "plug_bounty_elixir",
		OutputQty:    1,
		MinLevel:     1,
		XP:           35,
		Ingredients: []sdk.RecipeIngredient{
			{ItemID: "plug_bounty_ore", Quantity: 2},
		},
	})
	sdk.RegisterRecipe(sdk.RecipeSpec{
		ID:           "plug_craft_crown",
		Job:          "craftsman",
		Name:         "Forge Stalker Halo",
		Description:  "Hammer Gilded Ore around a Wakfu Crystal into a Relic crown.",
		OutputItemID: "plug_stalker_crown",
		OutputQty:    1,
		MinLevel:     1,
		XP:           60,
		Ingredients: []sdk.RecipeIngredient{
			{ItemID: "plug_bounty_ore", Quantity: 2},
			{ItemID: "mat_wakfu_crystal", Quantity: 1},
		},
	})
	sdk.RegisterRecipe(sdk.RecipeSpec{
		ID:           "plug_craft_halberd",
		Job:          "craftsman",
		Name:         "Forge Gilded Halberd",
		Description:  "Shape Gilded Ore and a Forged Ingot into a Mythic polearm.",
		OutputItemID: "plug_stalker_halberd",
		OutputQty:    1,
		MinLevel:     1,
		XP:           85,
		Ingredients: []sdk.RecipeIngredient{
			{ItemID: "plug_bounty_ore", Quantity: 3},
			{ItemID: "mat_forged_ingot", Quantity: 1},
		},
	})

	sdk.RegisterMonster(sdk.MonsterSpec{
		ID:          bountyMonsterArchID,
		Name:        "Gilded Stalker",
		Level:       5,
		MaxHP:       80,
		MovePoints:  8.5,
		AttackPts:   5,
		Damage:      14,
		Range:       3.2,
		GoldReward:  60,
		XPReward:    90,
		BodyColor:   [3]float32{0.94, 0.72, 0.18},
		TrimColor:   [3]float32{0.68, 0.16, 0.14},
		Scale:       1.32,
		ModelPath:   "assets/gilded_stalker.glb",
		AIStyle:     "caster",
		Description: "Molten-gold sentinel animated by crimson Wakfu cores. Guards hub veins and drops Relic forging materials.",
		Habitat:     "Sanctuary Courtyard (0,0) or Summoned near Hunters",
		Spells: []sdk.MonsterSpellSpec{
			{
				ID:         "gilded_cleave",
				Name:       "Gilded Cleave",
				APCost:     2,
				MinRange:   0,
				MaxRange:   3.4,
				TargetKind: "enemy",
				EffectKind: "damage",
				Power:      14,
				VFXKind:    "slash",
				VFXColor:   [4]float32{0.98, 0.80, 0.24, 1.0},
			},
			{
				ID:         "wakfu_lance",
				Name:       "Wakfu Lance",
				APCost:     2,
				MinRange:   2.0,
				MaxRange:   8.5,
				TargetKind: "enemy",
				EffectKind: "damage",
				Power:      13,
				VFXKind:    "bolt",
				VFXColor:   [4]float32{0.96, 0.24, 0.18, 1.0},
			},
			{
				ID:            "solar_cataclysm",
				Name:          "Solar Cataclysm",
				APCost:        3,
				MinRange:      0,
				MaxRange:      7.0,
				AoERadius:     2.8,
				CooldownTurns: 2,
				TargetKind:    "enemy",
				EffectKind:    "damage",
				Power:         18,
				VFXKind:       "nova",
				VFXColor:      [4]float32{1.0, 0.78, 0.20, 1.0},
			},
			{
				ID:            "ore_regeneration",
				Name:          "Ore Regeneration",
				APCost:        2,
				CooldownTurns: 2,
				TargetKind:    "self",
				EffectKind:    "heal",
				Power:         16,
				VFXKind:       "ring",
				VFXColor:      [4]float32{0.35, 0.96, 0.56, 1.0},
			},
		},
		LootTable: []sdk.MonsterLoot{
			{ItemID: "plug_bounty_ore", Chance: 0.90, MinQty: 1, MaxQty: 2},
			{ItemID: "plug_bounty_elixir", Chance: 0.45, MinQty: 1, MaxQty: 1},
			{ItemID: "plug_stalker_crown", Chance: 0.18, MinQty: 1, MaxQty: 1},
			{ItemID: "mat_wakfu_crystal", Chance: 0.70, MinQty: 1, MaxQty: 2},
			{ItemID: "portal_style_crimson", Chance: 0.15, MinQty: 1, MaxQty: 1},
		},
	})
	sdk.SpawnMonster(sdk.MonsterSpawnSpec{
		ID:          bountyHubSpawnID,
		ArchetypeID: bountyMonsterArchID,
		RoomX:       0,
		RoomY:       0,
		PosX:        -8.5,
		PosZ:        8.5,
		PackSize:    1,
		RespawnSec:  25,
	})
	sdk.SpawnMonster(sdk.MonsterSpawnSpec{
		ID:          "bounty_vault_boss",
		ArchetypeID: bountyMonsterArchID,
		RoomX:       0,
		RoomY:       3,
		PosX:        0,
		PosZ:        -3.5,
		PackSize:    2,
		RespawnSec:  30,
	})
	sdk.SpawnGatherNode(sdk.GatherNodeSpec{
		ID:           "bounty_vault_vein",
		Job:          "mineur",
		Name:         "Vault Gold Vein",
		RoomX:        0,
		RoomY:        3,
		PosX:         -7.5,
		PosZ:         3.5,
		RewardItemID: "plug_bounty_ore",
		MinLevel:     1,
		GatherSec:    1.4,
		XP:           40,
		RespawnSec:   12,
		VisualColor:  [4]float32{0.98, 0.80, 0.24, 1.0},
		ModelID:      "gilded_ore_item",
		ModelScale:   1.4,
	})

	hubZero := 0
	sdk.ConfigureCodex(sdk.CodexConfigSpec{
		Title:         "Bounty Realm Codex & Bestiary",
		BannerText:    "Sanctuary Bounty Edition - Search Gilded Stalker drops, Gilded Ore veins, and Relic gear.",
		ExtraTabLabel: "Bounty Guide",
	})
	sdk.RegisterCodexEntry(sdk.CodexEntrySpec{
		ID:          "bounty_forge_guide",
		Category:    "custom",
		KindBadge:   "BOUNTY GUIDE",
		Title:       "Gilded Mining & Relic Forge",
		Subtitle:    "Mine Gilded Ore in Sanctuary Courtyard (0,0)",
		Description: "Harvest the Gilded Ore Vein at (-3.5, -9.0) or defeat Gilded Stalkers for Gilded Ore. Brew Stalker Draft (Mineur Lv.1) or forge the Stalker Halo and Gilded Halberd (Craftsman Lv.1).",
		Location:    "Gilded Ore Vein (-3.5, -9.0)",
		RoomX:       &hubZero,
		RoomY:       &hubZero,
		IconItemID:  "plug_stalker_halberd",
		ItemIDs:     []string{"plug_bounty_ore", "plug_bounty_elixir", "plug_stalker_crown", "plug_stalker_halberd"},
		Tags:        []string{"gilded", "bounty", "forge", "guide", "stalker"},
	})
	s.loadPersistedState()
}

func (s *bountyServer) OnBeginCharacterCreate(ev sdk.BeginCharacterCreateEvent) (sdk.CharacterCreateConfig, bool) {
	return sdk.CharacterCreateConfig{
		Mode:     sdk.CharCreateModeVanillaExtended,
		Title:    "BOUNTY GUILD INITIATION",
		Subtitle: "Pick a Bounty Guild origin for " + ev.DisplayName + " (Vanilla Saved Looks apply)",
		Classes:  bountyStarterClasses(),
	}, true
}

func (s *bountyServer) OnFinalizeCharacterCreate(ev sdk.FinalizeCharacterCreateEvent) (sdk.CharacterCreateResult, error) {
	name := strings.TrimSpace(ev.HeroName)
	if len(name) < 3 {
		return sdk.CharacterCreateResult{}, errors.New("hero name must be at least 3 characters")
	}
	switch strings.ToLower(name) {
	case "stalker", "harbinger", "bounty", "gilded", "admin", "system":
		return sdk.CharacterCreateResult{}, errors.New("reserved guild name rejected by Bounty Board")
	}

	origin := bountyOrigins[0]
	matched := false
	for _, candidate := range bountyOrigins {
		if candidate.Class.ID == ev.ClassID {
			origin = candidate
			matched = true
			break
		}
	}
	if !matched && strings.TrimSpace(ev.ClassID) != "" {
		return sdk.CharacterCreateResult{}, errors.New("unknown bounty guild origin: " + ev.ClassID)
	}

	tpl := challenges[origin.StartChallenge%len(challenges)]
	rawHero, _ := json.Marshal(heroBountyState{
		HeroName:     name,
		OriginID:     origin.Class.ID,
		OriginName:   origin.Class.Name,
		ChallengeIdx: origin.StartChallenge,
		VisitedRooms: make(map[string]bool),
		LastEvent:    origin.Class.Name + " vows sworn -- Bounty: " + tpl.Title,
	})
	return sdk.CharacterCreateResult{
		HeroName:            name,
		Equipped:            origin.Class.Equipped,
		OverrideEquipped:    true,
		StarterBag:          origin.StarterBag,
		OverrideStarterBag:  true,
		StarterGold:         origin.StarterGold,
		OverrideStarterGold: true,
		BaseHP:              100 + origin.Class.BonusHP,
		BaseAP:              6 + origin.Class.BonusAP,
		BaseMP:              6 + origin.Class.BonusMP,
		CustomKV:            map[string][]byte{"hero:{hero_id}": rawHero},
		WelcomeNotice:       origin.Class.Name + " initiated! Active bounty: " + tpl.Title + ".",
	}, nil
}

func (s *bountyServer) ensureHero(heroID, heroName string) *heroBountyState {
	st, ok := s.heroes[heroID]
	if !ok {
		var kvSt heroBountyState
		if foundKV, err := sdk.KVGetJSON("hero:"+heroID, &kvSt); err == nil && foundKV &&
			(kvSt.HeroID != "" || kvSt.OriginID != "" || kvSt.OriginName != "") {
			kvSt.HeroID = heroID
			kvSt.Online = false
			if kvSt.VisitedRooms == nil {
				kvSt.VisitedRooms = make(map[string]bool)
			}
			if kvSt.OriginName == "" {
				kvSt.OriginName = "Freelance"
			}
			st = &kvSt
		} else {
			st = &heroBountyState{
				HeroID:       heroID,
				HeroName:     heroName,
				OriginName:   "Freelance",
				VisitedRooms: make(map[string]bool),
				LastEvent:    "Accepted bounty: " + challenges[0].Title,
			}
		}
		s.heroes[heroID] = st
	}
	if st.OriginName == "" {
		st.OriginName = "Freelance"
	}
	if heroName != "" {
		st.HeroName = heroName
	}
	return st
}

func rankForCompleted(n int) string {
	switch {
	case n >= 6:
		return "Grand Champion"
	case n >= 3:
		return "Seasoned Seeker"
	case n >= 1:
		return "Bounty Scout"
	default:
		return "Initiate"
	}
}

func (s *bountyServer) snapshot(st *heroBountyState) proto.BountySnapshot {
	tpl := challenges[st.ChallengeIdx%len(challenges)]
	return proto.BountySnapshot{
		Kind:           tpl.Kind,
		Title:          tpl.Title,
		Description:    tpl.Description,
		Progress:       st.Progress,
		Target:         tpl.Target,
		Completed:      st.Completed,
		CompletedCount: st.CompletedCount,
		RealmCheers:    s.realmCheers,
		OriginName:     st.OriginName,
		RankTitle:      rankForCompleted(st.CompletedCount),
		LastEvent:      st.LastEvent,
	}
}

func (s *bountyServer) pushHero(st *heroBountyState) {
	if st == nil || !st.Online {
		return
	}
	sdk.SendJSONToHero(st.HeroID, proto.ChannelBounty, s.snapshot(st))
}

func (s *bountyServer) advanceProgress(st *heroBountyState, delta int, evMsg string) {
	tpl := challenges[st.ChallengeIdx%len(challenges)]
	st.LastEvent = evMsg
	if !st.Completed {
		st.Progress += delta
		if st.Progress >= tpl.Target {
			st.Progress = tpl.Target
			st.Completed = true
			st.CompletedCount++
			rewardQty := 1
			if tpl.Kind == "hunt" {
				rewardQty = 2
			}
			sdk.GiveItemToHero(st.HeroID, "plug_bounty_ore", rewardQty)
			st.LastEvent = fmt.Sprintf("Completed %s! (+%d Gilded Ore)", tpl.Title, rewardQty)
			sdk.BroadcastNotice("info", "Bounty Complete",
				fmt.Sprintf("%s completed the '%s' bounty!", st.HeroName, tpl.Title))
		}
	}
	s.saveHeroState(st)
	s.pushHero(st)
}

func (s *bountyServer) OnPlayerJoin(hero sdk.HeroSnapshot) {
	st := s.ensureHero(hero.ID, hero.Name)
	st.Online = true
	st.RoomX = hero.RoomX
	st.RoomY = hero.RoomY
	st.PosX = hero.PosX
	st.PosZ = hero.PosZ
	roomKey := fmt.Sprintf("%d,%d", hero.RoomX, hero.RoomY)
	st.VisitedRooms[roomKey] = true
	s.saveHeroState(st)
	s.pushHero(st)
}

func (s *bountyServer) OnPlayerLeave(heroID string) {
	if st, ok := s.heroes[heroID]; ok {
		st.Online = false
	}
}

func (s *bountyServer) OnRoomEnter(ev sdk.RoomEvent) {
	st := s.ensureHero(ev.HeroID, ev.HeroName)
	st.Online = true
	st.RoomX = ev.RoomX
	st.RoomY = ev.RoomY
	tpl := challenges[st.ChallengeIdx%len(challenges)]
	roomKey := fmt.Sprintf("%d,%d", ev.RoomX, ev.RoomY)
	alreadyVisited := st.VisitedRooms[roomKey]
	st.VisitedRooms[roomKey] = true

	if tpl.Kind == "explore" && !st.Completed && !alreadyVisited {
		s.advanceProgress(st, 1, fmt.Sprintf("Explored chunk (%d, %d).", ev.RoomX, ev.RoomY))
		return
	}
	if !alreadyVisited {
		s.saveHeroState(st)
	}
	s.pushHero(st)
}

func (s *bountyServer) OnChat(ev sdk.ChatEvent) {
	st := s.ensureHero(ev.HeroID, ev.HeroName)
	tpl := challenges[st.ChallengeIdx%len(challenges)]
	if tpl.Kind == "chat" && !st.Completed {
		s.advanceProgress(st, 1, fmt.Sprintf("Spoke in %s chat.", ev.Channel))
	}
}

func (s *bountyServer) OnPacket(senderHeroID string, channel uint16, data []byte) {
	if channel == proto.ChannelSandbox {
		var req proto.SandboxProbeReq
		if err := json.Unmarshal(data, &req); err != nil {
			req = proto.SandboxProbeReq{Path: "/", Mode: "explore"}
		}
		resp := proto.RunSandboxProbes("server", req.Path, req.Mode)
		errPreview := sdk.SetCharacterCreatePreview(sdk.CharacterCreatePreviewSpec{CameraFocus: "face"})
		errSubmit := sdk.SubmitCharacterCreate(sdk.CharacterCreateSubmission{Name: "Escalate"})
		blockedClientHostcall := errPreview != nil && errSubmit != nil
		detail := fmt.Sprintf("Rejected (preview=%v, submit=%v): RoleServer blocked", errPreview != nil, errSubmit != nil)
		if !blockedClientHostcall {
			detail = fmt.Sprintf("UNEXPECTED CLIENT HOSTCALL SUCCEEDED (preview=%v, submit=%v)", errPreview, errSubmit)
			resp.AllBlocked = false
		}
		resp.Checks = append(resp.Checks, proto.SandboxProbeCheck{
			Name:    "Client Hostcalls",
			Target:  "CreatePreview/SubmitCreate",
			Blocked: blockedClientHostcall,
			Detail:  detail,
		})
		sdk.SendJSONToHero(senderHeroID, proto.ChannelSandbox, resp)
		return
	}
	if channel != proto.ChannelBounty {
		return
	}
	var act proto.ClientAction
	if err := json.Unmarshal(data, &act); err != nil {
		return
	}
	st := s.ensureHero(senderHeroID, "")
	st.Online = true

	switch act.Action {
	case "sync":
		s.pushHero(st)

	case "reroll":
		st.ChallengeIdx = (st.ChallengeIdx + 1) % len(challenges)
		st.Progress = 0
		st.Completed = false
		st.VisitedRooms = make(map[string]bool)
		tpl := challenges[st.ChallengeIdx]
		st.LastEvent = "Rolled new bounty: " + tpl.Title
		s.saveHeroState(st)
		s.pushHero(st)

	case "cheer":
		s.realmCheers++
		s.saveGlobalState()
		note := strings.TrimSpace(act.Note)
		if len(note) > 60 {
			note = note[:60]
		}
		banner := fmt.Sprintf("%s cheered for the realm! (Total Cheers: %d)", st.HeroName, s.realmCheers)
		if note != "" {
			banner = fmt.Sprintf("%s: %s (Cheers: %d)", st.HeroName, note, s.realmCheers)
		}
		sdk.BroadcastNotice("info", "Realm Cheer", banner)

		tpl := challenges[st.ChallengeIdx%len(challenges)]
		if tpl.Kind == "cheer" && !st.Completed {
			s.advanceProgress(st, 1, "Rallied the realm!")
		} else {
			st.LastEvent = fmt.Sprintf("Sent realm cheer #%d!", s.realmCheers)
			s.saveHeroState(st)
			s.pushHero(st)
		}
		// Refresh RealmCheers counter on all other connected players.
		for _, other := range s.heroes {
			if other != nil && other.Online && other.HeroID != st.HeroID {
				other.LastEvent = banner
				s.pushHero(other)
			}
		}

	case "summon_boss":
		spawnX := st.PosX + 4.0
		spawnZ := st.PosZ + 2.0
		if spawnX > 14 {
			spawnX = st.PosX - 4.0
		}
		spawnID := sanitizeSpawnID("bsum_" + st.HeroID)
		if res := sdk.SpawnMonster(sdk.MonsterSpawnSpec{
			ID:          spawnID,
			ArchetypeID: bountyMonsterArchID,
			RoomX:       st.RoomX,
			RoomY:       st.RoomY,
			PosX:        spawnX,
			PosZ:        spawnZ,
			PackSize:    1,
			RespawnSec:  -1,
		}); res == 0 {
			st.LastEvent = fmt.Sprintf("Summoned Gilded Stalker in room (%d, %d)!", st.RoomX, st.RoomY)
			sdk.BroadcastNotice("warn", "Gilded Stalker Spotted",
				fmt.Sprintf("%s summoned a Gilded Stalker in chunk (%d, %d)!", st.HeroName, st.RoomX, st.RoomY))
		} else {
			st.LastEvent = "Could not summon Gilded Stalker in this area."
		}
		s.saveHeroState(st)
		s.pushHero(st)
	}
}

func (s *bountyServer) OnBattleEnd(ev sdk.BattleEndEvent) {
	if !ev.Won || ev.ArchetypeID != bountyMonsterArchID {
		return
	}
	sdk.BroadcastNotice("info", "Gilded Stalker Slain",
		fmt.Sprintf("%s defeated the Gilded Stalker in chunk (%d, %d)!", strings.Join(ev.HeroNames, ", "), ev.RoomX, ev.RoomY))
	for _, heroID := range ev.HeroIDs {
		st, ok := s.heroes[heroID]
		if !ok {
			continue
		}
		tpl := challenges[st.ChallengeIdx%len(challenges)]
		if tpl.Kind == "hunt" && !st.Completed {
			s.advanceProgress(st, 1, "Slew the Gilded Stalker!")
		} else {
			st.LastEvent = "Assisted in defeating a Gilded Stalker!"
			s.saveHeroState(st)
			s.pushHero(st)
		}
	}
}

func (s *bountyServer) OnGather(ev sdk.GatherEvent) {
	st := s.ensureHero(ev.HeroID, ev.HeroName)
	st.Online = true
	tpl := challenges[st.ChallengeIdx%len(challenges)]
	msg := fmt.Sprintf("Gathered +%d %s.", ev.Quantity, ev.ItemID)
	if ev.NodeID == "bounty_gilded_vein" {
		msg = fmt.Sprintf("Mined +%d Gilded Ore from Gilded Ore Vein!", ev.Quantity)
	}
	if tpl.Kind == "gather" && !st.Completed {
		s.advanceProgress(st, ev.Quantity, msg)
		return
	}
	st.LastEvent = msg
	s.saveHeroState(st)
	s.pushHero(st)
}

func (s *bountyServer) OnCraft(ev sdk.CraftEvent) {
	st := s.ensureHero(ev.HeroID, ev.HeroName)
	st.Online = true
	st.LastEvent = fmt.Sprintf("Crafted %dx %s.", ev.Quantity, ev.ItemID)
	if ev.ItemID == "plug_stalker_crown" || ev.ItemID == "plug_stalker_halberd" {
		sdk.BroadcastNotice("info", "Relic Forged",
			fmt.Sprintf("%s forged %s via %s!", ev.HeroName, ev.ItemID, ev.Job))
	}
	s.saveHeroState(st)
	s.pushHero(st)
}

func (s *bountyServer) OnNPCInteract(ev sdk.NPCInteractEvent) (sdk.NPCDialogResult, bool) {
	return vanilla.HandleNPCInteract(ev)
}

func (s *bountyServer) OnInteract(ev sdk.InteractEvent) bool {
	if ev.InteractableID != "gilded_vault_altar" {
		return false
	}
	sdk.GiveItemToHero(ev.HeroID, "plug_bounty_ore", 1)
	sdk.SendNoticeToHero(ev.HeroID, "info", "Gilded Vault Altar", "Golden light condenses inside your satchel (+1 Gilded Ore).")
	return true
}

func init() {
	sdk.RegisterServer(newBountyServer())
}

func sanitizeSpawnID(raw string) string {
	var b strings.Builder
	for i := 0; i < len(raw) && b.Len() < 36; i++ {
		c := raw[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			b.WriteByte(c)
		}
	}
	if b.Len() < 3 {
		return "bsum_boss"
	}
	return b.String()
}

func main() {}
