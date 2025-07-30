package cube

import (
	"testing"
)

// Test parsing of advanced move notation
func TestAdvancedMoveNotationParsing(t *testing.T) {
	testCases := []struct {
		notation string
		expected Move
	}{
		// Middle slice moves
		{"M", Move{Slice: M_Slice, Clockwise: true}},
		{"M'", Move{Slice: M_Slice, Clockwise: false}},
		{"M2", Move{Slice: M_Slice, Clockwise: true, Double: true}},
		{"E", Move{Slice: E_Slice, Clockwise: true}},
		{"E'", Move{Slice: E_Slice, Clockwise: false}},
		{"S", Move{Slice: S_Slice, Clockwise: true}},
		{"S2", Move{Slice: S_Slice, Clockwise: true, Double: true}},

		// Wide moves
		{"Rw", Move{Face: Right, Wide: true, Clockwise: true, WideDepth: 2}},
		{"Rw'", Move{Face: Right, Wide: true, Clockwise: false, WideDepth: 2}},
		{"Fw2", Move{Face: Front, Wide: true, Clockwise: true, Double: true, WideDepth: 2}},
		{"Lw", Move{Face: Left, Wide: true, Clockwise: true, WideDepth: 2}},

		// Layer moves
		{"2R", Move{Face: Right, Layer: 1, Clockwise: true}}, // 2R = second layer from right
		{"3L", Move{Face: Left, Layer: 2, Clockwise: true}},  // 3L = third layer from left
		{"2R'", Move{Face: Right, Layer: 1, Clockwise: false}},
		{"3U2", Move{Face: Up, Layer: 2, Clockwise: true, Double: true}},

		// Cube rotations (case sensitive - lowercase only)
		{"x", Move{Rotation: X_Rotation, Clockwise: true}},
		{"x'", Move{Rotation: X_Rotation, Clockwise: false}},
		{"x2", Move{Rotation: X_Rotation, Clockwise: true, Double: true}},
		{"y", Move{Rotation: Y_Rotation, Clockwise: true}},
		{"y'", Move{Rotation: Y_Rotation, Clockwise: false}},
		{"z", Move{Rotation: Z_Rotation, Clockwise: true}},
		{"z2", Move{Rotation: Z_Rotation, Clockwise: true, Double: true}},
	}

	for _, tc := range testCases {
		t.Run(tc.notation, func(t *testing.T) {
			move, err := ParseMove(tc.notation)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", tc.notation, err)
			}

			// Compare relevant fields
			if move.Face != tc.expected.Face {
				t.Errorf("Face mismatch for %s: got %v, expected %v", tc.notation, move.Face, tc.expected.Face)
			}
			if move.Clockwise != tc.expected.Clockwise {
				t.Errorf("Clockwise mismatch for %s: got %v, expected %v", tc.notation, move.Clockwise, tc.expected.Clockwise)
			}
			if move.Double != tc.expected.Double {
				t.Errorf("Double mismatch for %s: got %v, expected %v", tc.notation, move.Double, tc.expected.Double)
			}
			if move.Wide != tc.expected.Wide {
				t.Errorf("Wide mismatch for %s: got %v, expected %v", tc.notation, move.Wide, tc.expected.Wide)
			}
			if move.Layer != tc.expected.Layer {
				t.Errorf("Layer mismatch for %s: got %v, expected %v", tc.notation, move.Layer, tc.expected.Layer)
			}
			if move.Slice != tc.expected.Slice {
				t.Errorf("Slice mismatch for %s: got %v, expected %v", tc.notation, move.Slice, tc.expected.Slice)
			}
			if move.Rotation != tc.expected.Rotation {
				t.Errorf("Rotation mismatch for %s: got %v, expected %v", tc.notation, move.Rotation, tc.expected.Rotation)
			}
		})
	}
}

