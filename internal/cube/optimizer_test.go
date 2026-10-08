package cube

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestOptimizePreservesFullNotation(t *testing.T) {
	for _, text := range []string{"x F", "R x R", "3Rw 3Rw", "Rw 3Rw"} {
		t.Run(text, func(t *testing.T) {
			assertOptimizationPreservesState(t, 4, text)
		})
	}
	rng := rand.New(rand.NewSource(20261007))
	for _, size := range []int{3, 4, 5} {
		var tokens []string
		for _, face := range []string{"R", "L", "U", "D", "F", "B"} {
			tokens = append(tokens, face, face+"w")
			for depth := 2; depth <= size; depth++ {
				tokens = append(tokens, fmt.Sprintf("%d%s", depth, face), fmt.Sprintf("%d%sw", depth, face))
			}
		}
		tokens = append(tokens, "x", "y", "z")
		// Even cubes have no single middle slice; those moves are rejected.
		if size%2 == 1 {
			tokens = append(tokens, "M", "E", "S")
		}
		for _, rotationHeavy := range []bool{false, true} {
			t.Run(fmt.Sprintf("size=%d/rotation-heavy=%t", size, rotationHeavy), func(t *testing.T) {
				for trial := 0; trial < 300; trial++ {
					var moves []Move
					for i := 0; i < 80; i++ {
						token := tokens[rng.Intn(len(tokens))]
						if rotationHeavy && rng.Intn(4) != 0 {
							token = []string{"x", "y", "z"}[rng.Intn(3)]
						}
						move, err := ParseMove(token + []string{"", "'", "2"}[rng.Intn(3)])
						if err != nil {
							t.Fatal(err)
						}
						moves = append(moves, move)
						if rng.Intn(3) == 0 {
							moves = append(moves, move)
						}
						if rotationHeavy && move.Rotation != NoRotation {
							// Mix directions and half turns in runs on the same axis.
							for run := rng.Intn(4); run > 0; run-- {
								move.Clockwise = rng.Intn(2) == 0
								move.Double = rng.Intn(3) == 0
								moves = append(moves, move)
							}
						}
					}
					assertOptimizationPreservesState(t, size, FormatMoves(moves))
				}
			})
		}
	}
}

func TestOptimizeRotationPairs(t *testing.T) {
	for _, axis := range []string{"x", "y", "z"} {
		suffixes := []string{"", "'", "2"}
		expected := [3][3]string{
			{axis + "2", "", axis + "'"},
			{"", axis + "2", axis},
			{axis + "'", axis, ""},
		}
		for i, first := range suffixes {
			for j, second := range suffixes {
				input := axis + first + " " + axis + second
				t.Run(input, func(t *testing.T) {
					result, err := OptimizeScramble(input)
					if err != nil {
						t.Fatal(err)
					}
					if result != expected[i][j] {
						t.Fatalf("OptimizeScramble(%q) = %q, want %q", input, result, expected[i][j])
					}
					for _, size := range []int{3, 4, 5} {
						assertOptimizationPreservesState(t, size, input)
					}
				})
			}
		}
	}
}

func assertOptimizationPreservesState(t *testing.T, size int, text string) {
	t.Helper()
	moves, err := ParseMoves(text)
	if err != nil {
		t.Fatal(err)
	}
	optimizedText, err := OptimizeScramble(text)
	if err != nil {
		t.Fatal(err)
	}
	optimized, err := ParseMoves(optimizedText)
	if err != nil {
		t.Fatal(err)
	}
	if len(optimized) > len(moves) {
		t.Fatalf("optimization grew %q to %q", text, optimizedText)
	}
	// Unique sticker labels also detect changes hidden by uniform face colors.
	original := NewCube(size)
	for face := range original.Faces {
		for row := range original.Faces[face] {
			for col := range original.Faces[face][row] {
				original.Faces[face][row][col] = Color(face*size*size + row*size + col)
			}
		}
	}
	result := original.clone()
	if err := original.ApplyMoves(moves); err != nil {
		t.Fatal(err)
	}
	if err := result.ApplyMoves(optimized); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(original.Faces, result.Faces) {
		t.Fatalf("%dx%d optimization changed state: %q -> %q", size, size, text, optimizedText)
	}
}

