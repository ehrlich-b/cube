package cube

// ============================================================================
// CUBIE-LEVEL COORDINATE REPRESENTATION (3x3 ONLY)
//
// This file is the cubie-level foundation that move tables and pruning tables
// for a real solver will eventually be built on. It mirrors the existing
// sticker engine (the sticker grid is the oracle) in a coordinate space that
// tracks physical cubies: corner permutation + orientation and edge
// permutation + orientation, exactly like Thistlethwaite/Kociemba/Korf solvers.
//
// It is 3x3 only by design. ToCubieState refuses non-3x3 cubes.
//
// ---------------------------------------------------------------------------
// 3D COORDINATE FRAME
//
// The cube is placed in a unit cube with faces along the axes:
//
//	+x = Right (R),  -x = Left (L)
//	+y = Up    (U),  -y = Down (D)
//	+z = Front (F),  -z = Back (B)
//
// Every word below ("clockwise", "Right", ...) is interpreted exactly as this
// repository's sticker engine toggles stickers (ring generators + face
// rotations in permutations.go). The move tables in this file are generated to
// reproduce the engine's sticker permutation one-to-one. Consequence: the whole
// construction is a mirror image of the WCA convention (every "clockwise"
// quarter turn is the reverse of the usual right-handed view), but that is only
// a labeling choice; all group properties below are unaffected and the tables
// are exactly consistent with the existing engine.
//
// A face turn is a rotation of one layer. As a 3x3 rotation applied to the
// coordinates (x, y, z) of the moved stickers' outward normals, the six quarter
// turns (and their powers) are:
//
//	R: (x, y, z) -> (x,  z, -y)
//	L: (x, y, z) -> (x, -z,  y)
//	U: (x, y, z) -> (-z, y,  x)
//	D: (x, y, z) -> (z,  y, -x)
//	F: (x, y, z) -> (y, -x,  z)
//	B: (x, y, z) -> (-y, x,  z)
//
// (R2/L2/U2/D2/F2/B2 are 180-degree variants; three applications of the table
// give the inverse quarter turn.) These six exact integer maps were checked to
// reproduce, sticker for sticker, the engine's own permutation. Keeping them as
// literal data means a reader can re-derive every row of the move tables below
// by hand.
//
// ---------------------------------------------------------------------------
// CORNER SLOTS (8)
//
// A corner is a body position on the cube. Slots are numbered 0..7 and named
// after the three faces that meet there. For each slot the three facelets are
// listed in the fixed order used all over this file:
//
//	  order:  X = the L/R facelet,  Y = the U/D facelet,  Z = the F/B facelet
//
//	slot  name   position     X facelet    Y facelet    Z facelet
//	0     URF    (+x,+y,+z)   R(0,0)       U(2,2)       F(0,2)
//	1     UFL    (-x,+y,+z)   L(0,2)       U(2,0)       F(0,0)
//	2     ULB    (-x,+y,-z)   L(0,0)       U(0,0)       B(0,2)
//	3     UBR    (+x,+y,-z)   R(0,2)       U(0,2)       B(0,0)
//	4     DFR    (+x,-y,+z)   R(2,0)       D(0,2)       F(2,2)
//	5     DFL    (-x,-y,+z)   L(2,2)       D(0,0)       F(2,0)
//	6     DBL    (-x,-y,-z)   L(2,0)       D(2,0)       B(2,2)
//	7     DBR    (+x,-y,-z)   R(2,2)       D(2,2)       B(2,0)
//
// ("U(2,2)" means the facelet at face Up, row 2, column 2, in the same
// (row, column) indexing the sticker engine and the ASCII display use.)
//
// CORNER ORIENTATION (CO[slot] in 0..2)
//
// The orientation reference is the U/D-colored sticker of the corner cubie
// (Yellow or White). For each slot fix a clockwise sense by looking along the
// slot's body diagonal (+diag) from the cube center toward the corner; define
// a +120-degree right-handed rotation about that diagonal as the "twist" step:
//
//   step 0 (CO=0): the cubie's U/D sticker sits on the Y facelet (U or D face)
//   step 1 (CO=1): ... on the successor facelet of Y in the +120 cycle
//   step 2 (CO=2): ... on the next successor
//
// The per-slot +120 cycles (successor of each facelet X, Y, Z) are:
//
//	slot  0  1  2  3  4  5  6  7
//	X->   Y  Z  Y  Z  Z  Y  Z  Y
//	Y->   Z  X  Z  X  X  Z  X  Z
//	Z->   X  Y  X  Y  Y  X  Y  X
//
// Equivalently, the table cornerTwistFromUD (see below) gives CO directly from
// "which of the three facelets (X,Y,Z) currently holds the U/D sticker".
//
//	slot:   0  1  2  3  4  5  6  7
//	X holds U/D: CO=2 1 2 1 1 2 1 2
//	Y holds U/D: CO=0 0 0 0 0 0 0 0
//	Z holds U/D: CO=1 2 1 2 2 1 2 1
//
// This is the standard twist convention (the 8 corners' CO sums to 0 mod 3 on
// every reachable state), with the twist sense fixed by the +diagonal choice.
//
// ---------------------------------------------------------------------------
// EDGE SLOTS (12)
//
//	slot  name  position      facelet 0      facelet 1
//	0     UF    (0,+y,+z)     U(2,1)         F(0,1)
//	1     UR    (+x,+y,0)     U(1,2)         R(0,1)
//	2     UB    (0,+y,-z)     U(0,1)         B(0,1)
//	3     UL    (-x,+y,0)     U(1,0)         L(0,1)
//	4     DF    (0,-y,+z)     D(0,1)         F(2,1)
//	5     DR    (+x,-y,0)     D(1,2)         R(2,1)
//	6     DB    (0,-y,-z)     D(2,1)         B(2,1)
//	7     DL    (-x,-y,0)     D(1,0)         L(2,1)
//	8     FR    (+x,0,+z)     F(1,2)         R(1,0)
//	9     FL    (-x,0,+z)     F(1,0)         L(1,2)
//	10    BL    (-x,0,-z)     B(1,2)         L(1,0)
//	11    BR    (+x,0,-z)     B(1,0)         R(1,2)
//
// The facelet-0 position of every slot is its "home"/reference facelet: the
// U/D facelet for the U/D slots (0..7) and the F/B facelet for the four
// equator slots (8..11).
//
// EDGE ORIENTATION (EO[slot] in 0..1)
//
// The reference sticker of an edge cubie is its U/D-colored sticker (Yellow or
// White); for the four equator cubies, which have no U/D sticker, the reference
// is its F/B-colored sticker (Blue or Green).
//
//	EO = 0  iff  the reference sticker sits on the slot's facelet 0.
//	EO = 1  iff  it sits on facelet 1.
//
// For a cubie in its home slot this agrees with the familiar reading ("U/D
// sticker on the U or D face" for U/D edges, "F/B sticker on the F or B face"
// for equator edges). The two readings differ whenever a U/D edge sits in an
// equator slot (or vice-versa), and this facelet-0 reading is the one with the
// clean linear move behaviour: every face turn flips a fixed set of edges
// (those whose reference facelet is mapped to a non-reference facelet), and the
// 12 edges' EO sums to 0 mod 2 on every reachable state.
//
// ---------------------------------------------------------------------------
// IDENTITY OF A CUBIE
//
// A cubie is identified by the COLOR SET of its stickers, never by its slot:
// the corner cubie with color set {Yellow, Red, Blue} is "the cubie that lives
// in slot 0" and therefore has identity 0 (CP = 0 wherever it is sitting), and
// likewise for the other seven corners and the twelve edges. In a solved cube
// every cubie sits in its own slot, so CP[i] == i and EP[i] == i.
//
// MOVE TABLES
//
// The table moveTables[face][qt] below describes a single quarter-turn family
// (qt = 1 cw, 2 = 180 deg, 3 = ccw). For the cubie sitting in slot i before the
// move:
//
//	cPerm[i]  = the slot the cubie moves to
//	cOrient[i] = the amount to ADD (mod 3) to its corner orientation
//	ePerm[i]  = the slot the edge moves to
//	eFlip[i]  = the amount to ADD (mod 2) to its edge orientation
//
// All rows were generated from the six rotation maps at the top and then
// verified differentially against the sticker engine (see coord_test.go).
// ============================================================================

