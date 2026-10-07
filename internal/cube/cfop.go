package cube

import (
	"container/heap"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// SolveStage is a verified CFOP checkpoint. Empty stages represent skips.
type SolveStage struct {
	Name  string
	Cases []string
	Moves []Move
	After *Cube
}

// TurnCount counts face, wide and slice turns; grip rotations are separate.
func TurnCount(moves []Move) int {
	n := 0
	for _, m := range moves {
		if m.Rotation == NoRotation {
			n++
		}
	}
	return n
}

func cfopTarget(slots uint8) *Cube {
	goal := NewCube(3)
	for f := range goal.Faces {
		for r := range goal.Faces[f] {
			for c := range goal.Faces[f][r] {
				goal.Faces[f][r][c] = Grey
			}
		}
	}
	home := NewCube(3)
	set := func(p Coord) { goal.Faces[p.Face][p.Row][p.Col] = sticker(home, p) }
	for _, coords := range edgeFacelets[4:8] {
		for _, p := range coords {
			set(p)
		}
	}
	for i := 0; i < 4; i++ {
		if slots&(1<<i) == 0 {
			continue
		}
		for _, p := range cornerFacelets[4+i] {
			set(p)
		}
		for _, p := range edgeFacelets[8+i] {
			set(p)
		}
	}
	return goal
}

// F2LSlotSolved checks a corner-edge pair in FR, FL, BL, BR order.
func F2LSlotSolved(c *Cube, slot int) bool {
	if !searchCubeShape(c) || c.Size != 3 || slot < 0 || slot >= 4 {
		return false
	}
	for _, p := range cornerFacelets[4+slot] {
		if sticker(c, p) != c.faceColor(p.Face) {
			return false
		}
	}
	for _, p := range edgeFacelets[8+slot] {
		if sticker(c, p) != c.faceColor(p.Face) {
			return false
		}
	}
	return true
}

func cfopSlots(c *Cube) uint8 {
	var mask uint8
	for i := 0; i < 4; i++ {
		if F2LSlotSolved(c, i) {
			mask |= 1 << i
		}
	}
	return mask
}

func OLLSolved(c *Cube) bool {
	if !searchCubeShape(c) || c.Size != 3 {
		return false
	}
	for _, row := range c.Faces[Up] {
		for _, color := range row {
			if color != c.faceColor(Up) {
				return false
			}
		}
	}
	return true
}

type cfopCase struct {
	pattern *Cube
	moves   []Move
	names   []string
	state   cubie
}

type cfopDatabase struct {
	f2l      []cfopCase
	oll, pll []cfopCase
	err      error
}

var cfopOnce sync.Once
var cfopDB cfopDatabase

// Express grip conjugations in the fixed starting frame. This keeps y setups
// from adding needless rotations to every slot/last-layer algorithm.
func fixedFrameMoves(moves []Move) []Move {
	frame := NewCube(3)
	var rotations, result []Move
	for _, m := range moves {
		if m.Rotation != NoRotation {
			frame.ApplyMove(m)
			rotations = append(rotations, m)
			continue
		}
		probe := frame.clone()
		probe.ApplyMove(m)
		probe.ApplyMoves(inverseSequence(rotations))
		found := false
		for _, candidate := range cfopMoveAlphabet() {
			test := NewCube(3)
			test.ApplyMove(candidate)
			if test.String() == probe.String() {
				result = append(result, candidate)
				found = true
				break
			}
		}
		if !found {
			return append([]Move(nil), moves...)
		}
	}
	_, restore, _ := canonical3x3(frame)
	result = append(result, inverseSequence(restore)...)
	return OptimizeMoves(result)
}

func cfopMoveAlphabet() []Move {
	result := append([]Move(nil), coordinateMoves[:]...)
	extra, _ := ParseMoves("M M2 M' E E2 E' S S2 S' Rw Rw2 Rw' Lw Lw2 Lw' Uw Uw2 Uw' Dw Dw2 Dw' Fw Fw2 Fw' Bw Bw2 Bw'")
	return append(result, extra...)
}

func yawMoves(turns int) []Move {
	if turns == 0 {
		return nil
	}
	return []Move{{Rotation: Y_Rotation, Clockwise: turns != 3, Double: turns == 2}}
}

func aufMoves(turns int) []Move {
	if turns == 0 {
		return nil
	}
	return []Move{{Face: Up, Clockwise: turns != 3, Double: turns == 2}}
}

func cfopCube(s cubie) *Cube {
	c, home := NewCube(3), NewCube(3)
	for slot, coords := range cornerFacelets {
		for i, p := range coords {
			c.Faces[p.Face][p.Row][p.Col] = sticker(home, cornerFacelets[s.cp[slot]][(i-int(s.co[slot])+3)%3])
		}
	}
	for slot, coords := range edgeFacelets {
		for i, p := range coords {
			c.Faces[p.Face][p.Row][p.Col] = sticker(home, edgeFacelets[s.ep[slot]][i^int(s.eo[slot])])
		}
	}
	return c
}

func llKey(s cubie, oll bool) string {
	if oll {
		return string(s.co[:4]) + string(s.eo[:4])
	}
	return string(s.cp[:4]) + string(s.ep[:4])
}

type llNode struct {
	state cubie
	moves []Move
	names []string
	cost  int
}
type llQueue []llNode

func (q llQueue) Len() int           { return len(q) }
func (q llQueue) Less(i, j int) bool { return q[i].cost < q[j].cost }
func (q llQueue) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q *llQueue) Push(x any)        { *q = append(*q, x.(llNode)) }
func (q *llQueue) Pop() any          { old := *q; n := old[len(old)-1]; *q = old[:len(old)-1]; return n }

