package cube

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"math/bits"
	"os"
	"time"
)

const (
	fullCornerStates    = 40320 * 2187
	sixEdgePermutations = 12 * 11 * 10 * 9 * 8 * 7
	sixEdgeStates       = sixEdgePermutations * 64
	optimalCacheBytes   = 64 + fullCornerStates/2 + sixEdgeStates + sixEdgePermutations*18*4
)

// Distances use four bits each. Six-edge moves separate the 12P6 positions
// from the 64 orientation masks, avoiding a 3 GB expanded transition table.
type optimalPatterns struct {
	corners []uint8
	edges   [2][]uint8
	moves   []uint32
}

var optimalLock = make(chan struct{}, 1)
var optimalDB *optimalPatterns

func nibbleDistance(data []uint8, x int) int { return int(data[x>>1] >> ((x & 1) * 4) & 15) }

func packDistances(data []uint8, deadline time.Time) []uint8 {
	if data == nil || tableDeadlineExceeded(deadline) {
		return nil
	}
	packed := make([]uint8, len(data)/2)
	for i := range packed {
		if i&65535 == 0 && tableDeadlineExceeded(deadline) {
			return nil
		}
		if data[i*2] >= 15 || data[i*2+1] >= 15 {
			return nil
		}
		packed[i] = data[i*2] | data[i*2+1]<<4
	}
	return packed
}

func sixEdgeRank(pos [6]uint8) int {
	x, used := 0, uint16(0)
	for i, p := range pos {
		x = x*(12-i) + int(p) - bits.OnesCount16(used&((1<<p)-1))
		used |= 1 << p
	}
	return x
}

func sixEdgeUnrank(x int) [6]uint8 {
	var digits [6]int
	for i := 5; i >= 0; i-- {
		digits[i] = x % (12 - i)
		x /= 12 - i
	}
	var pos [6]uint8
	used := uint16(0)
	for i, d := range digits {
		for p := uint8(0); p < 12; p++ {
			if used&(1<<p) != 0 {
				continue
			}
			if d == 0 {
				pos[i] = p
				used |= 1 << p
				break
			}
			d--
		}
	}
	return pos
}

func sixEdgeCoordinate(state cubie, group int) int {
	var pos [6]uint8
	flip := 0
	for p, id := range state.ep {
		if int(id)/6 == group {
			i := int(id) % 6
			pos[i] = uint8(p)
			flip |= int(state.eo[p]) << i
		}
	}
	return sixEdgeRank(pos)*64 + flip
}

func sixEdgeMoves(deadline time.Time) []uint32 {
	if tableDeadlineExceeded(deadline) {
		return nil
	}
	moves := make([]uint32, sixEdgePermutations*18)
	var dest [18][12]uint8
	for m, state := range cubieMoves {
		for p, source := range state.ep {
			dest[m][source] = uint8(p)
		}
	}
	for x := 0; x < sixEdgePermutations; x++ {
		if x&255 == 0 && tableDeadlineExceeded(deadline) {
			return nil
		}
		pos := sixEdgeUnrank(x)
		for m, state := range cubieMoves {
			var next [6]uint8
			flip := 0
			for i, p := range pos {
				next[i] = dest[m][p]
				flip |= int(state.eo[next[i]]) << i
			}
			moves[x*18+m] = uint32(sixEdgeRank(next)*64 + flip)
		}
	}
	return moves
}

func sixEdgePruning(moves []uint32, group int, deadline time.Time) []uint8 {
	if moves == nil || tableDeadlineExceeded(deadline) {
		return nil
	}
	d := make([]uint8, sixEdgeStates)
	for i := range d {
		d[i] = 255
	}
	goal := sixEdgeCoordinate(identityCubie(), group)
	d[goal] = 0
	queue := make([]uint32, 1, sixEdgeStates)
	queue[0] = uint32(goal)
	for head := 0; head < len(queue); head++ {
		if head&1023 == 0 && tableDeadlineExceeded(deadline) {
			return nil
		}
		x := int(queue[head])
		flip := uint32(x & 63)
		for _, y := range moves[x/64*18 : x/64*18+18] {
			y ^= flip
			if d[y] == 255 {
				d[y] = d[x] + 1
				queue = append(queue, y)
			}
		}
	}
	if len(queue) != sixEdgeStates {
		return nil
	}
	return packDistances(d, deadline)
}

