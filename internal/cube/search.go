package cube

// SearchToTarget runs an iterative-deepening depth-first search over the 18
// face turns (U D L R F B, each in three variants) looking for the shortest
// sequence whose end state satisfies isTarget. It returns ok=false if no such
// sequence is found within maxDepth moves.
//
// Two prunings keep the branching factor down without losing optimality:
//   - the same face is never turned twice in a row (U2 after U is redundant);
//   - each opposite-face pair is only searched in a single canonical order
//     (U-then-D is tried, D-then-U is skipped, since U D and D U commute).
//
// The caller's cube is never mutated: the search runs on a private copy.
// If isTarget(c) already holds, an empty (but non-nil) sequence and ok=true
// are returned.
func SearchToTarget(c *Cube, isTarget func(*Cube) bool, maxDepth int) ([]Move, bool) {
	if isTarget(c) {
		return []Move{}, true
	}

	work := c.clone()
	for depth := 1; depth <= maxDepth; depth++ {
		path := make([]Move, 0, depth)
		if searchDFS(work, isTarget, depth, &path, noFace) {
			return path, true
		}
	}
	return nil, false
}

// noFace is a sentinel used before the first move, when neither pruning rule
// has a previous face to check against.
const noFace Face = -1

// moveFaces lists the six face-turning axes in the fixed order the search tries
// them.
var moveFaces = []Face{Up, Down, Left, Right, Front, Back}

// faceMoves returns the three variants of a face turn in the order the search
// tries them: clockwise, half turn, counter-clockwise.
func faceMoves(face Face) []Move {
	return []Move{
		{Face: face, Clockwise: true},
		{Face: face, Double: true},
		{Face: face}, // Clockwise false == counter-clockwise
	}
}

// oppositeFaces reports whether two faces lie on opposite sides of the cube.
func oppositeFaces(a, b Face) bool {
	switch a {
	case Up:
		return b == Down
	case Down:
		return b == Up
	case Left:
		return b == Right
	case Right:
		return b == Left
	case Front:
		return b == Back
	case Back:
		return b == Front
	}
	return false
}

// searchDFS explores the pruned move tree exactly depthLeft moves deep from
// the current state, recording the moves taken in path. It reports whether a
// state satisfying isTarget was reached. On failure the cube is left exactly as
// it was found (each applied move is undone); on success the last move stays
// applied, which is fine because the search cube is a throwaway copy.
func searchDFS(c *Cube, isTarget func(*Cube) bool, depthLeft int, path *[]Move, prevFace Face) bool {
	if depthLeft == 0 {
		return isTarget(c)
	}

	for _, face := range moveFaces {
		// Turning the same face twice in a row is redundant: skip it.
		if prevFace != noFace && face == prevFace {
			continue
		}
		// Opposite faces commute, so force a single canonical order: the pairs
		// below all satisfy smaller-value-first (Front<Back, Left<Right,
		// Up<Down), so a face is only explored after its opposite if it is the
		// later member (e.g. after D, U is skipped).
		if prevFace != noFace && oppositeFaces(face, prevFace) && face < prevFace {
			continue
		}
		for _, mv := range faceMoves(face) {
			c.ApplyMove(mv)
			*path = append(*path, mv)
			if searchDFS(c, isTarget, depthLeft-1, path, mv.Face) {
				return true
			}
			*path = (*path)[:len(*path)-1]
			c.ApplyMove(invertMove(mv))
		}
	}
	return false
}

// invertMove returns the move that undoes mv. Only the simple face turns the
// search generates are supported.
func invertMove(mv Move) Move {
	if mv.Double {
		return mv // 180-degree turns are their own inverse
	}
	mv.Clockwise = !mv.Clockwise
	return mv
}

// clone returns a deep copy of the cube.
func (c *Cube) clone() *Cube {
	nc := &Cube{Size: c.Size}
	for face := 0; face < 6; face++ {
		nc.Faces[face] = make([][]Color, c.Size)
		for row := 0; row < c.Size; row++ {
			nc.Faces[face][row] = append([]Color(nil), c.Faces[face][row]...)
		}
	}
	return nc
}
