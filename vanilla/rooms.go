package vanilla

import (
	"fmt"

	sdk "github.com/mmoboy/hero-armory-sdk"
)

func gridPortals(x, y int) []sdk.RoomPortalSpec {
	var out []sdk.RoomPortalSpec
	if y < 2 {
		out = append(out, sdk.RoomPortalSpec{Edge: "north", TargetX: x, TargetY: y + 1})
	}
	if y > -2 {
		out = append(out, sdk.RoomPortalSpec{Edge: "south", TargetX: x, TargetY: y - 1})
	}
	if x < 2 {
		out = append(out, sdk.RoomPortalSpec{Edge: "east", TargetX: x + 1, TargetY: y})
	}
	if x > -2 {
		out = append(out, sdk.RoomPortalSpec{Edge: "west", TargetX: x - 1, TargetY: y})
	}
	return out
}

func rectOp(kind string, minX, minZ, maxX, maxZ, elev float32) sdk.RoomTerrainOpSpec {
	return sdk.RoomTerrainOpSpec{
		Op: "rect", Surface: kind, MinX: minX, MinZ: minZ, MaxX: maxX, MaxZ: maxZ, Elev: elev,
	}
}

func slopeOp(kind string, minX, minZ, maxX, maxZ float32, riseDir string, elev, elevEnd float32) sdk.RoomTerrainOpSpec {
	return sdk.RoomTerrainOpSpec{
		Op: "slope", Surface: kind, MinX: minX, MinZ: minZ, MaxX: maxX, MaxZ: maxZ,
		RiseDir: riseDir, Elev: elev, ElevEnd: elevEnd,
	}
}

var outerRoomNames = map[[2]int]string{
	{-2, 2}:  "Whispering Canopy",
	{-1, 2}:  "Mossbrook Clearing",
	{0, 2}:   "Sylvan Watchtower",
	{1, 2}:   "Frostglade Terrace",
	{2, 2}:   "Aether Spire Peak",
	{-2, 1}:  "Verdant Crossing",
	{2, 1}:   "Celestial Causeway",
	{-2, 0}:  "Redrock Canyon",
	{2, 0}:   "Nebula Foundry",
	{-2, -1}: "Dune Coliseum",
	{2, -1}:  "Starshard Vault",
	{-2, -2}: "Magma Caldera",
	{-1, -2}: "Ashen Bastion",
	{0, -2}:  "Amethyst Catacombs",
	{1, -2}:  "Obsidian Hollow",
	{2, -2}:  "Eclipse Sanctum",
}

func biomeHouseColors(biome string) (label string, wall [3]float32, roof [3]float32) {
	switch biome {
	case "emerald_glade":
		return "Emerald Glade", [3]float32{0.76, 0.90, 0.78}, [3]float32{0.28, 0.92, 0.46}
	case "astral_forge":
		return "Astral Forge", [3]float32{0.72, 0.80, 0.96}, [3]float32{0.24, 0.92, 1.00}
	case "sunset_arena":
		return "Sunset Dunes", [3]float32{0.94, 0.84, 0.74}, [3]float32{0.98, 0.64, 0.22}
	case "cindervault":
		return "Cindervault", [3]float32{0.90, 0.76, 0.72}, [3]float32{0.98, 0.42, 0.24}
	case "twilight":
		return "Twilight Crypt", [3]float32{0.78, 0.74, 0.92}, [3]float32{0.76, 0.46, 0.98}
	default:
		return "Sanctuary", [3]float32{0.82, 0.86, 0.94}, [3]float32{0.22, 0.86, 0.98}
	}
}

