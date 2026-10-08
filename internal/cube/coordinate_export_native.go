//go:build !js || !wasm

package cube

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"
)

// ExportBrowserCoordinates converts the checked native asset for the website.
// It never generates replacement tables or changes the native cache format.
func ExportBrowserCoordinates() ([]byte, error) {
	t := decodeCoordinateTables(embeddedCoordinates, time.Time{})
	if t == nil {
		return nil, fmt.Errorf("invalid native coordinate asset")
	}
	payload := []byte("CWB1")
	for _, field := range coordinateWordFields(t) {
		payload = binary.LittleEndian.AppendUint32(payload, uint32(len(*field)))
		for _, v := range *field {
			payload = binary.LittleEndian.AppendUint16(payload, v)
		}
	}
	for _, field := range coordinateByteFields(t) {
		payload = binary.LittleEndian.AppendUint32(payload, uint32(len(*field)))
		payload = append(payload, (*field)...)
	}
	payload = binary.LittleEndian.AppendUint32(payload, uint32(len(t.NearPhase1Keys)))
	for _, v := range t.NearPhase1Keys {
		payload = binary.LittleEndian.AppendUint32(payload, v)
	}
	var b bytes.Buffer
	z, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	if _, err := z.Write(payload); err != nil {
		return nil, err
	}
	if err := z.Close(); err != nil {
		return nil, err
	}
	fingerprint := edgeMoveFingerprint()
	data := append(fingerprint[:], b.Bytes()...)
	sum := sha256.Sum256(data)
	return append(sum[:], data...), nil
}