func optimalPatternTables(t *coordinateTables, deadline time.Time) *optimalPatterns {
	if !lockSearchTables(optimalLock, deadline) {
		return nil
	}
	defer func() { <-optimalLock }()
	if optimalDB != nil {
		return optimalDB
	}
	if cached := loadOptimalPatterns(deadline); cached != nil {
		optimalDB = cached
		return cached
	}
	fmt.Fprintln(os.Stderr, "Building optimal pattern databases once: 128.33 MiB of cache, about one minute; initialization counts toward --time-limit.")
	db := &optimalPatterns{}
	db.corners = packDistances(pairPruning(t.Corner, t.Twist, 2187, false, deadline), deadline)
	if db.corners == nil {
		return nil
	}
	db.moves = sixEdgeMoves(deadline)
	for g := range db.edges {
		db.edges[g] = sixEdgePruning(db.moves, g, deadline)
		if db.edges[g] == nil {
			return nil
		}
	}
	if tableDeadlineExceeded(deadline) {
		return nil
	}
	optimalDB = db
	// Persistence is optional, and only attempted with enough remaining time
	// to write the bounded cache. Incomplete builds are never published/saved.
	if deadline.IsZero() || time.Until(deadline) > time.Second {
		_ = saveOptimalPatterns(db, deadline) // Optional disk cache.
	}
	return db
}

func optimalCachePath() string {
	return tableCachePath("optimal-v1.bin")
}

func loadOptimalPatterns(deadline time.Time) *optimalPatterns {
	if tableDeadlineExceeded(deadline) {
		return nil
	}
	f, _, err := openTableCache(optimalCachePath(), optimalCacheBytes, optimalCacheBytes, deadline)
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
	hash := sha256.New()
	hash.Write(header[32:])
	data := make([]byte, fullCornerStates/2+sixEdgeStates)
	for offset := 0; offset < len(data); {
		if tableDeadlineExceeded(deadline) {
			return nil
		}
		n := min(1<<20, len(data)-offset)
		chunk := data[offset : offset+n]
		if _, err := io.ReadFull(reader, chunk); err != nil {
			return nil
		}
		hash.Write(chunk)
		offset += n
	}
	db := &optimalPatterns{corners: data[: fullCornerStates/2 : fullCornerStates/2]}
	pos := fullCornerStates / 2
	for g := range db.edges {
		db.edges[g] = data[pos : pos+sixEdgeStates/2 : pos+sixEdgeStates/2]
		pos += sixEdgeStates / 2
	}
	for i, d := range data {
		if i&65535 == 0 && tableDeadlineExceeded(deadline) {
			return nil
		}
		if d&15 == 15 || d>>4 == 15 {
			return nil
		}
	}
	if nibbleDistance(db.corners, 0) != 0 {
		return nil
	}
	for g := range db.edges {
		if nibbleDistance(db.edges[g], sixEdgeCoordinate(identityCubie(), g)) != 0 {
			return nil
		}
	}
	db.moves = make([]uint32, sixEdgePermutations*18)
	buffer := make([]byte, 65536*4)
	for offset := 0; offset < len(db.moves); {
		if tableDeadlineExceeded(deadline) {
			return nil
		}
		n := min(65536, len(db.moves)-offset)
		chunk := buffer[:n*4]
		if _, err := io.ReadFull(reader, chunk); err != nil {
			return nil
		}
		hash.Write(chunk)
		for i := 0; i < n; i++ {
			v := binary.LittleEndian.Uint32(chunk[i*4:])
			if v >= sixEdgeStates {
				return nil
			}
			db.moves[offset+i] = v
		}
		offset += n
	}
	var extra [1]byte
	if n, err := reader.Read(extra[:]); n != 0 || err != io.EOF {
		return nil
	}
	if !bytes.Equal(header[:32], hash.Sum(nil)) || tableDeadlineExceeded(deadline) {
		return nil
	}
	return db
}

func saveOptimalPatterns(db *optimalPatterns, deadline time.Time) error {
	return writeTableCache(optimalCachePath(), "optimal-*.bin", deadline, func(f *os.File) error {
		hash := sha256.New()
		fileWriter := tableDeadlineWriter{f, deadline}
		writer := io.MultiWriter(fileWriter, hash)
		if _, err := fileWriter.Write(make([]byte, 32)); err != nil {
			return err
		}
		fingerprint := edgeMoveFingerprint()
		if _, err := writer.Write(fingerprint[:]); err != nil {
			return err
		}
		for _, data := range [][]uint8{db.corners, db.edges[0], db.edges[1]} {
			if _, err := writer.Write(data); err != nil {
				return err
			}
		}
		buffer := make([]byte, 65536*4)
		for offset := 0; offset < len(db.moves); {
			if tableDeadlineExceeded(deadline) {
				return context.DeadlineExceeded
			}
			n := min(65536, len(db.moves)-offset)
			for i, v := range db.moves[offset : offset+n] {
				binary.LittleEndian.PutUint32(buffer[i*4:], v)
			}
			if _, err := writer.Write(buffer[:n*4]); err != nil {
				return err
			}
			offset += n
		}
		if tableDeadlineExceeded(deadline) {
			return context.DeadlineExceeded
		}
		_, err := f.WriteAt(hash.Sum(nil), 0)
		return err
	})
}
