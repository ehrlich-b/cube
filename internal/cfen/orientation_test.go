package cfen

import (
	"fmt"
	"testing"

	"github.com/ehrlich-b/cube/internal/cube"
)

func TestWGNormalizesToPhysicalCanonicalState(t *testing.T) {
	state, err := ParseCFEN("WG|W9/R9/G9/Y9/O9/B9")
	if err != nil {
		t.Fatal(err)
	}
	c, err := state.ToCube()
	if err != nil {
		t.Fatal(err)
	}
	if err := cube.Validate3x3(c); err != nil {
		t.Fatal(err)
	}
	text, err := GenerateCFEN(c)
	if err != nil || text != "YB|Y9/R9/B9/W9/O9/G9" {
		t.Fatalf("WG normalized to %q: %v", text, err)
	}
}

func TestAll24PhysicalOrientationRoundTrips(t *testing.T) {
	seen := make(map[CFENOrientation]bool)
	for x := 0; x < 4; x++ {
		for y := 0; y < 4; y++ {
			for z := 0; z < 4; z++ {
				text := ""
				for axis, n := range []int{x, y, z} {
					for i := 0; i < n; i++ {
						text += []string{"x ", "y ", "z "}[axis]
					}
				}
				rotations, _ := cube.ParseMoves(text)
				frame := cube.NewCube(3)
				frame.ApplyMoves(rotations)
				o := CFENOrientation{frame.Faces[cube.Up][1][1], frame.Faces[cube.Front][1][1]}
				if seen[o] {
					continue
				}
				seen[o] = true
				for _, size := range []int{2, 3, 4, 5} {
					original := scrambledCube(t, size, "R U F2 L' B Rw")
					state, err := FromCube(original, o)
					if err != nil {
						t.Fatal(err)
					}
					oriented := scrambledCube(t, size, "R U F2 L' B Rw")
					oriented.ApplyMoves(rotations)
					order := [...]cube.Face{cube.Up, cube.Right, cube.Front, cube.Down, cube.Left, cube.Back}
					for f, face := range order {
						for i, color := range state.Faces[f].Stickers {
							if color != oriented.Faces[face][i/size][i%size] {
								t.Fatalf("%s%s size %d face %s sticker %d has wrong orientation", o.Up, o.Front, size, face, i)
							}
						}
					}
					parsed, err := ParseCFEN(state.String())
					if err != nil {
						t.Fatal(err)
					}
					restored, err := parsed.ToCube()
					if err != nil || !cubesEqual(original, restored) {
						t.Fatalf("%s%s size %d round-trip failed: %v", o.Up, o.Front, size, err)
					}
					if size == 3 {
						if err := cube.Validate3x3(restored); err != nil {
							t.Fatal(err)
						}
					}
				}
			}
		}
	}
	if len(seen) != 24 {
		t.Fatal(fmt.Sprintf("tested %d orientations, want 24", len(seen)))
	}
}
