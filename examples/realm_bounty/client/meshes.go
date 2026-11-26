package main

import (
	"fmt"
	"math"

	sdk "github.com/mmoboy/hero-armory-sdk"
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

// Triangles are wound counter-clockwise seen from outside (the SDK front-face
// convention); meshes_test.go checks every helper.
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
		b.addTri(baseCenter, ring[i], ring[ni], colBase, colMid, colMid)
		b.addTri(ring[ni], ring[i], tip, colMid, colMid, colTip)
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
		b.addTri(lower[i], upper[ni], lower[ni], colRim, colTop, colRim)
		b.addTri(lower[i], upper[i], upper[ni], colRim, colTop, colTop)
		b.addTri(upper[ni], upper[i], capCenter, colTop, colTop, colTop)
		b.addTri(botCenter, lower[i], lower[ni], colRim, colRim, colRim)
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

func (b *meshBuilder) addBox(cx, cy, cz, hx, hy, hz float32, colBot, colTop [4]float32) {
	p := [8][3]float32{
		{cx - hx, cy - hy, cz - hz},
		{cx + hx, cy - hy, cz - hz},
		{cx + hx, cy - hy, cz + hz},
		{cx - hx, cy - hy, cz + hz},
		{cx - hx, cy + hy, cz - hz},
		{cx + hx, cy + hy, cz - hz},
		{cx + hx, cy + hy, cz + hz},
		{cx - hx, cy + hy, cz + hz},
	}
	quad := func(a, bb, c, d int, ca, cb, cc, cd [4]float32) {
		b.addTri(p[a], p[bb], p[c], ca, cb, cc)
		b.addTri(p[a], p[c], p[d], ca, cc, cd)
	}
	quad(3, 2, 6, 7, colBot, colBot, colTop, colTop) // front (+Z)
	quad(1, 0, 4, 5, colBot, colBot, colTop, colTop) // back (-Z)
	quad(2, 1, 5, 6, colBot, colBot, colTop, colTop) // right (+X)
	quad(0, 3, 7, 4, colBot, colBot, colTop, colTop) // left (-X)
	quad(7, 6, 5, 4, colTop, colTop, colTop, colTop) // top (+Y)
	quad(0, 1, 2, 3, colBot, colBot, colBot, colBot) // bottom (-Y)
}

func registerCustomRealmMeshes() {
	meshes := []sdk.CustomMeshSpec{
		buildGildedVeinMesh(),
		buildGildedOreItemMesh(),
		buildGildedAutomatonBodyMesh(),
		buildGildedHalberdMesh(),
		buildCrystalPickaxeMesh(),
		buildStalkerHaloMesh(),
	}
	for _, m := range meshes {
		if err := sdk.RegisterCustomMesh(m); err != nil {
			sdk.Log(fmt.Sprintf("RegisterCustomMesh(%s) error: %v", m.ID, err))
		}
	}
}

// buildGildedAutomatonBodyMesh builds a playable custom 3D basalt-and-gold Wakfu construct
// with articulated armored torso, broad pauldrons, hovering rune-crowned head, and glowing core.
func buildGildedAutomatonBodyMesh() sdk.CustomMeshSpec {
	armor := &meshBuilder{}
	basaltDark := [4]float32{0.16, 0.17, 0.22, 1}
	basaltMid := [4]float32{0.26, 0.27, 0.34, 1}
	goldTrim := [4]float32{0.94, 0.74, 0.22, 1}
	goldBright := [4]float32{1.0, 0.90, 0.48, 1}

	// Legs / armored greaves (Y: 0.0 -> 0.74)
	armor.addBox(-0.18, 0.36, 0.0, 0.11, 0.36, 0.12, basaltDark, basaltMid)
	armor.addBox(0.18, 0.36, 0.0, 0.11, 0.36, 0.12, basaltDark, basaltMid)
	armor.addBox(-0.18, 0.44, 0.09, 0.09, 0.08, 0.05, goldTrim, goldBright)
	armor.addBox(0.18, 0.44, 0.09, 0.09, 0.08, 0.05, goldTrim, goldBright)

	// Pelvis & armored chest torso (Y: 0.72 -> 1.40)
	armor.addBox(0.0, 0.80, 0.0, 0.26, 0.10, 0.15, basaltDark, goldTrim)
	armor.addBox(0.0, 1.14, 0.0, 0.33, 0.26, 0.19, basaltDark, basaltMid)
	armor.addBox(0.0, 1.37, 0.0, 0.36, 0.05, 0.21, goldTrim, goldBright)

	// Broad gilded pauldrons & armored arms
	armor.addBox(-0.44, 1.30, 0.0, 0.16, 0.12, 0.17, goldTrim, goldBright)
	armor.addBox(0.44, 1.30, 0.0, 0.16, 0.12, 0.17, goldTrim, goldBright)
	armor.addBox(-0.45, 0.96, 0.0, 0.10, 0.24, 0.10, basaltDark, basaltMid)
	armor.addBox(0.45, 0.96, 0.0, 0.10, 0.24, 0.10, basaltDark, basaltMid)

	// Floating rune-crowned automaton head & glowing chest Wakfu core
	coreCrystal := &meshBuilder{}
	cyanDeep := [4]float32{0.14, 0.68, 0.92, 1}
	cyanGlow := [4]float32{0.58, 0.96, 1.0, 1}
	coreCrystal.addBox(0.0, 1.58, 0.0, 0.17, 0.16, 0.16, basaltDark, goldTrim)
	coreCrystal.addBox(0.0, 1.60, 0.15, 0.13, 0.04, 0.03, cyanDeep, cyanGlow)
	coreCrystal.addDiamondSpire([3]float32{0.0, 1.72, 0.0}, 0.0, 0.0, 0.11, 0.24, 6, goldTrim, cyanDeep, cyanGlow)
	coreCrystal.addDiamondSpire([3]float32{0.0, 1.02, 0.18}, 0.0, 0.10, 0.10, 0.24, 6, cyanDeep, cyanGlow, goldBright)

	return sdk.CustomMeshSpec{
		ID: "gilded_automaton_body",
		Parts: []sdk.CustomSubMeshSpec{
			armor.toPart("automaton_armor", 0.42, 0.68, "", false, [4]float32{}, 0),
			coreCrystal.toPart("automaton_core_head", 0.18, 0.80, "bob", false, [4]float32{0.34, 0.88, 1.0, 1.0}, 0.72),
		},
	}
}

// buildGildedHalberdMesh builds a mythic gilded hunter halberd for the weapon bone socket.
func buildGildedHalberdMesh() sdk.CustomMeshSpec {
	haft := &meshBuilder{}
	darkWood := [4]float32{0.18, 0.14, 0.16, 1}
	brass := [4]float32{0.82, 0.62, 0.22, 1}
	haft.addBox(0.0, 0.42, 0.0, 0.032, 0.78, 0.032, darkWood, brass)
	haft.addBox(0.0, -0.36, 0.0, 0.048, 0.05, 0.048, brass, [4]float32{1.0, 0.88, 0.44, 1})

	blade := &meshBuilder{}
	crimson := [4]float32{0.82, 0.18, 0.16, 1}
	gold := [4]float32{0.98, 0.80, 0.24, 1}
	tip := [4]float32{1.0, 0.96, 0.68, 1}
	blade.addDiamondSpire([3]float32{0.0, 1.14, 0.0}, 0.0, 0.0, 0.09, 0.42, 4, crimson, gold, tip)
	blade.addBox(0.14, 1.04, 0.0, 0.14, 0.18, 0.024, crimson, gold)
	blade.addBox(-0.08, 1.06, 0.0, 0.07, 0.06, 0.024, gold, tip)

	return sdk.CustomMeshSpec{
		ID: "gilded_halberd_mesh",
		Parts: []sdk.CustomSubMeshSpec{
			haft.toPart("halberd_haft", 0.52, 0.45, "", false, [4]float32{}, 0),
			blade.toPart("halberd_blade", 0.18, 0.85, "", false, [4]float32{1.0, 0.78, 0.22, 1.0}, 0.58),
		},
	}
}

// buildCrystalPickaxeMesh builds a glowing Wakfu-veined mining pickaxe for the weapon bone socket.
func buildCrystalPickaxeMesh() sdk.CustomMeshSpec {
	handle := &meshBuilder{}
	ironDark := [4]float32{0.20, 0.22, 0.28, 1}
	ironLight := [4]float32{0.44, 0.48, 0.56, 1}
	handle.addBox(0.0, 0.30, 0.0, 0.035, 0.48, 0.035, ironDark, ironLight)

	head := &meshBuilder{}
	gold := [4]float32{0.96, 0.78, 0.22, 1}
	cyan := [4]float32{0.34, 0.90, 1.0, 1}
	head.addBox(0.0, 0.76, 0.0, 0.28, 0.065, 0.055, gold, cyan)
	head.addDiamondSpire([3]float32{0.24, 0.75, 0.0}, 0.45, 0.0, 0.06, 0.24, 4, gold, cyan, [4]float32{1, 1, 1, 1})
	head.addDiamondSpire([3]float32{-0.24, 0.75, 0.0}, -0.45, 0.0, 0.06, 0.24, 4, gold, cyan, [4]float32{1, 1, 1, 1})

	return sdk.CustomMeshSpec{
		ID: "crystal_pickaxe_mesh",
		Parts: []sdk.CustomSubMeshSpec{
			handle.toPart("pickaxe_handle", 0.60, 0.50, "", false, [4]float32{}, 0),
			head.toPart("pickaxe_crystal_head", 0.20, 0.80, "", false, [4]float32{0.34, 0.86, 1.0, 1.0}, 0.55),
		},
	}
}

// buildStalkerHaloMesh builds a floating gilded-crystal relic crown for the hat bone socket.
func buildStalkerHaloMesh() sdk.CustomMeshSpec {
	crown := &meshBuilder{}
	crimson := [4]float32{0.80, 0.16, 0.16, 1}
	gold := [4]float32{0.98, 0.78, 0.22, 1}
	bright := [4]float32{1.0, 0.96, 0.66, 1}
	crown.addBox(0.0, 0.08, 0.0, 0.22, 0.04, 0.22, crimson, gold)
	crown.addDiamondSpire([3]float32{0.0, 0.12, 0.18}, 0.0, 0.08, 0.06, 0.22, 5, crimson, gold, bright)
	crown.addDiamondSpire([3]float32{-0.18, 0.12, 0.04}, -0.15, 0.0, 0.05, 0.18, 4, crimson, gold, bright)
	crown.addDiamondSpire([3]float32{0.18, 0.12, 0.04}, 0.15, 0.0, 0.05, 0.18, 4, crimson, gold, bright)

	return sdk.CustomMeshSpec{
		ID: "stalker_halo_mesh",
		Parts: []sdk.CustomSubMeshSpec{
			crown.toPart("stalker_halo_crest", 0.18, 0.84, "bob", false, [4]float32{1.0, 0.80, 0.24, 1.0}, 0.62),
		},
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