import (
	"fmt"
)

// CubieState is the cubie-level coordinate representation of a 3x3 cube.
//
// CP[i] = identity of the corner cubie sitting in corner slot i.
// CO[i] = orientation of that corner (0..2), see header comment.
// EP[i] = identity of the edge cubie sitting in edge slot i.
// EO[i] = orientation of that edge (0 or 1), see header comment.
type CubieState struct {
	CP [8]int
	CO [8]int
	EP [12]int
	EO [12]int
}

// SolvedCubieState returns the cubie state of a solved cube: every cubie sits
// in its own slot, fully oriented.
func SolvedCubieState() CubieState {
	var s CubieState
	for i := 0; i < 8; i++ {
		s.CP[i] = i
	}
	for i := 0; i < 12; i++ {
		s.EP[i] = i
	}
	return s
}

// IsSolved reports whether the cubie state is the solved state.
func (s CubieState) IsSolved() bool {
	for i := 0; i < 8; i++ {
		if s.CP[i] != i || s.CO[i] != 0 {
			return false
		}
	}
	for i := 0; i < 12; i++ {
		if s.EP[i] != i || s.EO[i] != 0 {
			return false
		}
	}
	return true
}

// faceCell addresses one sticker. Order of fields: face, row, column.
type faceCell struct {
	face Face
	row  int
	col  int
}

