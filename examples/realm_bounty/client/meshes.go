package main

import (
	"fmt"
	"math"

	sdk "hero-armory-sdk"
)

type meshBuilder struct {
	positions [][3]float32
	colors    [][4]float32
	indices   []uint32
}

func (b *meshBuilder) addTri(p0, p1, p2 [3]float32, c0, c1, c2 [4]float32) {
	base := uint32(len(b.positions))
	b.positions = append(b.positions, p0, p1, p2)
	b.colors = append(b.colors, c0, c1, c2)
	b.indices = append(b.indices, base, base+1, base+2)
}

func (b *meshBuilder) addDiamondSpire(baseCenter [3]float32, tiltX, tiltZ, radius, height float32, sides int, colBase, colMid, colTip [4]float32) {
	if sides < 4 {
		sides = 4
	}
	ringY := height * 0.36
	ring := make([][3]float32, sides)
	for i := 0; i < sides; i++ {
		ang := 2 * math.Pi * float64(i) / float64(sides)
		rx := float32(math.Cos(ang)) * radius
		rz := float32(math.Sin(ang)) * radius
		ring[i] = [3]float32{
			baseCenter[0] + rx + tiltX*ringY,
			baseCenter[1] + ringY,
			baseCenter[2] + rz + tiltZ*ringY,
		}
	}
	tip := [3]float32{
		baseCenter[0] + tiltX*height,
		baseCenter[1] + height,
		baseCenter[2] + tiltZ*height,
	}
	for i := 0; i < sides; i++ {
		ni := (i + 1) % sides
		b.addTri(baseCenter, ring[ni], ring[i], colBase, colMid, colMid)
		b.addTri(ring[i], ring[ni], tip, colMid, colMid, colTip)
	}
}

func (b *meshBuilder) addFacetedMound(cx, cy, cz, rx, ry, rz float32, sides int, colRim, colTop [4]float32) {
	if sides < 5 {
		sides = 5
	}
	lower := make([][3]float32, sides)
	upper := make([][3]float32, sides)
	for i := 0; i < sides; i++ {
		ang := 2 * math.Pi * float64(i) / float64(sides)
		wobble := float32(0.84 + 0.16*math.Sin(float64(i)*2.3))
		ca := float32(math.Cos(ang)) * wobble
		sa := float32(math.Sin(ang)) * wobble
		lower[i] = [3]float32{cx + rx*ca, cy, cz + rz*sa}
		upper[i] = [3]float32{cx + rx*0.64*ca, cy + ry*(0.72+0.14*wobble), cz + rz*0.64*sa}
	}
	capCenter := [3]float32{cx, cy + ry, cz}
	botCenter := [3]float32{cx, cy, cz}
	for i := 0; i < sides; i++ {
		ni := (i + 1) % sides
		b.addTri(lower[i], lower[ni], upper[ni], colRim, colRim, colTop)
		b.addTri(lower[i], upper[ni], upper[i], colRim, colTop, colTop)
		b.addTri(upper[i], upper[ni], capCenter, colTop, colTop, colTop)
		b.addTri(botCenter, lower[ni], lower[i], colRim, colRim, colRim)
	}
}

func (b *meshBuilder) toPart(name string, roughness, metallic float32, animMode string, activeOnly bool, emissive [4]float32, emissiveInt float32) sdk.CustomSubMeshSpec {
	return sdk.CustomSubMeshSpec{
		Name:              name,
		Positions:         b.positions,
		Colors:            b.colors,
		Indices:           b.indices,
		Color:             [4]float32{1, 1, 1, 1},
		Roughness:         roughness,
		Metallic:          metallic,
		EmissiveColor:     emissive,
		EmissiveIntensity: emissiveInt,
		AnimMode:          animMode,
		ActiveOnly:        activeOnly,
	}
}

func registerCustomRealmMeshes() {
	if err := sdk.RegisterCustomMesh(buildGildedVeinMesh()); err != nil {
		sdk.Log(fmt.Sprintf("RegisterCustomMesh(gilded_vein_mesh) error: %v", err))
	}
	if err := sdk.RegisterCustomMesh(buildGildedOreItemMesh()); err != nil {
		sdk.Log(fmt.Sprintf("RegisterCustomMesh(gilded_ore_item) error: %v", err))
	}
}

