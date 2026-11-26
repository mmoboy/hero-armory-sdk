package vanilla

import (
	"fmt"

	sdk "github.com/mmoboy/hero-armory-sdk"
	"github.com/mmoboy/hero-armory-sdk/tactical"
)

// TacticalMonsters returns the 4 classic Sanctuary monsters adapted as
// tactical.MobTemplate archetypes with signature spells.
func TacticalMonsters() []tactical.MobTemplate {
	return []tactical.MobTemplate{
		{
			ID: "gobball", Name: "Wild Gobball", Role: "bruiser",
			BaseHP: 65, BaseDamage: 12, BaseAP: 4, BaseMP: 6.5, Range: 2.8,
			GoldReward: 18, XPReward: 28,
			BodyColor: [3]float32{0.92, 0.88, 0.78}, TrimColor: [3]float32{0.58, 0.36, 0.22}, Scale: 0.98,
			DefaultTier: tactical.TierNormal,
			Description: "Woolly horned grazer with a stubborn headbutt charge.",
			Spells: []tactical.MobSpellSpec{
				{
					ID: "wool_ram", Name: "Wool Ram", APCost: 3, Cooldown: 2, Range: 3.2,
					Damage: 18, EffectKind: "damage", VFXKind: "slash", VFXColor: [4]float32{0.94, 0.84, 0.56, 1},
				},
			},
		},
		{
			ID: "forest_slime", Name: "Verdant Slime", Role: "caster",
			BaseHP: 78, BaseDamage: 15, BaseAP: 5, BaseMP: 5.5, Range: 7.5,
			GoldReward: 24, XPReward: 40,
			BodyColor: [3]float32{0.24, 0.82, 0.44}, TrimColor: [3]float32{0.10, 0.48, 0.24}, Scale: 1.05,
			DefaultTier: tactical.TierNormal,
			Description: "Emerald ooze saturated with forest mana, lobbing acidic bolts.",
			Spells: []tactical.MobSpellSpec{
				{
					ID: "acid_glob", Name: "Acid Glob", APCost: 3, Cooldown: 2, Range: 8.0,
					Damage: 17, EffectKind: "damage",
					StatusID: "poison", StatusLabel: "Poison", StatusTurns: 2, StatusPower: 6, StatusColor: [3]float32{0.34, 0.96, 0.44},
					VFXKind: "bolt", VFXColor: [4]float32{0.30, 0.94, 0.46, 1},
				},
			},
		},
		{
			ID: "gobball_warchief", Name: "Gobball Warchief", Role: "bruiser",
			BaseHP: 118, BaseDamage: 20, BaseAP: 5, BaseMP: 7.0, Range: 3.2,
			GoldReward: 44, XPReward: 68,
			BodyColor: [3]float32{0.80, 0.30, 0.24}, TrimColor: [3]float32{0.92, 0.74, 0.28}, Scale: 1.28,
			DefaultTier: tactical.TierVeteran,
			Description: "Crimson-horned pack leader guarding the outer frontier.",
			Spells: []tactical.MobSpellSpec{
				{
					ID: "warchief_cleave", Name: "Warchief Cleave", APCost: 3, Cooldown: 2, Range: 3.6,
					Damage: 26, EffectKind: "damage", VFXKind: "slash", VFXColor: [4]float32{0.96, 0.42, 0.22, 1},
				},
				{
					ID: "war_cry_ward", Name: "War Cry Ward", APCost: 2, Cooldown: 3, Range: 0,
					Shield: 26, EffectKind: "shield", VFXKind: "shield", VFXColor: [4]float32{0.96, 0.76, 0.26, 1},
				},
			},
		},
		{
			ID: "twilight_wolf", Name: "Twilight Stalker", Role: "assassin",
			BaseHP: 128, BaseDamage: 24, BaseAP: 6, BaseMP: 8.5, Range: 3.4,
			GoldReward: 58, XPReward: 88,
			BodyColor: [3]float32{0.34, 0.24, 0.56}, TrimColor: [3]float32{0.82, 0.42, 0.96}, Scale: 1.18,
			DefaultTier: tactical.TierHeroic,
			Description: "Swift shadow-cursed hound prowling the outer perimeter.",
			Spells: []tactical.MobSpellSpec{
				{
					ID: "shadow_pounce", Name: "Shadow Pounce", APCost: 3, Cooldown: 2, Range: 4.5,
					Damage: 28, EffectKind: "damage",
					StatusID: "burn", StatusLabel: "Hex", StatusTurns: 2, StatusPower: 9, StatusColor: [3]float32{0.82, 0.36, 1.0},
					VFXKind: "nova", VFXColor: [4]float32{0.78, 0.36, 1.0, 1},
				},
			},
		},
	}
}