// colorAt reads the color at a faceCell.
func colorAt(c *Cube, fc faceCell) Color {
	return c.Faces[fc.face][fc.row][fc.col]
}

func isUDColor(c Color) bool { return c == White || c == Yellow }
func isFBColor(c Color) bool { return c == Blue || c == Green }

// ---------------------------------------------------------------------------
// Slot geometry tables (see the header comment).

// cornerFacelets lists, per corner slot, its three facelets in X,Y,Z order.
var cornerFacelets = [8][3]faceCell{
	{{Right, 0, 0}, {Up, 2, 2}, {Front, 0, 2}},   // 0 URF
	{{Left, 0, 2}, {Up, 2, 0}, {Front, 0, 0}},    // 1 UFL
	{{Left, 0, 0}, {Up, 0, 0}, {Back, 0, 2}},     // 2 ULB
	{{Right, 0, 2}, {Up, 0, 2}, {Back, 0, 0}},    // 3 UBR
	{{Right, 2, 0}, {Down, 0, 2}, {Front, 2, 2}}, // 4 DFR
	{{Left, 2, 2}, {Down, 0, 0}, {Front, 2, 0}},  // 5 DFL
	{{Left, 2, 0}, {Down, 2, 0}, {Back, 2, 2}},   // 6 DBL
	{{Right, 2, 2}, {Down, 2, 2}, {Back, 2, 0}},  // 7 DBR
}

// edgeFacelets lists, per edge slot, its two facelets: facelet 0 is the
// reference facelet (U/D for slots 0..7, F/B for slots 8..11).
var edgeFacelets = [12][2]faceCell{
	{{Up, 2, 1}, {Front, 0, 1}},    // 0 UF
	{{Up, 1, 2}, {Right, 0, 1}},    // 1 UR
	{{Up, 0, 1}, {Back, 0, 1}},     // 2 UB
	{{Up, 1, 0}, {Left, 0, 1}},     // 3 UL
	{{Down, 0, 1}, {Front, 2, 1}},  // 4 DF
	{{Down, 1, 2}, {Right, 2, 1}},  // 5 DR
	{{Down, 2, 1}, {Back, 2, 1}},   // 6 DB
	{{Down, 1, 0}, {Left, 2, 1}},   // 7 DL
	{{Front, 1, 2}, {Right, 1, 0}}, // 8 FR
	{{Front, 1, 0}, {Left, 1, 2}},  // 9 FL
	{{Back, 1, 2}, {Left, 1, 0}},   // 10 BL
	{{Back, 1, 0}, {Right, 1, 2}},  // 11 BR
}

// cornerTwistFromUD maps (slot, facelet index X/Y/Z of the U/D sticker) to CO.
var cornerTwistFromUD = [8][3]int{
	{2, 0, 1},
	{1, 0, 2},
	{2, 0, 1},
	{1, 0, 2},
	{1, 0, 2},
	{2, 0, 1},
	{1, 0, 2},
	{2, 0, 1},
}

// cornerHomeColors holds, per corner identity i, the sorted color triple of the
// three stickers of the corner cubie that lives in slot i.
var cornerHomeColors = [8][3]Color{
	{Yellow, Red, Blue},     // 0 URF:  U, R, F
	{Yellow, Orange, Blue},  // 1 UFL:  U, L, F
	{Yellow, Orange, Green}, // 2 ULB: U, L, B
	{Yellow, Red, Green},    // 3 UBR:  U, R, B
	{White, Red, Blue},      // 4 DFR:  D, R, F
	{White, Orange, Blue},   // 5 DFL:  D, L, F
	{White, Orange, Green},  // 6 DBL:  D, L, B
	{White, Red, Green},     // 7 DBR:  D, R, B
}

