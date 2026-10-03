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

// BeginnerSolver implements the complete 3x3 beginner lesson. Returned moves
// must satisfy the existing full-solver contract, not just a partial layer goal.
type BeginnerSolver struct{}

func (s *BeginnerSolver) Name() string {
	return "Beginner"
}

func (s *BeginnerSolver) Solve(c *Cube) (*SolverResult, error) {
	start := time.Now()
	lesson, err := PlanBeginner(c)
	if err != nil {
		return nil, err
	}
	moves := lesson.Moves()
	return &SolverResult{
		Solution: moves,
		Steps:    len(moves),
		Duration: time.Since(start),
	}, nil
}

// CFOP and Kociemba remain unimplemented stubs. Beginner is complete for 3x3.
//
// The CLI rejects these unavailable algorithms rather than presenting an empty
// result as success. Their legacy API behavior remains guarded by the invariant.

// TODO: All solver helper methods will be implemented with the new design

// CFOPSolver implements CFOP method (placeholder)
type CFOPSolver struct{}

func (s *CFOPSolver) Name() string {
	return "CFOP"
}

func (s *CFOPSolver) Solve(cube *Cube) (*SolverResult, error) {
	start := time.Now()

	// Check if cube is already solved
	if cube.IsSolved() {
		return &SolverResult{
			Solution: []Move{},
			Steps:    0,
			Duration: time.Since(start),
		}, nil
	}

	// TODO: Implement real CFOP solver
	// For now, return empty solution
	return &SolverResult{
		Solution: []Move{},
		Steps:    0,
		Duration: time.Since(start),
	}, nil
}

// KociembaSolver implements Kociemba's two-phase algorithm (placeholder)
type KociembaSolver struct{}

func (s *KociembaSolver) Name() string {
	return "Kociemba"
}

func (s *KociembaSolver) Solve(cube *Cube) (*SolverResult, error) {
	if cube.Size != 3 {
		return nil, fmt.Errorf("Kociemba algorithm only supports 3x3x3 cubes")
	}

	start := time.Now()

	// Check if cube is already solved
	if cube.IsSolved() {
		return &SolverResult{
			Solution: []Move{},
			Steps:    0,
			Duration: time.Since(start),
		}, nil
	}

	// TODO: Implement real Kociemba two-phase algorithm
	// For now, return empty solution
	return &SolverResult{
		Solution: []Move{},
		Steps:    0,
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
