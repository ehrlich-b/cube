package cube

import (
	"testing"
)

// TestPiecePredicatesSolvedCube verifies that every stage predicate is true on a
// freshly solved cube.
func TestPiecePredicatesSolvedCube(t *testing.T) {
	cube := NewCube(3)

	tests := []struct {
		name  string
		check func() bool
	}{
		{"white-blue edge solved", func() bool { return EdgeSolved(cube, White, Blue) }},
		{"white-orange edge solved", func() bool { return EdgeSolved(cube, White, Orange) }},
		{"white-red edge solved", func() bool { return EdgeSolved(cube, White, Red) }},
		{"white-green edge solved", func() bool { return EdgeSolved(cube, White, Green) }},
		{"white cross solved", func() bool { return WhiteCrossSolved(cube) }},
		{"white-blue-red corner solved", func() bool { return CornerSolved(cube, White, Blue, Red) }},
		{"white-blue-orange corner solved", func() bool { return CornerSolved(cube, White, Blue, Orange) }},
		{"white-green-red corner solved", func() bool { return CornerSolved(cube, White, Green, Red) }},
		{"white-green-orange corner solved", func() bool { return CornerSolved(cube, White, Green, Orange) }},
		{"first layer solved", func() bool { return FirstLayerSolved(cube) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check() {
				t.Errorf("%s should be true on a solved cube", tt.name)
			}
		})
	}
}

// sortedColors3 returns the three colors in ascending order.
func sortedColors3(a, b, c Color) [3]Color {
	if a > b {
		a, b = b, a
	}
	if b > c {
		b, c = c, b
	}
	if a > b {
		a, b = b, a
	}
	return [3]Color{a, b, c}
}

// assertEdgeHome checks that on a solved cube the returned edge's two stickers
// sit on faces whose centers are exactly colorA and colorB (in either order).
func assertEdgeHome(t *testing.T, c *Cube, e EdgePiece, colorA, colorB Color) {
	t.Helper()
	if e.A == e.B {
		t.Errorf("edge stickers must be distinct addresses, got %v and %v", e.A, e.B)
	}
	aColor := c.GetCubieColor(e.A)
	bColor := c.GetCubieColor(e.B)
	aFace, _, _ := CubieToFacePos(e.A, c.Size)
	bFace, _, _ := CubieToFacePos(e.B, c.Size)
	if aColor != c.Faces[aFace][c.Size/2][c.Size/2] {
		t.Errorf("sticker %v color %v does not match face center %v", e.A, aColor, c.Faces[aFace][c.Size/2][c.Size/2])
	}
	if bColor != c.Faces[bFace][c.Size/2][c.Size/2] {
		t.Errorf("sticker %v color %v does not match face center %v", e.B, bColor, c.Faces[bFace][c.Size/2][c.Size/2])
	}
	if !((aColor == colorA && bColor == colorB) || (aColor == colorB && bColor == colorA)) {
		t.Errorf("edge colors (%v,%v) do not match query (%v,%v)", aColor, bColor, colorA, colorB)
	}
}

// TestFindEdgeSolvedCube verifies FindEdge on a solved cube returns addressed
// whose face centers match the queried colors.
func TestFindEdgeSolvedCube(t *testing.T) {
	cube := NewCube(3)

	tests := []struct {
		name   string
		colorA Color
		colorB Color
	}{
		{"white-blue", White, Blue},
		{"blue-white in either order", Blue, White},
		{"white-orange", White, Orange},
		{"orange-white in either order", Orange, White},
		{"white-red", White, Red},
		{"white-green", White, Green},
		{"yellow-red", Yellow, Red},
		{"blue-yellow", Blue, Yellow},
		{"green-orange", Green, Orange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edge, err := FindEdge(cube, tt.colorA, tt.colorB)
			if err != nil {
				t.Fatalf("FindEdge(%v, %v) unexpected error: %v", tt.colorA, tt.colorB, err)
			}
			assertEdgeHome(t, cube, edge, tt.colorA, tt.colorB)
		})
	}
}

// assertCornerHome checks that on a solved cube the returned corner's stickers
// sit on faces whose centers are exactly the queried colors.
func assertCornerHome(t *testing.T, c *Cube, corner CornerPiece, colors [3]Color) {
	t.Helper()
	if corner.A == corner.B || corner.A == corner.C || corner.B == corner.C {
		t.Errorf("corner stickers must be distinct addresses, got %v", corner)
	}
	var got [3]Color
	for i, addr := range []CubieAddress{corner.A, corner.B, corner.C} {
		face, _, _ := CubieToFacePos(addr, c.Size)
		color := c.GetCubieColor(addr)
		if color != c.Faces[face][c.Size/2][c.Size/2] {
			t.Errorf("sticker %v color %v does not match face center %v", addr, color, c.Faces[face][c.Size/2][c.Size/2])
		}
		got[i] = color
	}
	if sortedColors3(got[0], got[1], got[2]) != sortedColors3(colors[0], colors[1], colors[2]) {
		t.Errorf("corner colors %v do not match query %v", got, colors)
	}
}

// TestFindCornerSolvedCube verifies FindCorner on a solved cube returns
// addressed whose face centers match the queried colors.
func TestFindCornerSolvedCube(t *testing.T) {
	cube := NewCube(3)

	tests := []struct {
		name   string
		colors [3]Color
	}{
		{"white-blue-red", [3]Color{White, Blue, Red}},
		{"white-red-blue in any order", [3]Color{White, Red, Blue}},
		{"white-blue-orange", [3]Color{White, Blue, Orange}},
		{"white-green-red", [3]Color{White, Green, Red}},
		{"white-green-orange", [3]Color{White, Green, Orange}},
		{"yellow-blue-red", [3]Color{Yellow, Blue, Red}},
		{"yellow-green-orange", [3]Color{Yellow, Green, Orange}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			corner, err := FindCorner(cube, tt.colors[0], tt.colors[1], tt.colors[2])
			if err != nil {
				t.Fatalf("FindCorner(%v) unexpected error: %v", tt.colors, err)
			}
			assertCornerHome(t, cube, corner, tt.colors)
		})
	}
}

