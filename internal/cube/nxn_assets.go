package cube

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
)

func nxnMoveCode(m Move) uint16 {
	code := uint16(m.Face) | uint16(m.Rotation)<<6 | uint16(m.Layer)<<8 | uint16(m.WideDepth)<<11 | uint16(m.Slice)<<14
	if m.Clockwise {
		code |= 1 << 3
	}
	if m.Double {
		code |= 1 << 4
	}
	if m.Wide {
		code |= 1 << 5
	}
	return code
}

// Bind the asset to the ordered generators, orbit coordinates and seed cycles.
func nxnTableSignature(t *reductionTables) uint32 {
	h := uint32(2166136261)
	add := func(v uint16) {
		h = (h ^ uint32(v&255)) * 16777619
		h = (h ^ uint32(v>>8)) * 16777619
	}
	for _, m := range t.moves {
		add(nxnMoveCode(m))
	}
	for _, orbits := range [][]*reductionOrbit{t.centers, t.wings} {
		for _, o := range orbits {
			for _, pos := range o.positions {
				add(uint16(pos))
			}
			for _, m := range o.cycle {
				add(nxnMoveCode(m))
			}
		}
	}
	return h
}

type nxnAssetReader struct {
	data []byte
	pos  int
	err  error
}

func (r *nxnAssetReader) take(n int) []byte {
	if r.err != nil {
		return nil
	}
	if n > len(r.data)-r.pos {
		r.err = fmt.Errorf("truncated reduction asset at byte %d", r.pos)
		return nil
	}
	data := r.data[r.pos : r.pos+n]
	r.pos += n
	return data
}

func (r *nxnAssetReader) byte() int {
	if data := r.take(1); data != nil {
		return int(data[0])
	}
	return 0
}

func (r *nxnAssetReader) word() int {
	if data := r.take(2); data != nil {
		return int(binary.LittleEndian.Uint16(data))
	}
	return 0
}

func (r *nxnAssetReader) moves(n int) []Move {
	moves := make([]Move, r.byte())
	for i := range moves {
		code := uint16(r.word())
		m := Move{Face: Face(code & 7), Clockwise: code&8 != 0, Double: code&16 != 0, Wide: code&32 != 0,
			Rotation: RotationType(code >> 6 & 3), Layer: int(code >> 8 & 7), WideDepth: int(code >> 11 & 7), Slice: SliceType(code >> 14)}
		if err := m.Validate(n); err != nil {
			r.err = fmt.Errorf("invalid reduction asset move: %w", err)
		}
		moves[i] = m
	}
	if len(moves) == 0 && r.err == nil {
		r.err = fmt.Errorf("empty reduction asset action")
	}
	return moves
}

func loadReductionTables(n int) (*reductionTables, error) {
	t, err := nxnBaseTables(n)
	if err != nil {
		return nil, err
	}
	asset := [8][]byte{4: nxnAsset4, 5: nxnAsset5, 6: nxnAsset6, 7: nxnAsset7}[n]
	z, err := gzip.NewReader(bytes.NewReader(asset))
	if err != nil {
		return nil, fmt.Errorf("dimension %d: reduction asset: %w", n, err)
	}
	if len(z.Extra) != 4 {
		z.Close()
		return nil, fmt.Errorf("dimension %d: reduction asset length missing", n)
	}
	length := binary.LittleEndian.Uint32(z.Extra)
	if length < 12 || length > 8<<20 {
		z.Close()
		return nil, fmt.Errorf("dimension %d: invalid reduction asset length", n)
	}
	data := make([]byte, length)
	_, err = io.ReadFull(z, data)
	if err == nil {
		var tail [1]byte
		written, readErr := z.Read(tail[:])
		if written != 0 || readErr != io.EOF {
			err = fmt.Errorf("dimension %d: reduction asset length mismatch", n)
		}
	}
	closeErr := z.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if err := decodeReductionTables(t, data); err != nil {
		return nil, fmt.Errorf("dimension %d: %w", n, err)
	}
	return t, nil
}

