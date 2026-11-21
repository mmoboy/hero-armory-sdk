package vanilla

import (
	"strings"

	sdk "github.com/mmoboy/hero-armory-sdk"
)

type setVisual struct {
	setID   string
	rarity  string
	element string
	prim    [4]float32
	acc     [4]float32
	glow    [4]float32
}

var setVisuals = map[string]setVisual{
	"kokiri": {
		setID: "kokiri", rarity: "Heroic", element: "Forest / Wind",
		prim: [4]float32{0.16, 0.68, 0.32, 1.0},
		acc:  [4]float32{0.92, 0.76, 0.22, 1.0},
		glow: [4]float32{0.35, 0.98, 0.55, 1.0},
	},
	"gobball": {
		setID: "gobball", rarity: "Royal", element: "Earth / Fire",
		prim: [4]float32{0.95, 0.90, 0.80, 1.0},
		acc:  [4]float32{0.92, 0.68, 0.18, 1.0},
		glow: [4]float32{1.00, 0.75, 0.25, 1.0},
	},
	"wakfu": {
		setID: "wakfu", rarity: "Mythic", element: "Wakfu / Stasis",
		prim: [4]float32{0.14, 0.24, 0.52, 1.0},
		acc:  [4]float32{0.92, 0.78, 0.28, 1.0},
		glow: [4]float32{0.22, 0.92, 1.00, 1.0},
	},
	"twilight": {
		setID: "twilight", rarity: "Relic", element: "Twilight / Arcane",
		prim: [4]float32{0.16, 0.18, 0.26, 1.0},
		acc:  [4]float32{1.00, 0.46, 0.16, 1.0},
		glow: [4]float32{0.24, 0.80, 1.00, 1.0},
	},
}

