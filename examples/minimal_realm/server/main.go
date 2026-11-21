package main

import (
	sdk "hero-armory-sdk"
)

// minimalRealmPlugin boots an empty-shell realm with nothing except a single
// flat center room at (0,0): no items, no NPCs, no monsters, no jobs, and no
// procedural props or player housing.
type minimalRealmPlugin struct {
	sdk.BaseServerPlugin
}

func (minimalRealmPlugin) Init() {
	sdk.Log("Minimal Realm server plugin initialized")

	sdk.RegisterRoom(sdk.RoomSpec{
		X:                0,
		Y:                0,
		Name:             "Center Arena",
		Biome:            "sanctuary",
		LightingPreset:   "daylight",
		HalfSize:         16.0,
		SafeZone:         true,
		HousingZone:      false,
		DisableAutoProps: true,
		FloorColor:       [3]float32{0.18, 0.21, 0.28},
		WallColor:        [3]float32{0.12, 0.14, 0.20},
		AccentColor:      [3]float32{0.25, 0.85, 0.96},
		SpawnPos:         [3]float32{0, 0, 0},
		TerrainOps: []sdk.RoomTerrainOpSpec{
			{
				Op:      "rect",
				Surface: "courtyard",
				MinX:    -16,
				MaxX:    16,
				MinZ:    -16,
				MaxZ:    16,
				Elev:    0,
			},
		},
	})

	sdk.SetSpawnRoom(0, 0, [3]float32{0, 0, 0})

	sdk.SetStarterConfig(sdk.StarterConfigSpec{
		SpawnRoomX:     0,
		SpawnRoomY:     0,
		SpawnPos:       [3]float32{0, 0, 0},
		StarterGold:    0,
		BaseHP:         100,
		BaseAP:         6,
		BaseMP:         6,
		DisableHousing: true,
	})
}

func init() {
	sdk.RegisterServer(minimalRealmPlugin{})
}

func main() {}
