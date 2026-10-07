package cube

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Small, exact pattern databases provide admissible bounds in the face-turn
// metric. No generated tables are repository artifacts. Cached bytes carry a
// version and checksum; invalid/unavailable caches are rebuilt in memory.
type coordinateTables struct {
	Twist, Flip, Slice, Corner                                               []uint16
	Edge2, Slice2                                                            []uint16
	TwistSlice, FlipSlice, TwistFlip, CornerSlice, EdgeSlice, CornerDistance []uint8
}

var tablesLock = make(chan struct{}, 1)
var tables *coordinateTables

var phase2Moves = []int{1, 4, 7, 10, 12, 13, 14, 15, 16, 17}

func solverTables() *coordinateTables {
	return solverTablesLimit(time.Time{})
}

func tableDeadlineExceeded(deadline time.Time) bool {
	return !deadline.IsZero() && !time.Now().Before(deadline)
}

// A timed caller can also stop waiting for another caller's initialization.
func lockSearchTables(lock chan struct{}, deadline time.Time) bool {
	if tableDeadlineExceeded(deadline) {
		return false
	}
	if deadline.IsZero() {
		lock <- struct{}{}
		return true
	}
	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()
	select {
	case lock <- struct{}{}:
		if tableDeadlineExceeded(deadline) {
			<-lock
			return false
		}
		return true
	case <-timer.C:
		return false
	}
}

func solverTablesLimit(deadline time.Time) *coordinateTables {
	if !lockSearchTables(tablesLock, deadline) {
		return nil
	}
	defer func() { <-tablesLock }()
	if tables != nil {
		return tables
	}
	if cached := loadCoordinateTablesLimit(deadline); cached != nil {
		if tableDeadlineExceeded(deadline) {
			return nil
		}
		tables = cached
		return tables
	}
	t := &coordinateTables{}
	t.Twist = makeMoveTable(2187, twistCubie, func(s cubie) int { return s.twist() }, false, deadline)
	t.Flip = makeMoveTable(2048, flipCubie, func(s cubie) int { return s.flip() }, false, deadline)
	t.Slice = makeMoveTable(495, sliceCubie, func(s cubie) int { return s.slice() }, false, deadline)
	t.Corner = makeMoveTable(40320, func(x int) cubie { s := identityCubie(); setPermutation(s.cp[:], x); return s }, func(s cubie) int { return permutationRank(s.cp[:]) }, false, deadline)
	t.Edge2 = makeMoveTable(40320, func(x int) cubie { s := identityCubie(); setPermutation(s.ep[:8], x); return s }, func(s cubie) int { return permutationRank(s.ep[:8]) }, true, deadline)
	t.Slice2 = makeMoveTable(24, func(x int) cubie {
		s := identityCubie()
		setPermutation(s.ep[8:], x)
		for i := 8; i < 12; i++ {
			s.ep[i] += 8
		}
		return s
	}, func(s cubie) int { return permutationRank(s.ep[8:]) }, true, deadline)
	if tableDeadlineExceeded(deadline) {
		return nil
	}
	t.TwistSlice = pairPruning(t.Twist, t.Slice, 495, false, deadline)
	t.FlipSlice = pairPruning(t.Flip, t.Slice, 495, false, deadline)
	t.TwistFlip = pairPruning(t.Twist, t.Flip, 2048, false, deadline)
	t.CornerSlice = pairPruning(t.Corner, t.Slice2, 24, true, deadline)
	t.EdgeSlice = pairPruning(t.Edge2, t.Slice2, 24, true, deadline)
	t.CornerDistance = pairPruning(t.Corner, make([]uint16, 18), 1, false, deadline)
	if tableDeadlineExceeded(deadline) {
		return nil
	}
	// Publish only complete tables so a cancelled build can be retried.
	tables = t
	// Optional disk persistence must not extend a timed solve's setup.
	if deadline.IsZero() {
		saveCoordinateTables(t)
	}
	return tables
}

func makeMoveTable(size int, decode func(int) cubie, encode func(cubie) int, phase2 bool, deadline time.Time) []uint16 {
	if tableDeadlineExceeded(deadline) {
		return nil
	}
	t := make([]uint16, size*18)
	for x := 0; x < size; x++ {
		if x&255 == 0 && tableDeadlineExceeded(deadline) {
			return nil
		}
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

func pairPruning(a, b []uint16, bSize int, phase2 bool, deadline time.Time) []uint8 {
	if tableDeadlineExceeded(deadline) {
		return nil
	}
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
		if head&1023 == 0 && tableDeadlineExceeded(deadline) {
			return nil
		}
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
	return filepath.Join(dir, "coordinates-v2.gob")
}

func loadCoordinateTables() *coordinateTables {
	return loadCoordinateTablesLimit(time.Time{})
}

func loadCoordinateTablesLimit(deadline time.Time) *coordinateTables {
	if tableDeadlineExceeded(deadline) {
		return nil
	}
	f, err := os.Open(coordinateCachePath())
	if err != nil {
		return nil
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() < 32 || info.Size() > 16<<20 || tableDeadlineExceeded(deadline) {
		return nil
	}
	// Allocate only the checked size, even if the file grows after Stat.
	data := make([]byte, int(info.Size()))
	if _, err := io.ReadFull(f, data); err != nil || tableDeadlineExceeded(deadline) {
		return nil
	}
	var extra [1]byte
	if n, err := f.Read(extra[:]); n != 0 || err != io.EOF {
		return nil
	}
	sum := sha256.Sum256(data[32:])
	if !bytes.Equal(data[:32], sum[:]) || tableDeadlineExceeded(deadline) {
		return nil
	}
	var t coordinateTables
	if gob.NewDecoder(bytes.NewReader(data[32:])).Decode(&t) != nil || tableDeadlineExceeded(deadline) {
		return nil
	}
	if len(t.Twist) != 2187*18 || len(t.Flip) != 2048*18 || len(t.Slice) != 495*18 || len(t.Corner) != 40320*18 || len(t.Edge2) != 40320*18 || len(t.Slice2) != 24*18 || len(t.TwistSlice) != 2187*495 || len(t.FlipSlice) != 2048*495 || len(t.TwistFlip) != 2187*2048 || len(t.CornerSlice) != 40320*24 || len(t.EdgeSlice) != 40320*24 || len(t.CornerDistance) != 40320 {
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
	return max(int(t.TwistSlice[co*495+sl]), int(t.FlipSlice[eo*495+sl]), int(t.TwistFlip[co*2048+eo]))
}

func (t *coordinateTables) phase2Bound(cp, ep, sp int) int {
	return max(int(t.CornerSlice[cp*24+sp]), int(t.EdgeSlice[ep*24+sp]))
}
