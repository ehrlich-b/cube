package cube

// FindPattern returns one shortest sequence in the supplied move alphabet.
// Grey target stickers are wildcards. Nil moves selects the 18 face turns.
// Unsupported dimensions/states/alphabets retain a sticker BFS fallback.
// Neither cube is mutated.
func FindPattern(start, target *Cube, moves []Move, maxDepth int) ([]Move, bool) {
	if !searchCubeShape(start) || !searchCubeShape(target) || start.Size != target.Size {
		return nil, false
	}
	if matchesStickerPattern(start, target) {
		return []Move{}, true
	}
	if maxDepth <= 0 {
		return nil, false
	}
	indices, supported := coordinateMoveIndices(moves)
	if start.Size != 3 || !supported || Validate3x3(start) != nil {
		return stickerPatternBFS(start, target, moves, maxDepth)
	}
	// Relabel colors relative to the current centers, preserving the physical
	// move frame. No extra rotations may enter a shortest face-turn answer.
	work, goal := start.clone(), target.clone()
	home := NewCube(3)
	var colors [7]Color
	colors[Grey] = Grey
	for face := range start.Faces {
		colors[start.Faces[face][1][1]] = home.Faces[face][1][1]
	}
	complete := true
	for face := range work.Faces {
		for row := range work.Faces[face] {
			for col := range work.Faces[face][row] {
				work.Faces[face][row][col] = colors[work.Faces[face][row][col]]
				color := goal.Faces[face][row][col]
				if color < White || color > Grey {
					return nil, false
				}
				goal.Faces[face][row][col] = colors[color]
				if color == Grey {
					complete = false
				}
			}
		}
		if goal.Faces[face][1][1] != Grey && goal.Faces[face][1][1] != home.Faces[face][1][1] {
			return nil, false
		}
	}
	state := readCubie(work)
	if complete {
		if Validate3x3(goal) != nil {
			return nil, false
		}
		return exactCoordinateSearch(readCubie(goal).inverse().mul(state), moves, maxDepth)
	}
	p := compileCubiePattern(goal)
	s := wildcardSearch{pattern: p, allowed: indices, path: make([]int, maxDepth)}
	for _, m := range indices {
		s.present[m] = true
	}
	for depth := p.bound(state); depth <= maxDepth; depth++ {
		if s.dfs(state, depth, 0, -1) {
			result := make([]Move, depth)
			for i, m := range s.path[:depth] {
				result[i] = coordinateMoves[m]
			}
			return result, true
		}
	}
	return nil, false
}

func searchCubeShape(c *Cube) bool {
	if c == nil || c.Size < 1 {
		return false
	}
	for _, face := range c.Faces {
		if len(face) != c.Size {
			return false
		}
		for _, row := range face {
			if len(row) != c.Size {
				return false
			}
		}
	}
	return true
}

func matchesStickerPattern(c, target *Cube) bool {
	for f := range target.Faces {
		for r, row := range target.Faces[f] {
			for col, color := range row {
				if color != Grey && c.Faces[f][r][col] != color {
					return false
				}
			}
		}
	}
	return true
}

type cubiePattern struct {
	corners [8][24]uint8
	edges   [12][24]uint8
}

// Each piece's table is a multi-source shortest-distance table to all positions
// and orientations compatible with the requested stickers. Taking their max
// is admissible: every real solution must satisfy each individual piece.
func compileCubiePattern(target *Cube) cubiePattern {
	var p cubiePattern
	home := NewCube(3)
	for id, colors := range cornerFacelets {
		var goals [24]bool
		for slot, coords := range cornerFacelets {
			for twist := 0; twist < 3; twist++ {
				ok := true
				for i, coord := range coords {
					color := sticker(target, coord)
					if color != Grey && color != sticker(home, colors[(i-twist+3)%3]) {
						ok = false
					}
				}
				goals[slot*3+twist] = ok
			}
		}
		p.corners[id] = pieceDistances(goals, true)
	}
	for id, colors := range edgeFacelets {
		var goals [24]bool
		for slot, coords := range edgeFacelets {
			for flip := 0; flip < 2; flip++ {
				ok := true
				for i, coord := range coords {
					color := sticker(target, coord)
					if color != Grey && color != sticker(home, colors[i^flip]) {
						ok = false
					}
				}
				goals[slot*2+flip] = ok
			}
		}
		p.edges[id] = pieceDistances(goals, false)
	}
	return p
}

func pieceDistances(goals [24]bool, corner bool) [24]uint8 {
	var distance [24]uint8
	queue := make([]int, 0, 24)
	for x, goal := range goals {
		distance[x] = 255
		if goal {
			distance[x] = 0
			queue = append(queue, x)
		}
	}
	orientations := 2
	if corner {
		orientations = 3
	}
	for head := 0; head < len(queue); head++ {
		x := queue[head]
		pos, ori := x/orientations, x%orientations
		for _, m := range cubieMoves {
			var next int
			if corner {
				for dest, source := range m.cp {
					if int(source) == pos {
						next = dest*3 + (ori+int(m.co[dest]))%3
						break
					}
				}
			} else {
				for dest, source := range m.ep {
					if int(source) == pos {
						next = dest*2 + (ori ^ int(m.eo[dest]))
						break
					}
				}
			}
			if distance[next] == 255 {
				distance[next] = distance[x] + 1
				queue = append(queue, next)
			}
		}
	}
	return distance
}

func (p *cubiePattern) bound(s cubie) int {
	h := 0
	for pos, id := range s.cp {
		h = max(h, int(p.corners[id][pos*3+int(s.co[pos])]))
	}
	for pos, id := range s.ep {
		h = max(h, int(p.edges[id][pos*2+int(s.eo[pos])]))
	}
	return h
}

type wildcardSearch struct {
	pattern cubiePattern
	allowed []int
	present [18]bool
	path    []int
}

func (s *wildcardSearch) dfs(state cubie, left, level, prev int) bool {
	h := s.pattern.bound(state)
	if h > left {
		return false
	}
	if left == 0 {
		return h == 0
	}
	prune := exactSearch{present: s.present}
	for _, m := range s.allowed {
		if prune.skip(m, prev) {
			continue
		}
		s.path[level] = m
		if s.dfs(state.mul(cubieMoves[m]), left-1, level+1, m) {
			return true
		}
	}
	return false
}

func stickerPatternBFS(start, target *Cube, moves []Move, maxDepth int) ([]Move, bool) {
	if moves == nil {
		moves = coordinateMoves[:]
	}
	type node struct {
		c    *Cube
		path []Move
	}
	queue := []node{{start.clone(), nil}}
	seen := map[string]bool{start.String(): true}
	for head := 0; head < len(queue); head++ {
		current := queue[head]
		queue[head] = node{}
		if len(current.path) >= maxDepth {
			continue
		}
		for _, m := range moves {
			next := current.c.clone()
			next.ApplyMove(m)
			key := next.String()
			if seen[key] {
				continue
			}
			seen[key] = true
			path := append(append([]Move(nil), current.path...), m)
			if matchesStickerPattern(next, target) {
				return path, true
			}
			if len(path) < maxDepth {
				queue = append(queue, node{next, path})
			}
		}
	}
	return nil, false
}
