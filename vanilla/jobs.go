package vanilla

import (
	"fmt"

	sdk "github.com/mmoboy/hero-armory-sdk"
)

type gatherTier struct {
	ItemID    string
	NodeName  string
	MinLevel  int
	XP        int64
	GatherSec float32
}

// Jobs returns all 5 classic Sanctuary gathering/crafting professions.
func Jobs() []sdk.JobDefSpec {
	return []sdk.JobDefSpec{
		{
			ID: "bucheron", Name: "Bucheron", ToolLabel: "Axe", ResourceID: "mat_ash_wood",
			Description: "Chops sanctuary trees to harvest Ash, Oak, and Elder timber.",
		},
		{
			ID: "mineur", Name: "Mineur", ToolLabel: "Pickaxe", ResourceID: "mat_iron_ore",
			Description: "Extracts Iron, Copper, and Wakfu Crystals from rocky veins.",
		},
		{
			ID: "herborist", Name: "Herborist", ToolLabel: "Sickle", ResourceID: "mat_wild_mint",
			Description: "Gathers Wild Mint, Moon Petals, and Starblooms from meadows.",
		},
		{
			ID: "fisherman", Name: "Fisherman", ToolLabel: "Fishing Rod", ResourceID: "mat_river_perch",
			Description: "Catches River Perch, Silver Trout, and Golden Koi in shoals.",
		},
		{
			ID: "craftsman", Name: "Craftsman", ToolLabel: "Workbench", ResourceID: "mat_refined_plank",
			Description: "Shapes planks, ingots, and artisan relic cores across benches.",
		},
	}
}

// Recipes returns all 10 classic Sanctuary crafting recipes.
func Recipes() []sdk.RecipeSpec {
	return []sdk.RecipeSpec{
		{
			ID: "craft_bucheron_plank", Job: "bucheron", Name: "Refined Plank", MinLevel: 1, XP: 30,
			OutputItemID: "mat_refined_plank", OutputQty: 1,
			Ingredients: []sdk.RecipeIngredient{{ItemID: "mat_ash_wood", Quantity: 2}},
			Description: "Plane Ash Timber into sturdy structural planks.",
		},
		{
			ID: "craft_bucheron_bundle", Job: "bucheron", Name: "Heartwood Core", MinLevel: 2, XP: 50,
			OutputItemID: "mat_artisan_core", OutputQty: 1,
			Ingredients: []sdk.RecipeIngredient{{ItemID: "mat_oak_wood", Quantity: 2}, {ItemID: "mat_ash_wood", Quantity: 1}},
			Description: "Bind seasoned Oak and Ash timber into an Artisan Core.",
		},
		{
			ID: "craft_mineur_ingot", Job: "mineur", Name: "Forged Ingot", MinLevel: 1, XP: 30,
			OutputItemID: "mat_forged_ingot", OutputQty: 1,
			Ingredients: []sdk.RecipeIngredient{{ItemID: "mat_iron_ore", Quantity: 2}},
			Description: "Smelt raw Iron Ore into a tempered Forged Ingot.",
		},
		{
			ID: "craft_mineur_sheikah_core", Job: "mineur", Name: "Sheikah Core", MinLevel: 2, XP: 55,
			OutputItemID: "mat_sheikah_core", OutputQty: 1,
			Ingredients: []sdk.RecipeIngredient{{ItemID: "mat_copper_ore", Quantity: 2}, {ItemID: "mat_wakfu_crystal", Quantity: 1}},
			Description: "Fuse Copper Ore with Wakfu Crystal into a Sheikah Core.",
		},
		{
			ID: "craft_herb_wakfu_vial", Job: "herborist", Name: "Wakfu Vial", MinLevel: 1, XP: 30,
			OutputItemID: "cons_wakfu_potion", OutputQty: 1,
			Ingredients: []sdk.RecipeIngredient{{ItemID: "mat_wild_mint", Quantity: 2}},
			Description: "Distill fresh Wild Mint into a restorative Wakfu Vial.",
		},
		{
			ID: "craft_herb_heart_elixir", Job: "herborist", Name: "Heart Elixir", MinLevel: 2, XP: 55,
			OutputItemID: "cons_red_heart_elixir", OutputQty: 1,
			Ingredients: []sdk.RecipeIngredient{{ItemID: "mat_moon_petal", Quantity: 2}, {ItemID: "mat_wild_mint", Quantity: 1}},
			Description: "Brew Moon Petals and Wild Mint into a potent Heart Elixir.",
		},
		{
			ID: "craft_fish_broth", Job: "fisherman", Name: "Tonic Fish Broth", MinLevel: 1, XP: 30,
			OutputItemID: "cons_wakfu_potion", OutputQty: 1,
			Ingredients: []sdk.RecipeIngredient{{ItemID: "mat_river_perch", Quantity: 2}},
			Description: "Simmer River Perch into an invigorating Wakfu tonic.",
		},
		{
			ID: "craft_fish_koi_elixir", Job: "fisherman", Name: "Golden Koi Elixir", MinLevel: 2, XP: 55,
			OutputItemID: "cons_red_heart_elixir", OutputQty: 1,
			Ingredients: []sdk.RecipeIngredient{{ItemID: "mat_silver_trout", Quantity: 1}, {ItemID: "mat_golden_koi", Quantity: 1}},
			Description: "Refine rare fish oils into a restorative Heart Elixir.",
		},
		{
			ID: "craft_craftsman_core", Job: "craftsman", Name: "Artisan Core", MinLevel: 1, XP: 35,
			OutputItemID: "mat_artisan_core", OutputQty: 1,
			Ingredients: []sdk.RecipeIngredient{{ItemID: "mat_refined_plank", Quantity: 1}, {ItemID: "mat_forged_ingot", Quantity: 1}},
			Description: "Assemble Refined Planks and Forged Ingots into an Artisan Core.",
		},
		{
			ID: "craft_craftsman_elixir", Job: "craftsman", Name: "Masterwork Elixirs", MinLevel: 2, XP: 60,
			OutputItemID: "cons_red_heart_elixir", OutputQty: 2,
			Ingredients: []sdk.RecipeIngredient{{ItemID: "mat_artisan_core", Quantity: 1}, {ItemID: "mat_wakfu_crystal", Quantity: 1}},
			Description: "Channel an Artisan Core and Wakfu Crystal into two Heart Elixirs.",
		},
	}
}