func buildGildedVeinMesh() sdk.CustomMeshSpec {
	basalt := &meshBuilder{}
	basaltRim := [4]float32{0.18, 0.19, 0.23, 1}
	basaltTop := [4]float32{0.31, 0.29, 0.34, 1}
	basalt.addFacetedMound(0, 0, 0, 1.08, 0.62, 0.98, 9, basaltRim, basaltTop)
	basalt.addFacetedMound(-0.52, 0, 0.38, 0.44, 0.36, 0.42, 6, basaltRim, basaltTop)
	basalt.addFacetedMound(0.54, 0, -0.30, 0.48, 0.42, 0.44, 6, basaltRim, basaltTop)

	crystals := &meshBuilder{}
	crimsonRoot := [4]float32{0.76, 0.18, 0.14, 1}
	goldBody := [4]float32{0.96, 0.76, 0.20, 1}
	goldCrown := [4]float32{1.0, 0.94, 0.56, 1}
	crystals.addDiamondSpire([3]float32{0.0, 0.48, 0.0}, 0.04, 0.03, 0.26, 1.18, 6, crimsonRoot, goldBody, goldCrown)
	crystals.addDiamondSpire([3]float32{-0.34, 0.40, 0.20}, -0.30, 0.22, 0.18, 0.86, 5, crimsonRoot, goldBody, goldCrown)
	crystals.addDiamondSpire([3]float32{0.34, 0.42, 0.16}, 0.32, 0.16, 0.17, 0.82, 5, crimsonRoot, goldBody, goldCrown)
	crystals.addDiamondSpire([3]float32{0.12, 0.42, -0.34}, 0.10, -0.32, 0.19, 0.90, 5, crimsonRoot, goldBody, goldCrown)
	crystals.addDiamondSpire([3]float32{-0.22, 0.36, -0.26}, -0.22, -0.24, 0.15, 0.72, 5, crimsonRoot, goldBody, goldCrown)

	return sdk.CustomMeshSpec{
		ID: "gilded_vein_mesh",
		Parts: []sdk.CustomSubMeshSpec{
			basalt.toPart("vein_basalt_base", 0.86, 0.08, "", false, [4]float32{}, 0),
			crystals.toPart("vein_gold_shards", 0.20, 0.74, "bob", true, [4]float32{0.98, 0.74, 0.18, 1.0}, 0.55),
		},
	}
}

func buildGildedOreItemMesh() sdk.CustomMeshSpec {
	matrix := &meshBuilder{}
	rockDark := [4]float32{0.24, 0.18, 0.16, 1}
	rockLight := [4]float32{0.40, 0.28, 0.20, 1}
	matrix.addFacetedMound(0, -0.24, 0, 0.44, 0.46, 0.40, 7, rockDark, rockLight)

	shards := &meshBuilder{}
	crimsonRoot := [4]float32{0.80, 0.16, 0.16, 1}
	goldBody := [4]float32{0.98, 0.80, 0.24, 1}
	goldTip := [4]float32{1.0, 0.96, 0.68, 1}
	shards.addDiamondSpire([3]float32{0, 0.08, 0}, 0.10, 0.05, 0.15, 0.54, 5, crimsonRoot, goldBody, goldTip)
	shards.addDiamondSpire([3]float32{-0.18, 0.04, 0.12}, -0.34, 0.20, 0.11, 0.40, 5, crimsonRoot, goldBody, goldTip)
	shards.addDiamondSpire([3]float32{0.18, 0.02, -0.10}, 0.30, -0.22, 0.12, 0.42, 5, crimsonRoot, goldBody, goldTip)

	return sdk.CustomMeshSpec{
		ID: "gilded_ore_item",
		Parts: []sdk.CustomSubMeshSpec{
			matrix.toPart("ore_chunk_matrix", 0.78, 0.18, "", false, [4]float32{}, 0),
			shards.toPart("ore_crystal_veins", 0.18, 0.72, "", false, [4]float32{0.98, 0.76, 0.22, 1.0}, 0.65),
		},
	}
}
