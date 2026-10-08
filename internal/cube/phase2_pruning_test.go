package cube

import (
	"math/rand"
	"testing"
)

func TestPhaseTwoCorrelatedPruning(t *testing.T) {
	tables := solverTables()
	combMoves := cornerCombinationMoves()
	r := rand.New(rand.NewSource(2026100712))
	stronger := false
	for trial := 0; trial < 1000; trial++ {
		state := identityCubie()
		for depth := 0; depth < 18; depth++ {
			cp, ep, sp := permutationRank(state.cp[:]), permutationRank(state.ep[:8]), permutationRank(state.ep[8:])
			old := max(tables.cornerSliceBound(cp, sp), tables.edgeSliceBound(ep, sp))
			bound := tables.phase2Bound(cp, ep, sp)
			if max(bound, tables.phase2InverseBound(cp, ep, sp)) > depth {
				t.Fatalf("inadmissible phase-two bound at depth %d", depth)
			}
			stronger = stronger || bound > old
			inv := state.inverse()
			if int(tables.PermInverse[cp]) != permutationRank(inv.cp[:]) || int(tables.PermInverse[ep]) != permutationRank(inv.ep[:8]) || int(tables.SliceInverse[sp]) != permutationRank(inv.ep[8:]) {
				t.Fatal("inverse coordinate disagrees with cubie inverse")
			}
			m := phase2Moves[r.Intn(len(phase2Moves))]
			next := state.mul(cubieMoves[m])
			if int(combMoves[cornerCombination(state)*18+m]) != cornerCombination(next) {
				t.Fatal("combination/parity move disagrees with cubie move")
			}
			state = next
		}
	}
	if !stronger {
		t.Fatal("correlated pruning never strengthens the previous bound")
	}
}
