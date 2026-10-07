package cube

// cubie coordinates use exactly the ordered facelets of physical validation.
// A move maps destination slots to source slots, with additive orientations.
type cubie struct {
	cp, co [8]uint8
	ep, eo [12]uint8
}

func identityCubie() cubie {
	var s cubie
	for i := range s.cp {
		s.cp[i] = uint8(i)
	}
	for i := range s.ep {
		s.ep[i] = uint8(i)
	}
	return s
}

// readCubie requires a validated cube whose centers are in the canonical frame.
func readCubie(c *Cube) cubie {
	home := NewCube(3)
	var s cubie
	for slot, coords := range cornerFacelets {
		for id, target := range cornerFacelets {
			for twist := 0; twist < 3; twist++ {
				if sticker(c, coords[twist]) == sticker(home, target[0]) &&
					sticker(c, coords[(twist+1)%3]) == sticker(home, target[1]) &&
					sticker(c, coords[(twist+2)%3]) == sticker(home, target[2]) {
					s.cp[slot], s.co[slot] = uint8(id), uint8(twist)
				}
			}
		}
	}
	for slot, coords := range edgeFacelets {
		for id, target := range edgeFacelets {
			for flip := 0; flip < 2; flip++ {
				if sticker(c, coords[flip]) == sticker(home, target[0]) &&
					sticker(c, coords[1-flip]) == sticker(home, target[1]) {
					s.ep[slot], s.eo[slot] = uint8(id), uint8(flip)
				}
			}
		}
	}
	return s
}

func (s cubie) mul(m cubie) cubie {
	var n cubie
	for i, p := range m.cp {
		n.cp[i], n.co[i] = s.cp[p], (s.co[p]+m.co[i])%3
	}
	for i, p := range m.ep {
		n.ep[i], n.eo[i] = s.ep[p], s.eo[p]^m.eo[i]
	}
	return n
}

func (s cubie) inverse() cubie {
	var n cubie
	for i, p := range s.cp {
		n.cp[p], n.co[p] = uint8(i), (3-s.co[i])%3
	}
	for i, p := range s.ep {
		n.ep[p], n.eo[p] = uint8(i), s.eo[i]
	}
	return n
}

var coordinateMoves = func() [18]Move {
	var moves [18]Move
	for f := Front; f <= Down; f++ {
		for p, m := range faceMoves(f) {
			moves[int(f)*3+p] = m
		}
	}
	return moves
}()

var cubieMoves = func() [18]cubie {
	var moves [18]cubie
	for i, m := range coordinateMoves {
		c := NewCube(3)
		c.ApplyMove(m)
		moves[i] = readCubie(c)
	}
	return moves
}()

func (s cubie) twist() int {
	x := 0
	for i := 0; i < 7; i++ {
		x = x*3 + int(s.co[i])
	}
	return x
}

func twistCubie(x int) cubie {
	s, sum := identityCubie(), 0
	for i := 6; i >= 0; i-- {
		s.co[i] = uint8(x % 3)
		sum += x % 3
		x /= 3
	}
	s.co[7] = uint8((3 - sum%3) % 3)
	return s
}

func (s cubie) flip() int {
	x := 0
	for i := 0; i < 11; i++ {
		x = x*2 + int(s.eo[i])
	}
	return x
}

func flipCubie(x int) cubie {
	s, sum := identityCubie(), 0
	for i := 10; i >= 0; i-- {
		s.eo[i] = uint8(x % 2)
		sum ^= x % 2
		x /= 2
	}
	s.eo[11] = uint8(sum)
	return s
}

// All 495 subsets of four edge positions, with the solved slice ranked zero.
var sliceMasks, sliceRanks = func() ([495]uint16, [4096]uint16) {
	var masks [495]uint16
	var ranks [4096]uint16
	n := 0
	for mask := 4095; mask >= 0; mask-- {
		count := 0
		for x := mask; x != 0; x &= x - 1 {
			count++
		}
		if count == 4 {
			masks[n] = uint16(mask)
			ranks[mask] = uint16(n)
			n++
		}
	}
	return masks, ranks
}()

func (s cubie) slice() int {
	mask := 0
	for i, e := range s.ep {
		if e >= 8 {
			mask |= 1 << i
		}
	}
	return int(sliceRanks[mask])
}

func sliceCubie(x int) cubie {
	s := identityCubie()
	a, b := uint8(0), uint8(8)
	for i := range s.ep {
		if sliceMasks[x]&(1<<i) != 0 {
			s.ep[i] = b
			b++
		} else {
			s.ep[i] = a
			a++
		}
	}
	return s
}

func permutationRank(p []uint8) int {
	x := 0
	for i := range p {
		x *= len(p) - i
		for j := i + 1; j < len(p); j++ {
			if p[j] < p[i] {
				x++
			}
		}
	}
	return x
}

func setPermutation(p []uint8, x int) {
	var digits [12]int
	for i := len(p) - 1; i >= 0; i-- {
		digits[i] = x % (len(p) - i)
		x /= len(p) - i
	}
	var available [12]uint8
	for i := range p {
		available[i] = uint8(i)
	}
	n := len(p)
	for i := range p {
		d := digits[i]
		p[i] = available[d]
		copy(available[d:], available[d+1:n])
		n--
	}
}

func skipCoordinateFace(move, prev int) bool {
	if prev < 0 {
		return false
	}
	f, p := Face(move/3), Face(prev/3)
	return f == p || (oppositeFaces(f, p) && f < p)
}

// The lesson's center-frame search records clockwise quarter rotations. A
// fast solver can express consecutive rotations as one half/inverse turn.
func compactGrip(rotations []Move) []Move {
	var moves []Move
	for _, m := range rotations {
		if len(moves) == 0 || moves[len(moves)-1].Rotation != m.Rotation {
			moves = append(moves, m)
			continue
		}
		prev := moves[len(moves)-1]
		turns := (moveToQuarterTurns(prev) + moveToQuarterTurns(m)) % 4
		moves = moves[:len(moves)-1]
		if turns != 0 {
			moves = append(moves, Move{Rotation: m.Rotation, Clockwise: turns == 1, Double: turns == 2})
		}
	}
	return moves
}
