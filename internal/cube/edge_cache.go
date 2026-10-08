package cube

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"os"
	"time"
)

const edgePatternSize = 190080
const edgeCacheBytes = 64 + edgePatternSize*(18*4+3)

func edgeCachePath() string {
	return tableCachePath("edges-v1.bin")
}

func edgeMoveFingerprint() [32]byte {
	var data []byte
	for _, m := range cubieMoves {
		data = append(data, m.cp[:]...)
		data = append(data, m.co[:]...)
		data = append(data, m.ep[:]...)
		data = append(data, m.eo[:]...)
	}
	return sha256.Sum256(data)
}

// A compact binary cache avoids regenerating the shared four-edge transitions
// in every CLI process. Checksums, dimensions, move fingerprints and bounds
// reject stale/corrupt data; optional I/O never prevents an in-memory rebuild.
func loadEdgePatterns(deadline time.Time) *edgePatterns {
	if tableDeadlineExceeded(deadline) {
		return nil
	}
	f, _, err := openTableCache(edgeCachePath(), edgeCacheBytes, edgeCacheBytes, deadline)
	if err != nil {
		return nil
	}
	defer f.Close()
	data := make([]byte, edgeCacheBytes)
	reader := tableDeadlineReader{f, deadline}
	if _, err = io.ReadFull(reader, data); err != nil || tableDeadlineExceeded(deadline) {
		return nil
	}
	var extra [1]byte
	if n, err := reader.Read(extra[:]); n != 0 || err != io.EOF {
		return nil
	}
	sum, fingerprint := sha256.Sum256(data[32:]), edgeMoveFingerprint()
	if !bytes.Equal(data[:32], sum[:]) || !bytes.Equal(data[32:64], fingerprint[:]) {
		return nil
	}
	db := &edgePatterns{moves: make([]uint32, edgePatternSize*18)}
	pos := 64
	for i := range db.moves {
		if i&16383 == 0 && tableDeadlineExceeded(deadline) {
			return nil
		}
		value := binary.LittleEndian.Uint32(data[pos:])
		if value >= edgePatternSize {
			return nil
		}
		db.moves[i] = value
		pos += 4
	}
	for g := range db.distance {
		db.distance[g] = append([]uint8(nil), data[pos:pos+edgePatternSize]...)
		if db.distance[g][edgeCoordinate(identityCubie(), g)] != 0 {
			return nil
		}
		for _, d := range db.distance[g] {
			if d > 12 {
				return nil
			}
		}
		pos += edgePatternSize
	}
	if tableDeadlineExceeded(deadline) {
		return nil
	}
	return db
}

func saveEdgePatterns(db *edgePatterns) error {
	data := make([]byte, edgeCacheBytes)
	fingerprint := edgeMoveFingerprint()
	copy(data[32:64], fingerprint[:])
	pos := 64
	for _, m := range db.moves {
		binary.LittleEndian.PutUint32(data[pos:], m)
		pos += 4
	}
	for _, distance := range db.distance {
		copy(data[pos:], distance)
		pos += len(distance)
	}
	sum := sha256.Sum256(data[32:])
	copy(data[:32], sum[:])
	return writeTableCache(edgeCachePath(), "edges-*.bin", time.Time{}, func(f *os.File) error {
		_, err := f.Write(data)
		return err
	})
}