func buildOuterRoomSpec(x, y int) sdk.RoomSpec {
	name := outerRoomNames[[2]int{x, y}]
	if name == "" {
		name = fmt.Sprintf("Frontier Sector (%d,%d)", x, y)
	}
	biome := "twilight"
	switch {
	case y >= 1 && x <= 0:
		biome = "emerald_glade"
	case x >= 1 && y >= 0:
		biome = "astral_forge"
	case x <= -1 && y <= 0 && y >= -1:
		biome = "sunset_arena"
	case y == -2 && x <= -1:
		biome = "cindervault"
	}
	label, houseWall, houseRoof := biomeHouseColors(biome)

	elev := float32(0.80)
	if biome == "sunset_arena" {
		elev = -0.60
	} else if biome == "cindervault" {
		elev = 0.65
	}

	ops := []sdk.RoomTerrainOpSpec{
		rectOp("courtyard", -20.0, -20.0, 20.0, 20.0, elev),
	}
	if y < 2 {
		if elev >= 0 {
			ops = append(ops, slopeOp("ramp", -6.0, -30.0, 6.0, -20.0, "south", 0.0, elev))
		} else {
			ops = append(ops, slopeOp("ramp", -6.0, -30.0, 6.0, -20.0, "north", elev, 0.0))
		}
	}
	if y > -2 {
		if elev >= 0 {
			ops = append(ops, slopeOp("ramp", -6.0, 20.0, 6.0, 30.0, "north", 0.0, elev))
		} else {
			ops = append(ops, slopeOp("ramp", -6.0, 20.0, 6.0, 30.0, "south", elev, 0.0))
		}
	}
	if x > -2 {
		if elev >= 0 {
			ops = append(ops, slopeOp("bridge", -30.0, -6.0, -20.0, 6.0, "east", 0.0, elev))
		} else {
			ops = append(ops, slopeOp("bridge", -30.0, -6.0, -20.0, 6.0, "west", elev, 0.0))
		}
	}
	if x < 2 {
		if elev >= 0 {
			ops = append(ops, slopeOp("bridge", 20.0, -6.0, 30.0, 6.0, "west", 0.0, elev))
		} else {
			ops = append(ops, slopeOp("bridge", 20.0, -6.0, 30.0, 6.0, "east", elev, 0.0))
		}
	}
	if biome == "emerald_glade" {
		ops = append(ops, rectOp("water", 10.0, 10.0, 16.0, 16.0, elev))
	}

	var houses []sdk.RoomHouseSpec
	var ias []sdk.RoomInteractableSpec
	if (x+y)%2 == 0 {
		houses = []sdk.RoomHouseSpec{{
			ID: fmt.Sprintf("outpost_%d_%d", x, y), Name: name + " Lodge",
			X: -13.0, Z: -13.0, W: 6.0, D: 6.0, DoorDir: "south",
			Color: houseWall, RoofColor: houseRoof,
		}}
		ias = []sdk.RoomInteractableSpec{
			{ID: fmt.Sprintf("chest_%d_%d", x, y), Kind: "chest", Label: "Frontier Chest", Description: "Supplies left by earlier explorers of " + name + ".", RewardItem: "cons_wakfu_potion", X: -13.0, Z: -14.5, Radius: 0.85},
			{ID: fmt.Sprintf("shrine_%d_%d", x, y), Kind: "shrine", Label: name + " Shrine", Description: "A serene beacon marking the outer realm border.", X: 13.0, Z: -13.0, Radius: 1.05},
		}
	} else {
		kind := "shrine"
		if biome == "astral_forge" || biome == "cindervault" {
			kind = "forge"
		}
		ias = []sdk.RoomInteractableSpec{
			{ID: fmt.Sprintf("feat_%d_%d", x, y), Kind: kind, Label: name + " Beacon", Description: "An ancient landmark in " + name + ".", X: 12.0, Z: -12.0, Radius: 1.0},
			{ID: fmt.Sprintf("sign_%d_%d", x, y), Kind: "signpost", Label: "Waystone", Description: "You stand in " + name + " (" + label + ").", X: -4.0, Z: 9.0, Radius: 0.7},
		}
	}

	return sdk.RoomSpec{
		X: x, Y: y, Name: name, Biome: biome, HalfSize: 30.0,
		SpawnPos:   [3]float32{0, 0, 2.2},
		TerrainOps: ops,
		Obstacles: []sdk.RoomObstacleSpec{
			{X: -8.0, Z: -8.0, Radius: 0.95, Height: 3.6},
			{X: 8.0, Z: -8.0, Radius: 0.95, Height: 3.6},
			{X: -8.0, Z: 8.0, Radius: 0.95, Height: 3.6},
			{X: 8.0, Z: 8.0, Radius: 0.95, Height: 3.6},
		},
		Portals:       gridPortals(x, y),
		Houses:        houses,
		Interactables: ias,
	}
}

