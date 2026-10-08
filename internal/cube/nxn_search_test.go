package cube

import (
	"math/bits"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func TestNxNWingScoresMatchFullPermutation(t *testing.T) {
	rng := rand.New(rand.NewSource(20261008))
	for _, n := range []int{4, 5, 6, 7} {
		tables, err := nxnTables(n)
		if err != nil {
			t.Fatal(err)
		}
		for _, o := range tables.wings {
			actions := nxnPairingActions(n, o)
			for trial := 0; trial < 20; trial++ {
				var state, mates [24]uint8
				for i, id := range rng.Perm(24) {
					state[i] = uint8(id)
				}
				for i, id := range state {
					mates[i] = o.mate[id]
				}
				for i := range actions {
					a := &actions[i]
					after := nxnWingAfter(state, *a)
					fixed, paired := 0, 0
					for pos, id := range after {
						if int(id) == pos {
							fixed++
						}
						if after[o.mate[pos]] == o.mate[id] {
							paired++
						}
					}
					if a.fixedScore(nxnPackWings(state)) != fixed || a.pairedScore(state, mates) != paired {
						t.Fatalf("%dx%d layer %d action %d: compact wing score differs", n, n, o.layer, i)
					}
				}
			}
		}
	}
}

func TestNxNBlockScoresMatchStickerPermutation(t *testing.T) {
	rng := rand.New(rand.NewSource(2026100801))
	for _, n := range []int{6, 7} {
		tables, err := nxnTables(n)
		if err != nil {
			t.Fatal(err)
		}
		actions := nxnCenterBlockActions(n, tables)
		var positions []int
		for _, o := range tables.centers {
			positions = append(positions, o.positions...)
		}
		var colors [144]Color
		var present [6][3]uint64
		var flat [6 * 7 * 7]Color
		home := NewCube(n)
		for i, pos := range positions {
			colors[i] = Color(rng.Intn(6))
			flat[pos] = colors[i]
			present[colors[i]][i/64] |= 1 << (i % 64)
		}
		for id := range actions {
			a := &actions[id]
			p := nxnPermutation(n, a.moves)
			gain := 0
			for _, src := range positions {
				dst := p[src]
				if flat[src] == nxnColor(home, dst) {
					gain++
				}
				if flat[src] == nxnColor(home, src) {
					gain--
				}
			}
			for _, transfer := range a.trans {
				if p[positions[transfer.src]] != positions[transfer.dst] {
					t.Fatalf("%dx%d action %d: range composition differs from stickers", n, n, id)
				}
			}
			if n%2 == 1 {
				for face := 0; face < 6; face++ {
					pos := face*n*n + (n/2)*n + n/2
					if p[pos] != pos {
						t.Fatalf("%dx%d action %d moved a fixed center", n, n, id)
					}
				}
			}
			if a.gain(&present) != gain {
				t.Fatalf("%dx%d action %d: bitset gain %d, sticker gain %d", n, n, id, a.gain(&present), gain)
			}
		}
		influences := nxnCenterInfluences(actions)
		gains := make([]int, len(actions))
		for id := range actions {
			gains[id] = actions[id].gain(&present)
		}
		for trial := 0; trial < 20; trial++ {
			pos, color := rng.Intn(len(positions)), Color(rng.Intn(6))
			for _, term := range influences[pos][colors[pos]] {
				gains[term>>1] -= 1 - 2*int(term&1)
			}
			for _, term := range influences[pos][color] {
				gains[term>>1] += 1 - 2*int(term&1)
			}
			present[colors[pos]][pos/64] &^= 1 << (pos % 64)
			present[color][pos/64] |= 1 << (pos % 64)
			colors[pos] = color
			for id := range actions {
				if gains[id] != actions[id].gain(&present) {
					t.Fatalf("%dx%d action %d: incremental gain drifted", n, n, id)
				}
			}
		}
	}
}

func TestNxNCenterBeamMatchesReference(t *testing.T) {
	rng := rand.New(rand.NewSource(2026100802))
	for _, n := range []int{4, 5} {
		tables, err := nxnTables(n)
		if err != nil {
			t.Fatal(err)
		}
		var dbs []*centerPatterns
		for _, o := range tables.centers {
			dbs = append(dbs, centerPatternTable(n, tables, o))
		}
		c := NewCube(n)
		for i := 0; i < 60; i++ {
			if err := c.ApplyMove(tables.moves[rng.Intn(len(tables.moves))]); err != nil {
				t.Fatal(err)
			}
		}
		for stage, faces := range [][]Face{{Up, Down}, {Left, Right, Front, Back}} {
			var allowed []int
			for g, m := range tables.moves {
				if n%2 == 1 && m.Layer == n/2 {
					continue
				}
				if stage == 0 || !m.Wide && (m.Layer == 0 || m.Layer == n-1) || nxnAxis(m.Face) == nxnAxis(Up) {
					allowed = append(allowed, g)
				}
			}
			got := nxnSearchCenters(c, tables, dbs, faces, allowed)
			want := referenceCenterSearch(c, tables, dbs, faces, allowed)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("%dx%d stage %d: beam changed generator ties\ngot %s\nwant %s", n, n, stage, FormatMoves(got), FormatMoves(want))
			}
		}
	}
}

