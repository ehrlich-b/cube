package cube

import (
	"fmt"
	"testing"
)

// Test that demonstrates the fundamental limitation:
// Current move system only works properly for 3x3x3 cubes
func TestMoveSystemLimitations(t *testing.T) {
	t.Run("3x3x3 moves work correctly", func(t *testing.T) {
		cube := NewCube(3)

		// Apply R move and verify specific positions changed
		rMove := Move{Face: Right, Clockwise: true}
		cube.ApplyMove(rMove)

		// In a proper R move on 3x3, the Front face right column should contain pieces from Down face
		// This is a basic sanity check that our 3x3 implementation works
		if cube.IsSolved() {
			t.Error("3x3 cube should not be solved after R move")
		}
	})

	t.Run("4x4x4 moves work correctly with multiple layers", func(t *testing.T) {
		cube := NewCube(4)
		originalState := cube.String()

		// Store original colors at key positions
		origFrontRight := [2]Color{cube.Faces[Front][0][2], cube.Faces[Front][0][3]}
		origFrontLeft := [2]Color{cube.Faces[Front][0][0], cube.Faces[Front][0][1]}

		// Apply R move to 4x4 cube
		rMove := Move{Face: Right, Clockwise: true}
		cube.ApplyMove(rMove)

		newState := cube.String()

		// The move should change the state
		if originalState == newState {
			t.Error("4x4 cube state should change after R move")
		}

		// For 4x4 cubes, R move should affect the two rightmost layers
		// Check that the Front face's right two columns changed (don't care what color)
		if cube.Faces[Front][0][2] == origFrontRight[0] || cube.Faces[Front][0][3] == origFrontRight[1] {
			t.Error("4x4 R move should change front face right columns")
		}

		// Verify the move affects exactly 2 layers (half of 4)
		// The leftmost 2 columns should remain unchanged
		if cube.Faces[Front][0][0] != origFrontLeft[0] || cube.Faces[Front][0][1] != origFrontLeft[1] {
			t.Error("4x4 R move should NOT change front face left columns")
		}
	})

	t.Run("2x2x2 moves work (simpler case)", func(t *testing.T) {
		cube := NewCube(2)
		originalState := cube.String()

		// Store original color at key position
		origFrontRight := cube.Faces[Front][0][1]

		// Apply R move
		rMove := Move{Face: Right, Clockwise: true}
		cube.ApplyMove(rMove)

		// Should not be solved after move
		if cube.IsSolved() {
			t.Error("2x2 cube should not be solved after R move")
		}

		// Should be different from original
		if originalState == cube.String() {
			t.Error("2x2 cube state should change after R move")
		}

		// For 2x2, R move affects 1 layer (half of 2)
		// Check Front face right column changed (don't care what color)
		if cube.Faces[Front][0][1] == origFrontRight {
			t.Error("2x2 R move should change front face right column")
		}
	})

	t.Run("5x5x5 moves preserve center layer", func(t *testing.T) {
		cube := NewCube(5)

		// Store original colors at key positions
		origFrontRight := [2]Color{cube.Faces[Front][0][3], cube.Faces[Front][0][4]}
		origFrontCenter := cube.Faces[Front][0][2]
		origFrontLeft := [2]Color{cube.Faces[Front][0][0], cube.Faces[Front][0][1]}

		// Apply R move to 5x5 cube
		rMove := Move{Face: Right, Clockwise: true}
		cube.ApplyMove(rMove)

		// For 5x5, R move should affect outer 2 layers, preserve center
		// Check that rightmost 2 columns changed (don't care what color)
		if cube.Faces[Front][0][3] == origFrontRight[0] || cube.Faces[Front][0][4] == origFrontRight[1] {
			t.Error("5x5 R move should change front face rightmost 2 columns")
		}

		// Check that center column (index 2) is unchanged
		if cube.Faces[Front][0][2] != origFrontCenter {
			t.Error("5x5 R move should NOT change front face center column")
		}

		// Check that leftmost 2 columns are unchanged
		if cube.Faces[Front][0][0] != origFrontLeft[0] || cube.Faces[Front][0][1] != origFrontLeft[1] {
			t.Error("5x5 R move should NOT change front face leftmost 2 columns")
		}
	})

	t.Run("comprehensive multi-size move testing", func(t *testing.T) {
		testSizes := []int{2, 3, 4, 5, 6}

		for _, size := range testSizes {
			t.Run(fmt.Sprintf("%dx%dx%d", size, size, size), func(t *testing.T) {
				cube := NewCube(size)
				originalState := cube.String()

				// Store original colors at all positions in the front face top row
				origFrontColors := make([]Color, size)
				for col := 0; col < size; col++ {
					origFrontColors[col] = cube.Faces[Front][0][col]
				}

				// Apply R move
				rMove := Move{Face: Right, Clockwise: true}
				cube.ApplyMove(rMove)

				// Basic checks
				if cube.IsSolved() {
					t.Errorf("%dx%d cube should not be solved after R move", size, size)
				}

				if originalState == cube.String() {
					t.Errorf("%dx%d cube state should change after R move", size, size)
				}

				// Check that proper number of layers moved
				expectedLayers := size / 2

				// For the Front face, check that the rightmost 'expectedLayers' columns changed
				for layer := 0; layer < expectedLayers; layer++ {
					col := size - 1 - layer
					if cube.Faces[Front][0][col] == origFrontColors[col] {
						t.Errorf("%dx%d cube: column %d should have changed after R move", size, size, col)
					}
				}

				// Check that the remaining columns did NOT change (if any)
				unchangedLayers := size - expectedLayers
				for layer := 0; layer < unchangedLayers; layer++ {
					if cube.Faces[Front][0][layer] != origFrontColors[layer] {
						t.Errorf("%dx%d cube: column %d should NOT have changed after R move", size, size, layer)
					}
				}
			})
		}
	})
}

