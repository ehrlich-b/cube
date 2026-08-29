package cube

import "testing"

// TestStickerIndexIndexToCoordRoundTrip verifies that stickerIndex and
// indexToCoord are exact inverses for every index in [0, 6*N*N) across the
// cube sizes the search can handle.
func TestStickerIndexIndexToCoordRoundTrip(t *testing.T) {
	for _, N := range []int{2, 3, 4, 5} {
		for idx := 0; idx < 6*N*N; idx++ {
			face, row, col := indexToCoord(idx, N)
			if back := stickerIndex(face, row, col, N); back != idx {
				t.Errorf("N=%d: indexToCoord(%d) = (%s,%d,%d), stickerIndex back = %d (want %d)",
					N, idx, face.String(), row, col, back, idx)
			}
		}
	}
}

// TestStickerIndexIndexToCoordBoundaries independently derives the two
// endpoints of the flat index space (face order Front, Back, Left, Right,
// Up, Down; row-major within a face) and checks both directions.
func TestStickerIndexIndexToCoordBoundaries(t *testing.T) {
	N := 4
	// Index 0: first sticker of the first face (Front, row 0, col 0).
	if face, row, col := indexToCoord(0, N); face != Front || row != 0 || col != 0 {
		t.Errorf("indexToCoord(0, %d) = (%s,%d,%d), want (Front,0,0)", N, face.String(), row, col)
	}
	if got := stickerIndex(Front, 0, 0, N); got != 0 {
		t.Errorf("stickerIndex(Front,0,0,%d) = %d, want 0", N, got)
	}

	// Index 6*N*N-1: last sticker of the last face (Down, row N-1, col N-1).
	last := 6*N*N - 1
	if face, row, col := indexToCoord(last, N); face != Down || row != N-1 || col != N-1 {
		t.Errorf("indexToCoord(%d, %d) = (%s,%d,%d), want (Down,%d,%d)",
			last, N, face.String(), row, col, N-1, N-1)
	}
	if got := stickerIndex(Down, N-1, N-1, N); got != last {
		t.Errorf("stickerIndex(Down,%d,%d,%d) = %d, want %d", N-1, N-1, N, got, last)
	}
}

// TestFaceMovesAllSixFaces checks that faceMoves returns exactly three moves
// in the clockwise/double/ccw order for every face, all on that face.
func TestFaceMovesAllSixFaces(t *testing.T) {
	for _, face := range []Face{Front, Back, Left, Right, Up, Down} {
		moves := faceMoves(face)
		if len(moves) != 3 {
			t.Fatalf("faceMoves(%s) returned %d moves, want 3", face.String(), len(moves))
		}

		for i, mv := range moves {
			if mv.Face != face {
				t.Errorf("faceMoves(%s)[%d].Face = %s, want %s", face.String(), i, mv.Face.String(), face.String())
			}
			// Order: index 0 clockwise, index 1 double, index 2 counter-clockwise.
			if want := i == 0; mv.Clockwise != want {
				t.Errorf("faceMoves(%s)[%d].Clockwise = %v, want %v", face.String(), i, mv.Clockwise, want)
			}
			if want := i == 1; mv.Double != want {
				t.Errorf("faceMoves(%s)[%d].Double = %v, want %v", face.String(), i, mv.Double, want)
			}
		}
	}
}

// TestFaceMovesStringForms asserts the rendered form of every face's moves.
func TestFaceMovesStringForms(t *testing.T) {
	want := map[Face][]string{
		Front: {"F", "F2", "F'"},
		Back:  {"B", "B2", "B'"},
		Left:  {"L", "L2", "L'"},
		Right: {"R", "R2", "R'"},
		Up:    {"U", "U2", "U'"},
		Down:  {"D", "D2", "D'"},
	}
	for _, face := range []Face{Front, Back, Left, Right, Up, Down} {
		moves := faceMoves(face)
		got := make([]string, 3)
		for i, mv := range moves {
			got[i] = mv.String()
		}
		for i := 0; i < 3; i++ {
			if got[i] != want[face][i] {
				t.Errorf("faceMoves(%s)[%d] = %q, want %q", face.String(), i, got[i], want[face][i])
			}
		}
	}
}

