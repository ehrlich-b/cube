package cube

import (
	"encoding/binary"
	"time"
)

// The browser format has a fixed schema, so it needs no gob type descriptors
// or reflection. Keep every coordinate; decoding uses the native validator.
func coordinateWordFields(t *coordinateTables) []*[]uint16 {
	return []*[]uint16{&t.TwistRepresentatives, &t.Twist, &t.Flip, &t.Slice, &t.Corner, &t.Edge2, &t.Slice2,
		&t.PermInverse, &t.SliceInverse, &t.PermSymCorner, &t.PermSymEdge, &t.SymSlice2, &t.SymCornerComb, &t.SymEdgeComb,
		&t.TwistSym, &t.SliceSym, &t.TwistFlipSym, &t.SymTwist, &t.SymFlip, &t.SymFlipDelta, &t.SymSlice}
}

func coordinateByteFields(t *coordinateTables) []*[]uint8 {
	return []*[]uint8{&t.NearPhase1Distances, &t.TwistSlice, &t.FlipSlice, &t.TwistFlip, &t.CornerSlice,
		&t.EdgeSlice, &t.CornerDistance, &t.CornerComb, &t.EdgeComb, &t.CornerEdgeComb}
}

func decodeBrowserCoordinates(data []byte, deadline time.Time) *coordinateTables {
	payload := coordinatePayload(data, deadline)
	if len(payload) < 4 || string(payload[:4]) != "CWB1" {
		return nil
	}
	payload = payload[4:]
	take := func(width int) []byte {
		if len(payload) < 4 {
			return nil
		}
		count := uint64(binary.LittleEndian.Uint32(payload))
		payload = payload[4:]
		length := count * uint64(width)
		if length > uint64(len(payload)) {
			return nil
		}
		data := payload[:int(length)]
		payload = payload[int(length):]
		return data
	}
	var t coordinateTables
	for _, field := range coordinateWordFields(&t) {
		data := take(2)
		if data == nil || tableDeadlineExceeded(deadline) {
			return nil
		}
		*field = make([]uint16, len(data)/2)
		for i := range *field {
			(*field)[i] = binary.LittleEndian.Uint16(data[2*i:])
		}
	}
	for _, field := range coordinateByteFields(&t) {
		data := take(1)
		if data == nil {
			return nil
		}
		*field = append([]uint8(nil), data...)
	}
	keys := take(4)
	if keys == nil || len(payload) != 0 || tableDeadlineExceeded(deadline) {
		return nil
	}
	t.NearPhase1Keys = make([]uint32, len(keys)/4)
	for i := range t.NearPhase1Keys {
		t.NearPhase1Keys[i] = binary.LittleEndian.Uint32(keys[4*i:])
	}
	return validateCoordinateTables(&t, deadline)
}
