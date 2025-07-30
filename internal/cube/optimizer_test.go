package cube

import (
	"testing"
)

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