// Test that our move parsing works for all standard notation
func TestCompleteMoveNotation(t *testing.T) {
	notations := []string{
		"F", "F'", "F2",
		"B", "B'", "B2",
		"R", "R'", "R2",
		"L", "L'", "L2",
		"U", "U'", "U2",
		"D", "D'", "D2",
	}

	for _, notation := range notations {
		t.Run(notation, func(t *testing.T) {
			move, err := ParseMove(notation)
			if err != nil {
				t.Errorf("Failed to parse %s: %v", notation, err)
			}

			// Verify the move can be applied to a cube
			cube := NewCube(3)
			originalState := cube.String()
			cube.ApplyMove(move)
			newState := cube.String()

			if originalState == newState {
				t.Errorf("Move %s should change cube state", notation)
			}
		})
	}
}

// Test move sequences that should return to solved state
func TestMoveSequenceInverses(t *testing.T) {
	sequences := []struct {
		name     string
		sequence string
		inverse  string
	}{
		{"Single R move", "R", "R'"},
		{"Double move", "R2", "R2"},
		{"Simple sequence", "R U", "U' R'"},
		{"Sexy move", "R U R' U'", "U R U' R'"},
	}

	for _, seq := range sequences {
		t.Run(seq.name, func(t *testing.T) {
			cube := NewCube(3)
			originalState := cube.String()

			// Apply sequence
			moves, err := ParseScramble(seq.sequence)
			if err != nil {
				t.Fatalf("Failed to parse sequence %s: %v", seq.sequence, err)
			}
			cube.ApplyMoves(moves)

			// Apply inverse
			inverseMoves, err := ParseScramble(seq.inverse)
			if err != nil {
				t.Fatalf("Failed to parse inverse %s: %v", seq.inverse, err)
			}
			cube.ApplyMoves(inverseMoves)

			finalState := cube.String()
			if originalState != finalState {
				t.Errorf("Sequence %s followed by %s should return to original state", seq.sequence, seq.inverse)
			}
		})
	}
}

// Stress test: Apply many random moves and verify cube state tracking
func TestMoveSequenceConsistency(t *testing.T) {
	cube := NewCube(3)

	// Apply a long sequence of moves
	longSequence := "R U R' U' R' F R F' R U R' U' R' F R F' R U R' U'"
	moves, err := ParseScramble(longSequence)
	if err != nil {
		t.Fatalf("Failed to parse long sequence: %v", err)
	}

	// Track that each move changes the state
	for i, move := range moves {
		beforeState := cube.String()
		cube.ApplyMove(move)
		afterState := cube.String()

		if beforeState == afterState {
			t.Errorf("Move %d (%s) in sequence did not change cube state", i, move.String())
		}
	}

	// Cube should definitely not be solved after this sequence
	if cube.IsSolved() {
		t.Error("Cube should not be solved after long random sequence")
	}
}

// Test edge case: empty scramble
func TestEmptyScramble(t *testing.T) {
	cube := NewCube(3)
	originalState := cube.String()

	moves, err := ParseScramble("")
	if err != nil {
		t.Fatalf("Empty scramble should not error: %v", err)
	}

	cube.ApplyMoves(moves)
	finalState := cube.String()

	if originalState != finalState {
		t.Error("Empty scramble should not change cube state")
	}

	if !cube.IsSolved() {
		t.Error("Cube should remain solved after empty scramble")
	}
}

// Benchmark move application performance
func BenchmarkSingleMove(b *testing.B) {
	cube := NewCube(3)
	move := Move{Face: Right, Clockwise: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cube.ApplyMove(move)
	}
}

func BenchmarkScrambleApplication(b *testing.B) {
	moves, _ := ParseScramble("R U R' U' R' F R F'")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cube := NewCube(3)
		cube.ApplyMoves(moves)
	}
}
