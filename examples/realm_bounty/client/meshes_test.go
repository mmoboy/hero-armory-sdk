package main

import "testing"

// assertOutward checks every triangle of b is wound counter-clockwise seen
// from outside: its normal points away from center (the SDK front-face
// convention; the host derives lighting normals from the winding).
func assertOutward(t *testing.T, name string, b *meshBuilder, center [3]float32) {
	t.Helper()
	sub := func(a, c [3]float32) [3]float32 { return [3]float32{a[0] - c[0], a[1] - c[1], a[2] - c[2]} }
	for i := 0; i+2 < len(b.indices); i += 3 {
		p0, p1, p2 := b.positions[b.indices[i]], b.positions[b.indices[i+1]], b.positions[b.indices[i+2]]
		e1, e2 := sub(p1, p0), sub(p2, p0)
		n := [3]float32{e1[1]*e2[2] - e1[2]*e2[1], e1[2]*e2[0] - e1[0]*e2[2], e1[0]*e2[1] - e1[1]*e2[0]}
		mid := [3]float32{(p0[0] + p1[0] + p2[0]) / 3, (p0[1] + p1[1] + p2[1]) / 3, (p0[2] + p1[2] + p2[2]) / 3}
		out := sub(mid, center)
		if n[0]*out[0]+n[1]*out[1]+n[2]*out[2] <= 0 {
			t.Fatalf("%s: triangle %d is wound clockwise (normal points inward)", name, i/3)
		}
	}
}

func TestMeshHelpersWindOutward(t *testing.T) {
	white := [4]float32{1, 1, 1, 1}

	var spire meshBuilder
	spire.addDiamondSpire([3]float32{0, 0, 0}, 0, 0, 0.5, 2, 6, white, white, white)
	assertOutward(t, "addDiamondSpire", &spire, [3]float32{0, 0.5, 0})

	var mound meshBuilder
	mound.addFacetedMound(0, 0, 0, 1, 1, 1, 7, white, white)
	assertOutward(t, "addFacetedMound", &mound, [3]float32{0, 0.4, 0})

	var box meshBuilder
	box.addBox(0, 0, 0, 1, 1, 1, white, white)
	assertOutward(t, "addBox", &box, [3]float32{0, 0, 0})
}
