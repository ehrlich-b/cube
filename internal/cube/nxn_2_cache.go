//go:build !wasm

package cube

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/gob"
	"io"
	"os"
	"time"
)

// The 2x2 keeps the exact synthetic-3x3 search and move order. Its short
// searches otherwise spend most of a fresh process inflating the compact
// asset. An optional stored-block gzip cache trades disk space for loading
// time without adding another embedded asset or changing other sizes' loads.
func nxnTwoByTwoTables() {
	tablesLock <- struct{}{}
	if tables != nil {
		<-tablesLock
		return
	}
	if cached := loadTwoByTwoTables(); cached != nil {
		tables = cached
		<-tablesLock
		return
	}
	<-tablesLock
	t := solverTables()
	_ = saveTwoByTwoTables(t) // Optional: corrupt/unwritable caches still solve.
}

func twoByTwoCachePath() string {
	return tableCachePath("coordinates-2x2-v1.bin")
}

func loadTwoByTwoTables() *coordinateTables {
	f, size, err := openTableCache(twoByTwoCachePath(), 96, 10<<20, time.Time{})
	if err != nil {
		return nil
	}
	defer f.Close()
	data := make([]byte, int(size))
	if _, err := io.ReadFull(f, data); err != nil {
		return nil
	}
	var extra [1]byte
	if n, err := f.Read(extra[:]); n != 0 || err != io.EOF {
		return nil
	}
	// Bind to the exact compact asset, as well as the decoder's checksum,
	// move fingerprint and complete table-dimension validation.
	if !bytes.Equal(data[:32], embeddedCoordinates[:32]) {
		return nil
	}
	return decodeCoordinateTables(data[32:], time.Time{})
}

func saveTwoByTwoTables(t *coordinateTables) error {
	var b bytes.Buffer
	z, err := gzip.NewWriterLevel(&b, gzip.NoCompression)
	if err != nil {
		return err
	}
	if err := gob.NewEncoder(z).Encode(t); err != nil {
		z.Close()
		return err
	}
	if err := z.Close(); err != nil {
		return err
	}
	fingerprint := edgeMoveFingerprint()
	payload := append(fingerprint[:], b.Bytes()...)
	sum := sha256.Sum256(payload)
	return writeTableCache(twoByTwoCachePath(), "coordinates-2x2-*.bin", time.Time{}, func(f *os.File) error {
		for _, part := range [][]byte{embeddedCoordinates[:32], sum[:], payload} {
			if _, err := f.Write(part); err != nil {
				return err
			}
		}
		return nil
	})
}