func decodeReductionTables(t *reductionTables, data []byte) error {
	if len(data) < 12 || string(data[:4]) != "NXN1" || int(data[4]) != t.size ||
		int(data[5]) != len(t.centers) || int(data[6]) != len(t.wings) ||
		binary.LittleEndian.Uint32(data[8:12]) != nxnTableSignature(t) {
		return fmt.Errorf("reduction asset does not match generators and orbits")
	}
	r := nxnAssetReader{data: data, pos: 12}
	const coordinates = 24 * 24 * 24
	for _, orbits := range [][]*reductionOrbit{t.centers, t.wings} {
		for _, o := range orbits {
			o.parent = make([]int16, coordinates)
			for i := range o.parent {
				o.parent[i] = int16(r.word())
				if o.parent[i] < -1 || int(o.parent[i]) >= coordinates {
					return fmt.Errorf("invalid reduction asset setup parent")
				}
			}
			o.via = append([]uint8(nil), r.take(coordinates)...)
			inverse := r.take(coordinates / 8)
			if r.err != nil {
				return r.err
			}
			o.inverseRoot = make([]bool, coordinates)
			for i := range o.inverseRoot {
				o.inverseRoot[i] = inverse[i/8]&(1<<(i%8)) != 0
			}
			o.costs = append([]uint8(nil), r.take(coordinates)...)
			if r.err != nil {
				return r.err
			}
			for key, parent := range o.parent {
				if parent >= 0 && (int(o.via[key]) >= len(t.moves) || o.costs[key] == 0) {
					return fmt.Errorf("invalid reduction asset setup move or cost")
				}
			}
		}
	}
	var targets []Color
	home := NewCube(t.size)
	for _, o := range t.centers {
		for _, pos := range o.positions {
			targets = append(targets, nxnColor(home, pos))
		}
	}
	t.blocks = make([]centerBlockAction, r.word())
	for i := range t.blocks {
		a := &t.blocks[i]
		a.moves = r.moves(t.size)
		a.trans = make([]centerTransfer, r.byte())
		for j := range a.trans {
			src, dst := r.byte(), r.byte()
			if src >= len(targets) || dst >= len(targets) {
				return fmt.Errorf("invalid reduction asset center transfer")
			}
			a.trans[j] = centerTransfer{uint8(src), uint8(dst)}
		}
		a.buildMasks(targets)
	}
	for _, o := range t.wings {
		o.actions = make([]wingAction, r.word())
		for i := range o.actions {
			a := &o.actions[i]
			a.moves = r.moves(t.size)
			copy(a.full[:], r.take(24))
			copy(a.outer[:], r.take(24))
			for _, perm := range [2][24]uint8{a.full, a.outer} {
				var seen [24]bool
				for _, dst := range perm {
					if dst >= 24 || seen[dst] {
						return fmt.Errorf("invalid reduction asset wing permutation")
					}
					seen[dst] = true
				}
			}
			a.prepare(o.mate)
		}
		o.actionsOnce.Do(func() {})
	}
	patterns := r.byte()
	if patterns != 0 && (t.size != 5 || patterns != len(t.centers)) {
		return fmt.Errorf("invalid reduction asset center patterns")
	}
	if patterns != 0 {
		t.patterns = make([]*centerPatterns, patterns)
		for orbit, o := range t.centers {
			db := centerPatternShape(t.size, t, o)
			for g := range db.next {
				raw := r.take(2 * nxnCenterCoordinates)
				if r.err != nil {
					return r.err
				}
				row := make([]uint16, nxnCenterCoordinates)
				previous := 0
				for key := range row {
					next := previous + int(int16(binary.LittleEndian.Uint16(raw[key*2:])))
					if next < 0 || next >= nxnCenterCoordinates {
						return fmt.Errorf("invalid reduction asset pattern transition")
					}
					row[key], previous = uint16(next), next
				}
				db.next[g] = row
			}
			db.distCache = make(map[string][]uint8, 12)
			for _, locked := range []bool{false, true} {
				allowed := nxnAllowedCenterMoves(t, locked)
				for face := Front; face <= Down; face++ {
					db.distCache[centerDistanceKey(face, allowed)] = append([]uint8(nil), r.take(nxnCenterCoordinates)...)
				}
			}
			t.patterns[orbit] = db
		}
		t.patternsOnce.Do(func() {})
	}
	if r.err != nil {
		return r.err
	}
	if r.pos != len(data) {
		return fmt.Errorf("trailing reduction asset data")
	}
	t.blockInfluences = nxnCenterInfluences(t.blocks)
	t.blocksOnce.Do(func() {})
	return nil
}