// Test string representation of advanced moves
func TestAdvancedMoveStringification(t *testing.T) {
	testCases := []struct {
		move     Move
		expected string
	}{
		// Middle slice moves
		{Move{Slice: M_Slice, Clockwise: true}, "M"},
		{Move{Slice: M_Slice, Clockwise: false}, "M'"},
		{Move{Slice: E_Slice, Clockwise: true, Double: true}, "E2"},

		// Wide moves
		{Move{Face: Right, Wide: true, Clockwise: true}, "Rw"},
		{Move{Face: Front, Wide: true, Clockwise: false}, "Fw'"},
		{Move{Face: Left, Wide: true, Clockwise: true, Double: true}, "Lw2"},

		// Layer moves
		{Move{Face: Right, Layer: 1, Clockwise: true}, "2R"},
		{Move{Face: Left, Layer: 2, Clockwise: false}, "3L'"},
		{Move{Face: Up, Layer: 3, Clockwise: true, Double: true}, "4U2"},

		// Cube rotations
		{Move{Rotation: X_Rotation, Clockwise: true}, "x"},
		{Move{Rotation: Y_Rotation, Clockwise: false}, "y'"},
		{Move{Rotation: Z_Rotation, Clockwise: true, Double: true}, "z2"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.move.String()
			if result != tc.expected {
				t.Errorf("String mismatch: got %s, expected %s", result, tc.expected)
			}
		})
	}
}

// Test middle slice moves on 3x3 cube
func TestMiddleSliceMoves3x3(t *testing.T) {
	cube := NewCube(3)
	originalState := cube.String()

	// Test M move
	mMove := Move{Slice: M_Slice, Clockwise: true}
	cube.ApplyMove(mMove)

	if cube.String() == originalState {
		t.Error("M move should change cube state")
	}

	// Check that middle column of front face changed
	if cube.Faces[Front][0][1] == White || cube.Faces[Front][1][1] == White || cube.Faces[Front][2][1] == White {
		// At least one should have changed from white
		t.Error("M move should affect middle column of front face")
	}

	// Test E move on fresh cube
	cube = NewCube(3)
	eMove := Move{Slice: E_Slice, Clockwise: true}
	cube.ApplyMove(eMove)

	// Check that middle row changed
	if cube.Faces[Front][1][0] == White && cube.Faces[Front][1][1] == White && cube.Faces[Front][1][2] == White {
		t.Error("E move should affect middle row of front face")
	}
}

// Test wide moves on 4x4 cube
func TestWideMoves4x4(t *testing.T) {
	cube := NewCube(4)
	originalState := cube.String()

	// Test Rw move (should affect 2 rightmost layers)
	rwMove := Move{Face: Right, Wide: true, Clockwise: true, WideDepth: 2}
	cube.ApplyMove(rwMove)

	if cube.String() == originalState {
		t.Error("Rw move should change cube state")
	}

	// Store original colors for comparison
	origFrontRight := [2]Color{Blue, Blue} // Front face is Blue in new orientation
	origFrontLeft := [2]Color{Blue, Blue}

	// Check that rightmost 2 columns of front face changed
	if cube.Faces[Front][0][2] == origFrontRight[0] || cube.Faces[Front][0][3] == origFrontRight[1] {
		// Should have changed from original
		t.Error("Rw move should affect rightmost 2 columns of front face")
	}

	// Check that leftmost 2 columns remained unchanged
	if cube.Faces[Front][0][0] != origFrontLeft[0] || cube.Faces[Front][0][1] != origFrontLeft[1] {
		t.Error("Rw move should NOT affect leftmost 2 columns of front face")
	}
}

// Test layer moves on 5x5 cube
func TestLayerMoves5x5(t *testing.T) {
	cube := NewCube(5)
	originalState := cube.String()

	// Test 2R move (should affect only second layer from right)
	layerMove := Move{Face: Right, Layer: 1, Clockwise: true}
	cube.ApplyMove(layerMove)

	if cube.String() == originalState {
		t.Error("2R move should change cube state")
	}

	// Check that only the second-from-right column changed (front face is Blue in new orientation)
	if cube.Faces[Front][0][3] == Blue {
		t.Error("2R move should affect column 3 (second from right) of front face")
	}

	// Check that other columns remained unchanged
	if cube.Faces[Front][0][0] != Blue || cube.Faces[Front][0][1] != Blue || cube.Faces[Front][0][2] != Blue || cube.Faces[Front][0][4] != Blue {
		t.Error("2R move should only affect column 3, leaving others unchanged")
	}
}

