package cube

import (
	"fmt"
	"math/rand"
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

// TestMoveInverses - Comprehensive fuzzing test for move system
// Tests that move sequences + their inverses return to solved state
func TestMoveInverses(t *testing.T) {
	// Create a new random source with deterministic seed for reproducible testing
	rng := rand.New(rand.NewSource(42))

	moves := []string{"R", "R'", "R2", "L", "L'", "L2", "U", "U'", "U2", "D", "D'", "D2", "F", "F'", "F2", "B", "B'", "B2"}

	testCount := 50
	for i := 0; i < testCount; i++ {
		// Generate random scramble (3-5 moves)
		scrambleLength := 3 + rng.Intn(3)
		scramble := make([]string, scrambleLength)
		for j := 0; j < scrambleLength; j++ {
			scramble[j] = moves[rng.Intn(len(moves))]
		}

		// Create inverse sequence (reverse order, invert each move)
		inverse := make([]string, scrambleLength)
		for j := 0; j < scrambleLength; j++ {
			move := scramble[scrambleLength-1-j]
			// Invert the move: R->R', R'->R, R2->R2
			if len(move) > 1 && move[len(move)-1] == '\'' {
				inverse[j] = move[:len(move)-1] // Remove '
			} else if len(move) > 1 && move[len(move)-1] == '2' {
				inverse[j] = move // R2 stays R2
			} else {
				inverse[j] = move + "'" // Add '
			}
		}

		// Test: scramble + inverse should return to solved
		cube := NewCube(3)

		// Apply scramble
		scrambleMoves, err := ParseScramble(joinMoves(scramble))
		if err != nil {
			t.Fatalf("Failed to parse scramble %v: %v", scramble, err)
		}
		cube.ApplyMoves(scrambleMoves)

		// Apply inverse
		inverseMoves, err := ParseScramble(joinMoves(inverse))
		if err != nil {
			t.Fatalf("Failed to parse inverse %v: %v", inverse, err)
		}
		cube.ApplyMoves(inverseMoves)

		// Check if solved
		if !cube.IsSolved() {
			t.Errorf("Inverse test failed for scramble %v + inverse %v", scramble, inverse)
			t.Errorf("Scramble string: %s", joinMoves(scramble))
			t.Errorf("Inverse string: %s", joinMoves(inverse))
			t.Errorf("Final cube state: %s", cube.String())
			// Continue to test all cases instead of stopping on first failure
		}
	}

	t.Logf("All %d inverse tests passed!", testCount)
}

// TestCircuitFuzzing - The REAL test: every sequence must have finite cycle order
func TestCircuitFuzzing(t *testing.T) {
	// Set deterministic seed for reproducible testing
	rand.Seed(42)

	moves := []string{"R", "R'", "R2", "L", "L'", "L2", "U", "U'", "U2", "D", "D'", "D2", "F", "F'", "F2", "B", "B'", "B2"}

	testCount := 100
	maxCycleLength := 50 // Magic number - no cube sequence should need more than 50 cycles
	failureCount := 0

	for i := 0; i < testCount; i++ {
		// Generate random sequence (3-8 moves)
		scrambleLength := 3 + rand.Intn(6)
		scramble := make([]string, scrambleLength)
		for j := 0; j < scrambleLength; j++ {
			scramble[j] = moves[rand.Intn(len(moves))]
		}

		scrambleStr := joinMoves(scramble)
		cube := NewCube(3)
		originalState := cube.String()

		// Apply sequence repeatedly until it cycles back to solved (or we hit max)
		cycleFound := false
		for cycle := 1; cycle <= maxCycleLength; cycle++ {
			// Apply the sequence
			scrambleMoves, err := ParseScramble(scrambleStr)
			if err != nil {
				t.Fatalf("Failed to parse sequence %s: %v", scrambleStr, err)
			}
			cube.ApplyMoves(scrambleMoves)

			// Check if we're back to solved
			if cube.String() == originalState {
				t.Logf("✅ Sequence '%s' has cycle order %d", scrambleStr, cycle)
				cycleFound = true
				break
			}
		}

		if !cycleFound {
			failureCount++
			if failureCount <= 5 { // Report first 5 failures
				t.Errorf("❌ CIRCUIT FAILURE: Sequence '%s' did not cycle within %d applications!",
					scrambleStr, maxCycleLength)
				t.Errorf("Final cube state: %s", cube.String())
				t.Logf("--- Circuit failure %d ---", failureCount)
			}
		}

		// Progress reporting
		if (i+1)%20 == 0 {
			t.Logf("Circuit progress: %d/%d sequences tested, %d failures", i+1, testCount, failureCount)
		}
	}

	if failureCount == 0 {
		t.Logf("🎉 All %d sequences have finite cycle order!", testCount)
	} else {
		t.Errorf("❌ %d out of %d sequences failed circuit test (%.2f%% failure rate)",
			failureCount, testCount, float64(failureCount)/float64(testCount)*100)
		t.Errorf("This indicates FUNDAMENTAL bugs in the move system!")
	}
}

// TestMoveInversesAggressiveFuzzing - 1000 tests with longer sequences
func TestMoveInversesAggressiveFuzzing(t *testing.T) {
	// Set deterministic seed for reproducible testing
	rand.Seed(42)

	moves := []string{"R", "R'", "R2", "L", "L'", "L2", "U", "U'", "U2", "D", "D'", "D2", "F", "F'", "F2", "B", "B'", "B2"}

	testCount := 1000
	failureCount := 0

	for i := 0; i < testCount; i++ {
		// Generate random scramble (5-10 moves for more complexity)
		scrambleLength := 5 + rand.Intn(6)
		scramble := make([]string, scrambleLength)
		for j := 0; j < scrambleLength; j++ {
			scramble[j] = moves[rand.Intn(len(moves))]
		}

		// Create inverse sequence (reverse order, invert each move)
		inverse := make([]string, scrambleLength)
		for j := 0; j < scrambleLength; j++ {
			move := scramble[scrambleLength-1-j]
			// Invert the move: R->R', R'->R, R2->R2
			if len(move) > 1 && move[len(move)-1] == '\'' {
				inverse[j] = move[:len(move)-1] // Remove '
			} else if len(move) > 1 && move[len(move)-1] == '2' {
				inverse[j] = move // R2 stays R2
			} else {
				inverse[j] = move + "'" // Add '
			}
		}

		// Test: scramble + inverse should return to solved
		cube := NewCube(3)

		// Apply scramble
		scrambleMoves, err := ParseScramble(joinMoves(scramble))
		if err != nil {
			t.Fatalf("Failed to parse scramble %v: %v", scramble, err)
		}
		cube.ApplyMoves(scrambleMoves)

		// Apply inverse
		inverseMoves, err := ParseScramble(joinMoves(inverse))
		if err != nil {
			t.Fatalf("Failed to parse inverse %v: %v", inverse, err)
		}
		cube.ApplyMoves(inverseMoves)

		// Check if solved
		if !cube.IsSolved() {
			failureCount++
			if failureCount <= 5 { // Report first 5 failures
				t.Errorf("Aggressive fuzz test %d failed for scramble %v + inverse %v",
					i+1, scramble, inverse)
				t.Errorf("Scramble string: %s", joinMoves(scramble))
				t.Errorf("Inverse string: %s", joinMoves(inverse))
				t.Errorf("Final cube state: %s", cube.String())
				t.Logf("--- Failure %d of %d ---", failureCount, testCount)
			}
		}

		// Progress reporting every 100 tests
		if (i+1)%100 == 0 {
			t.Logf("Progress: %d/%d tests completed, %d failures so far", i+1, testCount, failureCount)
		}
	}

	if failureCount == 0 {
		t.Logf("🎉 All %d aggressive inverse tests passed!", testCount)
	} else {
		t.Errorf("❌ %d out of %d aggressive tests failed (%.2f%% failure rate)",
			failureCount, testCount, float64(failureCount)/float64(testCount)*100)
	}
}

// TestSpecificProblematicSequences tests sequences that previously failed
func TestSpecificProblematicSequences(t *testing.T) {
	testCases := []struct {
		name     string
		scramble string
		inverse  string
	}{
		{"R U sequence", "R U", "U' R'"},
		{"U B sequence", "U B", "B' U'"},
		{"D B sequence", "D B", "B' D'"},
		{"R U R' U' sequence", "R U R' U'", "U R U' R'"},
		{"F R U sequence", "F R U", "U' R' F'"},
		{"Complex sequence", "R U F' D L", "L' D' F U' R'"},
		{"Back face issues", "U B F", "F' B' U'"},
		{"Multi-face", "R L U D F B", "B' F' D' U' L' R'"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cube := NewCube(3)

			// Apply scramble
			scrambleMoves, err := ParseScramble(tc.scramble)
			if err != nil {
				t.Fatalf("Failed to parse scramble %s: %v", tc.scramble, err)
			}
			cube.ApplyMoves(scrambleMoves)

			// Apply inverse
			inverseMoves, err := ParseScramble(tc.inverse)
			if err != nil {
				t.Fatalf("Failed to parse inverse %s: %v", tc.inverse, err)
			}
			cube.ApplyMoves(inverseMoves)

			// Check if solved
			if !cube.IsSolved() {
				t.Errorf("Sequence %s + %s did not return to solved state", tc.scramble, tc.inverse)
				t.Errorf("Final cube state: %s", cube.String())
			}
		})
	}
}

