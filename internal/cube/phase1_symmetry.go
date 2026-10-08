package cube

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const flipSliceStates = 2048 * 495

// Quotient corner orientation by the eight proper rotations preserving the
// U/D axis, and conjugate flip/slice into the same frame. Stabilizers of each
// representative are zero-cost equivalences during breadth-first generation.
type phase1Patterns struct {
	code        [2187]uint16
	reps        []uint16
	stabilizers []uint8
	twistMove   []uint16
	flipPerm    [2048 * 8]uint16
	flipDelta   [495 * 8]uint16
	sliceConj   [495 * 8]uint16
	distance    []uint8
}

var useLargePhase1 = os.Getenv("CUBE_LARGE_TABLES") == "1"

var phase1LargeLock sync.Mutex
var phase1LargeDB atomic.Pointer[phase1Patterns]

func (t *coordinateTables) twoPhaseBound(co, eo, sl int) int {
	if useLargePhase1 {
		if db := phase1LargeDB.Load(); db != nil {
			return db.bound(co, eo, sl)
		}
	}
	return t.phase1Bound(co, eo, sl)
}

func phase1Symmetries() [8]cubie {
	var sym [8]cubie
	for i := range sym {
		frame := NewCube(3)
		if i >= 4 {
			frame.ApplyMove(Move{Rotation: X_Rotation, Double: true})
		}
		for n := 0; n < i%4; n++ {
			frame.ApplyMove(Move{Rotation: Y_Rotation, Clockwise: true})
		}
		sym[i] = readCubie(frame)
	}
	return sym
}

func phase1SymmetryCoordinates(t *coordinateTables) *phase1Patterns {
	db := &phase1Patterns{}
	sym := phase1Symmetries()
	var twistConj [2187 * 8]uint16
	for i, rotation := range sym {
		inv := rotation.inverse()
		for x := 0; x < 2187; x++ {
			twistConj[x*8+i] = uint16(inv.mul(twistCubie(x)).mul(rotation).twist())
		}
		for x := 0; x < 2048; x++ {
			state := flipCubie(x)
			var next cubie
			for p, source := range rotation.ep {
				next.eo[p] = state.eo[source]
			}
			db.flipPerm[x*8+i] = uint16(next.flip())
		}
		for x := 0; x < 495; x++ {
			next := inv.mul(sliceCubie(x)).mul(rotation)
			db.flipDelta[x*8+i], db.sliceConj[x*8+i] = uint16(next.flip()), uint16(next.slice())
		}
	}
	var class [2187]uint16
	for x := range db.code {
		rep, selected := x, 0
		for i := 0; i < 8; i++ {
			if int(twistConj[x*8+i]) < rep {
				rep, selected = int(twistConj[x*8+i]), i
			}
		}
		if rep == x {
			class[x] = uint16(len(db.reps))
			db.reps = append(db.reps, uint16(x))
		}
		db.code[x] = class[rep]*8 + uint16(selected)
	}
	db.stabilizers = make([]uint8, len(db.reps))
	db.twistMove = make([]uint16, len(db.reps)*18)
	for c, rep := range db.reps {
		for i := 0; i < 8; i++ {
			if twistConj[int(rep)*8+i] == rep {
				db.stabilizers[c] |= 1 << i
			}
		}
		for m := 0; m < 18; m++ {
			db.twistMove[c*18+m] = db.code[t.Twist[int(rep)*18+m]]
		}
	}
	return db
}

func (db *phase1Patterns) conjugate(fs, sym int) int {
	flip, slice := fs/495, fs%495
	return int(db.flipPerm[flip*8+sym]^db.flipDelta[slice*8+sym])*495 + int(db.sliceConj[slice*8+sym])
}

func (db *phase1Patterns) bound(co, eo, sl int) int {
	code := int(db.code[co])
	return nibbleDistance(db.distance, code/8*flipSliceStates+db.conjugate(eo*495+sl, code&7))
}

