package cube

import "time"

var cornerMasks, cornerRanks = func() ([70]uint8, [256]uint8) {
	var masks [70]uint8
	var ranks [256]uint8
	n := 0
	for mask := 0; mask < 256; mask++ {
		bits := 0
		for x := mask; x != 0; x &= x - 1 {
			bits++
		}
		if bits == 4 {
			masks[n], ranks[mask] = uint8(mask), uint8(n)
			n++
		}
	}
	return masks, ranks
}()

// Track the positions of four U-layer corners and the full corner parity.
// Paired with the eight non-slice edges this preserves correlations omitted
// by either permutation/slice database, while using only 5,644,800 bytes.
func cornerCombination(s cubie) int {
	mask, parity := 0, 0
	for i, p := range s.cp {
		if p < 4 {
			mask |= 1 << i
		}
		for _, q := range s.cp[i+1:] {
			if p > q {
				parity ^= 1
			}
		}
	}
	return int(cornerRanks[mask])*2 + parity
}

func cornerCombinationMoves() []uint16 {
	return makeMoveTable(140, func(x int) cubie {
		s := identityCubie()
		a, b := uint8(0), uint8(4)
		for i := range s.cp {
			if cornerMasks[x/2]&(1<<i) != 0 {
				s.cp[i] = a
				a++
			} else {
				s.cp[i] = b
				b++
			}
		}
		if cornerCombination(s)%2 != x%2 {
			for i, p := range s.cp {
				if p == 0 {
					s.cp[i] = 1
				} else if p == 1 {
					s.cp[i] = 0
				}
			}
		}
		return s
	}, cornerCombination, true, time.Time{})
}

func phase2PermutationCoordinates(deadline time.Time) ([]uint8, []uint16, []uint16) {
	comb, inverse, slice := make([]uint8, 40320), make([]uint16, 40320), make([]uint16, 24)
	for x := range comb {
		if x&255 == 0 && tableDeadlineExceeded(deadline) {
			return nil, nil, nil
		}
		s := identityCubie()
		setPermutation(s.cp[:], x)
		inv := s.inverse()
		comb[x], inverse[x] = uint8(cornerCombination(s)), uint16(permutationRank(inv.cp[:]))
	}
	for x := range slice {
		var p, inv [4]uint8
		setPermutation(p[:], x)
		for i, v := range p {
			inv[v] = uint8(i)
		}
		slice[x] = uint16(permutationRank(inv[:]))
	}
	return comb, inverse, slice
}
