package cube

import (
	"fmt"
	"time"
)

// SolverResult represents the result of a solve attempt
type SolverResult struct {
	Solution []Move
	Steps    int
	Duration time.Duration
}

// Solver interface for different solving algorithms
type Solver interface {
	Solve(cube *Cube) (*SolverResult, error)
	Name() string
}

// BeginnerSolver implements a basic layer-by-layer method
type BeginnerSolver struct{}

func (s *BeginnerSolver) Name() string {
	return "Beginner"
}

func (s *BeginnerSolver) Solve(cube *Cube) (*SolverResult, error) {
	start := time.Now()

	// Check if cube is already solved
	if cube.IsSolved() {
		return &SolverResult{
			Solution: []Move{},
			Steps:    0,
			Duration: time.Since(start),
		}, nil
	}

	// Simple solver implementation
	// This is not a real algorithm but demonstrates working infrastructure
	var solution []Move

	// Try to solve using basic patterns
	// For now, just return the inverse of common beginner moves
	// This is a placeholder until we implement a real algorithm

	// Create a copy of the cube to test moves
	testCube := s.copyCube(cube)

	// Try a sequence of moves to see if they help
	basicMoves := []Move{
		{Face: Right, Clockwise: true},
		{Face: Up, Clockwise: true},
		{Face: Right, Clockwise: false},
		{Face: Up, Clockwise: false},
		{Face: Right, Clockwise: false},
		{Face: Front, Clockwise: true},
		{Face: Right, Clockwise: true},
		{Face: Front, Clockwise: false},
	}

	// Apply basic moves to test cube
	testCube.ApplyMoves(basicMoves)

	// If this sequence gets us closer to solved (very basic heuristic)
	// then use it, otherwise use a different pattern
	if s.countSolvedPieces(testCube) > s.countSolvedPieces(cube) {
		solution = basicMoves
	} else {
		// Try different moves
		solution = []Move{
			{Face: Front, Clockwise: true},
			{Face: Up, Clockwise: true},
			{Face: Right, Clockwise: true},
			{Face: Up, Clockwise: false},
			{Face: Right, Clockwise: false},
			{Face: Front, Clockwise: false},
		}
	}

	return &SolverResult{
		Solution: solution,
		Steps:    len(solution),
		Duration: time.Since(start),
	}, nil
}

// copyCube creates a deep copy of a cube for testing
func (s *BeginnerSolver) copyCube(original *Cube) *Cube {
	copy := &Cube{Size: original.Size}

	for face := 0; face < 6; face++ {
		copy.Faces[face] = make([][]Color, original.Size)
		for row := 0; row < original.Size; row++ {
			copy.Faces[face][row] = make([]Color, original.Size)
			for col := 0; col < original.Size; col++ {
				copy.Faces[face][row][col] = original.Faces[face][row][col]
			}
		}
	}

	return copy
}

// countSolvedPieces counts how many pieces are in correct position/orientation
func (s *BeginnerSolver) countSolvedPieces(cube *Cube) int {
	count := 0
	solvedColors := []Color{White, Yellow, Red, Orange, Blue, Green}

	for face := 0; face < 6; face++ {
		expectedColor := solvedColors[face]
		for row := 0; row < cube.Size; row++ {
			for col := 0; col < cube.Size; col++ {
				if cube.Faces[face][row][col] == expectedColor {
					count++
				}
			}
		}
	}

	return count
}

// CFOPSolver implements the CFOP method
type CFOPSolver struct{}

func (s *CFOPSolver) Name() string {
	return "CFOP"
}

func (s *CFOPSolver) Solve(cube *Cube) (*SolverResult, error) {
	start := time.Now()

	// Placeholder CFOP implementation
	// Real CFOP would implement:
	// 1. Cross
	// 2. F2L (First Two Layers)
	// 3. OLL (Orient Last Layer)
	// 4. PLL (Permute Last Layer)

	solution := []Move{
		{Face: Front, Clockwise: true},
		{Face: Right, Clockwise: true},
		{Face: Up, Clockwise: true},
		{Face: Right, Clockwise: false},
		{Face: Up, Clockwise: false},
		{Face: Front, Clockwise: false},
	}

	return &SolverResult{
		Solution: solution,
		Steps:    len(solution),
		Duration: time.Since(start),
	}, nil
}

// KociembaSolver implements Kociemba's two-phase algorithm
type KociembaSolver struct{}

func (s *KociembaSolver) Name() string {
	return "Kociemba"
}

func (s *KociembaSolver) Solve(cube *Cube) (*SolverResult, error) {
	if cube.Size != 3 {
		return nil, fmt.Errorf("Kociemba algorithm only supports 3x3x3 cubes")
	}

	start := time.Now()

	// Placeholder Kociemba implementation
	// Real Kociemba would implement:
	// Phase 1: Get to <U,D,R2,L2,F2,B2> subgroup
	// Phase 2: Solve within the subgroup

	solution := []Move{
		{Face: Up, Clockwise: true},
		{Face: Right, Clockwise: true},
		{Face: Up, Clockwise: false},
		{Face: Right, Clockwise: false},
		{Face: Up, Clockwise: false},
		{Face: Front, Clockwise: false},
		{Face: Up, Clockwise: true},
		{Face: Front, Clockwise: true},
	}

	return &SolverResult{
		Solution: solution,
		Steps:    len(solution),
		Duration: time.Since(start),
	}, nil
}

// GetSolver returns a solver by name
func GetSolver(name string) (Solver, error) {
	switch name {
	case "beginner":
		return &BeginnerSolver{}, nil
	case "cfop":
		return &CFOPSolver{}, nil
	case "kociemba":
		return &KociembaSolver{}, nil
	default:
		return nil, fmt.Errorf("unknown solver: %s", name)
	}
}
