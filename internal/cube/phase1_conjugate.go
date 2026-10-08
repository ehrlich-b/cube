package cube

import "time"

func phase1ConjugateCoordinates(deadline time.Time) ([]uint16, []uint16) {
	frame := NewCube(3)
	frame.ApplyMove(Move{Rotation: Y_Rotation, Clockwise: true})
	sym := readCubie(frame)
	inv := sym.inverse()
	twist, flip := make([]uint16, 2187), make([]uint16, 2048*495)
	for x := range twist {
		if x&255 == 0 && tableDeadlineExceeded(deadline) {
			return nil, nil
		}
		twist[x] = uint16(inv.mul(twistCubie(x)).mul(sym).twist())
	}
	for eo := 0; eo < 2048; eo++ {
		if tableDeadlineExceeded(deadline) {
			return nil, nil
		}
		orientation := flipCubie(eo).eo
		for sl := 0; sl < 495; sl++ {
			s := sliceCubie(sl)
			s.eo = orientation
			flip[eo*495+sl] = uint16(inv.mul(s).mul(sym).flip())
		}
	}
	return twist, flip
}
