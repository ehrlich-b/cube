//go:build !js || !wasm

package cube

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func fallbackCoordinateTables(deadline time.Time) *coordinateTables {
	return generateCoordinateTables(deadline)
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
	f, size, err := openTableCache(coordinateCachePath(), 64, 10<<20, deadline)
	if err != nil {
		return nil
	}
	defer f.Close()
	// Allocate only the checked size, even if the file grows after Stat.
	data := make([]byte, int(size))
	reader := tableDeadlineReader{f, deadline}
	if _, err := io.ReadFull(reader, data); err != nil || tableDeadlineExceeded(deadline) {
		return nil
	}
	var extra [1]byte
	if n, err := reader.Read(extra[:]); n != 0 || err != io.EOF {
		return nil
	}
	return decodeCoordinateTables(data, deadline)
}

func decodeCoordinateTables(data []byte, deadline time.Time) *coordinateTables {
	payload := coordinatePayload(data, deadline)
	if payload == nil {
		return nil
	}
	var t coordinateTables
	if gob.NewDecoder(bytes.NewReader(payload)).Decode(&t) != nil || tableDeadlineExceeded(deadline) {
		return nil
	}
	return validateCoordinateTables(&t, deadline)
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

func saveCoordinateTables(t *coordinateTables) error {
	return saveCoordinateBytes(encodeCoordinateTables(t))
}

func saveCoordinateBytes(data []byte) error {
	if data == nil {
		return fmt.Errorf("cannot encode coordinate tables")
	}
	return writeTableCache(coordinateCachePath(), "coordinates-*.gz", time.Time{}, func(f *os.File) error {
		_, err := f.Write(data)
		return err
	})
}