func titleCaseID(id string) string {
	parts := strings.Split(id, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

func makeEquipItem(id, slot, setKey string) sdk.ItemSpec {
	sv := setVisuals[setKey]
	return sdk.ItemSpec{
		ID:           id,
		Name:         titleCaseID(id),
		Kind:         "equipment",
		Slot:         slot,
		VisualPreset: id,
		MaxStack:     1,
		Rarity:       sv.rarity,
		Element:      sv.element,
		PrimaryCol:   sv.prim,
		AccentCol:    sv.acc,
		GlowCol:      sv.glow,
		SetID:        sv.setID,
	}
}

// EquipmentItems returns all 32 classic Sanctuary equipment items in canonical slot order.
func EquipmentItems() []sdk.ItemSpec {
	return []sdk.ItemSpec{
		// Hats
		makeEquipItem("hat_gobball_headgear", "hat", "gobball"),
		makeEquipItem("hat_kokiri_cap", "hat", "kokiri"),
		makeEquipItem("hat_twilight_crown", "hat", "twilight"),
		makeEquipItem("hat_wakfu_chronocowl", "hat", "wakfu"),
		// Amulets
		makeEquipItem("amulet_gobball_talisman", "amulet", "gobball"),
		makeEquipItem("amulet_kokiri_spiritual", "amulet", "kokiri"),
		makeEquipItem("amulet_twilight_tear", "amulet", "twilight"),
		makeEquipItem("amulet_wakfu_pendulum", "amulet", "wakfu"),
		// Chests
		makeEquipItem("chest_gobball_cuirass", "chest", "gobball"),
		makeEquipItem("chest_kokiri_tunic", "chest", "kokiri"),
		makeEquipItem("chest_twilight_hauberk", "chest", "twilight"),
		makeEquipItem("chest_wakfu_robe", "chest", "wakfu"),
		// Belts
		makeEquipItem("belt_gobball_cord", "belt", "gobball"),
		makeEquipItem("belt_kokiri_girdle", "belt", "kokiri"),
		makeEquipItem("belt_twilight_girdle", "belt", "twilight"),
		makeEquipItem("belt_wakfu_sash", "belt", "wakfu"),
		// Gloves
		makeEquipItem("gloves_gobball_mitts", "gloves", "gobball"),
		makeEquipItem("gloves_kokiri_gauntlets", "gloves", "kokiri"),
		makeEquipItem("gloves_twilight_grips", "gloves", "twilight"),
		makeEquipItem("gloves_wakfu_bracers", "gloves", "wakfu"),
		// Rings
		makeEquipItem("ring_gobball_chanoine", "ring", "gobball"),
		makeEquipItem("ring_kokiri_fairy", "ring", "kokiri"),
		makeEquipItem("ring_twilight_magnesis", "ring", "twilight"),
		makeEquipItem("ring_wakfu_chanoine", "ring", "wakfu"),
		// Boots
		makeEquipItem("boots_gobball_sabatons", "boots", "gobball"),
		makeEquipItem("boots_kokiri_pegasus", "boots", "kokiri"),
		makeEquipItem("boots_twilight_greaves", "boots", "twilight"),
		makeEquipItem("boots_wakfu_striders", "boots", "wakfu"),
		// Weapons
		makeEquipItem("weapon_goultard_cleaver", "weapon", "gobball"),
		makeEquipItem("weapon_guardian_halberd", "weapon", "twilight"),
		makeEquipItem("weapon_master_blade", "weapon", "kokiri"),
		makeEquipItem("weapon_xelor_staff", "weapon", "wakfu"),
	}
}

// StackableItems returns all 20 classic Sanctuary consumables and crafting materials.
func StackableItems() []sdk.ItemSpec {
	return []sdk.ItemSpec{
		{
			ID: "cons_wakfu_potion", Name: "Wakfu Stasis Elixir", Kind: "consumable", MaxStack: 20,
			Rarity: "Rare", Element: "Wakfu / Stasis",
			Description: "Condensed azure Wakfu flask distilled by Eliatrope alchemists. Restores spirit and cools channel focus.",
			PrimaryCol:  [4]float32{0.14, 0.78, 0.96, 1.0}, AccentCol: [4]float32{0.90, 0.76, 0.28, 1.0}, GlowCol: [4]float32{0.25, 0.95, 1.00, 1.0},
		},
		{
			ID: "cons_red_heart_elixir", Name: "Kokiri Vitality Vial", Kind: "consumable", MaxStack: 20,
			Rarity: "Uncommon", Element: "Forest / Vitality",
			Description: "Crimson herbal draft corked with Deku bark. Keeps vanguard knights standing through fierce skirmishes.",
			PrimaryCol:  [4]float32{0.92, 0.22, 0.28, 1.0}, AccentCol: [4]float32{0.94, 0.80, 0.42, 1.0}, GlowCol: [4]float32{1.00, 0.36, 0.44, 1.0},
		},
		{
			ID: "mat_wakfu_crystal", Name: "Refined Wakfu Prism", Kind: "material", MaxStack: 99,
			Rarity: "Mythic", Element: "Wakfu",
			Description: "Crystallized ambient energy harvested near Zaap pillars. Used to attune Xelor and Eliatrope artifacts.",
			PrimaryCol:  [4]float32{0.16, 0.64, 0.94, 1.0}, AccentCol: [4]float32{0.62, 0.96, 1.00, 1.0}, GlowCol: [4]float32{0.22, 0.92, 1.00, 1.0},
		},
		{
			ID: "mat_gobball_wool", Name: "Royal Gobball Fleece", Kind: "material", MaxStack: 99,
			Rarity: "Common", Element: "Earth",
			Description: "Dense sun-warmed Astrub pasture wool prized by tailors and armorsmiths alike.",
			PrimaryCol:  [4]float32{0.95, 0.90, 0.78, 1.0}, AccentCol: [4]float32{0.88, 0.66, 0.20, 1.0}, GlowCol: [4]float32{0.98, 0.78, 0.34, 1.0},
		},
		{
			ID: "mat_kokiri_emerald", Name: "Sacred Forest Rupee", Kind: "material", MaxStack: 99,
			Rarity: "Heroic", Element: "Forest / Wind",
			Description: "Faceted emerald trade crystal accepted by sanctuary smiths and wandering merchants.",
			PrimaryCol:  [4]float32{0.16, 0.82, 0.42, 1.0}, AccentCol: [4]float32{0.62, 1.00, 0.78, 1.0}, GlowCol: [4]float32{0.28, 0.98, 0.56, 1.0},
		},
		{
			ID: "mat_sheikah_core", Name: "Ancient Sheikah Core", Kind: "material", MaxStack: 99,
			Rarity: "Relic", Element: "Twilight / Arcane",
			Description: "Humming obsidian resonator powered by amber constellation circuitry.",
			PrimaryCol:  [4]float32{0.18, 0.20, 0.26, 1.0}, AccentCol: [4]float32{1.00, 0.52, 0.16, 1.0}, GlowCol: [4]float32{1.00, 0.58, 0.22, 1.0},
		},
		{
			ID: "mat_ash_wood", Name: "Ash Timber", Kind: "material", MaxStack: 99,
			Rarity: "Common", Element: "Forest",
			Description: "Seasoned hardwood harvested by a Bucheron from sanctuary ash trees.",
			PrimaryCol:  [4]float32{0.52, 0.34, 0.18, 1.0}, AccentCol: [4]float32{0.76, 0.56, 0.34, 1.0}, GlowCol: [4]float32{0.46, 0.84, 0.36, 1.0},
		},
		{
			ID: "mat_iron_ore", Name: "Iron Ore", Kind: "material", MaxStack: 99,
			Rarity: "Common", Element: "Earth",
			Description: "Heavy mineral ore extracted from rocky veins by a Mineur.",
			PrimaryCol:  [4]float32{0.44, 0.50, 0.58, 1.0}, AccentCol: [4]float32{0.76, 0.84, 0.94, 1.0}, GlowCol: [4]float32{0.52, 0.78, 0.98, 1.0},
		},
		{
			ID: "mat_oak_wood", Name: "Oak Heartwood", Kind: "material", MaxStack: 99,
			Rarity: "Uncommon", Element: "Forest",
			Description: "Dense amber-grained timber felled by an apprentice Bucheron.",
			PrimaryCol:  [4]float32{0.44, 0.24, 0.10, 1.0}, AccentCol: [4]float32{0.84, 0.56, 0.24, 1.0}, GlowCol: [4]float32{0.92, 0.68, 0.24, 1.0},
		},
		{
			ID: "mat_elder_sap", Name: "Elder Bark", Kind: "material", MaxStack: 99,
			Rarity: "Rare", Element: "Forest",
			Description: "Luminous ancient bark steeped in forest Wakfu.",
			PrimaryCol:  [4]float32{0.20, 0.42, 0.26, 1.0}, AccentCol: [4]float32{0.58, 0.96, 0.64, 1.0}, GlowCol: [4]float32{0.42, 1.00, 0.70, 1.0},
		},
		{
			ID: "mat_copper_ore", Name: "Copper Ore", Kind: "material", MaxStack: 99,
			Rarity: "Uncommon", Element: "Earth",
			Description: "Warm-toned conductive ore prized by veteran smiths.",
			PrimaryCol:  [4]float32{0.62, 0.30, 0.16, 1.0}, AccentCol: [4]float32{0.96, 0.60, 0.34, 1.0}, GlowCol: [4]float32{1.00, 0.58, 0.28, 1.0},
		},
		{
			ID: "mat_wild_mint", Name: "Wild Mint", Kind: "material", MaxStack: 99,
			Rarity: "Common", Element: "Forest",
			Description: "Fragrant sanctuary herb gathered by an Herborist.",
			PrimaryCol:  [4]float32{0.24, 0.64, 0.32, 1.0}, AccentCol: [4]float32{0.58, 0.96, 0.52, 1.0}, GlowCol: [4]float32{0.56, 0.98, 0.48, 1.0},
		},
		{
			ID: "mat_moon_petal", Name: "Moon Petal", Kind: "material", MaxStack: 99,
			Rarity: "Uncommon", Element: "Arcane",
			Description: "Luminescent blossom harvested under twilight by skilled Herborists.",
			PrimaryCol:  [4]float32{0.46, 0.32, 0.78, 1.0}, AccentCol: [4]float32{0.82, 0.68, 1.00, 1.0}, GlowCol: [4]float32{0.78, 0.62, 1.00, 1.0},
		},
		{
			ID: "mat_star_bloom", Name: "Starbloom", Kind: "material", MaxStack: 99,
			Rarity: "Rare", Element: "Forest",
			Description: "Radiant golden-leafed flower infused with pure Wakfu.",
			PrimaryCol:  [4]float32{0.72, 0.56, 0.14, 1.0}, AccentCol: [4]float32{1.00, 0.92, 0.46, 1.0}, GlowCol: [4]float32{1.00, 0.88, 0.38, 1.0},
		},
		{
			ID: "mat_river_perch", Name: "River Perch", Kind: "material", MaxStack: 99,
			Rarity: "Common", Element: "Water",
			Description: "Freshwater fish reeled in by a Fisherman.",
			PrimaryCol:  [4]float32{0.18, 0.52, 0.68, 1.0}, AccentCol: [4]float32{0.52, 0.88, 0.98, 1.0}, GlowCol: [4]float32{0.34, 0.86, 0.96, 1.0},
		},
		{
			ID: "mat_silver_trout", Name: "Silver Trout", Kind: "material", MaxStack: 99,
			Rarity: "Uncommon", Element: "Water",
			Description: "Gleaming silver trout caught from deep plaza currents.",
			PrimaryCol:  [4]float32{0.44, 0.58, 0.72, 1.0}, AccentCol: [4]float32{0.84, 0.94, 1.00, 1.0}, GlowCol: [4]float32{0.62, 0.90, 1.00, 1.0},
		},
		{
			ID: "mat_golden_koi", Name: "Golden Koi", Kind: "material", MaxStack: 99,
			Rarity: "Rare", Element: "Water",
			Description: "Prized golden koi sought by master Fishermen.",
			PrimaryCol:  [4]float32{0.82, 0.48, 0.12, 1.0}, AccentCol: [4]float32{1.00, 0.84, 0.32, 1.0}, GlowCol: [4]float32{1.00, 0.78, 0.28, 1.0},
		},
		{
			ID: "mat_refined_plank", Name: "Refined Plank", Kind: "material", MaxStack: 99,
			Rarity: "Common", Element: "Earth",
			Description: "Planed and braced timber shaped at an Artisan Workbench.",
			PrimaryCol:  [4]float32{0.58, 0.38, 0.18, 1.0}, AccentCol: [4]float32{0.88, 0.68, 0.38, 1.0}, GlowCol: [4]float32{0.96, 0.72, 0.34, 1.0},
		},
		{
			ID: "mat_forged_ingot", Name: "Forged Ingot", Kind: "material", MaxStack: 99,
			Rarity: "Uncommon", Element: "Fire",
			Description: "Tempered alloy bar hammered by an apprentice Craftsman.",
			PrimaryCol:  [4]float32{0.56, 0.28, 0.18, 1.0}, AccentCol: [4]float32{1.00, 0.62, 0.32, 1.0}, GlowCol: [4]float32{1.00, 0.56, 0.24, 1.0},
		},
		{
			ID: "mat_artisan_core", Name: "Artisan Core", Kind: "material", MaxStack: 99,
			Rarity: "Rare", Element: "Arcane",
			Description: "Precision-crafted relic core assembled by a master Craftsman.",
			PrimaryCol:  [4]float32{0.34, 0.28, 0.54, 1.0}, AccentCol: [4]float32{1.00, 0.82, 0.38, 1.0}, GlowCol: [4]float32{1.00, 0.76, 0.30, 1.0},
		},
	}
}

// PortalStyleItems returns all 4 classic Sanctuary house portal style items.
func PortalStyleItems() []sdk.ItemSpec {
	return []sdk.ItemSpec{
		{
			ID: "portal_style_crimson", Name: "Crimson Singularity", Kind: "portal_style", MaxStack: 1,
			Rarity: "Relic", Element: "Stasis / Crimson",
			Description: "Socket into your House Portal Style slot while inside your house to dye your black hole molten crimson.",
			PrimaryCol:  [4]float32{0.82, 0.14, 0.22, 1.0}, AccentCol: [4]float32{1.00, 0.48, 0.22, 1.0}, GlowCol: [4]float32{1.00, 0.26, 0.34, 1.0},
		},
		{
			ID: "portal_style_emerald", Name: "Emerald Rift", Kind: "portal_style", MaxStack: 1,
			Rarity: "Heroic", Element: "Forest / Wakfu",
			Description: "Socket into your House Portal Style slot while inside your house to infuse your black hole with jade vortex energy.",
			PrimaryCol:  [4]float32{0.10, 0.72, 0.44, 1.0}, AccentCol: [4]float32{0.46, 1.00, 0.76, 1.0}, GlowCol: [4]float32{0.22, 0.96, 0.64, 1.0},
		},
		{
			ID: "portal_style_solar", Name: "Solar Eclipse", Kind: "portal_style", MaxStack: 1,
			Rarity: "Mythic", Element: "Solar / Gold",
			Description: "Socket into your House Portal Style slot while inside your house to surround your black hole with a blazing solar corona.",
			PrimaryCol:  [4]float32{0.94, 0.62, 0.12, 1.0}, AccentCol: [4]float32{1.00, 0.92, 0.54, 1.0}, GlowCol: [4]float32{1.00, 0.78, 0.24, 1.0},
		},
		{
			ID: "portal_style_frost", Name: "Astral Frost", Kind: "portal_style", MaxStack: 1,
			Rarity: "Rare", Element: "Frost / Astral",
			Description: "Socket into your House Portal Style slot while inside your house to chill your black hole into a glacial sapphire rift.",
			PrimaryCol:  [4]float32{0.18, 0.58, 0.96, 1.0}, AccentCol: [4]float32{0.70, 0.96, 1.00, 1.0}, GlowCol: [4]float32{0.34, 0.88, 1.00, 1.0},
		},
	}
}

// RegisterItems registers all Sanctuary equipment, consumables, materials,
// portal styles, and the Sanctuary starter hero configuration.
func RegisterItems() {
	for _, it := range EquipmentItems() {
		sdk.RegisterItem(it)
	}
	for _, it := range StackableItems() {
		sdk.RegisterItem(it)
	}
	for _, it := range PortalStyleItems() {
		sdk.RegisterItem(it)
	}
	sdk.SetStarterConfig(SanctuaryStarterConfig())
}