// Test cube rotations
func TestCubeRotations(t *testing.T) {
	cube := NewCube(3)
	originalFront := cube.Faces[Front][0][0]
	originalUp := cube.Faces[Up][0][0]
	originalBack := cube.Faces[Back][0][0]
	originalDown := cube.Faces[Down][0][0]

	// Test x rotation (around R axis)
	xMove := Move{Rotation: X_Rotation, Clockwise: true}
	cube.ApplyMove(xMove)

	// After x rotation: F→D, U→F, B→U, D→B
	if cube.Faces[Down][0][0] != originalFront {
		t.Error("After x rotation, Down face should contain original Front")
	}
	if cube.Faces[Front][0][0] != originalUp {
		t.Error("After x rotation, Front face should contain original Up")
	}
	if cube.Faces[Up][0][0] != originalBack {
		t.Error("After x rotation, Up face should contain original Back")
	}
	if cube.Faces[Back][0][0] != originalDown {
		t.Error("After x rotation, Back face should contain original Down")
	}
}

// Test complex sequences with advanced notation
func TestAdvancedNotationSequences(t *testing.T) {
	sequences := []string{
		"M E S",         // All slice moves
		"Rw Fw Uw",      // All wide moves
		"2R 3L 2F",      // All layer moves
		"x y z",         // All rotations
		"R M U Rw x",    // Mixed notation
		"2R' M2 Fw' x'", // With modifiers
	}

	for _, seq := range sequences {
		t.Run(seq, func(t *testing.T) {
			moves, err := ParseScramble(seq)
			if err != nil {
				t.Fatalf("Failed to parse sequence '%s': %v", seq, err)
			}

			// Apply to cube and verify changes
			cube := NewCube(3)
			if seq == "2R 3L 2F" { // Layer moves need bigger cube
				cube = NewCube(5)
			}
			originalState := cube.String()

			cube.ApplyMoves(moves)

			// Most sequences should change the cube state
			// Exception: some rotation combinations might return to solved
			if cube.String() == originalState && seq != "x y z" {
				t.Errorf("Sequence '%s' should change cube state", seq)
			}
		})
	}
}

// Test that slice moves are invalid on even cubes
func TestSliceMovesEvenCubes(t *testing.T) {
	cube := NewCube(4) // Even cube
	originalState := cube.String()

	// M move should have no effect on 4x4
	mMove := Move{Slice: M_Slice, Clockwise: true}
	cube.ApplyMove(mMove)

	if cube.String() != originalState {
		t.Error("M move should have no effect on even-sized (4x4) cube")
	}

	// E move should have no effect on 4x4
	eMove := Move{Slice: E_Slice, Clockwise: true}
	cube.ApplyMove(eMove)

	if cube.String() != originalState {
		t.Error("E move should have no effect on even-sized (4x4) cube")
	}
}

// Performance test for advanced moves
func BenchmarkAdvancedMoveParsing(b *testing.B) {
	notations := []string{"M", "Rw", "2R", "x", "M'", "Fw2", "3L'", "y2"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		notation := notations[i%len(notations)]
		_, err := ParseMove(notation)
		if err != nil {
			b.Fatalf("Failed to parse %s: %v", notation, err)
		}
	}
}

func BenchmarkAdvancedMoveApplication(b *testing.B) {
	cube := NewCube(4)
	moves := []Move{
		{Slice: M_Slice, Clockwise: true},
		{Face: Right, Wide: true, Clockwise: true},
		{Face: Right, Layer: 1, Clockwise: true},
		{Rotation: X_Rotation, Clockwise: true},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		move := moves[i%len(moves)]
		cube.ApplyMove(move)
	}
}
