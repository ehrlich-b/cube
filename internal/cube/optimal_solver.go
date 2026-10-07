package cube

import (
	"fmt"
	"time"
)

// SolveOptimal performs a face-turn-metric IDA* search with a time limit.
// Rotated grips are normalized before search; those grip rotations are not
// part of the face-turn distance. A timeout is an error, never an unproved
// "optimal" answer. Table initialization is included in the limit.
func SolveOptimal(c *Cube, limit time.Duration) (*SolverResult, error) {
	start := time.Now()
	if limit <= 0 {
		return nil, fmt.Errorf("optimal search time limit must be positive")
	}
	if err := Validate3x3(c); err != nil {
		return nil, err
	}
	if c.IsSolved() {
		return &SolverResult{Solution: []Move{}, Duration: time.Since(start)}, nil
	}
	work, rotations, err := canonical3x3(c)
	if err != nil {
		return nil, err
	}
	// IDA* proves every smaller depth impossible before returning its answer.
	result, ok, timedOut := exactCoordinateSearchLimit(readCubie(work), nil, 22, start.Add(limit))
	if timedOut {
		return nil, fmt.Errorf("optimal search time limit exceeded (%s); use --method kociemba for a fast solution", limit)
	}
	if !ok {
		return nil, fmt.Errorf("no optimal solution found within 22 face turns")
	}
	moves := append(compactGrip(rotations), result...)
	check := c.clone()
	check.ApplyMoves(moves)
	if !check.IsSolved() {
		return nil, fmt.Errorf("optimal solution failed full-cube verification")
	}
	return &SolverResult{Solution: moves, Steps: len(moves), Duration: time.Since(start)}, nil
}
