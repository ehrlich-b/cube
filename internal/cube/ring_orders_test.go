package cube

import "testing"

// pairOrder applies seq repeatedly to a 3x3 cube and returns the smallest
// positive i such that applying the sequence i times returns to solved, or -1
// if it never does within maxIter iterations.
func pairOrder(t *testing.T, seq string) int {
	mv, err := ParseMoves(seq)
	if err != nil {
		t.Fatalf("ParseMoves(%q): %v", seq, err)
	}
	c := NewCube(3)
	for i := 1; i <= 2000; i++ {
		c.ApplyMoves(mv)
		if c.IsSolved() {
			return i
		}
	}
	return -1
}

// TestFaceTurnPairOrders pins the order of every two-face product on a 3x3.
// A correct engine gives exactly 105 for adjacent faces and 4 for opposite
// faces. This is the oracle for the ring-generator bug: wrong L/B rings used
// to produce 455, 63, 77, etc.
func TestFaceTurnPairOrders(t *testing.T) {
	adjacent := []string{"R U", "R F", "R B", "R D", "U F", "U B",
		"L U", "L F", "L B", "L D", "F D", "B D"}
	opposite := []string{"R L", "U D", "F B"}
	for _, s := range adjacent {
		if o := pairOrder(t, s); o != 105 {
			t.Errorf("ADJACENT %q order = %d, want 105", s, o)
		}
	}
	for _, s := range opposite {
		if o := pairOrder(t, s); o != 4 {
			t.Errorf("OPPOSITE %q order = %d, want 4", s, o)
		}
	}
}

// TestKnownAlgorithmOrders pins the fix with real algorithms whose orders are
// known on a physical cube, so the pair table alone cannot drift.
func TestKnownAlgorithmOrders(t *testing.T) {
	cases := []struct {
		seq   string
		order int
	}{
		{"R U R' U'", 6}, // sexy move
		{"R B R' B'", 6}, // reverse sexy move on B, exercises the fixed B ring
	}
	for _, tc := range cases {
		if o := pairOrder(t, tc.seq); o != tc.order {
			t.Errorf("algorithm %q order = %d, want %d", tc.seq, o, tc.order)
		}
	}
}
