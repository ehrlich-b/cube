package cube

import (
	"fmt"
	"strconv"
	"strings"
)

// Physical case coordinates, independent of algorithms_data.json, its moves,
// and its inverse recognition patterns. Numbering and factual piece states:
// https://github.com/cs0x7f/cstimer/blob/master/src/js/scramble/scramble_333_edit.js
// (f2l_map, oll_map, pll_map, accessed 2026-10-07). Nibbles run UR/UF/UL/UB
// and URF/UFL/ULB/UBR, least significant first, as in standard cubie notation.
// OLL fixes orientation only; PLL fixes permutation only; F2L fixes the FR
// corner/edge locations and orientations. No reference algorithm is replayed.
var standardOLL = [58][2]uint16{
	{}, {0x1111, 0x1212}, {0x1111, 0x1122}, {0x1111, 0x0222}, {0x1111, 0x0111},
	{0x0011, 0x2022}, {0x0011, 0x1011}, {0x0011, 0x2202}, {0x0011, 0x0111},
	{0x0011, 0x1110}, {0x0011, 0x2220}, {0x0011, 0x0222}, {0x0011, 0x1101},
	{0x0101, 0x2022}, {0x0101, 0x0111}, {0x0101, 0x0222}, {0x0101, 0x1011},
	{0x1111, 0x0102}, {0x1111, 0x0012}, {0x1111, 0x0021}, {0x1111, 0x0000},
	{0x0000, 0x1212}, {0x0000, 0x1122}, {0x0000, 0x0012}, {0x0000, 0x0021},
	{0x0000, 0x0102}, {0x0000, 0x0111}, {0x0000, 0x0222}, {0x0011, 0x0000},
	{0x0011, 0x0210}, {0x0011, 0x2100}, {0x0011, 0x0021}, {0x0011, 0x1002},
	{0x0101, 0x0021}, {0x0101, 0x0210}, {0x0011, 0x1020}, {0x0011, 0x0102},
	{0x0011, 0x2010}, {0x0011, 0x0201}, {0x0101, 0x1020}, {0x0101, 0x0102},
	{0x0011, 0x1200}, {0x0011, 0x0120}, {0x0011, 0x0012}, {0x0011, 0x2001},
	{0x0101, 0x0012}, {0x0101, 0x0120}, {0x0011, 0x1221}, {0x0011, 0x1122},
	{0x0011, 0x2112}, {0x0011, 0x2211}, {0x0101, 0x1221}, {0x0101, 0x1122},
	{0x0011, 0x2121}, {0x0011, 0x1212}, {0x0101, 0x2121}, {0x0101, 0x1212},
	{0x0101, 0x0000},
}

var standardPLL = map[string][2]uint16{
	"H": {0x1032, 0x3210}, "Ua": {0x3102, 0x3210}, "Ub": {0x3021, 0x3210}, "Z": {0x2301, 0x3210},
	"Aa": {0x3210, 0x3021}, "Ab": {0x3210, 0x3102}, "E": {0x3210, 0x2301}, "F": {0x3012, 0x3201},
	"Ga": {0x2130, 0x3021}, "Gb": {0x1320, 0x3102}, "Gc": {0x3021, 0x3102}, "Gd": {0x3102, 0x3021},
	"Ja": {0x3201, 0x3201}, "Jb": {0x3120, 0x3201}, "Na": {0x1230, 0x3012}, "Nb": {0x3012, 0x3012},
	"Ra": {0x0213, 0x3201}, "Rb": {0x2310, 0x3201}, "T": {0x1230, 0x3201}, "V": {0x3120, 0x3012},
	"Y": {0x3201, 0x3012},
}

// Nibbles are corner twist, corner position, edge flip, edge position.
var standardF2L = [42]uint16{
	0, 0x2000, 0x1011, 0x2012, 0x1003, 0x2003, 0x1012, 0x2002, 0x1013,
	0x2013, 0x1002, 0x2010, 0x1001, 0x2011, 0x1000, 0x2001, 0x1010,
	0x0000, 0x0011, 0x0003, 0x0012, 0x0002, 0x0013, 0x0001, 0x0010,
	0x0400, 0x0411, 0x1400, 0x2411, 0x1411, 0x2400, 0x0018, 0x0008,
	0x2008, 0x1008, 0x2018, 0x1018, 0x0418, 0x1408, 0x2408, 0x1418, 0x2418,
}

