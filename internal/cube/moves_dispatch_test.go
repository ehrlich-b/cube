package cube

import (
	"fmt"
	"reflect"
	"testing"
)

// These tests exercise getAffectedLayers and moveToMoveType directly —
// the two dispatcher functions every single ApplyMove call runs through.
// They assert the functions' own return values, not their combined effect
// on a cube's stickers.

func TestGetAffectedLayersDocumentedCases(t *testing.T) {
	cases := []struct {
		name string
		move Move
		dim  int
		want []int
	}{
		{"slice odd N middle layer", Move{Slice: M_Slice}, 3, []int{1}},
		{"slice even N undefined", Move{Slice: M_Slice}, 4, []int{}},
		{"rotation affects every layer", Move{Rotation: X_Rotation}, 3, []int{0, 1, 2}},
		{"wide default depth", Move{Wide: true}, 5, []int{0, 1}},
		{"wide explicit depth", Move{Wide: true, WideDepth: 3}, 5, []int{0, 1, 2}},
		{"wide non-positive depth falls back to 2", Move{Wide: true, WideDepth: -1}, 5, []int{0, 1}},
		{"numbered layer move only that layer", Move{Layer: 2}, 5, []int{2}},
		{"plain face move outer layer only", Move{Face: Right}, 3, []int{0}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := getAffectedLayers(tc.move, tc.dim)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("getAffectedLayers(%+v, %d) = %v, want %v", tc.move, tc.dim, got, tc.want)
			}
		})
	}
}

func TestGetAffectedLayersSliceMovesAllDimensions(t *testing.T) {
	for N := 2; N <= 6; N++ {
		t.Run(fmt.Sprintf("N=%d", N), func(t *testing.T) {
			got := getAffectedLayers(Move{Slice: M_Slice}, N)
			if N%2 == 0 {
				if len(got) != 0 {
					t.Errorf("slice move on even N=%d: got %v, want []", N, got)
				}
				return
			}
			want := []int{N / 2}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("slice move on odd N=%d: got %v, want %v", N, got, want)
			}
		})
	}
}

func TestGetAffectedLayersRotationsAllLayers(t *testing.T) {
	for _, N := range []int{2, 5} {
		t.Run(fmt.Sprintf("N=%d", N), func(t *testing.T) {
			got := getAffectedLayers(Move{Rotation: X_Rotation}, N)
			want := make([]int, N)
			for i := range want {
				want[i] = i
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("rotation on N=%d: got %v, want %v", N, got, want)
			}
		})
	}
}

func TestMoveToMoveTypeFaceDispatch(t *testing.T) {
	faceTypePairs := []struct {
		face Face
		mt   MoveType
	}{
		{Right, MoveR},
		{Left, MoveL},
		{Up, MoveU},
		{Down, MoveD},
		{Front, MoveF},
		{Back, MoveB},
	}

	for _, p := range faceTypePairs {
		t.Run(p.face.String(), func(t *testing.T) {
			mt, turns := moveToMoveType(Move{Face: p.face, Clockwise: true})
			if mt != p.mt || turns != 1 {
				t.Errorf("moveToMoveType(Face=%s, cw) = (%v, %d), want (%v, 1)", p.face, mt, turns, p.mt)
			}
		})
	}
}

func TestMoveToMoveTypeQuarterTurnCounts(t *testing.T) {
	cases := []struct {
		name  string
		move  Move
		wantT int
	}{
		{"clockwise", Move{Face: Right, Clockwise: true}, 1},
		{"counter-clockwise is 3 quarter turns, not -1", Move{Face: Right, Clockwise: false}, 3},
		{"double is 2 quarter turns", Move{Face: Right, Double: true}, 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mt, turns := moveToMoveType(tc.move)
			if mt != MoveR {
				t.Errorf("expected MoveR, got %v", mt)
			}
			if turns != tc.wantT {
				t.Errorf("quarter turns = %d, want %d", turns, tc.wantT)
			}
		})
	}
}

func TestMoveToMoveTypeSliceAndRotationDispatch(t *testing.T) {
	faceTypes := map[MoveType]bool{
		MoveR: true, MoveL: true, MoveU: true,
		MoveD: true, MoveF: true, MoveB: true,
	}

	cases := []struct {
		name string
		move Move
		want MoveType
	}{
		{"M slice", Move{Slice: M_Slice, Clockwise: true}, MoveM},
		{"E slice", Move{Slice: E_Slice, Clockwise: true}, MoveE},
		{"S slice", Move{Slice: S_Slice, Clockwise: true}, MoveS},
		{"X rotation", Move{Rotation: X_Rotation, Clockwise: true}, MoveX},
		{"Y rotation", Move{Rotation: Y_Rotation, Clockwise: true}, MoveY},
		{"Z rotation", Move{Rotation: Z_Rotation, Clockwise: true}, MoveZ},
	}

	seen := map[MoveType]bool{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mt, turns := moveToMoveType(tc.move)
			if mt != tc.want {
				t.Errorf("got (%v, %d), want (%v, 1)", mt, turns, tc.want)
			}
			if turns != 1 {
				t.Errorf("quarter turns = %d, want 1", turns)
			}
			if faceTypes[mt] {
				t.Errorf("%v collides with a face MoveType (%v)", mt, tc.move)
			}
			if seen[mt] {
				t.Errorf("%v returned for two different inputs", mt)
			}
			seen[mt] = true
		})
	}
}

func TestMoveToMoveTypeFallback(t *testing.T) {
	// The zero-value Move{} has Face == Front (iota 0), a VALID face — so
	// it is NOT a fallback. Clockwise/Double are false, so it dispatches to
	// (MoveF, 3).
	t.Run("zero value is a valid face, not a fallback", func(t *testing.T) {
		mt, turns := moveToMoveType(Move{})
		if mt != MoveF || turns != 3 {
			t.Errorf("Move{} = (%v, %d), want (MoveF, 3)", mt, turns)
		}
	})

	// A Face value outside the Front..Down range is genuinely unmatched and
	// hits the face switch's `default: return MoveR, 0`.
	t.Run("unmatched face value hits default", func(t *testing.T) {
		mt, turns := moveToMoveType(Move{Face: Face(6)})
		if mt != MoveR || turns != 0 {
			t.Errorf("Move{Face: Face(6)} = (%v, %d), want (MoveR, 0)", mt, turns)
		}
	})

	// Same for the slice switch's and rotation switch's default branches.
	t.Run("unmatched slice type hits default", func(t *testing.T) {
		mt, turns := moveToMoveType(Move{Slice: SliceType(99)})
		if mt != MoveR || turns != 0 {
			t.Errorf("Move{Slice: SliceType(99)} = (%v, %d), want (MoveR, 0)", mt, turns)
		}
	})

	t.Run("unmatched rotation type hits default", func(t *testing.T) {
		mt, turns := moveToMoveType(Move{Rotation: RotationType(99)})
		if mt != MoveR || turns != 0 {
			t.Errorf("Move{Rotation: RotationType(99)} = (%v, %d), want (MoveR, 0)", mt, turns)
		}
	})
}