// The original exhaustive beam expansion: Go maps, unpruned candidates, and
// stable comparison sorting. Keep this independent of optimized selection.
func referenceCenterSearch(c *Cube, t *reductionTables, dbs []*centerPatterns, faces []Face, allowed []int) []Move {
	var start [8]uint16
	var dists [][]uint8
	home := NewCube(c.Size)
	for orbit, o := range t.centers {
		for k, face := range faces {
			var mask uint32
			for i, pos := range o.positions {
				if nxnColor(c, pos) == home.Faces[face][0][0] {
					mask |= 1 << i
				}
			}
			start[len(faces)*orbit+k] = centerRank(mask)
			dists = append(dists, dbs[orbit].distances(face, allowed))
		}
	}
	score := func(key [8]uint16) int {
		h, wrong := 0, 0
		for i, dist := range dists {
			h += int(dist[key[i]])
			db := dbs[i/len(faces)]
			wrong += 4 - bits.OnesCount32(db.masks[key[i]]&db.goal[faces[i%len(faces)]])
		}
		return h*16 + wrong
	}
	nodes := []centerSearchNode{{key: start, parent: -1, move: -1, score: score(start)}}
	beam, best := []int{0}, 0
	seen := map[[8]uint16]bool{start: true}
	for depth := 0; depth < 36 && nodes[best].score != 0; depth++ {
		var candidates []centerSearchNode
		levelSeen := map[[8]uint16]bool{}
		for _, parent := range beam {
			prev := nodes[parent]
			for _, g := range allowed {
				if prev.move >= 0 && g/3 == prev.move/3 {
					continue
				}
				next := centerSearchNode{parent: parent, move: g}
				for i := range dists {
					next.key[i] = dbs[i/len(faces)].next[g][prev.key[i]]
				}
				if seen[next.key] || levelSeen[next.key] {
					continue
				}
				levelSeen[next.key] = true
				next.score = score(next.key)
				candidates = append(candidates, next)
			}
		}
		if len(candidates) == 0 {
			break
		}
		sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].score < candidates[j].score })
		if len(candidates) > 512 {
			candidates = candidates[:512]
		}
		beam = beam[:0]
		for _, next := range candidates {
			seen[next.key] = true
			id := len(nodes)
			nodes = append(nodes, next)
			beam = append(beam, id)
			if next.score < nodes[best].score {
				best = id
			}
		}
	}
	var reversed []Move
	for best != 0 {
		reversed = append(reversed, t.moves[nodes[best].move])
		best = nodes[best].parent
	}
	part := make([]Move, len(reversed))
	for i, m := range reversed {
		part[len(part)-1-i] = m
	}
	return part
}
