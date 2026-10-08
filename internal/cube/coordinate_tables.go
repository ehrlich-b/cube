package cube

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/gob"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Small, exact pattern databases provide admissible bounds in the face-turn
// metric. A reproducible compact asset ships with the executable; first use
// copies it to the optional cache. Checksums, dimensions and move fingerprints
// guard cached bytes; a corrupt/unavailable cache falls back to the asset.
type coordinateTables struct {
	TwistRepresentatives                                                        []uint16
	NearPhase1Keys                                                              []uint32
	NearPhase1Distances                                                         []uint8
	nearPhase1                                                                  map[uint32]uint8
	Twist, Flip, Slice, Corner                                                  []uint16
	Edge2, Slice2                                                               []uint16
	TwistSlice, FlipSlice, TwistFlip, CornerSlice, EdgeSlice, CornerDistance    []uint8
	CornerComb, EdgeComb, CornerEdgeComb                                        []uint8
	PermInverse, SliceInverse                                                   []uint16
	PermSymCorner, PermSymEdge, SymSlice2, SymCornerComb, SymEdgeComb           []uint16
	TwistSym, SliceSym, TwistFlipSym, SymTwist, SymFlip, SymFlipDelta, SymSlice []uint16
}

//go:embed tables/coordinates-v5.bin.gz
var embeddedCoordinates []byte

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
	if t := decodeCoordinateTables(embeddedCoordinates, deadline); t != nil {
		tables = t
		if deadline.IsZero() {
			saveCoordinateBytes(embeddedCoordinates)
		}
		return t
	}
	t := generateCoordinateTables(deadline)
	if t == nil {
		return nil
	}
	tables = t
	if deadline.IsZero() {
		saveCoordinateTables(t)
	}
	return tables
}