// Rooms returns all 25 classic 5x5 Sanctuary RoomSpecs.
func Rooms() []sdk.RoomSpec {
	inner := map[[2]int]sdk.RoomSpec{
		{0, 0}: {
			X: 0, Y: 0, Name: "Sanctuary Courtyard", Biome: "sanctuary", LightingPreset: "daylight", HalfSize: 30.0,
			FloorColor: [3]float32{0.20, 0.23, 0.30}, WallColor: [3]float32{0.14, 0.17, 0.24}, AccentColor: [3]float32{0.22, 0.84, 0.96},
			SpawnPos: [3]float32{0, 0, 2.2},
			Obstacles: []sdk.RoomObstacleSpec{
				{X: -6.5, Z: -6.5, Radius: 0.95, Height: 3.6},
				{X: 6.5, Z: -6.5, Radius: 0.95, Height: 3.6},
				{X: -6.5, Z: 6.5, Radius: 0.95, Height: 3.6},
				{X: 6.5, Z: 6.5, Radius: 0.95, Height: 3.6},
			},
			Houses: []sdk.RoomHouseSpec{
				{ID: "sanctuary_guildhall", Name: "Sanctuary Guildhouse", X: 13.0, Z: -13.0, W: 8.0, D: 6.0, DoorDir: "south", Color: [3]float32{0.82, 0.80, 0.74}, RoofColor: [3]float32{0.18, 0.36, 0.58}},
				{ID: "sanctuary_armory_hall", Name: "Wayfarer Inn", X: -13.0, Z: -13.0, W: 6.0, D: 6.0, DoorDir: "east", Color: [3]float32{0.80, 0.76, 0.68}, RoofColor: [3]float32{0.22, 0.42, 0.62}},
			},
			Interactables: []sdk.RoomInteractableSpec{
				{ID: "hub_chest", Kind: "chest", Label: "Wayfarer's Chest", Description: "A brass-bound chest stocked with traveller supplies.", RewardItem: "cons_wakfu_potion", X: 13.0, Z: -14.5, Radius: 0.85},
				{ID: "hub_shrine", Kind: "shrine", Label: "Sanctuary Fountain", Description: "Cool astral water shimmers with restorative energy.", X: 4.5, Z: -11.5, Radius: 1.1},
				{ID: "hub_sign", Kind: "signpost", Label: "Courtyard Signpost", Description: "North: Emerald Glade | East: Astral Forge | West: Sunset Arena | South: Twilight Crypt", X: -4.0, Z: 9.0, Radius: 0.7},
			},
			TerrainOps: []sdk.RoomTerrainOpSpec{
				rectOp("courtyard", -20.0, -20.0, 20.0, 20.0, 0.90),
				slopeOp("ramp", -6.0, -30.0, 6.0, -20.0, "south", 0.0, 0.90),
				slopeOp("ramp", -6.0, 20.0, 6.0, 30.0, "north", 0.0, 0.90),
				slopeOp("ramp", -30.0, -6.0, -20.0, 6.0, "east", 0.0, 0.90),
				slopeOp("ramp", 20.0, -6.0, 30.0, 6.0, "west", 0.0, 0.90),
				rectOp("water", -16.0, 8.0, -10.0, 14.0, 0.90),
				rectOp("water", 10.0, 8.0, 16.0, 14.0, 0.90),
			},
			Portals: gridPortals(0, 0),
		},
		{0, 1}: {
			X: 0, Y: 1, Name: "Kokiri Emerald Glade", Biome: "emerald_glade", LightingPreset: "daylight", HalfSize: 30.0,
			FloorColor: [3]float32{0.14, 0.24, 0.18}, WallColor: [3]float32{0.11, 0.18, 0.14}, AccentColor: [3]float32{0.24, 0.88, 0.44},
			SpawnPos: [3]float32{0, 0, 2.2},
			Obstacles: []sdk.RoomObstacleSpec{
				{X: -7.0, Z: -7.0, Radius: 1.00, Height: 3.8},
				{X: 7.0, Z: -7.0, Radius: 1.00, Height: 3.8},
				{X: -8.0, Z: 5.0, Radius: 0.90, Height: 3.0},
				{X: 8.0, Z: 5.0, Radius: 0.90, Height: 3.0},
			},
			Houses: []sdk.RoomHouseSpec{
				{ID: "glade_cottage", Name: "Herbalist Cottage", X: -15.0, Z: -11.0, W: 6.0, D: 6.0, DoorDir: "east", Color: [3]float32{0.84, 0.78, 0.66}, RoofColor: [3]float32{0.16, 0.48, 0.24}},
			},
			Interactables: []sdk.RoomInteractableSpec{
				{ID: "glade_chest", Kind: "chest", Label: "Ranger's Cache", Description: "Hidden beside the mossy cottage hearth.", RewardItem: "cons_health_elixir", X: -16.2, Z: -11.0, Radius: 0.85},
				{ID: "glade_shrine", Kind: "shrine", Label: "Spirit Tree Shrine", Description: "Ancient emerald leaves whisper of vitality.", X: 14.0, Z: -6.0, Radius: 1.1},
			},
			TerrainOps: []sdk.RoomTerrainOpSpec{
				rectOp("dais", -22.0, -20.0, 22.0, 12.0, 1.0),
				slopeOp("bridge", -6.0, 12.0, 6.0, 30.0, "north", 0.0, 1.0),
				slopeOp("ramp", -6.0, -30.0, 6.0, -20.0, "south", 0.0, 1.0),
				slopeOp("ramp", -30.0, -6.0, -22.0, 6.0, "east", 0.0, 1.0),
				slopeOp("ramp", 22.0, -6.0, 30.0, 6.0, "west", 0.0, 1.0),
				rectOp("water", 10.0, -16.0, 18.0, -10.0, 1.0),
			},
			Portals: gridPortals(0, 1),
		},
		{1, 0}: {
			X: 1, Y: 0, Name: "Astral Forge Hall", Biome: "astral_forge", LightingPreset: "astral", HalfSize: 30.0,
			FloorColor: [3]float32{0.12, 0.16, 0.26}, WallColor: [3]float32{0.09, 0.12, 0.20}, AccentColor: [3]float32{0.20, 0.90, 1.00},
			SpawnPos: [3]float32{0, 0, 3.0},
			Obstacles: []sdk.RoomObstacleSpec{
				{X: -8.0, Z: -8.0, Radius: 0.95, Height: 4.2},
				{X: 8.0, Z: -8.0, Radius: 0.95, Height: 4.2},
				{X: -8.0, Z: 8.0, Radius: 0.95, Height: 4.2},
				{X: 0.0, Z: 0.0, Radius: 1.10, Height: 2.6},
			},
			Houses: []sdk.RoomHouseSpec{
				{ID: "astral_smithy", Name: "Master Smithy", X: 13.0, Z: 13.0, W: 8.0, D: 6.0, DoorDir: "west", Color: [3]float32{0.44, 0.50, 0.64}, RoofColor: [3]float32{0.14, 0.46, 0.66}},
			},
			Interactables: []sdk.RoomInteractableSpec{
				{ID: "forge_anvil", Kind: "forge", Label: "Astral Forge Anvil", Description: "Blue starlight embers crackle over the runic anvil.", X: 14.0, Z: 13.0, Radius: 0.95},
				{ID: "forge_chest", Kind: "chest", Label: "Ingot Strongbox", Description: "Packed with tempered astral metal.", RewardItem: "cons_wakfu_potion", X: -13.0, Z: -14.0, Radius: 0.85},
			},
			TerrainOps: []sdk.RoomTerrainOpSpec{
				rectOp("courtyard", -18.0, -20.0, 20.0, 20.0, 0.85),
				slopeOp("bridge", -30.0, -6.0, -18.0, 6.0, "east", 0.0, 0.85),
				slopeOp("bridge", 20.0, -6.0, 30.0, 6.0, "west", 0.0, 0.85),
				slopeOp("ramp", -6.0, -30.0, 6.0, -20.0, "south", 0.0, 0.85),
				slopeOp("ramp", -6.0, 20.0, 6.0, 30.0, "north", 0.0, 0.85),
			},
			Portals: gridPortals(1, 0),
		},
		{-1, 0}: {
			X: -1, Y: 0, Name: "Gobball Warchief Arena", Biome: "sunset_arena", LightingPreset: "sunset", HalfSize: 30.0,
			FloorColor: [3]float32{0.28, 0.20, 0.16}, WallColor: [3]float32{0.20, 0.14, 0.12}, AccentColor: [3]float32{0.96, 0.62, 0.20},
			SpawnPos: [3]float32{0, 0, 2.2},
			Obstacles: []sdk.RoomObstacleSpec{
				{X: -8.0, Z: -8.0, Radius: 1.00, Height: 3.4},
				{X: 8.0, Z: -8.0, Radius: 1.00, Height: 3.4},
				{X: -8.0, Z: 8.0, Radius: 1.00, Height: 3.4},
				{X: 8.0, Z: 8.0, Radius: 1.00, Height: 3.4},
			},
			Houses: []sdk.RoomHouseSpec{
				{ID: "arena_barracks", Name: "Champion's Lodge", X: -15.0, Z: -13.0, W: 6.0, D: 6.0, DoorDir: "south", Color: [3]float32{0.82, 0.66, 0.50}, RoofColor: [3]float32{0.72, 0.26, 0.16}},
			},
			Interactables: []sdk.RoomInteractableSpec{
				{ID: "arena_chest", Kind: "chest", Label: "Warchief Spoils", Description: "A scarred arena chest brimming with tonics.", RewardItem: "cons_health_elixir", X: -15.0, Z: -14.5, Radius: 0.85},
				{ID: "arena_sign", Kind: "signpost", Label: "Arena Notice", Description: "Enter the sunken pit to test your mettle!", X: 10.0, Z: -5.0, Radius: 0.7},
			},
			TerrainOps: []sdk.RoomTerrainOpSpec{
				rectOp("courtyard", -20.0, -20.0, 18.0, 20.0, -0.75),
				slopeOp("ramp", -6.0, -30.0, 6.0, -20.0, "north", -0.75, 0.0),
				slopeOp("ramp", -6.0, 20.0, 6.0, 30.0, "south", -0.75, 0.0),
				slopeOp("bridge", 18.0, -6.0, 30.0, 6.0, "east", -0.75, 0.0),
				slopeOp("bridge", -30.0, -6.0, -20.0, 6.0, "west", -0.75, 0.0),
			},
			Portals: gridPortals(-1, 0),
		},
		{0, -1}: {
			X: 0, Y: -1, Name: "Twilight Crypt", Biome: "twilight", LightingPreset: "astral", HalfSize: 30.0,
			FloorColor: [3]float32{0.16, 0.13, 0.23}, WallColor: [3]float32{0.11, 0.09, 0.17}, AccentColor: [3]float32{0.72, 0.44, 0.96},
			SpawnPos: [3]float32{0, 0, 2.2},
			Obstacles: []sdk.RoomObstacleSpec{
				{X: -7.5, Z: -7.5, Radius: 0.92, Height: 3.8},
				{X: 7.5, Z: -7.5, Radius: 0.92, Height: 3.8},
				{X: -7.5, Z: 7.5, Radius: 0.92, Height: 3.8},
				{X: 7.5, Z: 7.5, Radius: 0.92, Height: 3.8},
			},
			Houses: []sdk.RoomHouseSpec{
				{ID: "crypt_chapel", Name: "Crypt Keeper's Chapel", X: 0.0, Z: 15.0, W: 8.0, D: 6.0, DoorDir: "north", Color: [3]float32{0.44, 0.38, 0.56}, RoofColor: [3]float32{0.42, 0.20, 0.62}},
			},
			Interactables: []sdk.RoomInteractableSpec{
				{ID: "crypt_shrine", Kind: "shrine", Label: "Twilight Obelisk", Description: "Amethyst runes hum softly in the dark.", X: 0.0, Z: 15.5, Radius: 1.0},
			},
			TerrainOps: []sdk.RoomTerrainOpSpec{
				rectOp("courtyard", -18.0, -20.0, 18.0, 20.0, 0.75),
				slopeOp("bridge", -6.0, -30.0, 6.0, -20.0, "south", 0.0, 0.75),
				slopeOp("bridge", -6.0, 20.0, 6.0, 30.0, "north", 0.0, 0.75),
				slopeOp("alcove", -30.0, -6.0, -18.0, 6.0, "east", 0.0, 0.75),
				slopeOp("alcove", 18.0, -6.0, 30.0, 6.0, "west", 0.0, 0.75),
			},
			Portals: gridPortals(0, -1),
		},
		{-1, 1}: {
			X: -1, Y: 1, Name: "Elder Grove Bastion", Biome: "emerald_glade", LightingPreset: "daylight", HalfSize: 30.0,
			FloorColor: [3]float32{0.18, 0.23, 0.17}, WallColor: [3]float32{0.13, 0.17, 0.13}, AccentColor: [3]float32{0.58, 0.90, 0.34},
			SpawnPos: [3]float32{0, 0, 2.2},
			Obstacles: []sdk.RoomObstacleSpec{
				{X: -6.8, Z: -6.8, Radius: 0.95, Height: 3.5},
				{X: 8.6, Z: 8.6, Radius: 0.95, Height: 3.5},
			},
			Houses: []sdk.RoomHouseSpec{
				{ID: "elder_lodge", Name: "Elder Ranger Lodge", X: -15.0, Z: -13.0, W: 6.0, D: 6.0, DoorDir: "south", Color: [3]float32{0.78, 0.68, 0.54}, RoofColor: [3]float32{0.22, 0.50, 0.20}},
			},
			Interactables: []sdk.RoomInteractableSpec{
				{ID: "elder_chest", Kind: "chest", Label: "Elder Chest", Description: "Carved from living oak.", RewardItem: "cons_wakfu_potion", X: -15.0, Z: -14.5, Radius: 0.85},
			},
			TerrainOps: []sdk.RoomTerrainOpSpec{
				rectOp("courtyard", -20.0, -20.0, 18.0, 18.0, 0.70),
				slopeOp("bridge", -6.0, 18.0, 6.0, 30.0, "north", 0.0, 0.70),
				slopeOp("bridge", -6.0, -30.0, 6.0, -20.0, "south", 0.0, 0.70),
				slopeOp("wing", 18.0, -6.0, 30.0, 6.0, "west", 0.0, 0.70),
				slopeOp("wing", -30.0, -6.0, -20.0, 6.0, "east", 0.0, 0.70),
			},
			Portals: gridPortals(-1, 1),
		},
		{1, 1}: {
			X: 1, Y: 1, Name: "Starlight Observatory", Biome: "astral_forge", LightingPreset: "astral", HalfSize: 30.0,
			FloorColor: [3]float32{0.13, 0.19, 0.27}, WallColor: [3]float32{0.09, 0.14, 0.20}, AccentColor: [3]float32{0.42, 0.88, 1.00},
			SpawnPos: [3]float32{0, 0, 2.2},
			Obstacles: []sdk.RoomObstacleSpec{
				{X: -8.6, Z: 8.6, Radius: 0.95, Height: 3.8},
				{X: 8.6, Z: -8.6, Radius: 0.95, Height: 3.8},
			},
			Interactables: []sdk.RoomInteractableSpec{
				{ID: "obs_shrine", Kind: "shrine", Label: "Starlight Orrery", Description: "Celestial rings align with the northern sky.", X: 12.0, Z: -12.0, Radius: 1.1},
			},
			TerrainOps: []sdk.RoomTerrainOpSpec{
				rectOp("courtyard", -20.0, -20.0, 20.0, 20.0, 0.85),
				slopeOp("bridge", -6.0, 20.0, 6.0, 30.0, "north", 0.0, 0.85),
				slopeOp("bridge", -6.0, -30.0, 6.0, -20.0, "south", 0.0, 0.85),
				slopeOp("bridge", -30.0, -6.0, -20.0, 6.0, "east", 0.0, 0.85),
				slopeOp("bridge", 20.0, -6.0, 30.0, 6.0, "west", 0.0, 0.85),
			},
			Portals: gridPortals(1, 1),
		},
		{-1, -1}: {
			X: -1, Y: -1, Name: "Cindervault Cloister", Biome: "cindervault", LightingPreset: "sunset", HalfSize: 30.0,
			FloorColor: [3]float32{0.25, 0.17, 0.17}, WallColor: [3]float32{0.18, 0.12, 0.13}, AccentColor: [3]float32{0.94, 0.42, 0.26},
			SpawnPos: [3]float32{0, 0, 2.2},
			Obstacles: []sdk.RoomObstacleSpec{
				{X: -8.6, Z: 8.6, Radius: 0.95, Height: 3.4},
				{X: 8.6, Z: -8.6, Radius: 0.95, Height: 3.4},
			},
			Interactables: []sdk.RoomInteractableSpec{
				{ID: "cinder_forge", Kind: "forge", Label: "Ember Brazier Forge", Description: "Volcanic heat radiates from the basalt hearth.", X: -11.0, Z: 0.0, Radius: 1.0},
			},
			TerrainOps: []sdk.RoomTerrainOpSpec{
				rectOp("courtyard", -20.0, -18.0, 20.0, 20.0, 0.65),
				slopeOp("bridge", -6.0, -30.0, 6.0, -18.0, "south", 0.0, 0.65),
				slopeOp("bridge", -6.0, 20.0, 6.0, 30.0, "north", 0.0, 0.65),
				slopeOp("bridge", 20.0, -6.0, 30.0, 6.0, "west", 0.0, 0.65),
				slopeOp("bridge", -30.0, -6.0, -20.0, 6.0, "east", 0.0, 0.65),
			},
			Portals: gridPortals(-1, -1),
		},
		{1, -1}: {
			X: 1, Y: -1, Name: "Chrono Sanctum", Biome: "twilight", LightingPreset: "astral", HalfSize: 30.0,
			FloorColor: [3]float32{0.14, 0.15, 0.25}, WallColor: [3]float32{0.10, 0.11, 0.19}, AccentColor: [3]float32{0.48, 0.68, 1.00},
			SpawnPos: [3]float32{0, 0, 2.2},
			Obstacles: []sdk.RoomObstacleSpec{
				{X: -8.6, Z: -8.6, Radius: 0.95, Height: 3.8},
				{X: 8.6, Z: 8.6, Radius: 0.95, Height: 3.8},
			},
			Interactables: []sdk.RoomInteractableSpec{
				{ID: "chrono_chest", Kind: "chest", Label: "Chrono Reliquary", Description: "Time-locked bronze chest.", RewardItem: "cons_health_elixir", X: 12.0, Z: 0.0, Radius: 0.85},
			},
			TerrainOps: []sdk.RoomTerrainOpSpec{
				rectOp("courtyard", -20.0, -20.0, 20.0, 20.0, 0.75),
				slopeOp("bridge", -6.0, -30.0, 6.0, -20.0, "south", 0.0, 0.75),
				slopeOp("bridge", -6.0, 20.0, 6.0, 30.0, "north", 0.0, 0.75),
				slopeOp("bridge", -30.0, -6.0, -20.0, 6.0, "east", 0.0, 0.75),
				slopeOp("bridge", 20.0, -6.0, 30.0, 6.0, "west", 0.0, 0.75),
			},
			Portals: gridPortals(1, -1),
		},
	}

	out := make([]sdk.RoomSpec, 0, 25)
	for y := -2; y <= 2; y++ {
		for x := -2; x <= 2; x++ {
			var sp sdk.RoomSpec
			if s, ok := inner[[2]int{x, y}]; ok {
				sp = s
			} else {
				sp = buildOuterRoomSpec(x, y)
			}
			enrichSanctuaryRoom(&sp)
			out = append(out, sp)
		}
	}
	return out
}

