package cube

import (
	"fmt"
	"sync"
)

// Each movable center orbit has 24 positions. A wing has two sticker orbits
// of 24, with one sticker per physical piece in each orbit. Keeping an ordered
// color pair distinguishes the two wings of an edge without invented flips.
type reductionOrbit struct {
	positions []int
	partners  []int // wings only, indexed like positions
	layer     int   // a slice whose quarter turn changes this wing orbit's parity
	cycle     []Move
	root      int
	parent    []int16
	via       []uint8
}

type reductionTables struct {
	centers []*reductionOrbit
	wings   []*reductionOrbit
	moves   []Move
	perms   []Permutation
}

var reductionCache struct {
	sync.Mutex
	tables [8]*reductionTables
}

func nxnTables(n int) (*reductionTables, error) {
	reductionCache.Lock()
	defer reductionCache.Unlock()
	if t := reductionCache.tables[n]; t != nil {
		return t, nil
	}
	t, err := buildReductionTables(n)
	if err == nil {
		reductionCache.tables[n] = t
	}
	return t, err
}

func nxnInverse(moves []Move) []Move {
	result := make([]Move, len(moves))
	for i, m := range moves {
		if !m.Double {
			m.Clockwise = !m.Clockwise
		}
		result[len(moves)-1-i] = m
	}
	return result
}

func nxnCommutator(a, b []Move) []Move {
	result := append(append([]Move{}, a...), b...)
	result = append(result, nxnInverse(a)...)
	return append(result, nxnInverse(b)...)
}

func nxnPermutation(n int, moves []Move) Permutation {
	p := make(Permutation, 6*n*n)
	for i := range p {
		p[i] = i
	}
	for _, m := range moves {
		kind, turns := moveToMoveType(m)
		for _, layer := range getAffectedLayers(m, n) {
			q := getPermutation(n, kind, layer, turns)
			for i := range p {
				p[i] = q[p[i]]
			}
		}
	}
	return p
}

func buildReductionTables(n int) (*reductionTables, error) {
	t := &reductionTables{}
	// R/U/F through all layers include the opposite faces. Half/inverse turns
	// shorten setup paths; they do not change the generated permutation group.
	for _, f := range []Face{Right, Up, Front} {
		for layer := 0; layer < n; layer++ {
			for turns := 1; turns <= 3; turns++ {
				m := Move{Face: f, Layer: layer, Clockwise: turns == 1, Double: turns == 2}
				t.moves = append(t.moves, m)
				t.perms = append(t.perms, nxnPermutation(n, []Move{m}))
			}
		}
	}
	owner := make([]*reductionOrbit, 6*n*n)
	for start := range owner {
		if owner[start] != nil {
			continue
		}
		o := &reductionOrbit{positions: []int{start}}
		owner[start] = o
		for head := 0; head < len(o.positions); head++ {
			for _, p := range t.perms {
				dst := p[o.positions[head]]
				if owner[dst] == nil {
					owner[dst] = o
					o.positions = append(o.positions, dst)
				}
			}
		}
		f, r, c := indexToCoord(start, n)
		if !isBoundaryCell(f, r, c, n) && len(o.positions) == 24 {
			t.centers = append(t.centers, o)
		}
	}
	// [[inner R, U], inner F] is a pure center 3-cycle when the two
	// slices are not opposite intersections. Discover the seed for each orbit
	// from its complete sticker permutation and verify its exact support.
	u := []Move{{Face: Up, Clockwise: true}}
	for i := 1; i < n-1; i++ {
		for j := 1; j < n-1; j++ {
			r := []Move{{Face: Right, Layer: i, Clockwise: true}}
			f := []Move{{Face: Front, Layer: j, Clockwise: true}}
			moves := nxnCommutator(nxnCommutator(r, u), f)
			p := nxnPermutation(n, moves)
			support := nxnSupport(p)
			if len(support) != 3 {
				continue
			}
			o := owner[support[0]]
			face, row, col := indexToCoord(support[0], n)
			if !isBoundaryCell(face, row, col, n) && o.cycle == nil {
				o.cycle = moves
			}
		}
	}
	// [inner R, [R,U]] is a pure wing 3-cycle: the outer commutator
	// fixes every center, and the inner turn fixes every corner and midge.
	outer := nxnCommutator([]Move{{Face: Right, Clockwise: true}}, u)
	for layer := 1; layer < n/2; layer++ {
		moves := nxnCommutator([]Move{{Face: Right, Layer: layer, Clockwise: true}}, outer)
		p := nxnPermutation(n, moves)
		support := nxnSupport(p)
		if len(support) != 6 {
			return nil, fmt.Errorf("dimension %d: wing commutator has unexpected support", n)
		}
		o := owner[support[0]]
		o.cycle, o.layer = moves, layer
		for _, pos := range o.positions {
			f, r, c := indexToCoord(pos, n)
			partner, err := boundaryPartner(f, r, c, n)
			if err != nil {
				return nil, err
			}
			o.partners = append(o.partners, int(partner)-1)
		}
		for _, pos := range support {
			if owner[pos] != o && owner[pos] != owner[o.partners[0]] {
				return nil, fmt.Errorf("dimension %d: wing commutator crosses orbits", n)
			}
		}
		t.wings = append(t.wings, o)
	}
	for _, orbits := range [][]*reductionOrbit{t.centers, t.wings} {
		for _, o := range orbits {
			if err := o.buildSetups(n, t); err != nil {
				return nil, err
			}
		}
	}
	return t, nil
}