// TestOppositeFacesExhaustive iterates all 36 ordered pairs and verifies the
// symmetry-agnostic canonical opposite relationship.
func TestOppositeFacesExhaustive(t *testing.T) {
	faces := []Face{Front, Back, Left, Right, Up, Down}
	opposite := func(a, b Face) bool {
		return (a == Up && b == Down) || (a == Down && b == Up) ||
			(a == Left && b == Right) || (a == Right && b == Left) ||
			(a == Front && b == Back) || (a == Back && b == Front)
	}
	for _, a := range faces {
		for _, b := range faces {
			got := oppositeFaces(a, b)
			want := opposite(a, b)
			if got != want {
				t.Errorf("oppositeFaces(%s,%s) = %v, want %v", a.String(), b.String(), got, want)
			}
		}
	}
}

// TestInvertMoveExactValues asserts the exact resulting Move (not just its
// string) for each modifier combination, plus the involution property.
func TestInvertMoveExactValues(t *testing.T) {
	cases := []struct {
		name string
		in   Move
		want Move
	}{
		{"clockwise", Move{Face: Right, Clockwise: true}, Move{Face: Right}},
		{"ccw", Move{Face: Right}, Move{Face: Right, Clockwise: true}},
		{"double", Move{Face: Right, Double: true}, Move{Face: Right, Double: true}},
	}
	for _, tc := range cases {
		got := invertMove(tc.in)
		if got != tc.want {
			t.Errorf("%s: invertMove(%q) = %q (struct %+v), want %q (struct %+v)",
				tc.name, tc.in.String(), got.String(), got, tc.want.String(), tc.want)
		}
		// Involution: inverting twice returns the original.
		if back := invertMove(got); back != tc.in {
			t.Errorf("%s: invertMove(invertMove(%q)) = %q, want %q", tc.name, tc.in.String(), back.String(), tc.in.String())
		}
	}
}

// TestInvertMoveStringForms cross-checks the string rendering of inverted
// moves for all six faces in all three variants.
func TestInvertMoveStringForms(t *testing.T) {
	// Key is the RESULT of invertMove, value is that move's rendered form.
	want := map[Move]string{
		{Face: Front}:                  "F'",
		{Face: Back}:                   "B'",
		{Face: Left}:                   "L'",
		{Face: Right}:                  "R'",
		{Face: Up}:                     "U'",
		{Face: Down}:                   "D'",
		{Face: Front, Clockwise: true}: "F",
		{Face: Back, Clockwise: true}:  "B",
		{Face: Left, Clockwise: true}:  "L",
		{Face: Right, Clockwise: true}: "R",
		{Face: Up, Clockwise: true}:    "U",
		{Face: Down, Clockwise: true}:  "D",
		{Face: Front, Double: true}:    "F2",
		{Face: Back, Double: true}:     "B2",
		{Face: Left, Double: true}:     "L2",
		{Face: Right, Double: true}:    "R2",
		{Face: Up, Double: true}:       "U2",
		{Face: Down, Double: true}:     "D2",
	}
	for _, face := range []Face{Front, Back, Left, Right, Up, Down} {
		for _, mv := range faceMoves(face) {
			inv := invertMove(mv)
			wanted, ok := want[inv]
			if !ok {
				t.Fatalf("missing expectation for invertMove(%q) = %q", mv.String(), inv.String())
			}
			if inv.String() != wanted {
				t.Errorf("invertMove(%q) = %q, want %q", mv.String(), inv.String(), wanted)
			}
		}
	}
}

// TestInvertMovePreservesNonModifierFields builds moves carrying wide/layer
// modifier fields and confirms inverting only touches Clockwise.
func TestInvertMovePreservesNonModifierFields(t *testing.T) {
	// Clockwise move: only Clockwise should flip.
	cw := Move{Face: Right, Clockwise: true, Wide: true, WideDepth: 2, Layer: 1}
	got := invertMove(cw)
	if got != (Move{Face: Right, Wide: true, WideDepth: 2, Layer: 1}) {
		t.Errorf("invertMove(clockwise+wide) = %+v, want Clockwise flipped only", got)
	}
	if got.Clockwise != false || got.Wide != true || got.WideDepth != 2 || got.Layer != 1 || got.Face != Right {
		t.Errorf("invertMove(clockwise+wide) modified non-Clockwise field: %+v", got)
	}
	if got == cw {
		t.Errorf("invertMove should differ from a non-double input, got identical %+v", got)
	}

	// Double move: returned unchanged, all modifier fields intact.
	db := Move{Face: Right, Double: true, Wide: true, WideDepth: 2, Layer: 1}
	if back := invertMove(db); back != db {
		t.Errorf("invertMove(double+wide) = %+v, want unchanged %+v", back, db)
	}
}