func enrichSanctuaryRoom(r *sdk.RoomSpec) {
	switch r.Biome {
	case "sanctuary":
		r.MapColor = [3]float32{0.24, 0.74, 0.92}
		r.MapIcon = "star"
		r.Atmosphere = &sdk.RoomAtmosphereSpec{
			SkyZenith:    [3]float32{0.16, 0.38, 0.68},
			SkyHorizon:   [3]float32{0.58, 0.78, 0.92},
			FogColor:     [3]float32{0.52, 0.70, 0.84},
			FogDensity:   0.008,
			SunColor:     [3]float32{1.00, 0.95, 0.84},
			SunIntensity: 1.15,
			SunPitch:     48,
			SunYaw:       35,
			AmbientColor: [3]float32{0.44, 0.56, 0.68},
			Weather:      "none",
		}
		r.Lights = []sdk.RoomLightSpec{
			{ID: "courtyard_beacon", X: 0, Y: 4.5, Z: 0, Color: [3]float32{0.35, 0.88, 1.00}, Intensity: 1.4, Range: 16, PulseSpeed: 0.8},
		}
		r.Emitters = []sdk.RoomEmitterSpec{
			{ID: "courtyard_motes", Kind: "sparkles", X: 0, Y: 2.2, Z: 0, Spread: [3]float32{14, 2.5, 14}, Color: [4]float32{0.55, 0.94, 1.00, 0.85}, Rate: 8, Size: 0.16, Speed: 0.45, Lifetime: 3.5},
		}
	case "emerald_glade":
		r.MapColor = [3]float32{0.22, 0.76, 0.38}
		r.MapIcon = "leaf"
		r.Atmosphere = &sdk.RoomAtmosphereSpec{
			SkyZenith:        [3]float32{0.14, 0.42, 0.46},
			SkyHorizon:       [3]float32{0.48, 0.80, 0.62},
			FogColor:         [3]float32{0.34, 0.62, 0.46},
			FogDensity:       0.012,
			SunColor:         [3]float32{0.96, 1.00, 0.82},
			SunIntensity:     1.10,
			SunPitch:         44,
			SunYaw:           20,
			AmbientColor:     [3]float32{0.32, 0.54, 0.38},
			Weather:          "spores",
			WeatherColor:     [4]float32{0.52, 1.00, 0.58, 0.85},
			WeatherIntensity: 0.45,
		}
		r.Lights = []sdk.RoomLightSpec{
			{ID: "glade_canopy", X: 0, Y: 3.6, Z: 0, Color: [3]float32{0.36, 0.96, 0.52}, Intensity: 1.2, Range: 14, PulseSpeed: 0.6},
		}
		r.Emitters = []sdk.RoomEmitterSpec{
			{ID: "glade_fireflies", Kind: "fireflies", X: 0, Y: 1.8, Z: 0, Spread: [3]float32{16, 2.0, 16}, Color: [4]float32{0.68, 1.00, 0.44, 0.9}, Rate: 10, Size: 0.18, Speed: 0.55, Lifetime: 4.0},
		}
	case "astral_forge":
		r.MapColor = [3]float32{0.22, 0.62, 0.96}
		r.MapIcon = "compass"
		r.Atmosphere = &sdk.RoomAtmosphereSpec{
			SkyZenith:        [3]float32{0.05, 0.09, 0.24},
			SkyHorizon:       [3]float32{0.14, 0.32, 0.56},
			FogColor:         [3]float32{0.12, 0.24, 0.44},
			FogDensity:       0.014,
			SunColor:         [3]float32{0.62, 0.88, 1.00},
			SunIntensity:     0.95,
			SunPitch:         52,
			SunYaw:           -40,
			AmbientColor:     [3]float32{0.22, 0.34, 0.54},
			Weather:          "starfall",
			WeatherColor:     [4]float32{0.40, 0.90, 1.00, 0.9},
			WeatherIntensity: 0.40,
		}
		r.Lights = []sdk.RoomLightSpec{
			{ID: "astral_core", X: 0, Y: 4.0, Z: 0, Color: [3]float32{0.28, 0.86, 1.00}, Intensity: 1.6, Range: 15, PulseSpeed: 1.1},
		}
		r.Emitters = []sdk.RoomEmitterSpec{
			{ID: "astral_sparks", Kind: "sparkles", X: 0, Y: 2.0, Z: 0, Spread: [3]float32{14, 2.5, 14}, Color: [4]float32{0.38, 0.90, 1.00, 0.9}, Rate: 11, Size: 0.16, Speed: 0.65, Lifetime: 3.2},
		}
		if r.X == 1 && r.Y == 1 {
			r.Liquids = []sdk.RoomLiquidSpec{
				{ID: "observatory_aether", Kind: "aether", MinX: -28, MaxX: 28, MinZ: -28, MaxZ: 28, Y: 0.08, Color: [4]float32{0.18, 0.66, 0.96, 0.72}, Opacity: 0.72, WaveSpeed: 1.1, EmissiveIntensity: 0.55},
			}
		}
	case "sunset_arena":
		r.MapColor = [3]float32{0.92, 0.56, 0.22}
		r.MapIcon = "sword"
		r.Atmosphere = &sdk.RoomAtmosphereSpec{
			SkyZenith:        [3]float32{0.34, 0.16, 0.24},
			SkyHorizon:       [3]float32{0.94, 0.52, 0.24},
			FogColor:         [3]float32{0.68, 0.38, 0.22},
			FogDensity:       0.011,
			SunColor:         [3]float32{1.00, 0.74, 0.42},
			SunIntensity:     1.20,
			SunPitch:         22,
			SunYaw:           -65,
			AmbientColor:     [3]float32{0.52, 0.34, 0.24},
			Weather:          "ash",
			WeatherColor:     [4]float32{0.98, 0.72, 0.38, 0.75},
			WeatherIntensity: 0.30,
		}
		r.Lights = []sdk.RoomLightSpec{
			{ID: "arena_brazier", X: 0, Y: 3.2, Z: 0, Color: [3]float32{1.00, 0.62, 0.22}, Intensity: 1.5, Range: 15, PulseSpeed: 1.8},
		}
	case "cindervault":
		r.MapColor = [3]float32{0.88, 0.30, 0.20}
		r.MapIcon = "shield"
		r.Atmosphere = &sdk.RoomAtmosphereSpec{
			SkyZenith:        [3]float32{0.18, 0.05, 0.06},
			SkyHorizon:       [3]float32{0.62, 0.20, 0.10},
			FogColor:         [3]float32{0.38, 0.14, 0.10},
			FogDensity:       0.018,
			SunColor:         [3]float32{1.00, 0.48, 0.24},
			SunIntensity:     1.05,
			SunPitch:         26,
			SunYaw:           -110,
			AmbientColor:     [3]float32{0.42, 0.18, 0.14},
			Weather:          "embers",
			WeatherColor:     [4]float32{1.00, 0.44, 0.14, 0.95},
			WeatherIntensity: 0.65,
		}
		r.Lights = []sdk.RoomLightSpec{
			{ID: "cinder_hearth", X: 0, Y: 3.0, Z: 0, Color: [3]float32{1.00, 0.40, 0.14}, Intensity: 1.8, Range: 16, PulseSpeed: 2.2},
		}
		r.Emitters = []sdk.RoomEmitterSpec{
			{ID: "cinder_embers", Kind: "embers", X: 0, Y: 0.8, Z: 0, Spread: [3]float32{16, 1.5, 16}, Color: [4]float32{1.00, 0.46, 0.16, 0.95}, Rate: 14, Size: 0.18, Speed: 1.1, Lifetime: 3.0},
		}
		if r.X == -1 && r.Y == -1 || (r.X == -2 && r.Y == -2) {
			r.Liquids = []sdk.RoomLiquidSpec{
				{ID: "caldera_lava", Kind: "lava", MinX: -28, MaxX: 28, MinZ: -28, MaxZ: 28, Y: 0.05, Color: [4]float32{0.96, 0.30, 0.08, 0.92}, Opacity: 0.92, WaveSpeed: 0.8, EmissiveIntensity: 1.4},
			}
		}
	default: // "twilight"
		r.MapColor = [3]float32{0.58, 0.36, 0.88}
		r.MapIcon = "book"
		r.Atmosphere = &sdk.RoomAtmosphereSpec{
			SkyZenith:        [3]float32{0.07, 0.04, 0.16},
			SkyHorizon:       [3]float32{0.26, 0.14, 0.44},
			FogColor:         [3]float32{0.18, 0.11, 0.32},
			FogDensity:       0.017,
			SunColor:         [3]float32{0.76, 0.56, 1.00},
			SunIntensity:     0.85,
			SunPitch:         38,
			SunYaw:           130,
			AmbientColor:     [3]float32{0.26, 0.18, 0.40},
			Weather:          "void",
			WeatherColor:     [4]float32{0.72, 0.44, 1.00, 0.85},
			WeatherIntensity: 0.35,
		}
		r.Lights = []sdk.RoomLightSpec{
			{ID: "twilight_lantern", X: 0, Y: 3.6, Z: 0, Color: [3]float32{0.72, 0.42, 1.00}, Intensity: 1.5, Range: 15, PulseSpeed: 0.9},
		}
		r.Emitters = []sdk.RoomEmitterSpec{
			{ID: "twilight_mist", Kind: "mist", X: 0, Y: 1.2, Z: 0, Spread: [3]float32{15, 1.2, 15}, Color: [4]float32{0.56, 0.36, 0.88, 0.65}, Rate: 9, Size: 0.55, Speed: 0.35, Lifetime: 4.5},
		}
	}
}

// RegisterRooms registers all 25 classic Sanctuary rooms and sets (0,0) as spawn room.
func RegisterRooms() {
	for _, r := range Rooms() {
		sdk.RegisterRoom(r)
	}
	sdk.SetSpawnRoom(0, 0, [3]float32{0, 0, 2.2})
}
