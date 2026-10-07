package cli

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/ehrlich-b/cube/internal/cube"
)

func init() {
	if os.Getenv("CUBE_CACHE_DIR") == "" {
		root, _ := filepath.Abs("../../.scratch/cube-cache")
		_ = os.Setenv("CUBE_CACHE_DIR", root)
	}
}

func TestCoordinateSearchMatchesOldBFS(t *testing.T) {
	for depth := 1; depth <= 4; depth++ {
		for seed := int64(1); seed <= 10; seed++ {
			start, _ := scrambleAndTarget(t, seed+200, depth)
			bfs := breadthFirstSearch(start, func(c *cube.Cube) bool { return c.IsSolved() }, depth)
			result, ok := cube.FindPattern(start, cube.NewCube(3), nil, depth)
			if len(bfs) == 0 || !ok || len(result) != len(bfs[0].moves) {
				t.Fatalf("depth %d seed %d: old BFS/IDA* disagree", depth, seed)
			}
		}
	}
}

// Meet two independent sticker BFS frontiers of radius three. This gives the
// same shortest distance as the old forward BFS through depth six without its
// multi-gigabyte frontier. No cubie coordinates or search heuristics are used.
func stickerBFSRadius(c *cube.Cube, depth int) map[string]int {
	type node struct {
		c     *cube.Cube
		depth int
	}
	queue := []node{{copyCube(c), 0}}
	seen := map[string]int{c.String(): 0}
	moves, _ := cube.ParseMoves("R R2 R' L L2 L' U U2 U' D D2 D' F F2 F' B B2 B'")
	for head := 0; head < len(queue); head++ {
		current := queue[head]
		if current.depth == depth {
			continue
		}
		for _, m := range moves {
			next := copyCube(current.c)
			next.ApplyMove(m)
			key := next.String()
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = current.depth + 1
			queue = append(queue, node{next, current.depth + 1})
		}
	}
	return seen
}

func TestCoordinateSearchMatchesBFSDepthSix(t *testing.T) {
	r := rand.New(rand.NewSource(2026100704))
	moves, _ := cube.ParseMoves("R R2 R' L L2 L' U U2 U' D D2 D' F F2 F' B B2 B'")
	for depth := 1; depth <= 6; depth++ {
		for n := 0; n < 10; n++ {
			start := cube.NewCube(3)
			for i := 0; i < 12; i++ {
				start.ApplyMove(moves[r.Intn(18)])
			}
			target := copyCube(start)
			for i := 0; i < depth; i++ {
				target.ApplyMove(moves[r.Intn(18)])
			}
			a, b := stickerBFSRadius(start, 3), stickerBFSRadius(target, 3)
			shortest := 7
			for key, da := range a {
				if db, ok := b[key]; ok && da+db < shortest {
					shortest = da + db
				}
			}
			result, ok := cube.FindPattern(start, target, nil, 6)
			if !ok || len(result) != shortest {
				t.Fatalf("depth %d case %d: BFS=%d IDA*=%d (%v)", depth, n, shortest, len(result), ok)
			}
			start.ApplyMoves(result)
			if start.String() != target.String() {
				t.Fatal("answer does not reach concrete target")
			}
		}
	}
	t.Log("60 random start/target pairs match sticker BFS through depth six")
}
