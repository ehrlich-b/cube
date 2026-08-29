package cli

// show_highlight_test.go pins the behavior of shouldHighlight, the pure
// predicate behind `cube show`'s --highlight-* modes. It takes raw ints, not
// cube.Face values: face indices follow the package's iota order
// (Front=0, Back=1, Left=2, Right=3, Up=4, Down=5). The four named modes are
// "cross", "oll", "pll", "f2l"; anything else falls through to default-false.

import "testing"

func TestShouldHighlightCross(t *testing.T) {
	cases := []struct {
		name      string
		face, row int
		col, size int
		want      bool
	}{
		{"down face center", 5, 1, 1, 3, true},
		{"down face corner", 5, 0, 0, 3, false},
		{"down face opposite corner", 5, 2, 2, 3, false},
		{"down face edge cells", 5, 0, 1, 3, true},
		{"down face left edge", 5, 1, 0, 3, true},
		{"down face right edge", 5, 1, 2, 3, true},
		{"down face bottom edge", 5, 2, 1, 3, true},
		{"up face never highlighted", 4, 1, 1, 3, false},
		{"front bottom-center adjacent edge", 0, 2, 1, 3, true},
		{"back bottom-center adjacent edge", 1, 2, 1, 3, true},
		{"left bottom-center adjacent edge", 2, 2, 1, 3, true},
		{"right bottom-center adjacent edge", 3, 2, 1, 3, true},
		{"front not bottom row", 0, 1, 1, 3, false},
		{"left not center column", 2, 2, 0, 3, false},
		{"back not bottom row", 1, 0, 1, 3, false},
		{"right corner", 3, 2, 2, 3, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldHighlight(tc.face, tc.row, tc.col, tc.size, "cross"); got != tc.want {
				t.Errorf("shouldHighlight(%d, %d, %d, %d, \"cross\") = %v, want %v",
					tc.face, tc.row, tc.col, tc.size, got, tc.want)
			}
		})
	}
}

func TestShouldHighlightOLL(t *testing.T) {
	cases := []struct {
		name      string
		face, row int
		col, size int
		want      bool
	}{
		{"up face any cell", 4, 2, 2, 3, true},
		{"front top row", 0, 0, 1, 3, true},
		{"front middle row not highlighted", 0, 1, 1, 3, false},
		{"front bottom row not highlighted", 0, 2, 0, 3, false},
		{"down face never highlighted", 5, 0, 0, 3, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldHighlight(tc.face, tc.row, tc.col, tc.size, "oll"); got != tc.want {
				t.Errorf("shouldHighlight(%d, %d, %d, %d, \"oll\") = %v, want %v",
					tc.face, tc.row, tc.col, tc.size, got, tc.want)
			}
		})
	}

	t.Run("every up face cell highlighted", func(t *testing.T) {
		for r := 0; r < 3; r++ {
			for c := 0; c < 3; c++ {
				if got := shouldHighlight(4, r, c, 3, "oll"); !got {
					t.Errorf("shouldHighlight(4, %d, %d, 3, \"oll\") = false, want true", r, c)
				}
			}
		}
	})

	t.Run("only top row highlighted on side faces", func(t *testing.T) {
		for face := 0; face <= 3; face++ {
			for r := 0; r < 3; r++ {
				for c := 0; c < 3; c++ {
					want := r == 0
					if got := shouldHighlight(face, r, c, 3, "oll"); got != want {
						t.Errorf("shouldHighlight(%d, %d, %d, 3, \"oll\") = %v, want %v",
							face, r, c, got, want)
					}
				}
			}
		}
	})
}

// TestShouldHighlightPLLMatchesOLL pins the current source as observed: the
// "pll" case is word-for-word identical to "oll" today, so the two modes
// agree on every input. This documents that fact; it is not an assertion
// that they always must agree.
func TestShouldHighlightPLLMatchesOLL(t *testing.T) {
	inputs := []struct {
		face     int
		row, col int
		size     int
	}{
		{4, 0, 0, 3}, // Up face corner
		{4, 2, 2, 3}, // Up face far corner
		{4, 1, 0, 3}, // Up face edge
		{0, 0, 2, 3}, // Front side top row
		{1, 0, 0, 3}, // Back side top row
		{3, 0, 1, 3}, // Right side top row
		{0, 1, 1, 3}, // Front non-zero row
		{2, 2, 0, 3}, // Left non-zero row
		{1, 2, 2, 3}, // Back non-zero row
		{5, 0, 0, 3}, // Down face
	}
	for _, in := range inputs {
		oll := shouldHighlight(in.face, in.row, in.col, in.size, "oll")
		pll := shouldHighlight(in.face, in.row, in.col, in.size, "pll")
		if oll != pll {
			t.Errorf("modes agree? no: shouldHighlight(%d, %d, %d, %d): \"oll\" = %v, \"pll\" = %v (they share identical logic today)",
				in.face, in.row, in.col, in.size, oll, pll)
		}
	}
}