// edgeHomeColors holds, per edge identity i, the sorted color pair of the two
// stickers of the edge cubie that lives in slot i.
var edgeHomeColors = [12][2]Color{
	{Yellow, Blue},   // 0 UF
	{Yellow, Red},    // 1 UR
	{Yellow, Green},  // 2 UB
	{Yellow, Orange}, // 3 UL
	{White, Blue},    // 4 DF
	{White, Red},     // 5 DR
	{White, Green},   // 6 DB
	{White, Orange},  // 7 DL
	{Blue, Red},      // 8 FR
	{Blue, Orange},   // 9 FL
	{Green, Orange},  // 10 BL
	{Green, Red},     // 11 BR
}

// ToCubieState derives the cubie-level state from a sticker cube. It returns an
// error for any cube whose size is not 3 (the cubie representation is 3x3 only)
// or that cannot be represented (a sticker layout that no legal cube produces).
func ToCubieState(c *Cube) (CubieState, error) {
	if c.Size != 3 {
		return CubieState{}, fmt.Errorf("ToCubieState requires a 3x3 cube, got size %d", c.Size)
	}
	var s CubieState

	for slot := 0; slot < 8; slot++ {
		var col [3]Color
		for i := 0; i < 3; i++ {
			col[i] = colorAt(c, cornerFacelets[slot][i])
		}
		// Locate the U/D-colored sticker to read the orientation.
		udIdx := -1
		for i := 0; i < 3; i++ {
			if isUDColor(col[i]) {
				udIdx = i
				break
			}
		}
		if udIdx < 0 {
			return CubieState{}, fmt.Errorf("no U/D sticker on corner slot %d", slot)
		}
		s.CO[slot] = cornerTwistFromUD[slot][udIdx]
		id, ok := cornerIdentity(col[0], col[1], col[2])
		if !ok {
			return CubieState{}, fmt.Errorf("unrecognized corner cubie in slot %d (colors %s %s %s)", slot, col[0], col[1], col[2])
		}
		s.CP[slot] = id
	}

	for slot := 0; slot < 12; slot++ {
		c0 := colorAt(c, edgeFacelets[slot][0])
		c1 := colorAt(c, edgeFacelets[slot][1])
		// Reference sticker: the U/D-colored one if present, else the F/B one.
		eo := 0
		switch {
		case isUDColor(c0):
			eo = 0
		case isUDColor(c1):
			eo = 1
		case isFBColor(c0):
			eo = 0
		default:
			eo = 1
		}
		s.EO[slot] = eo
		id, ok := edgeIdentity(c0, c1)
		if !ok {
			return CubieState{}, fmt.Errorf("unrecognized edge cubie in slot %d (colors %s %s)", slot, c0, c1)
		}
		s.EP[slot] = id
	}

	return s, nil
}

// cornerIdentity identifies a corner cubie by its three sticker colors.
func cornerIdentity(a, b, c Color) (int, bool) {
	sa, sb, sc := a, b, c
	if sa > sb {
		sa, sb = sb, sa
	}
	if sb > sc {
		sb, sc = sc, sb
	}
	if sa > sb {
		sa, sb = sb, sa
	}
	for id := 0; id < 8; id++ {
		t := cornerHomeColors[id]
		ta, tb, tc := t[0], t[1], t[2]
		if ta > tb {
			ta, tb = tb, ta
		}
		if tb > tc {
			tb, tc = tc, tb
		}
		if ta > tb {
			ta, tb = tb, ta
		}
		if sa == ta && sb == tb && sc == tc {
			return id, true
		}
	}
	return 0, false
}

// edgeIdentity identifies an edge cubie by its two sticker colors.
func edgeIdentity(a, b Color) (int, bool) {
	sa, sb := a, b
	if sa > sb {
		sa, sb = sb, sa
	}
	for id := 0; id < 12; id++ {
		t := edgeHomeColors[id]
		ta, tb := t[0], t[1]
		if ta > tb {
			ta, tb = tb, ta
		}
		if sa == ta && sb == tb {
			return id, true
		}
	}
	return 0, false
}