func generateCoordinateTables(deadline time.Time) *coordinateTables {
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
	t.CornerComb, t.PermInverse, t.SliceInverse = phase2PermutationCoordinates(deadline)
	t.EdgeComb = pairPruning(t.Edge2, cornerCombinationMoves(), 140, true, deadline)

	if tableDeadlineExceeded(deadline) {
		return nil
	}
	t.CornerEdgeComb = pairPruning(t.Corner, edgeCombinationMoves(), 140, true, deadline)
	if tableDeadlineExceeded(deadline) {
		return nil
	}
	t.compactPhase1()
	t.compactPhase2()
	t.generateNearPhase1()
	if tableDeadlineExceeded(deadline) {
		return nil
	}
	return t
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
		if i&65535 == 0 && tableDeadlineExceeded(deadline) {
			return nil
		}
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
	return filepath.Join(dir, "coordinates-v5.bin.gz")
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
	if err != nil || !info.Mode().IsRegular() || info.Size() < 64 || info.Size() > 10<<20 || tableDeadlineExceeded(deadline) {
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
	return decodeCoordinateTables(data, deadline)
}

func decodeCoordinateTables(data []byte, deadline time.Time) *coordinateTables {
	if len(data) < 64 || tableDeadlineExceeded(deadline) {
		return nil
	}
	sum := sha256.Sum256(data[32:])
	if !bytes.Equal(data[:32], sum[:]) {
		return nil
	}
	fingerprint := edgeMoveFingerprint()
	if !bytes.Equal(data[32:64], fingerprint[:]) {
		return nil
	}
	z, err := gzip.NewReader(bytes.NewReader(data[64:]))
	if err != nil {
		return nil
	}
	defer z.Close()
	payload, err := io.ReadAll(io.LimitReader(tableDeadlineReader{z, deadline}, 32<<20))
	if err != nil || len(payload) == 32<<20 || tableDeadlineExceeded(deadline) {
		return nil
	}
	var t coordinateTables
	if gob.NewDecoder(bytes.NewReader(payload)).Decode(&t) != nil || tableDeadlineExceeded(deadline) {
		return nil
	}
	if len(t.Twist) != 2187*18 || len(t.Flip) != 2048*18 || len(t.Slice) != 495*18 || len(t.Corner) != 40320*18 || len(t.Edge2) != 40320*18 || len(t.Slice2) != 24*18 || len(t.TwistFlip) != 324*2048/2 || len(t.CornerDistance) != 40320 {
		return nil
	}
	if len(t.CornerComb) != 40320 || len(t.PermInverse) != 40320 || len(t.SliceInverse) != 24 {
		return nil
	}
	if len(t.TwistRepresentatives) != 168 || len(t.TwistSym) != 2187 || len(t.TwistFlipSym) != 2187 || len(t.SliceSym) != 495 || len(t.SymTwist) != 2187*16 || len(t.SymFlip) != 2048*16 || len(t.SymFlipDelta) != 495*16 || len(t.SymSlice) != 495*16 {
		return nil
	}
	if len(t.TwistSlice) != (int(maxSliceClass(t.TwistSym)+1)*495+1)/2 || len(t.FlipSlice) != (int(maxSliceClass(t.SliceSym))+1)*2048/2 {
		return nil
	}
	if len(t.PermSymCorner) != 40320 || len(t.PermSymEdge) != 40320 || len(t.SymSlice2) != 24*16 || len(t.SymCornerComb) != 140*16 || len(t.SymEdgeComb) != 140*16 {
		return nil
	}
	cn, en := int(maxSliceClass(t.PermSymCorner)+1), int(maxSliceClass(t.PermSymEdge)+1)
	if len(t.CornerSlice) != cn*24/2 || len(t.EdgeSlice) != en*24/2 || len(t.EdgeComb) != en*140/2 || len(t.CornerEdgeComb) != cn*140/2 {
		return nil
	}
	if len(t.NearPhase1Keys) == 0 || len(t.NearPhase1Keys) != len(t.NearPhase1Distances) || len(t.NearPhase1Keys) > 1_000_000 {
		return nil
	}
	t.nearPhase1 = make(map[uint32]uint8, len(t.NearPhase1Keys))
	for i, key := range t.NearPhase1Keys {
		if i&4095 == 0 && tableDeadlineExceeded(deadline) {
			return nil
		}
		t.nearPhase1[key] = t.NearPhase1Distances[i]
	}
	return &t
}

func encodeCoordinateTables(t *coordinateTables) []byte {
	var b bytes.Buffer
	z, _ := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if gob.NewEncoder(z).Encode(t) != nil || z.Close() != nil {
		return nil
	}
	fingerprint := edgeMoveFingerprint()
	payload := append(fingerprint[:], b.Bytes()...)
	sum := sha256.Sum256(payload)
	return append(sum[:], payload...)
}

func saveCoordinateTables(t *coordinateTables) { saveCoordinateBytes(encodeCoordinateTables(t)) }

func saveCoordinateBytes(data []byte) {
	path := coordinateCachePath()
	if path == "" || data == nil || os.MkdirAll(filepath.Dir(path), 0700) != nil {
		return
	}
	f, err := os.CreateTemp(filepath.Dir(path), "coordinates-*.gz")
	if err != nil {
		return
	}
	name := f.Name()
	defer os.Remove(name)
	_, e1 := f.Write(data)
	e2 := f.Close()
	if e1 == nil && e2 == nil {
		_ = os.Rename(name, path)
	}
}

func (t *coordinateTables) phase1Bound(co, eo, sl int) int {
	conjCo := int(t.SymTwist[co*16+1])
	conjEo := int(t.SymFlip[eo*16+1] ^ t.SymFlipDelta[sl*16+1])
	return max(t.twistSliceBound(co, sl), t.flipSliceBound(eo, sl), t.twistFlipBound(co, eo), t.twistFlipBound(conjCo, conjEo))
}

func (t *coordinateTables) phase2Bound(cp, ep, sp int) int {
	return max(t.cornerSliceBound(cp, sp), t.edgeSliceBound(ep, sp), t.edgeCornerBound(ep, int(t.CornerComb[cp])), t.edgeCornerBound(int(t.PermInverse[ep]), int(t.CornerComb[t.PermInverse[cp]])), t.cornerEdgeBound(cp, int(t.CornerComb[ep])))
}
func (t *coordinateTables) phase2Pruned(cp, ep, sp, left int) bool {
	return t.cornerSliceBound(cp, sp) > left || t.edgeSliceBound(ep, sp) > left || t.edgeCornerBound(ep, int(t.CornerComb[cp])) > left || t.edgeCornerBound(int(t.PermInverse[ep]), int(t.CornerComb[t.PermInverse[cp]])) > left || t.cornerEdgeBound(cp, int(t.CornerComb[ep])) > left
}
func (t *coordinateTables) phase2InverseBound(cp, ep, sp int) int {
	return t.phase2Bound(int(t.PermInverse[cp]), int(t.PermInverse[ep]), int(t.SliceInverse[sp]))
}

// For search, any admissible bound above left suffices. Most shallow children
// use the exact frontier; other children can stop after a compact pair check.
func (t *coordinateTables) phase1Prune(co, eo, sl, left int) int {
	if left <= 4 {
		return t.nearPhase1Bound(co, eo, sl)
	}
	h := t.twistFlipBound(co, eo)
	if h > left {
		return h
	}
	h = max(h, t.twistFlipBound(int(t.SymTwist[co*16+1]), int(t.SymFlip[eo*16+1]^t.SymFlipDelta[sl*16+1])))
	if h > left {
		return h
	}
	h = max(h, t.twistSliceBound(co, sl))
	if h > left {
		return h
	}
	return max(h, t.flipSliceBound(eo, sl))
}

// Bound decompression work between deadline checks, including tiny --optimal
// budgets. A cancelled load is never published to the shared table pointer.
type tableDeadlineReader struct {
	r        io.Reader
	deadline time.Time
}

func (r tableDeadlineReader) Read(p []byte) (int, error) {
	if tableDeadlineExceeded(r.deadline) {
		return 0, context.DeadlineExceeded
	}
	if len(p) > 64<<10 {
		p = p[:64<<10]
	}
	return r.r.Read(p)
}
