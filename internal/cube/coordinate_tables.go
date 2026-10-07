package cube

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"os"
	"path/filepath"
	"sync"
)

// Small, exact pattern databases provide admissible bounds in the face-turn
// metric. No generated tables are repository artifacts. Cached bytes carry a
// version and checksum; invalid/unavailable caches are rebuilt in memory.
type coordinateTables struct {
	Twist, Flip, Slice, Corner                                    []uint16
	Edge2, Slice2                                                 []uint16
	TwistSlice, FlipSlice, CornerSlice, EdgeSlice, CornerDistance []uint8
}

var tablesOnce sync.Once
var tables *coordinateTables

var phase2Moves = []int{1, 4, 7, 10, 12, 13, 14, 15, 16, 17}

func solverTables() *coordinateTables {
	tablesOnce.Do(func() {
		if cached := loadCoordinateTables(); cached != nil {
			tables = cached
			return
		}
		t := &coordinateTables{}
		t.Twist = makeMoveTable(2187, twistCubie, func(s cubie) int { return s.twist() }, false)
		t.Flip = makeMoveTable(2048, flipCubie, func(s cubie) int { return s.flip() }, false)
		t.Slice = makeMoveTable(495, sliceCubie, func(s cubie) int { return s.slice() }, false)
		t.Corner = makeMoveTable(40320, func(x int) cubie { s := identityCubie(); setPermutation(s.cp[:], x); return s }, func(s cubie) int { return permutationRank(s.cp[:]) }, false)
		t.Edge2 = makeMoveTable(40320, func(x int) cubie { s := identityCubie(); setPermutation(s.ep[:8], x); return s }, func(s cubie) int { return permutationRank(s.ep[:8]) }, true)
		t.Slice2 = makeMoveTable(24, func(x int) cubie {
			s := identityCubie()
			setPermutation(s.ep[8:], x)
			for i := 8; i < 12; i++ {
				s.ep[i] += 8
			}
			return s
		}, func(s cubie) int { return permutationRank(s.ep[8:]) }, true)
		t.TwistSlice = pairPruning(t.Twist, t.Slice, 495, false)
		t.FlipSlice = pairPruning(t.Flip, t.Slice, 495, false)
		t.CornerSlice = pairPruning(t.Corner, t.Slice2, 24, true)
		t.EdgeSlice = pairPruning(t.Edge2, t.Slice2, 24, true)
		t.CornerDistance = pairPruning(t.Corner, make([]uint16, 18), 1, false)
		tables = t
		saveCoordinateTables(t)
	})
	return tables
}

func makeMoveTable(size int, decode func(int) cubie, encode func(cubie) int, phase2 bool) []uint16 {
	t := make([]uint16, size*18)
	for x := 0; x < size; x++ {
		s := decode(x)
		for m := range cubieMoves {
			if phase2 && m < 12 && m%3 != 1 {
				continue
			}
			t[x*18+m] = uint16(encode(s.mul(cubieMoves[m])))
		}
	}
	return t
}

func pairPruning(a, b []uint16, bSize int, phase2 bool) []uint8 {
	dist := make([]uint8, len(a)/18*bSize)
	for i := range dist {
		dist[i] = 255
	}
	dist[0] = 0
	queue := make([]uint32, 1, len(dist))
	queue[0] = 0
	moves := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}
	if phase2 {
		moves = phase2Moves
	}
	for head := 0; head < len(queue); head++ {
		x := int(queue[head])
		ax, bx := x/bSize*18, x%bSize*18
		for _, m := range moves {
			y := int(a[ax+m])*bSize + int(b[bx+m])
			if dist[y] == 255 {
				dist[y] = dist[x] + 1
				queue = append(queue, uint32(y))
			}
		}
	}
	return dist
}

func coordinateCachePath() string {
	dir := os.Getenv("CUBE_CACHE_DIR")
	if dir == "" {
		base, err := os.UserCacheDir()
		if err != nil {
			return ""
		}
		dir = filepath.Join(base, "cube")
	}
	return filepath.Join(dir, "coordinates-v1.gob")
}

func loadCoordinateTables() *coordinateTables {
	data, err := os.ReadFile(coordinateCachePath())
	if err != nil || len(data) < 32 || len(data) > 16<<20 {
		return nil
	}
	sum := sha256.Sum256(data[32:])
	if !bytes.Equal(data[:32], sum[:]) {
		return nil
	}
	var t coordinateTables
	if gob.NewDecoder(bytes.NewReader(data[32:])).Decode(&t) != nil {
		return nil
	}
	if len(t.Twist) != 2187*18 || len(t.Flip) != 2048*18 || len(t.Slice) != 495*18 || len(t.Corner) != 40320*18 || len(t.Edge2) != 40320*18 || len(t.Slice2) != 24*18 || len(t.TwistSlice) != 2187*495 || len(t.FlipSlice) != 2048*495 || len(t.CornerSlice) != 40320*24 || len(t.EdgeSlice) != 40320*24 || len(t.CornerDistance) != 40320 {
		return nil
	}
	return &t
}

func saveCoordinateTables(t *coordinateTables) {
	path := coordinateCachePath()
	if path == "" {
		return
	}
	var b bytes.Buffer
	if gob.NewEncoder(&b).Encode(t) != nil {
		return
	}
	sum := sha256.Sum256(b.Bytes())
	if os.MkdirAll(filepath.Dir(path), 0700) != nil {
		return
	}
	f, err := os.CreateTemp(filepath.Dir(path), "coordinates-*.gob")
	if err != nil {
		return
	}
	name := f.Name()
	defer os.Remove(name)
	_, e1 := f.Write(sum[:])
	_, e2 := f.Write(b.Bytes())
	e3 := f.Close()
	if e1 == nil && e2 == nil && e3 == nil {
		_ = os.Rename(name, path)
	}
}

func (t *coordinateTables) phase1Bound(co, eo, sl int) int {
	return max(int(t.TwistSlice[co*495+sl]), int(t.FlipSlice[eo*495+sl]))
}

func (t *coordinateTables) phase2Bound(cp, ep, sp int) int {
	return max(int(t.CornerSlice[cp*24+sp]), int(t.EdgeSlice[ep*24+sp]))
}