// ---------------------------------------------------------------------------
// Move tables. moveTables[face][qt-1] where face is a Face (Front=0 ... Down=5)
// and qt is 1 (cw), 2 (180) or 3 (ccw). Entries for slots outside the turned
// layer are the identity (== their index) with orientation changes of 0.

type moveTable struct {
	cPerm   [8]int
	cOrient [8]int
	ePerm   [12]int
	eFlip   [12]int
}

var moveTables = [6][3]moveTable{
	// F (Face 0)
	{
		{ // qt=1
			cPerm:   [8]int{4, 0, 2, 3, 5, 1, 6, 7},
			cOrient: [8]int{1, 2, 0, 0, 2, 1, 0, 0},
			ePerm:   [12]int{8, 1, 2, 3, 9, 5, 6, 7, 4, 0, 10, 11},
			eFlip:   [12]int{1, 0, 0, 0, 1, 0, 0, 0, 1, 1, 0, 0},
		},
		{ // qt=2
			cPerm:   [8]int{5, 4, 2, 3, 1, 0, 6, 7},
			cOrient: [8]int{0, 0, 0, 0, 0, 0, 0, 0},
			ePerm:   [12]int{4, 1, 2, 3, 0, 5, 6, 7, 9, 8, 10, 11},
			eFlip:   [12]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{ // qt=3
			cPerm:   [8]int{1, 5, 2, 3, 0, 4, 6, 7},
			cOrient: [8]int{1, 2, 0, 0, 2, 1, 0, 0},
			ePerm:   [12]int{9, 1, 2, 3, 8, 5, 6, 7, 0, 4, 10, 11},
			eFlip:   [12]int{1, 0, 0, 0, 1, 0, 0, 0, 1, 1, 0, 0},
		},
	},
	// B (Face 1)
	{
		{ // qt=1
			cPerm:   [8]int{0, 1, 6, 2, 4, 5, 7, 3},
			cOrient: [8]int{0, 0, 1, 2, 0, 0, 2, 1},
			ePerm:   [12]int{0, 1, 10, 3, 4, 5, 11, 7, 8, 9, 6, 2},
			eFlip:   [12]int{0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 1},
		},
		{ // qt=2
			cPerm:   [8]int{0, 1, 7, 6, 4, 5, 3, 2},
			cOrient: [8]int{0, 0, 0, 0, 0, 0, 0, 0},
			ePerm:   [12]int{0, 1, 6, 3, 4, 5, 2, 7, 8, 9, 11, 10},
			eFlip:   [12]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{ // qt=3
			cPerm:   [8]int{0, 1, 3, 7, 4, 5, 2, 6},
			cOrient: [8]int{0, 0, 1, 2, 0, 0, 2, 1},
			ePerm:   [12]int{0, 1, 11, 3, 4, 5, 10, 7, 8, 9, 2, 6},
			eFlip:   [12]int{0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 1},
		},
	},
	// L (Face 2)
	{
		{ // qt=1
			cPerm:   [8]int{0, 5, 1, 3, 4, 6, 2, 7},
			cOrient: [8]int{0, 1, 2, 0, 0, 2, 1, 0},
			ePerm:   [12]int{0, 1, 2, 9, 4, 5, 6, 10, 8, 7, 3, 11},
			eFlip:   [12]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{ // qt=2
			cPerm:   [8]int{0, 6, 5, 3, 4, 2, 1, 7},
			cOrient: [8]int{0, 0, 0, 0, 0, 0, 0, 0},
			ePerm:   [12]int{0, 1, 2, 7, 4, 5, 6, 3, 8, 10, 9, 11},
			eFlip:   [12]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{ // qt=3
			cPerm:   [8]int{0, 2, 6, 3, 4, 1, 5, 7},
			cOrient: [8]int{0, 1, 2, 0, 0, 2, 1, 0},
			ePerm:   [12]int{0, 1, 2, 10, 4, 5, 6, 9, 8, 3, 7, 11},
			eFlip:   [12]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	},
	// R (Face 3)
	{
		{ // qt=1
			cPerm:   [8]int{3, 1, 2, 7, 0, 5, 6, 4},
			cOrient: [8]int{2, 0, 0, 1, 1, 0, 0, 2},
			ePerm:   [12]int{0, 11, 2, 3, 4, 8, 6, 7, 1, 9, 10, 5},
			eFlip:   [12]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{ // qt=2
			cPerm:   [8]int{7, 1, 2, 4, 3, 5, 6, 0},
			cOrient: [8]int{0, 0, 0, 0, 0, 0, 0, 0},
			ePerm:   [12]int{0, 5, 2, 3, 4, 1, 6, 7, 11, 9, 10, 8},
			eFlip:   [12]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{ // qt=3
			cPerm:   [8]int{4, 1, 2, 0, 7, 5, 6, 3},
			cOrient: [8]int{2, 0, 0, 1, 1, 0, 0, 2},
			ePerm:   [12]int{0, 8, 2, 3, 4, 11, 6, 7, 5, 9, 10, 1},
			eFlip:   [12]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	},
	// U (Face 4)
	{
		{ // qt=1
			cPerm:   [8]int{1, 2, 3, 0, 4, 5, 6, 7},
			cOrient: [8]int{0, 0, 0, 0, 0, 0, 0, 0},
			ePerm:   [12]int{3, 0, 1, 2, 4, 5, 6, 7, 8, 9, 10, 11},
			eFlip:   [12]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{ // qt=2
			cPerm:   [8]int{2, 3, 0, 1, 4, 5, 6, 7},
			cOrient: [8]int{0, 0, 0, 0, 0, 0, 0, 0},
			ePerm:   [12]int{2, 3, 0, 1, 4, 5, 6, 7, 8, 9, 10, 11},
			eFlip:   [12]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{ // qt=3
			cPerm:   [8]int{3, 0, 1, 2, 4, 5, 6, 7},
			cOrient: [8]int{0, 0, 0, 0, 0, 0, 0, 0},
			ePerm:   [12]int{1, 2, 3, 0, 4, 5, 6, 7, 8, 9, 10, 11},
			eFlip:   [12]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	},
	// D (Face 5)
	{
		{ // qt=1
			cPerm:   [8]int{0, 1, 2, 3, 7, 4, 5, 6},
			cOrient: [8]int{0, 0, 0, 0, 0, 0, 0, 0},
			ePerm:   [12]int{0, 1, 2, 3, 5, 6, 7, 4, 8, 9, 10, 11},
			eFlip:   [12]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{ // qt=2
			cPerm:   [8]int{0, 1, 2, 3, 6, 7, 4, 5},
			cOrient: [8]int{0, 0, 0, 0, 0, 0, 0, 0},
			ePerm:   [12]int{0, 1, 2, 3, 6, 7, 4, 5, 8, 9, 10, 11},
			eFlip:   [12]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{ // qt=3
			cPerm:   [8]int{0, 1, 2, 3, 5, 6, 7, 4},
			cOrient: [8]int{0, 0, 0, 0, 0, 0, 0, 0},
			ePerm:   [12]int{0, 1, 2, 3, 7, 4, 5, 6, 8, 9, 10, 11},
			eFlip:   [12]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	},
}

// quarterTurns returns 1, 2 or 3 for a clockwise, double or counter-clockwise
// face turn, mirroring moveToMoveType in moves.go.
func quarterTurns(mv Move) int {
	if mv.Double {
		return 2
	}
	if mv.Clockwise {
		return 1
	}
	return 3
}

// ApplyMove applies a single face turn to the cubie state, returning a new
// value. Only the 18 face turns are supported by the cubie representation; any
// other move type (wide, slice, layer, rotation) leaves the state unchanged.
func (s CubieState) ApplyMove(mv Move) CubieState {
	if mv.Face < Front || mv.Face > Down || mv.Wide || mv.Slice != NoSlice || mv.Rotation != NoRotation || mv.Layer != 0 {
		return s
	}
	tbl := &moveTables[mv.Face][quarterTurns(mv)-1]
	var ns CubieState
	for i := 0; i < 8; i++ {
		ns.CP[tbl.cPerm[i]] = s.CP[i]
		ns.CO[tbl.cPerm[i]] = (s.CO[i] + tbl.cOrient[i]) % 3
	}
	for i := 0; i < 12; i++ {
		ns.EP[tbl.ePerm[i]] = s.EP[i]
		ns.EO[tbl.ePerm[i]] = (s.EO[i] + tbl.eFlip[i]) % 2
	}
	return ns
}

// ApplyMoves applies a sequence of face turns to the cubie state.
func (s CubieState) ApplyMoves(mvs []Move) CubieState {
	for _, mv := range mvs {
		s = s.ApplyMove(mv)
	}
	return s
}