func generatePhase1Pruning(t *coordinateTables, db *phase1Patterns) []uint8 {
	size := len(db.reps) * flipSliceStates
	d := make([]uint8, size)
	for i := range d {
		d[i] = 255
	}
	d[0] = 0
	queue := make([]uint32, 1, size)
	fsMoves := make([]uint32, flipSliceStates*18)
	for x := 0; x < flipSliceStates; x++ {
		for m := 0; m < 18; m++ {
			fsMoves[x*18+m] = uint32(int(t.Flip[x/495*18+m])*495 + int(t.Slice[x%495*18+m]))
		}
	}
	depth := 0
	for head := 0; head < len(queue); head++ {
		x := int(queue[head])
		if int(d[x]) > depth {
			depth = int(d[x])
			// Once most states are known, scanning the remaining states backwards
			// avoids expanding the enormous final breadth-first frontier.
			if len(queue) > size/2 {
				break
			}
		}
		c, fs := x/flipSliceStates, x%flipSliceStates
		for m, next := range fsMoves[fs*18 : fs*18+18] {
			code := int(db.twistMove[c*18+m])
			nc, nf := code/8, db.conjugate(int(next), code&7)
			y := nc*flipSliceStates + nf
			if d[y] != 255 {
				continue
			}
			d[y] = d[x] + 1
			queue = append(queue, uint32(y))
			for sym := 1; sym < 8; sym++ {
				if db.stabilizers[nc]&(1<<sym) == 0 {
					continue
				}
				z := nc*flipSliceStates + db.conjugate(nf, sym)
				if d[z] == 255 {
					d[z] = d[y]
					queue = append(queue, uint32(z))
				}
			}
		}
	}
	queue = nil
	for depth++; depth < 15; depth++ {
		remaining := 0
		for x, distance := range d {
			if distance != 255 {
				continue
			}
			c, fs := x/flipSliceStates, x%flipSliceStates
			found := false
			for m, next := range fsMoves[fs*18 : fs*18+18] {
				code := int(db.twistMove[c*18+m])
				y := code/8*flipSliceStates + db.conjugate(int(next), code&7)
				if int(d[y]) == depth-1 {
					d[x] = uint8(depth)
					found = true
					break
				}
			}
			if !found {
				remaining++
			}
		}
		if remaining == 0 {
			break
		}
	}
	return packDistances(d, time.Time{})
}

func phase1PatternTables(t *coordinateTables) *phase1Patterns {
	db, _ := phase1PatternTablesPersist(t, false) // Solving permits an in-memory table.
	return db
}

func phase1PatternTablesPersist(t *coordinateTables, persist bool) (*phase1Patterns, error) {
	// Browser workers retain the smaller coordinate databases: a native cache
	// and a gigabyte-scale temporary BFS frontier are unsuitable there.
	if runtime.GOARCH == "wasm" || os.Getenv("CUBE_SMALL_TABLES") == "1" {
		return nil, fmt.Errorf("large phase-one tables are unavailable on this platform")
	}
	if db := phase1LargeDB.Load(); db != nil && !persist {
		return db, nil
	}
	phase1LargeLock.Lock()
	defer phase1LargeLock.Unlock()
	if db := phase1LargeDB.Load(); db != nil {
		if persist {
			return db, savePackedPattern("phase1-sym8-v1.bin", db.distance)
		}
		return db, nil
	}
	db := phase1SymmetryCoordinates(t)
	filename := "phase1-sym8-v1.bin"
	db.distance = loadPackedPattern(filename, len(db.reps)*flipSliceStates/2)
	if db.distance == nil {
		db.distance = generatePhase1Pruning(t, db)
		if db.distance == nil {
			return nil, fmt.Errorf("large phase-one table generation failed")
		}
		phase1LargeDB.Store(db)
		return db, savePackedPattern(filename, db.distance)
	}
	phase1LargeDB.Store(db)
	return db, nil
}

func loadPackedPattern(filename string, size int) []uint8 {
	return loadPackedPatternLimit(filename, size, time.Time{})
}

