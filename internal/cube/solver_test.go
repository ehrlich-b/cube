package cube

import (
	"testing"
)

func TestGetSolver(t *testing.T) {
	tests := []struct {
		name      string
		algorithm string
		wantName  string
		wantErr   bool
	}{
		{"Beginner solver", "beginner", "Beginner", false},
		{"CFOP solver", "cfop", "CFOP", false},
		{"Kociemba solver", "kociemba", "Kociemba", false},
		{"Invalid solver", "invalid", "", true},
		{"Empty string", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			solver, err := GetSolver(tt.algorithm)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSolver(%q) error = %v, wantErr %v", tt.algorithm, err, tt.wantErr)
				return
			}
			if !tt.wantErr && solver.Name() != tt.wantName {
				t.Errorf("GetSolver(%q).Name() = %q, want %q", tt.algorithm, solver.Name(), tt.wantName)
			}
		})
	}
}

func TestBeginnerSolverOnSolvedCube(t *testing.T) {
	cube := NewCube(3)
	solver := &BeginnerSolver{}

	result, err := solver.Solve(cube)
	if err != nil {
		t.Fatalf("BeginnerSolver.Solve() error = %v", err)
	}

	// Solved cube should return empty solution
	if len(result.Solution) != 0 {
		t.Errorf("BeginnerSolver on solved cube should return empty solution, got %d moves", len(result.Solution))
	}

	if result.Steps != 0 {
		t.Errorf("BeginnerSolver on solved cube should return 0 steps, got %d", result.Steps)
	}
}

func TestBeginnerSolverOnScrambledCube(t *testing.T) {
	cube := NewCube(3)

	// Apply R U R' U' scramble
	moves, err := ParseScramble("R U R' U'")
	if err != nil {
		t.Fatalf("Failed to parse scramble: %v", err)
	}
	cube.ApplyMoves(moves)

	solver := &BeginnerSolver{}
	result, err := solver.Solve(cube)
	if err != nil {
		t.Fatalf("BeginnerSolver.Solve() error = %v", err)
	}

	// Scrambled cube should return non-empty solution
	if len(result.Solution) == 0 {
		t.Error("BeginnerSolver on scrambled cube should return non-empty solution")
	}

	if result.Steps != len(result.Solution) {
		t.Errorf("Steps (%d) should equal solution length (%d)", result.Steps, len(result.Solution))
	}

	// Duration should be measured
	if result.Duration <= 0 {
		t.Error("Duration should be positive")
	}
}

func TestSolverResultConsistency(t *testing.T) {
	solvers := []struct {
		name      string
		algorithm string
	}{
		{"beginner", "beginner"},
		{"cfop", "cfop"},
		{"kociemba", "kociemba"},
	}

	for _, solverTest := range solvers {
		t.Run(solverTest.name, func(t *testing.T) {
			cube := NewCube(3)

			// Apply scramble
			moves, err := ParseScramble("R U R' U'")
			if err != nil {
				t.Fatalf("Failed to parse scramble: %v", err)
			}
			cube.ApplyMoves(moves)

			solver, err := GetSolver(solverTest.algorithm)
			if err != nil {
				t.Fatalf("Failed to get solver %s: %v", solverTest.algorithm, err)
			}

			result, err := solver.Solve(cube)
			if err != nil {
				t.Fatalf("%s solver error: %v", solverTest.algorithm, err)
			}

			// Basic consistency checks
			if result.Steps != len(result.Solution) {
				t.Errorf("%s: Steps (%d) != Solution length (%d)", solverTest.algorithm, result.Steps, len(result.Solution))
			}

			if result.Duration < 0 {
				t.Errorf("%s: Duration should not be negative", solverTest.algorithm)
			}
		})
	}
}

func TestKociembaSolver4x4Rejection(t *testing.T) {
	cube := NewCube(4) // 4x4x4 cube
	solver := &KociembaSolver{}

	_, err := solver.Solve(cube)
	if err == nil {
		t.Error("KociembaSolver should reject 4x4x4 cubes")
	}
}

func TestCopyCubeFunction(t *testing.T) {
	original := NewCube(3)

	// Apply some moves to original
	moves, err := ParseScramble("R U R'")
	if err != nil {
		t.Fatalf("Failed to parse moves: %v", err)
	}
	original.ApplyMoves(moves)

	solver := &BeginnerSolver{}
	copy := solver.copyCube(original)

	// Copy should have same state initially
	if original.String() != copy.String() {
		t.Error("Copied cube should have same state as original")
	}

	// Modifying copy shouldn't affect original
	copyMove := Move{Face: Front, Clockwise: true}
	copy.ApplyMove(copyMove)

	if original.String() == copy.String() {
		t.Error("Modifying copied cube should not affect original")
	}
}

func TestCountSolvedPieces(t *testing.T) {
	solver := &BeginnerSolver{}

	// Solved cube should have all pieces in correct position
	solvedCube := NewCube(3)
	solvedCount := solver.countSolvedPieces(solvedCube)
	expectedSolved := 6 * 3 * 3 // 6 faces, 3x3 each

	if solvedCount != expectedSolved {
		t.Errorf("Solved cube should have %d pieces in correct position, got %d", expectedSolved, solvedCount)
	}

	// Scrambled cube should have fewer solved pieces
	scrambledCube := NewCube(3)
	moves, err := ParseScramble("R U R' U'")
	if err != nil {
		t.Fatalf("Failed to parse scramble: %v", err)
	}
	scrambledCube.ApplyMoves(moves)

	scrambledCount := solver.countSolvedPieces(scrambledCube)
	if scrambledCount >= solvedCount {
		t.Errorf("Scrambled cube should have fewer solved pieces than solved cube (%d >= %d)", scrambledCount, solvedCount)
	}
}