// TestWhiteCrossAfterFMove verifies which white edges become unsolved by a
// single F move. The F move sweeps the Down face's top row, so it only disturbs
// the white-blue edge (whose Down sticker sits on the D-F boundary).
func TestWhiteCrossAfterFMove(t *testing.T) {
	cube := NewCube(3)
	moves, err := ParseScramble("F")
	if err != nil {
		t.Fatalf("failed to parse F: %v", err)
	}
	cube.ApplyMoves(moves)

	if WhiteCrossSolved(cube) {
		t.Error("WhiteCrossSolved should be false after a single F move")
	}

	tests := []struct {
		name   string
		colorA Color
		colorB Color
		want   bool
	}{
		{"white-blue edge disturbed by F", White, Blue, false},
		{"white-orange edge unaffected by F", White, Orange, true},
		{"white-red edge unaffected by F", White, Red, true},
		{"white-green edge unaffected by F", White, Green, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EdgeSolved(cube, tt.colorA, tt.colorB); got != tt.want {
				t.Errorf("EdgeSolved(%v, %v) after F = %v, want %v", tt.colorA, tt.colorB, got, tt.want)
			}
		})
	}
}

// TestPiecePredicatesAfterSexySix verifies that (R U R' U') x6 returns the cube
// to solved, so every predicate is true again.
func TestPiecePredicatesAfterSexySix(t *testing.T) {
	cube := NewCube(3)
	moves, err := ParseScramble("R U R' U' R U R' U' R U R' U' R U R' U' R U R' U' R U R' U'")
	if err != nil {
		t.Fatalf("failed to parse sexy move x6: %v", err)
	}
	cube.ApplyMoves(moves)

	if !WhiteCrossSolved(cube) {
		t.Error("WhiteCrossSolved should be true after (R U R' U') x6")
	}
	if !FirstLayerSolved(cube) {
		t.Error("FirstLayerSolved should be true after (R U R' U') x6")
	}
	if !EdgeSolved(cube, White, Blue) || !EdgeSolved(cube, White, Orange) ||
		!EdgeSolved(cube, White, Red) || !EdgeSolved(cube, White, Green) {
		t.Error("all white edges should be solved after (R U R' U') x6")
	}
	if !CornerSolved(cube, White, Blue, Red) || !CornerSolved(cube, White, Blue, Orange) ||
		!CornerSolved(cube, White, Green, Red) || !CornerSolved(cube, White, Green, Orange) {
		t.Error("all white corners should be solved after (R U R' U') x6")
	}
}

// TestFirstLayerWithInverse verifies that F R U breaks the first layer and that
// applying the exact inverse (U' R' F') restores it.
func TestFirstLayerWithInverse(t *testing.T) {
	cube := NewCube(3)
	moves, err := ParseScramble("F R U")
	if err != nil {
		t.Fatalf("failed to parse F R U: %v", err)
	}
	cube.ApplyMoves(moves)

	if FirstLayerSolved(cube) {
		t.Error("FirstLayerSolved should be false after F R U")
	}

	inverseMoves, err := ParseScramble("U' R' F'")
	if err != nil {
		t.Fatalf("failed to parse U' R' F': %v", err)
	}
	cube.ApplyMoves(inverseMoves)

	if !FirstLayerSolved(cube) {
		t.Error("FirstLayerSolved should be true after applying the inverse of F R U")
	}
}

// TestFindEdgeErrors verifies the documented error cases for FindEdge.
func TestFindEdgeErrors(t *testing.T) {
	cube := NewCube(3)

	tests := []struct {
		name   string
		colorA Color
		colorB Color
	}{
		{"same color twice", White, White},
		{"white and yellow share no edge", White, Yellow},
		{"red and orange share no edge", Red, Orange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := FindEdge(cube, tt.colorA, tt.colorB); err == nil {
				t.Errorf("FindEdge(%v, %v) should return an error", tt.colorA, tt.colorB)
			}
		})
	}
}

// TestFindCornerErrors verifies the documented error cases for FindCorner.
func TestFindCornerErrors(t *testing.T) {
	cube := NewCube(3)

	tests := []struct {
		name   string
		colors [3]Color
	}{
		{"duplicate colors", [3]Color{White, White, Blue}},
		{"white and yellow with blue is impossible", [3]Color{White, Yellow, Blue}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := FindCorner(cube, tt.colors[0], tt.colors[1], tt.colors[2]); err == nil {
				t.Errorf("FindCorner(%v) should return an error", tt.colors)
			}
		})
	}
}

// TestPieceFindersRejectNon3Cube verifies that FindEdge and FindCorner return an
// error (rather than panicking) for cube sizes other than 3.
func TestPieceFindersRejectNon3Cube(t *testing.T) {
	for _, size := range []int{2, 4, 5} {
		t.Run("edge size", func(t *testing.T) {
			c := NewCube(size)
			if _, err := FindEdge(c, White, Blue); err == nil {
				t.Errorf("FindEdge on %dx%d should return an error", size, size)
			}
		})
		t.Run("corner size", func(t *testing.T) {
			c := NewCube(size)
			if _, err := FindCorner(c, White, Blue, Red); err == nil {
				t.Errorf("FindCorner on %dx%d should return an error", size, size)
			}
		})
	}
}
