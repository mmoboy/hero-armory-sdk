// Package vanilla provides composable, opt-in classic Sanctuary realm content
// (rooms, equipment sets, portal styles, jobs, craft recipes, gather nodes,
// monsters, tactical skills, and NPCs) implemented cleanly through hero-armory-sdk.
package vanilla

import (
	sdk "github.com/mmoboy/hero-armory-sdk"
)

// RegisterAll registers the full Sanctuary world: 25 rooms, all items & starter
// configuration, jobs/recipes/gather nodes, tactical skills & monsters, and NPCs.
func RegisterAll() {
	RegisterRooms()
	RegisterItems()
	RegisterJobsAndGatherNodes()
	RegisterCombat()
	RegisterNPCs()
}

// KokiriGear returns the Kokiri Guardian 8-slot equipment ID map.
func KokiriGear() map[string]string {
	return map[string]string{
		"hat":    "hat_kokiri_cap",
		"amulet": "amulet_kokiri_spiritual",
		"chest":  "chest_kokiri_tunic",
		"belt":   "belt_kokiri_girdle",
		"gloves": "gloves_kokiri_gauntlets",
		"ring":   "ring_kokiri_fairy",
		"boots":  "boots_kokiri_pegasus",
		"weapon": "weapon_master_blade",
	}
}

// GobballGear returns the Gobball Warchief 8-slot equipment ID map.
func GobballGear() map[string]string {
	return map[string]string{
		"hat":    "hat_gobball_headgear",
		"amulet": "amulet_gobball_talisman",
		"chest":  "chest_gobball_cuirass",
		"belt":   "belt_gobball_cord",
		"gloves": "gloves_gobball_mitts",
		"ring":   "ring_gobball_chanoine",
		"boots":  "boots_gobball_sabatons",
		"weapon": "weapon_goultard_cleaver",
	}
}

// WakfuGear returns the Wakfu Chronomancer 8-slot equipment ID map.
func WakfuGear() map[string]string {
	return map[string]string{
		"hat":    "hat_wakfu_chronocowl",
		"amulet": "amulet_wakfu_pendulum",
		"chest":  "chest_wakfu_robe",
		"belt":   "belt_wakfu_sash",
		"gloves": "gloves_wakfu_bracers",
		"ring":   "ring_wakfu_chanoine",
		"boots":  "boots_wakfu_striders",
		"weapon": "weapon_xelor_staff",
	}
}

// TwilightGear returns the Twilight Sanctuary 8-slot equipment ID map.
func TwilightGear() map[string]string {
	return map[string]string{
		"hat":    "hat_twilight_crown",
		"amulet": "amulet_twilight_tear",
		"chest":  "chest_twilight_hauberk",
		"belt":   "belt_twilight_girdle",
		"gloves": "gloves_twilight_grips",
		"ring":   "ring_twilight_magnesis",
		"boots":  "boots_twilight_greaves",
		"weapon": "weapon_guardian_halberd",
	}
}

// SanctuaryStarterConfig returns the classic Sanctuary starter hero configuration.
func SanctuaryStarterConfig() sdk.StarterConfigSpec {
	tailItems := []string{
		"mat_wild_mint", "mat_moon_petal", "mat_star_bloom",
		"mat_river_perch", "mat_silver_trout", "mat_golden_koi",
		"mat_refined_plank", "mat_forged_ingot", "mat_artisan_core",
		"mat_oak_wood", "mat_elder_sap", "mat_copper_ore",
		"mat_ash_wood", "mat_iron_ore",
	}
	tailBag := make([]sdk.ItemStackSpec, 0, len(tailItems))
	for _, id := range tailItems {
		tailBag = append(tailBag, sdk.ItemStackSpec{ItemID: id, Quantity: 2})
	}
	return sdk.StarterConfigSpec{
		SpawnRoomX:      0,
		SpawnRoomY:      0,
		SpawnPos:        [3]float32{0, 0, 2.2},
		StarterGold:     2500,
		StarterEquipped: KokiriGear(),
		StarterBag: []sdk.ItemStackSpec{
			{ItemID: "cons_wakfu_potion", Quantity: 5},
			{ItemID: "cons_red_heart_elixir", Quantity: 3},
			{ItemID: "mat_wakfu_crystal", Quantity: 12},
			{ItemID: "mat_gobball_wool", Quantity: 18},
			{ItemID: "mat_kokiri_emerald", Quantity: 25},
			{ItemID: "mat_sheikah_core", Quantity: 4},
		},
		StarterTailBag:      tailBag,
		IncludeAllEquip:     true,
		IncludePortalStyles: true,
		BaseHP:              100,
		BaseMP:              6.0,
		BaseAP:              6,
	}
}
