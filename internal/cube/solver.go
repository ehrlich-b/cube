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
	Stages   []SolveStage // CFOP checkpoints, in playback order.
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

// CFOPSolver implements cross, paired F2L, database OLL and database PLL.
type CFOPSolver struct{}

func (s *CFOPSolver) Name() string {
	return "CFOP"
}

func (s *CFOPSolver) Solve(c *Cube) (*SolverResult, error) {
	return solveCFOP(c)
}

// KociembaSolver implements Kociemba's two-phase algorithm.
type KociembaSolver struct{}

func (s *KociembaSolver) Name() string {
	return "Kociemba"
}

func (s *KociembaSolver) Solve(c *Cube) (*SolverResult, error) {
	return solveTwoPhase(c)
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
