package vanilla

import (
	sdk "github.com/mmoboy/hero-armory-sdk"
)

// CodexEntries returns classic Sanctuary realm guide entries for the Codex.
func CodexEntries() []sdk.CodexEntrySpec {
	zero := 0
	return []sdk.CodexEntrySpec{
		{
			ID:          "sanctuary_guide_portals",
			Category:    "custom",
			Title:       "Sanctuary 5x5 Portals",
			Subtitle:    "Exploring the Realm Grid",
			Description: "Sanctuary spans a 5x5 grid from (-2,-2) to (2,2) across Emerald Glade, Astral Forge, Sunset Dunes, Cindervault, and Twilight Crypt. Step onto glowing doorway pads to cross rooms, press [M] for the map, or type /sanctuary in chat to return to (0,0).",
			KindBadge:   "GUIDE",
			Location:    "Sanctuary Courtyard (0,0)",
			RoomX:       &zero,
			RoomY:       &zero,
			Tags:        []string{"portals", "map", "navigation", "sanctuary"},
			AccentColor: [4]float32{0.24, 0.86, 0.98, 1.0},
			SortOrder:   10,
		},
		{
			ID:          "sanctuary_guide_sets",
			Category:    "custom",
			Title:       "Relic Armor Sets",
			Subtitle:    "Kokiri, Gobball, Wakfu & Twilight",
			Description: "Speak with Forgekeeper Lyra in Sanctuary Courtyard or biome stewards in the cardinal rooms to swap between the four classic 8-slot Sanctuary sets, or browse Merchant Orion for gear, elixirs, and Black Hole Portal Styles.",
			KindBadge:   "GEAR",
			Location:    "Sanctuary Courtyard (0,0)",
			RoomX:       &zero,
			RoomY:       &zero,
			IconItemID:  "weapon_master_blade",
			ItemIDs:     []string{"weapon_master_blade", "weapon_goultard_cleaver", "weapon_xelor_staff", "weapon_guardian_halberd"},
			Tags:        []string{"armor", "sets", "merchant", "gear"},
			AccentColor: [4]float32{0.92, 0.76, 0.22, 1.0},
			SortOrder:   20,
		},
		{
			ID:          "sanctuary_guide_jobs",
			Category:    "custom",
			Title:       "Professions & Gathering",
			Subtitle:    "Five Sanctuary Trades",
			Description: "Level Bucheron, Mineur, Herborist, Fisherman, and Craftsman across all 25 rooms. Higher-tier trees, veins, herbs, shoals, and benches unlock as your trade level rises and feed into crafting recipes.",
			KindBadge:   "JOBS",
			Location:    "All Sanctuary Rooms",
			IconItemID:  "mat_wakfu_crystal",
			ItemIDs:     []string{"mat_ash_wood", "mat_iron_ore", "mat_wild_mint", "mat_river_perch", "mat_artisan_core"},
			Tags:        []string{"jobs", "gathering", "crafting", "recipes"},
			AccentColor: [4]float32{0.28, 0.88, 0.48, 1.0},
			SortOrder:   30,
		},
	}
}

// RegisterCodex registers the classic Sanctuary Codex configuration and guide entries.
func RegisterCodex() {
	sdk.ConfigureCodex(sdk.CodexConfigSpec{
		Title:         "Sanctuary Codex",
		BannerText:    "Classic 5x5 Sanctuary Bestiary, Resources, Gear & Realm Guides",
		ExtraTabLabel: "Guides",
	})
	for _, entry := range CodexEntries() {
		sdk.RegisterCodexEntry(entry)
	}
}