// Reverse Dijkstra closes the incomplete CSV sets under verified database
// algorithms and AUF. OLL has 216 orientation states; PLL has 288 permutations.
// Every resulting recognition pattern is an inverse-to-goal state, just like
// the imported CFEN patterns. No generic full-cube solver supplies LL moves.
func compileLastLayer(macros []cfopCase, oll bool) []cfopCase {
	queue := &llQueue{{state: identityCubie()}}
	heap.Init(queue)
	distance := map[string]int{llKey(identityCubie(), oll): 0}
	var cases []cfopCase
	for queue.Len() > 0 {
		current := heap.Pop(queue).(llNode)
		if distance[llKey(current.state, oll)] != current.cost {
			continue
		}
		pattern := cfopCube(current.state)
		if oll {
			for f := range pattern.Faces {
				for r := range pattern.Faces[f] {
					for c, color := range pattern.Faces[f][r] {
						if color != Yellow {
							pattern.Faces[f][r][c] = Grey
						}
					}
				}
			}
		}
		cases = append(cases, cfopCase{pattern: pattern, moves: current.moves, names: current.names, state: current.state})
		for _, macro := range macros {
			next := current.state.mul(macro.state.inverse())
			key := llKey(next, oll)
			cost := current.cost + TurnCount(macro.moves)
			if old, ok := distance[key]; ok && old <= cost {
				continue
			}
			distance[key] = cost
			moves := append(append([]Move(nil), macro.moves...), current.moves...)
			names := append(append([]string(nil), macro.names...), current.names...)
			heap.Push(queue, llNode{next, moves, names, cost})
		}
	}
	return cases
}