// TacticalEncounters returns escalating API Level 5/6 tactical encounters
// across all 25 classic Sanctuary rooms (-2..2, -2..2).
func TacticalEncounters() []tactical.EncounterSpec {
	out := make([]tactical.EncounterSpec, 0, 49)
	for y := -2; y <= 2; y++ {
		for x := -2; x <= 2; x++ {
			tag := fmt.Sprintf("tac_%d_%d", x+2, y+2)
			if x == 0 && y == 0 {
				out = append(out, tactical.EncounterSpec{
					ID: tag + "_0", RoomX: 0, RoomY: 0, PosX: 11.5, PosZ: 11.5,
					TemplateID: "gobball", Level: 2, PackSize: 1, RespawnSec: 18,
				})
				continue
			}
			dist := absInt(x) + absInt(y)
			switch dist {
			case 1:
				out = append(out,
					tactical.EncounterSpec{ID: tag + "_0", RoomX: x, RoomY: y, PosX: -8.5, PosZ: -7.5, TemplateID: "gobball", Level: 3, PackSize: 2, RespawnSec: 20},
					tactical.EncounterSpec{ID: tag + "_1", RoomX: x, RoomY: y, PosX: 9.0, PosZ: 8.0, TemplateID: "forest_slime", Level: 4, PackSize: 2, RespawnSec: 22},
				)
			case 2:
				out = append(out,
					tactical.EncounterSpec{ID: tag + "_0", RoomX: x, RoomY: y, PosX: -8.5, PosZ: -7.5, TemplateID: "gobball_warchief", Level: 5, Tier: tactical.TierVeteran, PackSize: 2, RespawnSec: 25},
					tactical.EncounterSpec{ID: tag + "_1", RoomX: x, RoomY: y, PosX: 9.0, PosZ: 8.0, TemplateID: "frost_sentinel", Level: 6, Tier: tactical.TierVeteran, PackSize: 1, RespawnSec: 25},
				)
			case 3:
				out = append(out,
					tactical.EncounterSpec{ID: tag + "_0", RoomX: x, RoomY: y, PosX: -8.5, PosZ: -7.5, TemplateID: "twilight_wolf", Level: 7, Tier: tactical.TierHeroic, PackSize: 2, RespawnSec: 28},
					tactical.EncounterSpec{ID: tag + "_1", RoomX: x, RoomY: y, PosX: 9.0, PosZ: 8.0, TemplateID: "ember_golem", Level: 8, Tier: tactical.TierHeroic, PackSize: 1, RespawnSec: 28},
				)
			default: // dist == 4 (outer corners)
				out = append(out,
					tactical.EncounterSpec{ID: tag + "_0", RoomX: x, RoomY: y, PosX: -8.5, PosZ: -7.5, TemplateID: "shadow_assassin", Level: 9, Tier: tactical.TierHeroic, PackSize: 2, RespawnSec: 30},
					tactical.EncounterSpec{ID: tag + "_1", RoomX: x, RoomY: y, PosX: 8.5, PosZ: 7.5, TemplateID: "astral_overlord", Level: 10, Tier: tactical.TierMythic, PackSize: 1, RespawnSec: 40},
				)
			}
		}
	}
	return out
}

// RegisterTacticalEncounters registers Sanctuary's tactical monster templates
// and spawns all 25-room tactical encounters onto eng.
func RegisterTacticalEncounters(eng *tactical.Engine) {
	for _, m := range TacticalMonsters() {
		eng.RegisterMonster(m)
	}
	for _, enc := range TacticalEncounters() {
		_ = eng.AddEncounter(enc)
	}
}

func itemNameByID(itemID string) string {
	for _, it := range StackableItems() {
		if it.ID == itemID {
			return it.Name
		}
	}
	for _, it := range EquipmentItems() {
		if it.ID == itemID {
			return it.Name
		}
	}
	for _, it := range PortalStyleItems() {
		if it.ID == itemID {
			return it.Name
		}
	}
	return titleCaseID(itemID)
}

// AwardTacticalLoot grants guaranteed Sanctuary material/consumable drops when
// a hero wins a tactical battle and returns the awarded drops for the Victory
// Loot Summary screen.
func AwardTacticalLoot(res tactical.BattleResult) []tactical.LootDrop {
	if !res.Won || res.HeroID == "" {
		return nil
	}
	var pairs []sdk.ItemStackSpec
	switch res.TemplateID {
	case "gobball":
		pairs = []sdk.ItemStackSpec{{ItemID: "mat_gobball_wool", Quantity: 2}}
	case "forest_slime":
		pairs = []sdk.ItemStackSpec{{ItemID: "mat_kokiri_emerald", Quantity: 2}}
	case "gobball_warchief":
		pairs = []sdk.ItemStackSpec{
			{ItemID: "mat_gobball_wool", Quantity: 3},
			{ItemID: "mat_wakfu_crystal", Quantity: 1},
		}
	case "twilight_wolf":
		pairs = []sdk.ItemStackSpec{
			{ItemID: "mat_sheikah_core", Quantity: 1},
			{ItemID: "mat_wakfu_crystal", Quantity: 2},
		}
	case "astral_overlord":
		pairs = []sdk.ItemStackSpec{
			{ItemID: "mat_sheikah_core", Quantity: 2},
			{ItemID: "cons_red_heart_elixir", Quantity: 1},
		}
	default:
		pairs = []sdk.ItemStackSpec{{ItemID: "mat_wakfu_crystal", Quantity: 1}}
	}
	drops := make([]tactical.LootDrop, 0, len(pairs))
	for _, p := range pairs {
		sdk.GiveItemToHero(res.HeroID, p.ItemID, p.Quantity)
		drops = append(drops, tactical.LootDrop{
			ItemID: p.ItemID,
			Name:   itemNameByID(p.ItemID),
			Qty:    p.Quantity,
		})
	}
	return drops
}
