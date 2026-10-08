package cube

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"reflect"
	"testing"
	"time"
)

func TestBrowserCoordinateAsset(t *testing.T) {
	data, err := ExportBrowserCoordinates()
	if err != nil {
		t.Fatal(err)
	}
	want := decodeCoordinateTables(embeddedCoordinates, time.Time{})
	got := decodeBrowserCoordinates(data, time.Time{})
	if got == nil || !reflect.DeepEqual(got, want) {
		t.Fatal("browser coordinates differ from verified native tables")
	}
	for _, damaged := range [][]byte{nil, data[:63], data[:len(data)-1], append(append([]byte(nil), data...), 0)} {
		if decodeBrowserCoordinates(damaged, time.Time{}) != nil {
			t.Fatal("malformed browser coordinates accepted")
		}
	}
	if decodeBrowserCoordinates(data, time.Now().Add(-time.Second)) != nil {
		t.Fatal("expired browser table load accepted")
	}
	z, err := gzip.NewReader(bytes.NewReader(data[64:]))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := io.ReadAll(z)
	z.Close()
	if err != nil {
		t.Fatal(err)
	}
	// Re-sign malformed payloads to test the parser beyond its hash gate.
	for _, mutate := range []func([]byte){
		func(p []byte) { p[0] ^= 1 },
		func(p []byte) { binary.LittleEndian.PutUint32(p[4:], ^uint32(0)) },
		func(p []byte) { binary.LittleEndian.PutUint32(p[4:], 0) },
	} {
		broken := append([]byte(nil), payload...)
		mutate(broken)
		var b bytes.Buffer
		writer := gzip.NewWriter(&b)
		writer.Write(broken)
		writer.Close()
		body := append(append([]byte(nil), data[32:64]...), b.Bytes()...)
		sum := sha256.Sum256(body)
		if decodeBrowserCoordinates(append(sum[:], body...), time.Time{}) != nil {
			t.Fatal("re-signed invalid browser schema accepted")
		}
	}
}