func jobTiers(job string) []gatherTier {
	switch job {
	case "bucheron":
		return []gatherTier{
			{ItemID: "mat_ash_wood", NodeName: "Ash Tree", MinLevel: 1, XP: 25, GatherSec: 2.2},
			{ItemID: "mat_oak_wood", NodeName: "Oak Tree", MinLevel: 2, XP: 40, GatherSec: 3.0},
			{ItemID: "mat_elder_sap", NodeName: "Elder Tree", MinLevel: 3, XP: 60, GatherSec: 3.8},
		}
	case "mineur":
		return []gatherTier{
			{ItemID: "mat_iron_ore", NodeName: "Iron Vein", MinLevel: 1, XP: 25, GatherSec: 2.2},
			{ItemID: "mat_copper_ore", NodeName: "Copper Vein", MinLevel: 2, XP: 40, GatherSec: 3.0},
			{ItemID: "mat_wakfu_crystal", NodeName: "Wakfu Crystal", MinLevel: 3, XP: 60, GatherSec: 3.8},
		}
	case "herborist":
		return []gatherTier{
			{ItemID: "mat_wild_mint", NodeName: "Mint Patch", MinLevel: 1, XP: 25, GatherSec: 2.0},
			{ItemID: "mat_moon_petal", NodeName: "Moonflower", MinLevel: 2, XP: 40, GatherSec: 2.8},
			{ItemID: "mat_star_bloom", NodeName: "Starbloom", MinLevel: 3, XP: 60, GatherSec: 3.6},
		}
	case "fisherman":
		return []gatherTier{
			{ItemID: "mat_river_perch", NodeName: "River Shoal", MinLevel: 1, XP: 25, GatherSec: 2.4},
			{ItemID: "mat_silver_trout", NodeName: "Trout Riffle", MinLevel: 2, XP: 40, GatherSec: 3.2},
			{ItemID: "mat_golden_koi", NodeName: "Koi Pool", MinLevel: 3, XP: 60, GatherSec: 4.0},
		}
	case "craftsman":
		return []gatherTier{
			{ItemID: "mat_refined_plank", NodeName: "Wood Bench", MinLevel: 1, XP: 25, GatherSec: 2.2},
			{ItemID: "mat_forged_ingot", NodeName: "Anvil Forge", MinLevel: 2, XP: 40, GatherSec: 3.0},
			{ItemID: "mat_artisan_core", NodeName: "Runic Table", MinLevel: 3, XP: 60, GatherSec: 3.8},
		}
	default:
		return nil
	}
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// GatherNodesForRoom returns classic gatherable nodes for any Sanctuary room coordinate (x, y).
func GatherNodesForRoom(x, y int) []sdk.GatherNodeSpec {
	tag := fmt.Sprintf("gather_%d_%d", x, y)
	if x == 0 && y == 0 {
		return []sdk.GatherNodeSpec{
			{ID: tag + "_tree_0", Job: "bucheron", Name: "Ash Tree", RoomX: 0, RoomY: 0, PosX: -11.0, PosZ: -9.5, RewardItemID: "mat_ash_wood", MinLevel: 1, GatherSec: 2.2, XP: 25, RespawnSec: 22},
			{ID: tag + "_tree_1", Job: "bucheron", Name: "Ash Tree", RoomX: 0, RoomY: 0, PosX: 16.0, PosZ: -6.5, RewardItemID: "mat_ash_wood", MinLevel: 1, GatherSec: 2.2, XP: 25, RespawnSec: 22},
			{ID: tag + "_tree_2", Job: "bucheron", Name: "Oak Tree", RoomX: 0, RoomY: 0, PosX: 8.5, PosZ: -12.5, RewardItemID: "mat_oak_wood", MinLevel: 2, GatherSec: 3.0, XP: 40, RespawnSec: 22},
			{ID: tag + "_tree_3", Job: "bucheron", Name: "Elder Tree", RoomX: 0, RoomY: 0, PosX: -4.5, PosZ: -13.0, RewardItemID: "mat_elder_sap", MinLevel: 3, GatherSec: 3.8, XP: 60, RespawnSec: 22},
			{ID: tag + "_ore_0", Job: "mineur", Name: "Iron Vein", RoomX: 0, RoomY: 0, PosX: -13.0, PosZ: 9.0, RewardItemID: "mat_iron_ore", MinLevel: 1, GatherSec: 2.2, XP: 25, RespawnSec: 22},
			{ID: tag + "_ore_1", Job: "mineur", Name: "Copper Vein", RoomX: 0, RoomY: 0, PosX: 10.5, PosZ: 4.5, RewardItemID: "mat_copper_ore", MinLevel: 2, GatherSec: 3.0, XP: 40, RespawnSec: 22},
			{ID: tag + "_ore_2", Job: "mineur", Name: "Wakfu Crystal", RoomX: 0, RoomY: 0, PosX: 6.0, PosZ: 12.5, RewardItemID: "mat_wakfu_crystal", MinLevel: 3, GatherSec: 3.8, XP: 60, RespawnSec: 22},
			{ID: tag + "_herb_0", Job: "herborist", Name: "Mint Patch", RoomX: 0, RoomY: 0, PosX: -7.0, PosZ: 5.5, RewardItemID: "mat_wild_mint", MinLevel: 1, GatherSec: 2.0, XP: 25, RespawnSec: 22},
			{ID: tag + "_fish_0", Job: "fisherman", Name: "River Shoal", RoomX: 0, RoomY: 0, PosX: -15.5, PosZ: 6.0, RewardItemID: "mat_river_perch", MinLevel: 1, GatherSec: 2.4, XP: 25, RespawnSec: 22},
			{ID: tag + "_craft_0", Job: "craftsman", Name: "Wood Bench", RoomX: 0, RoomY: 0, PosX: 6.5, PosZ: -5.5, RewardItemID: "mat_refined_plank", MinLevel: 1, GatherSec: 2.2, XP: 25, RespawnSec: 22},
		}
	}

	bTiers := jobTiers("bucheron")
	mTiers := jobTiers("mineur")
	hTiers := jobTiers("herborist")
	fTiers := jobTiers("fisherman")
	cTiers := jobTiers("craftsman")
	bPick := bTiers[absInt(x+y)%len(bTiers)]
	mPick := mTiers[absInt(x*2+y)%len(mTiers)]
	hPick := hTiers[absInt(x+y*2+1)%len(hTiers)]
	fPick := fTiers[absInt(x*3+y)%len(fTiers)]
	cPick := cTiers[absInt(x+y*3+2)%len(cTiers)]

	return []sdk.GatherNodeSpec{
		{ID: tag + "_tree_0", Job: "bucheron", Name: "Ash Tree", RoomX: x, RoomY: y, PosX: -11.5, PosZ: 7.5, RewardItemID: "mat_ash_wood", MinLevel: 1, GatherSec: 2.2, XP: 25, RespawnSec: 22},
		{ID: tag + "_tree_1", Job: "bucheron", Name: bPick.NodeName, RoomX: x, RoomY: y, PosX: 11.0, PosZ: 9.5, RewardItemID: bPick.ItemID, MinLevel: bPick.MinLevel, GatherSec: bPick.GatherSec, XP: int64(bPick.XP), RespawnSec: 22},
		{ID: tag + "_ore_0", Job: "mineur", Name: "Iron Vein", RoomX: x, RoomY: y, PosX: 11.5, PosZ: -7.5, RewardItemID: "mat_iron_ore", MinLevel: 1, GatherSec: 2.2, XP: 25, RespawnSec: 22},
		{ID: tag + "_ore_1", Job: "mineur", Name: mPick.NodeName, RoomX: x, RoomY: y, PosX: -10.5, PosZ: -8.5, RewardItemID: mPick.ItemID, MinLevel: mPick.MinLevel, GatherSec: mPick.GatherSec, XP: int64(mPick.XP), RespawnSec: 22},
		{ID: tag + "_herb_0", Job: "herborist", Name: hPick.NodeName, RoomX: x, RoomY: y, PosX: -5.0, PosZ: 11.5, RewardItemID: hPick.ItemID, MinLevel: hPick.MinLevel, GatherSec: hPick.GatherSec, XP: int64(hPick.XP), RespawnSec: 22},
		{ID: tag + "_fish_0", Job: "fisherman", Name: fPick.NodeName, RoomX: x, RoomY: y, PosX: 5.5, PosZ: 11.5, RewardItemID: fPick.ItemID, MinLevel: fPick.MinLevel, GatherSec: fPick.GatherSec, XP: int64(fPick.XP), RespawnSec: 22},
		{ID: tag + "_craft_0", Job: "craftsman", Name: cPick.NodeName, RoomX: x, RoomY: y, PosX: 0.0, PosZ: -11.5, RewardItemID: cPick.ItemID, MinLevel: cPick.MinLevel, GatherSec: cPick.GatherSec, XP: int64(cPick.XP), RespawnSec: 22},
	}
}

// RegisterJobsAndGatherNodes registers all Sanctuary jobs, recipes, and 5x5 gatherable nodes.
func RegisterJobsAndGatherNodes() {
	for _, j := range Jobs() {
		sdk.RegisterJob(j)
	}
	for _, r := range Recipes() {
		sdk.RegisterRecipe(r)
	}
	for y := -2; y <= 2; y++ {
		for x := -2; x <= 2; x++ {
			for _, gn := range GatherNodesForRoom(x, y) {
				sdk.SpawnGatherNode(gn)
			}
		}
	}
}
