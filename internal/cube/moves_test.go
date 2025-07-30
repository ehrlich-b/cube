package cube

import (
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

	t.Run("4x4x4 moves expose implementation problems", func(t *testing.T) {
		cube := NewCube(4)
		originalState := cube.String()

		// Apply R move to 4x4 cube
		rMove := Move{Face: Right, Clockwise: true}
		cube.ApplyMove(rMove)

		newState := cube.String()

		// The move should change the state (this will pass)
		if originalState == newState {
			t.Error("4x4 cube state should change after R move")
		}

		// But the inner layers are NOT handled correctly by our current implementation
		// Our rotateAdjacentEdges function assumes only edge rows/columns exist
		// For 4x4, there should be inner edge pieces that move differently

		// This test documents the limitation - we can't easily test the correctness
		// without implementing the proper 4x4 move system first
		t.Log("WARNING: 4x4x4 move implementation is incomplete - inner layers not handled properly")
	})

	t.Run("2x2x2 moves work (simpler case)", func(t *testing.T) {
		cube := NewCube(2)

		// 2x2 should work because it's simpler than 3x3 (no edges, only corners)
		rMove := Move{Face: Right, Clockwise: true}
		cube.ApplyMove(rMove)

		if cube.IsSolved() {
			t.Error("2x2 cube should not be solved after R move")
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
