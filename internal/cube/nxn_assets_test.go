package cube

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func encodeReductionTables(t *reductionTables) []byte {
	data := []byte{'N', 'X', 'N', '1', byte(t.size), byte(len(t.centers)), byte(len(t.wings)), 0, 0, 0, 0, 0}
	binary.LittleEndian.PutUint32(data[8:12], nxnTableSignature(t))
	word := func(v int) { data = append(data, byte(v), byte(v>>8)) }
	moves := func(moves []Move) {
		data = append(data, byte(len(moves)))
		for _, m := range moves {
			word(int(nxnMoveCode(m)))
		}
	}
	for _, orbits := range [][]*reductionOrbit{t.centers, t.wings} {
		for _, o := range orbits {
			for _, parent := range o.parent {
				word(int(uint16(parent)))
			}
			data = append(data, o.via...)
			inverse := make([]byte, len(o.parent)/8)
			for i, flag := range o.inverseRoot {
				if flag {
					inverse[i/8] |= 1 << (i % 8)
				}
			}
			data = append(data, inverse...)
			var buffer [64]Move
			for key, parent := range o.parent {
				cost := 0
				if parent >= 0 {
					cost = len(o.cycleInto(t, key/576, key/24%24, key%24, buffer[:0]))
				}
				data = append(data, byte(cost))
			}
		}
	}
	var blocks []centerBlockAction
	if t.size >= 6 {
		blocks = nxnCenterBlockActions(t.size, t)
	}
	word(len(blocks))
	for _, a := range blocks {
		moves(a.moves)
		data = append(data, byte(len(a.trans)))
		for _, step := range a.trans {
			data = append(data, step.src, step.dst)
		}
	}
	for _, o := range t.wings {
		actions := nxnPairingActions(t.size, o)
		word(len(actions))
		for _, a := range actions {
			moves(a.moves)
			data = append(data, a.full[:]...)
			data = append(data, a.outer[:]...)
		}
	}
	patterns := 0
	if t.size == 5 {
		patterns = len(t.centers)
	}
	data = append(data, byte(patterns))
	if patterns != 0 {
		for _, o := range t.centers {
			db := centerPatternTable(t.size, t, o)
			for _, row := range db.next {
				previous := 0
				for _, next := range row {
					word(int(uint16(int16(int(next) - previous))))
					previous = int(next)
				}
			}
			for _, locked := range []bool{false, true} {
				allowed := nxnAllowedCenterMoves(t, locked)
				for face := Front; face <= Down; face++ {
					data = append(data, db.distances(face, allowed)...)
				}
			}
		}
	}
	return data
}

func TestNxNEmbeddedTables(t *testing.T) {
	assets := [8][]byte{4: nxnAsset4, 5: nxnAsset5, 6: nxnAsset6, 7: nxnAsset7}
	for _, n := range []int{4, 5, 6, 7} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			fresh, err := buildReductionTables(n)
			if err != nil {
				t.Fatal(err)
			}
			want := encodeReductionTables(fresh)
			if os.Getenv("CUBE_GENERATE_NXN") == "1" {
				var compressed bytes.Buffer
				z, err := gzip.NewWriterLevel(&compressed, gzip.BestCompression)
				if err != nil {
					t.Fatal(err)
				}
				z.Extra = make([]byte, 4)
				binary.LittleEndian.PutUint32(z.Extra, uint32(len(want)))
				if _, err := z.Write(want); err != nil {
					t.Fatal(err)
				}
				if err := z.Close(); err != nil {
					t.Fatal(err)
				}
				name := filepath.Join("tables", fmt.Sprintf("nxn-%d-v1.bin.gz", n))
				if err := os.WriteFile(name, compressed.Bytes(), 0644); err != nil {
					t.Fatal(err)
				}
				t.Logf("%s: %d raw bytes, %d compressed", name, len(want), compressed.Len())
				return
			}
			z, err := gzip.NewReader(bytes.NewReader(assets[n]))
			if err != nil {
				t.Fatal(err)
			}
			got, err := io.ReadAll(z)
			closeErr := z.Close()
			if err != nil || closeErr != nil {
				t.Fatalf("read embedded tables: %v / %v", err, closeErr)
			}
			if !bytes.Equal(got, want) {
				t.Fatal("embedded NxN tables differ from complete independent regeneration")
			}
			loaded, err := nxnBaseTables(n)
			if err != nil {
				t.Fatal(err)
			}
			if err := decodeReductionTables(loaded, got); err != nil {
				t.Fatal(err)
			}
			for orbit, db := range loaded.patterns {
				freshDB := centerPatternTable(n, fresh, fresh.centers[orbit])
				if !reflect.DeepEqual(db.next, freshDB.next) || !reflect.DeepEqual(db.goal, freshDB.goal) {
					t.Fatal("decoded pattern transitions differ from regeneration")
				}
				for _, locked := range []bool{false, true} {
					allowed := nxnAllowedCenterMoves(fresh, locked)
					for face := Front; face <= Down; face++ {
						if !bytes.Equal(db.distances(face, allowed), freshDB.distances(face, allowed)) {
							t.Fatal("decoded pattern distances differ from fresh BFS")
						}
					}
				}
			}
			for _, broken := range [][]byte{got[:11], got[:len(got)-1], append(append([]byte(nil), got...), 0)} {
				if err := decodeReductionTables(loaded, broken); err == nil {
					t.Fatal("malformed reduction asset accepted")
				}
			}
		})
	}
}