func loadCFOPDatabase() {
	var oll, pll []cfopCase
	for _, a := range AlgorithmDatabase {
		if a.Dimension != 3 || (!a.HasCategory("F2L") && !a.HasCategory("OLL") && !a.HasCategory("PLL")) {
			continue
		}
		moves, err := ParseMoves(a.Moves)
		if err != nil {
			cfopDB.err = err
			return
		}
		recognition, err := readAlgorithmPattern(a.Pattern)
		if err != nil {
			cfopDB.err = fmt.Errorf("%s: %w", a.CaseID, err)
			return
		}
		replay := recognition.clone()
		replay.ApplyMoves(moves)
		if replay.String() != NewCube(3).String() {
			cfopDB.err = fmt.Errorf("%s recognition does not verify", a.CaseID)
			return
		}
		end := NewCube(3)
		end.ApplyMoves(moves)
		_, restore, err := canonical3x3(end)
		if err != nil {
			cfopDB.err = err
			return
		}
		moves = append(moves, restore...)
		frame := NewCube(3)
		frame.ApplyMoves(inverseSequence(restore))
		recognition = relabelPattern(recognition, frame)
		for yaw := 0; yaw < 4; yaw++ {
			seq := append(append(yawMoves(yaw), moves...), inverseSequence(yawMoves(yaw))...)
			seq = fixedFrameMoves(seq)
			after := NewCube(3)
			after.ApplyMoves(seq)
			if !WhiteCrossSolved(after) {
				continue
			}
			frame := NewCube(3)
			frame.ApplyMoves(yawMoves(yaw))
			before := relabelPattern(recognition, frame)
			before.ApplyMoves(inverseSequence(yawMoves(yaw)))
			replay := before.clone()
			replay.ApplyMoves(seq)
			if replay.String() != NewCube(3).String() {
				cfopDB.err = fmt.Errorf("%s rotated recognition does not verify", a.CaseID)
				return
			}
			candidate := cfopCase{moves: seq, names: []string{a.CaseID + " · " + a.Name}, state: readCubie(before).inverse()}
			if a.HasCategory("F2L") {
				// A safe one-slot algorithm may affect this pair and the top layer.
				if cfopSlots(before) != 15 {
					missing := 15 ^ cfopSlots(before)
					if missing == 0 || missing&(missing-1) != 0 {
						continue
					}
					pattern := before.clone()
					// Keep only the target corner and edge's stickers, identified by cubie identity.
					state := readCubie(before)
					for f := range pattern.Faces {
						for r := range pattern.Faces[f] {
							for c := range pattern.Faces[f][r] {
								pattern.Faces[f][r][c] = Grey
							}
						}
					}
					for pos, id := range state.cp {
						if id >= 4 && missing&(1<<(id-4)) != 0 {
							for _, p := range cornerFacelets[pos] {
								pattern.Faces[p.Face][p.Row][p.Col] = sticker(before, p)
							}
						}
					}
					for pos, id := range state.ep {
						if id >= 8 && missing&(1<<(id-8)) != 0 {
							for _, p := range edgeFacelets[pos] {
								pattern.Faces[p.Face][p.Row][p.Col] = sticker(before, p)
							}
						}
					}
					candidate.pattern = pattern
					candidate.names = []string{cfopAlgorithmName(a, "F2L")}
					cfopDB.f2l = append(cfopDB.f2l, candidate)
				}
			}
			if a.HasCategory("OLL") && cfopSlots(after) == 15 {
				candidate.names = []string{cfopAlgorithmName(a, "OLL")}
				oll = append(oll, candidate)
			}
			if a.HasCategory("PLL") && cfopSlots(after) == 15 && OLLSolved(after) {
				candidate.names = []string{cfopAlgorithmName(a, "PLL")}
				pll = append(pll, candidate)
			}
		}
	}
	// U/U2/U' are explicit AUF edges, so before/after alignment is covered.
	for turns := 1; turns < 4; turns++ {
		seq := aufMoves(turns)
		c := NewCube(3)
		c.ApplyMoves(seq)
		candidate := cfopCase{moves: seq, names: []string{"AUF"}, state: readCubie(c)}
		oll = append(oll, candidate)
		pll = append(pll, candidate)
	}
	cfopDB.oll = compileLastLayer(oll, true)
	cfopDB.pll = compileLastLayer(pll, false)
	if len(cfopDB.oll) != 216 || len(cfopDB.pll) != 288 {
		cfopDB.err = fmt.Errorf("database LL coverage: OLL %d/216, PLL %d/288", len(cfopDB.oll), len(cfopDB.pll))
	}
	sort.SliceStable(cfopDB.f2l, func(i, j int) bool { return TurnCount(cfopDB.f2l[i].moves) < TurnCount(cfopDB.f2l[j].moves) })
}