func standardCaseState(id string) (cubie, string, error) {
	s := identityCubie()
	id = strings.TrimSuffix(id, "-CSV")
	category, name, _ := strings.Cut(id, "-")
	switch category {
	case "PLL":
		mask, ok := standardPLL[name]
		if !ok {
			break
		}
		for i := 0; i < 4; i++ {
			s.ep[i], s.cp[i] = uint8(mask[0]>>(i*4)&15), uint8(mask[1]>>(i*4)&15)
		}
		return s, category, nil
	case "OLL":
		// Historical Cross OLL is the full OLL-45 algorithm, not a universal
		// dot/line/L solver. Retain this lookup alias for older clients.
		if name == "CROSS" {
			name = "45"
		}
		n, err := strconv.Atoi(name)
		if err != nil || n < 1 || n > 57 {
			break
		}
		for i := 0; i < 4; i++ {
			s.eo[i], s.co[i] = uint8(standardOLL[n][0]>>(i*4)&15), uint8(standardOLL[n][1]>>(i*4)&15)
		}
		return s, category, nil
	case "F2L":
		n, err := strconv.Atoi(name)
		if err != nil || n < 1 || n > 41 {
			break
		}
		mask := standardF2L[n]
		cp, co, ep, eo := int(mask>>8&15), uint8(mask>>12), int(mask&15), uint8(mask>>4&1)
		s.cp[cp], s.cp[4] = s.cp[4], s.cp[cp]
		s.ep[ep], s.ep[8] = s.ep[8], s.ep[ep]
		s.co[cp], s.eo[ep] = co, eo
		// Complete unspecified top pieces into a physically legal cube, with
		// all three other pairs and the cross solved. Completion is independent
		// of the candidate algorithm and leaves the target piece untouched.
		corner, edge := (cp+1)%4, (ep+1)%4
		s.co[corner], s.eo[edge] = (3-co)%3, eo
		if (cp != 4) != (ep != 8) {
			other := (corner + 1) % 4
			if other == cp {
				other = (other + 1) % 4
			}
			s.cp[corner], s.cp[other] = s.cp[other], s.cp[corner]
		}
		return s, category, nil
	}
	return cubie{}, category, fmt.Errorf("no independent standard fixture for %s", id)
}

// VerifyStandardCase physically applies an algorithm to a named fixture. AUF
// and yaw are allowed setups; PLL also allows a final AUF. First-layer pieces
// must remain solved. It never reads Algorithm.Pattern or derives an inverse.
func VerifyStandardCase(a Algorithm, id string) error {
	s, category, err := standardCaseState(id)
	if err != nil {
		return err
	}
	moves, err := ParseMoves(a.Moves)
	if err != nil {
		return err
	}
	fixture := cfopCube(s)
	for yaw := 0; yaw < 4; yaw++ {
		for auf := 0; auf < 4; auf++ {
			c := fixture.clone()
			c.ApplyMoves(yawMoves(yaw))
			c.ApplyMoves(aufMoves(auf))
			c.ApplyMoves(moves)
			c, _, err = canonical3x3(c)
			if err != nil || !WhiteCrossSolved(c) || cfopSlots(c) != 15 {
				continue
			}
			if category == "F2L" || category == "OLL" && OLLSolved(c) {
				return nil
			}
			if category == "PLL" {
				for end := 0; end < 4; end++ {
					if c.IsSolved() {
						return nil
					}
					c.ApplyMoves(aufMoves(1))
				}
			}
		}
	}
	return fmt.Errorf("algorithm does not solve independent standard %s fixture under yaw/AUF", id)
}

// VerifyAlgorithmCases checks primary IDs and case-ID aliases in every CFOP
// category, including records merged across categories by the importer.
func VerifyAlgorithmCases(a Algorithm) error {
	if a.Dimension != 3 {
		return nil
	}
	for _, category := range []string{"OLL", "PLL", "F2L"} {
		if !a.HasCategory(category) {
			continue
		}
		found := false
		for _, id := range append([]string{a.CaseID}, a.Aliases...) {
			if !strings.HasPrefix(id, category+"-") {
				continue
			}
			found = true
			if err := VerifyStandardCase(a, id); err != nil {
				return err
			}
		}
		if !found {
			return fmt.Errorf("missing independent %s case ID", category)
		}
	}
	return nil
}