func loadPackedPatternLimit(filename string, size int, deadline time.Time) []uint8 {
	if tableDeadlineExceeded(deadline) {
		return nil
	}
	f, _, err := openTableCache(tableCachePath(filename), int64(size+64), int64(size+64), deadline)
	if err != nil {
		return nil
	}
	defer f.Close()
	reader := tableDeadlineReader{f, deadline}
	var header [64]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return nil
	}
	fingerprint := edgeMoveFingerprint()
	if !bytes.Equal(header[32:], fingerprint[:]) || tableDeadlineExceeded(deadline) {
		return nil
	}
	d := make([]uint8, size)
	hash := sha256.New()
	hash.Write(header[32:])
	for offset := 0; offset < len(d); {
		if tableDeadlineExceeded(deadline) {
			return nil
		}
		n := min(1<<20, len(d)-offset)
		chunk := d[offset : offset+n]
		if _, err := io.ReadFull(reader, chunk); err != nil {
			return nil
		}
		hash.Write(chunk)
		offset += n
	}
	var extra [1]byte
	if n, err := reader.Read(extra[:]); n != 0 || err != io.EOF {
		return nil
	}
	if !bytes.Equal(header[:32], hash.Sum(nil)) || d[0]&15 != 0 || tableDeadlineExceeded(deadline) {
		return nil
	}
	for i, value := range d {
		if i&65535 == 0 && tableDeadlineExceeded(deadline) {
			return nil
		}
		if value&15 == 15 || value>>4 == 15 {
			return nil
		}
	}
	return d
}

// Deep optimal searches can reuse an existing native phase-one cache, but do
// not start a second expensive generation when that optional cache is absent.
func cachedPhase1PatternTables(t *coordinateTables, deadline time.Time) *phase1Patterns {
	if db := phase1LargeDB.Load(); db != nil {
		return db
	}
	if runtime.GOARCH == "wasm" || tableDeadlineExceeded(deadline) {
		return nil
	}
	db := phase1SymmetryCoordinates(t)
	db.distance = loadPackedPatternLimit("phase1-sym8-v1.bin", len(db.reps)*flipSliceStates/2, deadline)
	if db.distance == nil {
		return nil
	}
	phase1LargeDB.CompareAndSwap(nil, db)
	return phase1LargeDB.Load()
}

func savePackedPattern(filename string, data []uint8) error {
	return writeTableCache(tableCachePath(filename), "pattern-*.bin", time.Time{}, func(f *os.File) error {
		fingerprint := edgeMoveFingerprint()
		hash := sha256.New()
		hash.Write(fingerprint[:])
		hash.Write(data)
		for _, part := range [][]byte{hash.Sum(nil), fingerprint[:], data} {
			if _, err := f.Write(part); err != nil {
				return err
			}
		}
		return nil
	})
}

// BuildLargePhase1Tables is an explicit opt-in to the optional 140.67 MiB
// database. Ordinary two-phase solves never call this builder.
func BuildLargePhase1Tables() (int, error) {
	if runtime.GOARCH == "wasm" || os.Getenv("CUBE_SMALL_TABLES") == "1" {
		return 0, fmt.Errorf("large phase-one tables are unavailable on this platform")
	}
	// Fail before spending minutes generating a table that cannot be persisted.
	f, err := createTableCache(tableCachePath("phase1-sym8-v1.bin"), "pattern-*.bin", time.Time{})
	if err != nil {
		return 0, fmt.Errorf("persist large phase-one table: %w", err)
	}
	closeErr := f.Close()
	os.Remove(f.Name())
	if closeErr != nil {
		return 0, fmt.Errorf("persist large phase-one table: %w", closeErr)
	}
	db, err := phase1PatternTablesPersist(solverTables(), true)
	if err != nil {
		return 0, fmt.Errorf("persist large phase-one table: %w", err)
	}
	return len(db.distance) + 64, nil
}