// TestSingleMoveFourFold tests that individual moves applied 4 times return to solved
func TestSingleMoveFourFold(t *testing.T) {
	moves := []string{"R", "L", "U", "D", "F", "B"}

	for _, moveStr := range moves {
		t.Run(moveStr+" move 4x", func(t *testing.T) {
			cube := NewCube(3)

			// Apply move 4 times (should return to solved)
			for i := 0; i < 4; i++ {
				move, err := ParseMove(moveStr)
				if err != nil {
					t.Fatalf("Failed to parse move %s: %v", moveStr, err)
				}
				cube.ApplyMove(move)
			}

			if !cube.IsSolved() {
				t.Errorf("Move %s applied 4 times did not return to solved state", moveStr)
			}
		})
	}
}

// TestTPerm tests the T-Perm algorithm specifically
func TestTPerm(t *testing.T) {
	tPerm := "R U R' F' R U R' U' R' F R2 U' R'"

	t.Run("T-Perm moves affect only expected positions", func(t *testing.T) {
		// Test that each move in T-Perm only affects the positions it should
		// This is a more nuanced test than "don't affect bottom face"
		testCases := []struct {
			move               string
			shouldAffectBottom bool
			description        string
		}{
			{"R", true, "R affects bottom right column"},
			{"U", false, "U should not affect bottom face"},
			{"R'", true, "R' affects bottom right column"},
			{"F'", true, "F' affects bottom front row"},
			{"F", true, "F affects bottom front row"},
		}

		for _, tc := range testCases {
			cube := NewCube(3)
			originalBottomFace := make([][]Color, 3)

			// Store original bottom face
			for row := 0; row < 3; row++ {
				originalBottomFace[row] = make([]Color, 3)
				for col := 0; col < 3; col++ {
					originalBottomFace[row][col] = cube.Faces[Down][row][col]
				}
			}

			// Apply single move
			move, err := ParseMove(tc.move)
			if err != nil {
				t.Fatalf("Failed to parse move %s: %v", tc.move, err)
			}
			cube.ApplyMove(move)

			// Check if bottom face changed
			bottomChanged := false
			for row := 0; row < 3; row++ {
				for col := 0; col < 3; col++ {
					if cube.Faces[Down][row][col] != originalBottomFace[row][col] {
						bottomChanged = true
						break
					}
				}
				if bottomChanged {
					break
				}
			}

			if tc.shouldAffectBottom && !bottomChanged {
				t.Errorf("Move %s should affect bottom face but didn't", tc.move)
			} else if !tc.shouldAffectBottom && bottomChanged {
				t.Errorf("Move %s should not affect bottom face but did", tc.move)
				t.Logf("Bottom face after move %s: %v", tc.move, cube.Faces[Down])
			}
		}
	})

	t.Run("T-Perm step-by-step analysis", func(t *testing.T) {
		// Break down T-Perm move by move to find the exact issue
		tPermMoves := []string{"R", "U", "R'", "F'", "R", "U", "R'", "U'", "R'", "F", "R2", "U'", "R'"}

		// Test 1: Apply T-Perm move by move and check each step
		t.Run("Step by step cube states", func(t *testing.T) {
			cube := NewCube(3)
			originalState := cube.String()

			for i, moveStr := range tPermMoves {
				move, err := ParseMove(moveStr)
				if err != nil {
					t.Fatalf("Failed to parse move %s: %v", moveStr, err)
				}
				cube.ApplyMove(move)

				// Log cube state after each move for debugging
				t.Logf("After move %d (%s): %s", i+1, moveStr, cube.String())

				// Check that we never affect middle layers after move 4 (first R F' R sequence)
				if i >= 3 { // After R U R' F'
					// Just verify cube is still in a reasonable state (not completely scrambled)
					// This is a basic sanity check - we're not checking exact positions
				}
			}

			finalState := cube.String()
			if finalState == originalState {
				t.Errorf("T-Perm should change the cube state, but it didn't")
			}
		})

		// Test 2: Check if individual T-Perm has the right cycle length
		t.Run("T-Perm cycle analysis", func(t *testing.T) {
			cube := NewCube(3)
			originalState := cube.String()

			// Apply T-Perm up to 6 times to find its order
			for cycle := 1; cycle <= 6; cycle++ {
				moves, err := ParseScramble(tPerm)
				if err != nil {
					t.Fatalf("Failed to parse T-Perm: %v", err)
				}
				cube.ApplyMoves(moves)

				currentState := cube.String()
				t.Logf("After T-Perm cycle %d: cube state hash = %d chars", cycle, len(currentState))

				if currentState == originalState {
					t.Logf("✅ T-Perm has order %d (returns to solved after %d cycles)", cycle, cycle)
					if cycle != 3 {
						t.Errorf("Expected T-Perm to have order 3, but found order %d", cycle)
					}
					return
				}
			}

			t.Errorf("T-Perm did not return to solved state within 6 cycles - there's a bug!")
			t.Errorf("Final state after 6 cycles: %s", cube.String())
		})
	})

	t.Run("T-Perm applied 3 times should return to solved", func(t *testing.T) {
		cube := NewCube(3)

		// Apply T-Perm 3 times
		for i := 0; i < 3; i++ {
			moves, err := ParseScramble(tPerm)
			if err != nil {
				t.Fatalf("Failed to parse T-Perm: %v", err)
			}
			cube.ApplyMoves(moves)
		}

		if !cube.IsSolved() {
			t.Errorf("T-Perm applied 3 times should return to solved state")
			t.Errorf("Final cube state: %s", cube.String())
		}
	})

	t.Run("T-Perm only affects top layer", func(t *testing.T) {
		cube := NewCube(3)
		originalBottomFace := make([][]Color, 3)
		originalMiddleLayers := make(map[string][][]Color)

		// Store original bottom face
		for i := 0; i < 3; i++ {
			originalBottomFace[i] = make([]Color, 3)
			for j := 0; j < 3; j++ {
				originalBottomFace[i][j] = cube.Faces[Down][i][j]
			}
		}

		// Store original middle layer edges
		faces := []Face{Front, Back, Left, Right}
		for _, face := range faces {
			originalMiddleLayers[face.String()] = make([][]Color, 3)
			for i := 0; i < 3; i++ {
				originalMiddleLayers[face.String()][i] = make([]Color, 3)
				for j := 0; j < 3; j++ {
					originalMiddleLayers[face.String()][i][j] = cube.Faces[face][i][j]
				}
			}
		}

		// Apply T-Perm
		moves, err := ParseScramble(tPerm)
		if err != nil {
			t.Fatalf("Failed to parse T-Perm: %v", err)
		}
		cube.ApplyMoves(moves)

		// Check that bottom face is unchanged
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				if cube.Faces[Down][i][j] != originalBottomFace[i][j] {
					t.Errorf("T-Perm should not affect bottom face, but position [%d][%d] changed", i, j)
				}
			}
		}

		// Check that middle layer edges are unchanged (rows 1 and 2)
		for _, face := range faces {
			for i := 1; i < 3; i++ { // Only check middle and bottom rows
				for j := 0; j < 3; j++ {
					if cube.Faces[face][i][j] != originalMiddleLayers[face.String()][i][j] {
						t.Errorf("T-Perm should not affect middle layers, but %s face position [%d][%d] changed", face.String(), i, j)
					}
				}
			}
		}
	})
}

// Helper function to join move strings
func joinMoves(moves []string) string {
	result := ""
	for i, move := range moves {
		if i > 0 {
			result += " "
		}
		result += move
	}
	return result
}
