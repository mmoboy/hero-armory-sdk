package vanilla

import (
	"fmt"

	"hero-armory-sdk"
)

// Skills returns the 3 classic Sanctuary tactical player combat skills.
func Skills() []sdk.PlayerSkillSpec {
	return []sdk.PlayerSkillSpec{
		{
			ID:          "strike",
			Name:        "Heavy Strike",
			IconName:    "ha_spell_strike",
			APCost:      2,
			Range:       2.8,
			AoERadius:   0,
			BaseDamage:  20,
			Pose:        "slash",
			Description: "Melee strike (2 AP, 2.8m)",
			TargetKind:  "enemy",
			EffectKind:  "damage",
			VFXKind:     "slash",
		},
		{
			ID:          "arcane_bolt",
			Name:        "Arcane Bolt",
			IconName:    "ha_spell_bolt",
			APCost:      3,
			Range:       11.5,
			AoERadius:   0,
			BaseDamage:  24,
			Pose:        "channel",
			Description: "Ranged bolt (3 AP, 11.5m)",
			TargetKind:  "enemy",
			EffectKind:  "damage",
			VFXKind:     "bolt",
		},
		{
			ID:          "arcane_nova",
			Name:        "Starlight Nova",
			IconName:    "ha_spell_nova",
			APCost:      4,
			Range:       7.0,
			AoERadius:   3.2,
			BaseDamage:  18,
			Pose:        "heroic",
			Description: "AoE blast (4 AP, 7m, 3.2m AoE)",
			TargetKind:  "enemy",
			EffectKind:  "damage",
			VFXKind:     "nova",
		},
	}
}

// Monsters returns the 4 classic Sanctuary monster archetypes.
func Monsters() []sdk.MonsterSpec {
	return []sdk.MonsterSpec{
		{
			ID: "gobball", Name: "Wild Gobball", Level: 2, MaxHP: 55, MovePoints: 8.0,
			AttackPts: 4, Damage: 12, Range: 2.6, GoldReward: 15, XPReward: 25,
			BodyColor: [3]float32{0.92, 0.88, 0.78}, TrimColor: [3]float32{0.58, 0.36, 0.22}, Scale: 0.95,
			Description: "Woolly horned grazer with a surprisingly stubborn headbutt. A staple sparring target for new adventurers.",
			Habitat:     "Sanctuary Courtyard & Inner Meadows",
			LootTable: []sdk.MonsterLoot{
				{ItemID: "mat_gobball_wool", Chance: 0.75, MinQty: 1, MaxQty: 3},
				{ItemID: "cons_wakfu_potion", Chance: 0.25, MinQty: 1, MaxQty: 1},
				{ItemID: "hat_gobball_headgear", Chance: 0.10, MinQty: 1, MaxQty: 1},
			},
		},
		{
			ID: "forest_slime", Name: "Verdant Slime", Level: 4, MaxHP: 72, MovePoints: 7.0,
			AttackPts: 4, Damage: 15, Range: 5.5, GoldReward: 22, XPReward: 38,
			BodyColor: [3]float32{0.24, 0.82, 0.44}, TrimColor: [3]float32{0.10, 0.48, 0.24}, Scale: 1.05,
			Description: "Emerald gelatinous ooze saturated with forest mana, capable of lobbing acidic bolts from medium range.",
			Habitat:     "Inner Glades & Verdant Crossing",
			LootTable: []sdk.MonsterLoot{
				{ItemID: "mat_kokiri_emerald", Chance: 0.70, MinQty: 1, MaxQty: 2},
				{ItemID: "cons_red_heart_elixir", Chance: 0.30, MinQty: 1, MaxQty: 1},
				{ItemID: "portal_style_emerald", Chance: 0.08, MinQty: 1, MaxQty: 1},
			},
		},
		{
			ID: "gobball_warchief", Name: "Gobball Warchief", Level: 6, MaxHP: 110, MovePoints: 9.0,
			AttackPts: 6, Damage: 20, Range: 2.8, GoldReward: 40, XPReward: 65,
			BodyColor: [3]float32{0.80, 0.30, 0.24}, TrimColor: [3]float32{0.92, 0.74, 0.28}, Scale: 1.28,
			Description: "Crimson-horned pack leader that charges relentlessly and guards rare crimson portal resonance shards.",
			Habitat:     "Outer Ring Frontier (Dist >= 2)",
			LootTable: []sdk.MonsterLoot{
				{ItemID: "mat_gobball_wool", Chance: 0.90, MinQty: 2, MaxQty: 4},
				{ItemID: "mat_wakfu_crystal", Chance: 0.45, MinQty: 1, MaxQty: 2},
				{ItemID: "weapon_goultard_cleaver", Chance: 0.18, MinQty: 1, MaxQty: 1},
				{ItemID: "portal_style_crimson", Chance: 0.10, MinQty: 1, MaxQty: 1},
			},
		},
		{
			ID: "twilight_wolf", Name: "Twilight Stalker", Level: 8, MaxHP: 125, MovePoints: 11.0,
			AttackPts: 6, Damage: 24, Range: 3.0, GoldReward: 55, XPReward: 85,
			BodyColor: [3]float32{0.34, 0.24, 0.56}, TrimColor: [3]float32{0.82, 0.42, 0.96}, Scale: 1.18,
			Description: "Swift shadow-cursed hound prowling the outer perimeter. Coveted for Sheikah cores and Twilight arms.",
			Habitat:     "Twilight Frontier & Abyssal Sectors (Dist >= 2)",
			LootTable: []sdk.MonsterLoot{
				{ItemID: "mat_sheikah_core", Chance: 0.65, MinQty: 1, MaxQty: 2},
				{ItemID: "mat_wakfu_crystal", Chance: 0.50, MinQty: 1, MaxQty: 2},
				{ItemID: "portal_style_frost", Chance: 0.12, MinQty: 1, MaxQty: 1},
				{ItemID: "weapon_guardian_halberd", Chance: 0.14, MinQty: 1, MaxQty: 1},
			},
		},
	}
}

