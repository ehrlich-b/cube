package cube

import (
	"fmt"
)

// whiteCrossEdges lists the four white-adjacent edges in the fixed order the
// solver places them. In the canonical orientation White sits on the Down face
// and Blue faces Front, so the white-blue edge belongs to the front side of the
// cross, white-orange to the left, white-red to the right and white-green to
// the back.
var whiteCrossEdges = [4][2]Color{
	{White, Blue},
	{White, Orange},
	{White, Red},
	{White, Green},
}

// maxCrossMoves is the hard cap on the length of a returned cross solution. A
// beginner-length cross never needs anywhere near this many moves; the cap is a
// safety net so a pathological search can never return an unbounded sequence.
const maxCrossMoves = 100

// SolveWhiteCross returns a sequence of moves that, applied to c, makes
// WhiteCrossSolved true. It works from any reachable 3x3 state, does not
// mutate its input, and never returns a sequence longer than maxCrossMoves.
//
// The cross is solved one edge at a time with incremental predicates: after
// stage k the first k white edges are all solved. Each stage uses a bounded
// IDDFS (SearchToTarget) with increasingly relaxed predicates, rather than
// searching for the whole cross at once (which needs ~8 moves and 18^8 nodes
// and would never finish).
func SolveWhiteCross(c *Cube) ([]Move, error) {
	if c.Size != 3 {
		return nil, fmt.Errorf("SolveWhiteCross requires a 3x3 cube, got size %d", c.Size)
	}

	if WhiteCrossSolved(c) {
		return []Move{}, nil
	}

	work := c.clone()
	var solution []Move

	for k := 1; k <= 4; k++ {
		seq, ok := solveEdgeStage(work, whiteCrossEdges[:k])
		if !ok {
			return nil, fmt.Errorf("SolveWhiteCross: failed to solve edge stage %d", k)
		}
		work.ApplyMoves(seq)
		solution = append(solution, seq...)
	}

	if len(solution) > maxCrossMoves {
		return nil, fmt.Errorf("SolveWhiteCross: solution of %d moves exceeds cap of %d", len(solution), maxCrossMoves)
	}

	return solution, nil
}

// edgesSolved reports whether every edge among the given color pairs sits in
// its home slot with the correct orientation.
func edgesSolved(c *Cube, edges [][2]Color) bool {
	for _, e := range edges {
		if !EdgeSolved(c, e[0], e[1]) {
			return false
		}
	}
	return true
}

// solveEdgeStage finds a short move sequence that solves edge k (the last edge
// in targets) while keeping every edge in targets already solved. It runs on a
// private copy and leaves the caller's cube untouched.
//
// It tries, in order:
//  1. the strict target ("these k edges solved") at depth 5;
//  2. the strict target at depth 7;
//  3. a relaxed target ("edge k solved and at least k-1 of the k edges solved")
//     at depth 7, followed by a re-run of the strict target at depth 5.
//
// A single white edge is always only a handful of moves from home, so the
// simple strict search almost always succeeds; the relaxed fallback exists for
// the rare states where preserving every previously placed edge at once needs a
// bit more slack.
func solveEdgeStage(start *Cube, targets [][2]Color) ([]Move, bool) {
	k := len(targets)
	strict := func(c *Cube) bool { return edgesSolved(c, targets) }

	for depth := 5; depth <= 7; depth += 2 {
		if seq, ok := SearchToTarget(start, strict, depth); ok {
			return seq, true
		}
	}

	relaxed := func(c *Cube) bool {
		if !EdgeSolved(c, targets[k-1][0], targets[k-1][1]) {
			return false
		}
		solved := 0
		for _, e := range targets {
			if EdgeSolved(c, e[0], e[1]) {
				solved++
			}
		}
		return solved >= k-1
	}

	if seq, ok := SearchToTarget(start, relaxed, 7); ok {
		work := start.clone()
		work.ApplyMoves(seq)
		if tail, ok2 := SearchToTarget(work, strict, 5); ok2 {
			combined := append(seq, tail...)
			return combined, true
		}
	}

	return nil, false
}
