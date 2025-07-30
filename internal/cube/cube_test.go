package cube

import (
	"testing"
)

func TestNewCube(t *testing.T) {
	tests := []struct {
		name string
		size int
		want int
	}{
		{"2x2x2 cube", 2, 2},
		{"3x3x3 cube", 3, 3},
		{"4x4x4 cube", 4, 4},
		{"5x5x5 cube", 5, 5},
		{"Invalid size should default to 2", 1, 2},
		{"Invalid size should default to 2", 0, 2},
		{"Invalid size should default to 2", -1, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cube := NewCube(tt.size)
			if cube.Size != tt.want {
				t.Errorf("NewCube(%d).Size = %d, want %d", tt.size, cube.Size, tt.want)
			}

			// Verify cube is solved initially
			if !cube.IsSolved() {
				t.Errorf("NewCube(%d) should be solved initially", tt.size)
			}
		})
	}
}

func TestCubeIsSolved(t *testing.T) {
	// Test solved 3x3x3 cube
	cube := NewCube(3)
	if !cube.IsSolved() {
		t.Error("New 3x3x3 cube should be solved")
	}

	// Apply a move and verify it's no longer solved
	move := Move{Face: Right, Clockwise: true}
	cube.ApplyMove(move)
	if cube.IsSolved() {
		t.Error("Cube should not be solved after applying move R")
	}
}

func TestParseMove(t *testing.T) {
	tests := []struct {
		notation string
		want     Move
		wantErr  bool
	}{
		{"R", Move{Face: Right, Clockwise: true}, false},
		{"R'", Move{Face: Right, Clockwise: false}, false},
		{"R2", Move{Face: Right, Clockwise: true, Double: true}, false},
		{"U", Move{Face: Up, Clockwise: true}, false},
		{"U'", Move{Face: Up, Clockwise: false}, false},
		{"U2", Move{Face: Up, Clockwise: true, Double: true}, false},
		{"F", Move{Face: Front, Clockwise: true}, false},
		{"B", Move{Face: Back, Clockwise: true}, false},
		{"L", Move{Face: Left, Clockwise: true}, false},
		{"D", Move{Face: Down, Clockwise: true}, false},
		{"", Move{}, true},   // Empty notation should error
		{"X", Move{}, true},  // Invalid face should error
		{"R3", Move{}, true}, // Invalid modifier should error
	}

	for _, tt := range tests {
		t.Run(tt.notation, func(t *testing.T) {
			got, err := ParseMove(tt.notation)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMove(%q) error = %v, wantErr %v", tt.notation, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseMove(%q) = %v, want %v", tt.notation, got, tt.want)
			}
		})
	}
}

func TestParseScramble(t *testing.T) {
	tests := []struct {
		scramble string
		wantLen  int
		wantErr  bool
	}{
		{"", 0, false}, // Empty scramble is valid
		{"R", 1, false},
		{"R U R' U'", 4, false},
		{"R U R' U' R' F R F'", 8, false},
		{"R X", 0, true}, // Invalid move should error
		{"R U2 R' D'", 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.scramble, func(t *testing.T) {
			got, err := ParseScramble(tt.scramble)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseScramble(%q) error = %v, wantErr %v", tt.scramble, err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("ParseScramble(%q) length = %d, want %d", tt.scramble, len(got), tt.wantLen)
			}
		})
	}
}

// Test that moves actually change cube state
func TestMovesChangeState(t *testing.T) {
	cube := NewCube(3)
	originalState := cube.String()

	// Apply R move
	rMove := Move{Face: Right, Clockwise: true}
	cube.ApplyMove(rMove)

	afterRMove := cube.String()
	if originalState == afterRMove {
		t.Error("R move should change cube state")
	}

	// Apply U move
	uMove := Move{Face: Up, Clockwise: true}
	cube.ApplyMove(uMove)

	afterUMove := cube.String()
	if afterRMove == afterUMove {
		t.Error("U move should change cube state")
	}
}

// Test that R U R' U' actually scrambles the cube
func TestRURPrimeUPrimeScramble(t *testing.T) {
	cube := NewCube(3)
	originalState := cube.String()

	// Apply the famous R U R' U' sequence
	moves, err := ParseScramble("R U R' U'")
	if err != nil {
		t.Fatalf("Failed to parse R U R' U': %v", err)
	}

	cube.ApplyMoves(moves)
	scrambledState := cube.String()

	if originalState == scrambledState {
		t.Error("R U R' U' should scramble the cube - state should be different from solved")
	}

	// Verify cube is not solved
	if cube.IsSolved() {
		t.Error("Cube should not be solved after R U R' U' scramble")
	}
}

// Test double moves (R2, U2, etc.)
func TestDoubleMoves(t *testing.T) {
	cube1 := NewCube(3)
	cube2 := NewCube(3)

	// Apply R2 to cube1
	r2Move := Move{Face: Right, Clockwise: true, Double: true}
	cube1.ApplyMove(r2Move)

	// Apply R R to cube2
	rMove := Move{Face: Right, Clockwise: true}
	cube2.ApplyMove(rMove)
	cube2.ApplyMove(rMove)

	// States should be identical
	if cube1.String() != cube2.String() {
		t.Error("R2 should be equivalent to R R")
	}
}

// Test inverse moves (R R' should return to original state)
func TestInverseMoves(t *testing.T) {
	cube := NewCube(3)
	originalState := cube.String()

	// Apply R then R'
	rMove := Move{Face: Right, Clockwise: true}
	rPrimeMove := Move{Face: Right, Clockwise: false}

	cube.ApplyMove(rMove)
	cube.ApplyMove(rPrimeMove)

	finalState := cube.String()
	if originalState != finalState {
		t.Error("R R' should return cube to original state")
	}

	if !cube.IsSolved() {
		t.Error("Cube should be solved after R R'")
	}
}

// Test all faces can be rotated
func TestAllFacesRotate(t *testing.T) {
	faces := []Face{Front, Back, Left, Right, Up, Down}
	faceNames := []string{"Front", "Back", "Left", "Right", "Up", "Down"}

	for i, face := range faces {
		t.Run(faceNames[i], func(t *testing.T) {
			cube := NewCube(3)
			originalState := cube.String()

			move := Move{Face: face, Clockwise: true}
			cube.ApplyMove(move)

			newState := cube.String()
			if originalState == newState {
				t.Errorf("%s face rotation should change cube state", faceNames[i])
			}
		})
	}
}