func solveCFOP(c *Cube) (*SolverResult, error) {
	started := time.Now()
	if err := Validate3x3(c); err != nil {
		return nil, err
	}
	work, rotations, _ := canonical3x3(c)
	result := &SolverResult{}
	add := func(name string, names []string, moves []Move) {
		// Returned stage slices must not expose the immutable recognition cache.
		moves = append([]Move(nil), moves...)
		names = append([]string(nil), names...)
		work.ApplyMoves(moves)
		result.Solution = append(result.Solution, moves...)
		result.Stages = append(result.Stages, SolveStage{name, names, moves, work.clone()})
	}
	cross, ok := FindPattern(work, cfopTarget(0), nil, 8)
	if !ok {
		return nil, fmt.Errorf("CFOP cross exceeded eight turns")
	}
	// The search used the canonical copy; replay the grip change exactly once
	// from the user's original frame when recording the Cross checkpoint.
	work = c.clone()
	add("Cross", []string{"Optimal white cross"}, append(rotations, cross...))
	cfopOnce.Do(loadCFOPDatabase)
	if cfopDB.err != nil {
		return nil, cfopDB.err
	}
	for n := 1; n <= 4; n++ {
		protected := cfopSlots(work)
		if protected == 15 {
			add(fmt.Sprintf("F2L %d", n), []string{"Skip"}, nil)
			continue
		}
		var chosen []Move
		var names []string
		found := false
		for _, candidate := range cfopDB.f2l {
			for auf := 0; auf < 4; auf++ {
				probe := work.clone()
				setup := aufMoves(auf)
				probe.ApplyMoves(setup)
				if !matchesStickerPattern(probe, candidate.pattern) {
					continue
				}
				probe.ApplyMoves(candidate.moves)
				if slots := cfopSlots(probe); slots&protected != protected || slots == protected || !WhiteCrossSolved(probe) {
					continue
				}
				seq := OptimizeMoves(append(setup, candidate.moves...))
				if !found || TurnCount(seq) < TurnCount(chosen) {
					chosen = seq
					names = candidate.names
					found = true
				}
			}
		}
		if !found {
			// Select the next unsolved pair and preserve every existing slot.
			for slot := 0; slot < 4; slot++ {
				if protected&(1<<slot) == 0 {
					chosen, found = FindPattern(work, cfopTarget(protected|(1<<slot)), nil, 12)
					names = []string{fmt.Sprintf("Pattern search · %s", []string{"FR", "FL", "BL", "BR"}[slot])}
					break
				}
			}
		}
		if !found {
			return nil, fmt.Errorf("CFOP F2L %d exceeded search bound", n)
		}
		add(fmt.Sprintf("F2L %d", n), names, chosen)
		if cfopSlots(work)&protected != protected || cfopSlots(work) == protected || !WhiteCrossSolved(work) {
			return nil, fmt.Errorf("CFOP F2L checkpoint failed")
		}
	}
	for _, stage := range []struct {
		name  string
		cases []cfopCase
	}{{"OLL", cfopDB.oll}, {"PLL", cfopDB.pll}} {
		found := false
		for _, candidate := range stage.cases {
			if matchesStickerPattern(work, candidate.pattern) {
				names := candidate.names
				if len(names) == 0 {
					names = []string{"Skip"}
				}
				add(stage.name, names, candidate.moves)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("CFOP %s pattern not found", stage.name)
		}
		if cfopSlots(work) != 15 || !OLLSolved(work) {
			return nil, fmt.Errorf("CFOP %s checkpoint failed", stage.name)
		}
	}
	replay := c.clone()
	replay.ApplyMoves(result.Solution)
	if !work.IsSolved() || !replay.IsSolved() {
		return nil, fmt.Errorf("CFOP returned an incomplete solution")
	}
	result.Steps = len(result.Solution)
	result.Duration = time.Since(started)
	return result, nil
}

func (s SolveStage) CaseName() string { return strings.Join(s.Cases, " + ") }

func cfopAlgorithmName(a Algorithm, category string) string {
	if strings.HasPrefix(a.CaseID, category+"-") {
		return a.CaseID + " · " + a.Name
	}
	for i, alias := range a.Aliases {
		if strings.HasPrefix(alias, category+"-") {
			name := a.Name
			if i+1 < len(a.Aliases) {
				name = a.Aliases[i+1]
			}
			return alias + " · " + name
		}
	}
	return a.CaseID + " · " + a.Name
}