func TestShouldHighlightF2L(t *testing.T) {
	cases := []struct {
		name      string
		face, row int
		col, size int
		want      bool
	}{
		{"down face any cell", 5, 0, 0, 3, true},
		{"front row 0 below boundary", 0, 0, 0, 3, false},
		{"front row 1 at boundary", 0, 1, 0, 3, true},
		{"front row 2 above boundary", 0, 2, 0, 3, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldHighlight(tc.face, tc.row, tc.col, tc.size, "f2l"); got != tc.want {
				t.Errorf("shouldHighlight(%d, %d, %d, %d, \"f2l\") = %v, want %v",
					tc.face, tc.row, tc.col, tc.size, got, tc.want)
			}
		})
	}

	t.Run("every down face cell highlighted", func(t *testing.T) {
		for r := 0; r < 3; r++ {
			for c := 0; c < 3; c++ {
				if got := shouldHighlight(5, r, c, 3, "f2l"); !got {
					t.Errorf("shouldHighlight(5, %d, %d, 3, \"f2l\") = false, want true", r, c)
				}
			}
		}
	})

	// The row >= size/3 boundary must be exact under integer division: for a
	// size-3 cube the threshold is 1, for a size-6 cube it is 2.
	t.Run("boundary exact at size 3 and size 6", func(t *testing.T) {
		for _, size := range []int{3, 6} {
			threshold := size / 3
			for r := 0; r < size; r++ {
				want := r >= threshold
				if got := shouldHighlight(0, r, 0, size, "f2l"); got != want {
					t.Errorf("shouldHighlight(0, %d, 0, %d, \"f2l\") = %v, want %v (threshold size/3 = %d)",
						r, size, got, want, threshold)
				}
			}
		}
	})
}

func TestShouldHighlightUnknownAndEmptyModes(t *testing.T) {
	modes := []string{"bogus", "", "zz_top"}
	inputs := []struct {
		face     int
		row, col int
		size     int
	}{
		{0, 0, 0, 3}, // Front
		{4, 0, 0, 3}, // Up
		{5, 2, 2, 3}, // Down
		{2, 1, 1, 3}, // Left
		{1, 2, 1, 3}, // Back
	}
	for _, mode := range modes {
		name := mode
		if mode == "" {
			name = "(empty)"
		}
		t.Run("mode_"+name, func(t *testing.T) {
			for _, in := range inputs {
				if got := shouldHighlight(in.face, in.row, in.col, in.size, mode); got {
					t.Errorf("shouldHighlight(%d, %d, %d, %d, %q) = true, want false",
						in.face, in.row, in.col, in.size, mode)
				}
			}
		})
	}
}

// TestShouldHighlightUnmentionedFacesForMode asserts that a face the switch
// never mentions for a given mode stays false for cells that match none of
// that mode's conditions.
func TestShouldHighlightUnmentionedFacesForMode(t *testing.T) {
	t.Run("cross up face never highlighted", func(t *testing.T) {
		for r := 0; r < 3; r++ {
			for c := 0; c < 3; c++ {
				if got := shouldHighlight(4, r, c, 3, "cross"); got {
					t.Errorf("shouldHighlight(4, %d, %d, 3, \"cross\") = true, want false", r, c)
				}
			}
		}
	})
	t.Run("cross back face outside adjacent-edge rule", func(t *testing.T) {
		if got := shouldHighlight(1, 0, 1, 3, "cross"); got {
			t.Errorf("shouldHighlight(1, 0, 1, 3, \"cross\") = true, want false")
		}
	})
	t.Run("oll down face never highlighted", func(t *testing.T) {
		for r := 0; r < 3; r++ {
			for c := 0; c < 3; c++ {
				if got := shouldHighlight(5, r, c, 3, "oll"); got {
					t.Errorf("shouldHighlight(5, %d, %d, 3, \"oll\") = true, want false", r, c)
				}
			}
		}
	})
	t.Run("pll down face never highlighted", func(t *testing.T) {
		if got := shouldHighlight(5, 2, 0, 3, "pll"); got {
			t.Errorf("shouldHighlight(5, 2, 0, 3, \"pll\") = true, want false")
		}
	})
	t.Run("f2l up face never highlighted", func(t *testing.T) {
		for r := 0; r < 3; r++ {
			for c := 0; c < 3; c++ {
				if got := shouldHighlight(4, r, c, 3, "f2l"); got {
					t.Errorf("shouldHighlight(4, %d, %d, 3, \"f2l\") = true, want false", r, c)
				}
			}
		}
	})
}