// MonsterSpawnsForRoom returns deterministic monster packs for Sanctuary room (x, y).
func MonsterSpawnsForRoom(x, y int) []sdk.MonsterSpawnSpec {
	tag := fmt.Sprintf("mob_%d_%d", x, y)
	if x == 0 && y == 0 {
		return []sdk.MonsterSpawnSpec{
			{
				ID: tag + "_0", ArchetypeID: "gobball",
				RoomX: 0, RoomY: 0, PosX: 11.5, PosZ: 11.5, PackSize: 2,
			},
		}
	}
	dist := absInt(x) + absInt(y)
	first := "gobball"
	second := "forest_slime"
	if dist >= 2 {
		first = "gobball_warchief"
		second = "twilight_wolf"
	}
	return []sdk.MonsterSpawnSpec{
		{ID: tag + "_0", ArchetypeID: first, RoomX: x, RoomY: y, PosX: -8.5, PosZ: -7.5, PackSize: 2},
		{ID: tag + "_1", ArchetypeID: second, RoomX: x, RoomY: y, PosX: 9.0, PosZ: 8.0, PackSize: 3},
	}
}

// RegisterCombat registers all Sanctuary tactical skills, monster archetypes, and 5x5 monster packs.
func RegisterCombat() {
	for _, sk := range Skills() {
		sdk.RegisterPlayerSkill(sk)
	}
	for _, m := range Monsters() {
		sdk.RegisterMonster(m)
	}
	for y := -2; y <= 2; y++ {
		for x := -2; x <= 2; x++ {
			for _, sp := range MonsterSpawnsForRoom(x, y) {
				sdk.SpawnMonster(sp)
			}
		}
	}
}