func nxnSupport(p Permutation) []int {
	var support []int
	for i, dst := range p {
		if i != dst {
			support = append(support, i)
		}
	}
	return support
}

func tripleKey(a, b, c int) int { return (a*24+b)*24 + c }

// Conjugating a verified pure cycle preserves its support size. BFS over just
// three positions supplies a short setup for every ordered triple (24*23*22),
// rather than searching whole cube states or depending on scramble history.
func (o *reductionOrbit) buildSetups(n int, t *reductionTables) error {
	if len(o.positions) != 24 || len(o.cycle) == 0 {
		return fmt.Errorf("dimension %d: no pure cycle for a reduction orbit", n)
	}
	local := make([]int, 6*n*n)
	for i := range local {
		local[i] = -1
	}
	for i, pos := range o.positions {
		local[pos] = i
	}
	p := nxnPermutation(n, o.cycle)
	var support []int
	for _, pos := range o.positions {
		if p[pos] != pos {
			support = append(support, pos)
		}
	}
	if len(support) != 3 || p[p[p[support[0]]]] != support[0] {
		return fmt.Errorf("dimension %d: reduction seed is not a 3-cycle", n)
	}
	a := support[0]
	o.root = tripleKey(local[a], local[p[a]], local[p[p[a]]])
	o.parent = make([]int16, 24*24*24)
	o.via = make([]uint8, len(o.parent))
	for i := range o.parent {
		o.parent[i] = -1
	}
	o.parent[o.root] = int16(o.root)
	queue := make([]int, 1, 24*23*22)
	queue[0] = o.root
	transitions := make([][24]int, len(t.perms))
	for g, perm := range t.perms {
		for i, pos := range o.positions {
			transitions[g][i] = local[perm[pos]]
		}
	}
	for head := 0; head < len(queue); head++ {
		key := queue[head]
		a, b, c := key/(24*24), key/24%24, key%24
		for g, trans := range transitions {
			next := tripleKey(trans[a], trans[b], trans[c])
			if o.parent[next] < 0 {
				o.parent[next], o.via[next] = int16(key), uint8(g)
				queue = append(queue, next)
			}
		}
	}
	if len(queue) != 24*23*22 {
		return fmt.Errorf("dimension %d: incomplete reduction setups (%d triples)", n, len(queue))
	}
	return nil
}

func (o *reductionOrbit) cycleMoves(t *reductionTables, a, b, c int) []Move {
	var reversed []Move
	for key := tripleKey(a, b, c); key != o.root; key = int(o.parent[key]) {
		reversed = append(reversed, t.moves[o.via[key]])
	}
	setup := make([]Move, len(reversed))
	for i, m := range reversed {
		setup[len(setup)-1-i] = m
	}
	moves := append(nxnInverse(setup), o.cycle...)
	return OptimizeMoves(append(moves, setup...))
}