func TestOptimizeMoves(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple doubling - R R",
			input:    "R R",
			expected: "R2",
		},
		{
			name:     "Triple move - R R R",
			input:    "R R R",
			expected: "R'",
		},
		{
			name:     "Quadruple move - R R R R",
			input:    "R R R R",
			expected: "",
		},
		{
			name:     "Canceling moves - R R'",
			input:    "R R'",
			expected: "",
		},
		{
			name:     "Canceling moves reverse - R' R",
			input:    "R' R",
			expected: "",
		},
		{
			name:     "Double move canceling - R2 R2",
			input:    "R2 R2",
			expected: "",
		},
		{
			name:     "Double plus single - R2 R",
			input:    "R2 R",
			expected: "R'",
		},
		{
			name:     "Double plus counter - R2 R'",
			input:    "R2 R'",
			expected: "R",
		},
		{
			name:     "No optimization possible",
			input:    "R U R' U'",
			expected: "R U R' U'",
		},
		{
			name:     "Mixed optimization",
			input:    "R R U U' F F F",
			expected: "R2 F'",
		},
		{
			name:     "Adjacent same-face only",
			input:    "R U R R U' F F'",
			expected: "R U R2 U'",
		},
		{
			name:     "Wide moves",
			input:    "Rw Rw",
			expected: "Rw2",
		},
		{
			name:     "Layer moves",
			input:    "2R 2R 2R",
			expected: "2R'",
		},
		{
			name:     "Rotation doubling",
			input:    "x x",
			expected: "x2",
		},
		{
			name:     "Rotation cancellation",
			input:    "x x'",
			expected: "",
		},
		{
			name:     "Y rotation cancellation",
			input:    "y y'",
			expected: "",
		},
		{
			name:     "Double rotation plus single",
			input:    "x2 x",
			expected: "x'",
		},
		{
			name:     "Four rotations cancel",
			input:    "z z z z",
			expected: "",
		},
		{
			name:     "Documented rotation and wide turns",
			input:    "x x 3Rw 3Rw",
			expected: "x2 3Rw2",
		},
		{
			name:     "Rotations inside face turns",
			input:    "R x x R'",
			expected: "R x2 R'",
		},
		{
			name:     "Face turns exposed by canceled rotations",
			input:    "F x x' F'",
			expected: "",
		},
		{
			name:     "Rotations exposed by canceled face turns",
			input:    "x F F' x",
			expected: "x2",
		},
		{
			name:     "Nested cancellations",
			input:    "F x y R R' y' x' F'",
			expected: "",
		},
		{
			name:     "Different axes preserve order",
			input:    "x y x'",
			expected: "x y x'",
		},
		{
			name:     "Noncommuting face preserves rotation order",
			input:    "x F x'",
			expected: "x F x'",
		},
		{
			name:     "Rotation separates face turns",
			input:    "F x F'",
			expected: "F x F'",
		},
		{
			name:     "Wide turn separates rotations",
			input:    "x Uw x'",
			expected: "x Uw x'",
		},
		{
			name:     "Layer turn separates rotations",
			input:    "x 2U x'",
			expected: "x 2U x'",
		},
		{
			name:     "Slice separates rotations",
			input:    "x E x'",
			expected: "x E x'",
		},
		{
			name:     "Empty sequence",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := OptimizeScramble(tc.input)
			if err != nil {
				t.Fatalf("Error optimizing %s: %v", tc.input, err)
			}

			if result != tc.expected {
				t.Errorf("OptimizeScramble(%s) = %s, expected %s", tc.input, result, tc.expected)
			}
		})
	}
}

func TestMoveToQuarterTurns(t *testing.T) {
	testCases := []struct {
		move     Move
		expected int
	}{
		{Move{Face: Right, Clockwise: true, Double: false}, 1},
		{Move{Face: Right, Clockwise: true, Double: true}, 2},
		{Move{Face: Right, Clockwise: false, Double: false}, 3},
	}

	for _, tc := range testCases {
		result := moveToQuarterTurns(tc.move)
		if result != tc.expected {
			t.Errorf("moveToQuarterTurns(%v) = %d, expected %d", tc.move, result, tc.expected)
		}
	}
}

func TestQuarterTurnsToMove(t *testing.T) {
	testCases := []struct {
		quarterTurns int
		expected     Move
	}{
		{1, Move{Face: Right, Clockwise: true, Double: false}},
		{2, Move{Face: Right, Clockwise: true, Double: true}},
		{3, Move{Face: Right, Clockwise: false, Double: false}},
	}

	for _, tc := range testCases {
		result := quarterTurnsToMove(Right, false, 0, tc.quarterTurns)
		if result == nil {
			t.Errorf("quarterTurnsToMove(%d) returned nil", tc.quarterTurns)
			continue
		}

		if result.Face != tc.expected.Face ||
			result.Clockwise != tc.expected.Clockwise ||
			result.Double != tc.expected.Double {
			t.Errorf("quarterTurnsToMove(%d) = %v, expected %v", tc.quarterTurns, *result, tc.expected)
		}
	}
}

func TestIsCancellingSequence(t *testing.T) {
	testCases := []struct {
		name     string
		sequence string
		expected bool
	}{
		{"Canceling pair", "R R'", true},
		{"Canceling quadruple", "R R R R", true},
		{"Double canceling", "R2 R2", true},
		{"Non-canceling", "R U R' U'", false},
		{"Empty sequence", "", true},
		{"Single move", "R", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			moves, err := ParseScramble(tc.sequence)
			if err != nil {
				t.Fatalf("Error parsing %s: %v", tc.sequence, err)
			}

			result := IsCancellingSequence(moves)
			if result != tc.expected {
				t.Errorf("IsCancellingSequence(%s) = %v, expected %v", tc.sequence, result, tc.expected)
			}
		})
	}
}

func TestGetMoveCount(t *testing.T) {
	testCases := []struct {
		name     string
		sequence string
		expected int
	}{
		{"Simple optimization", "R R", 1},
		{"Complete cancellation", "R R'", 0},
		{"No optimization", "R U", 2},
		{"Mixed sequence", "R R U U'", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			moves, err := ParseScramble(tc.sequence)
			if err != nil {
				t.Fatalf("Error parsing %s: %v", tc.sequence, err)
			}

			result := GetMoveCount(moves)
			if result != tc.expected {
				t.Errorf("GetMoveCount(%s) = %d, expected %d", tc.sequence, result, tc.expected)
			}
		})
	}
}
