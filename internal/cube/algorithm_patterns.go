package cube

import (
	"fmt"
	"strings"
)

// readAlgorithmPattern reads the generated concrete 3x3 YB subset of CFEN.
// The general CFEN package depends on cube, so the solver cannot import it.
// Import-time verification and corpus tests also exercise the general parser.
func readAlgorithmPattern(text string) (*Cube, error) {
	if len(text) > 128 || !strings.HasPrefix(text, "YB|") {
		return nil, fmt.Errorf("algorithm recognition needs concrete 3x3 YB CFEN")
	}
	fields := strings.Split(strings.TrimPrefix(text, "YB|"), "/")
	if len(fields) != 6 {
		return nil, fmt.Errorf("algorithm recognition needs six faces")
	}
	c := NewCube(3)
	order := [...]Face{Up, Right, Front, Down, Left, Back}
	for f, text := range fields {
		pos := 0
		for i := 0; i < len(text); {
			color := strings.IndexByte("WYROBG", text[i])
			i++
			if color < 0 {
				return nil, fmt.Errorf("invalid recognition color")
			}
			count := 1
			if i < len(text) && text[i] >= '1' && text[i] <= '9' {
				count = int(text[i] - '0')
				i++
			}
			if pos+count > 9 {
				return nil, fmt.Errorf("recognition face exceeds nine stickers")
			}
			for n := 0; n < count; n++ {
				c.Faces[order[f]][pos/3][pos%3] = Color(color)
				pos++
			}
		}
		if pos != 9 {
			return nil, fmt.Errorf("recognition face needs nine stickers")
		}
	}
	return c, nil
}

// A color relabeling commutes with sticker moves. Recoloring an inverse CFEN
// pattern to a rotated solved frame therefore moves a grip setup from the end
// of its algorithm to the start of its recognition state, without changing
// piece identities or adding a dependency from cube back to cfen.
func relabelPattern(pattern, solvedFrame *Cube) *Cube {
	var colors [6]Color
	home := NewCube(3)
	for f := range home.Faces {
		colors[home.faceColor(Face(f))] = solvedFrame.faceColor(Face(f))
	}
	result := pattern.clone()
	for f := range result.Faces {
		for r := range result.Faces[f] {
			for c, color := range result.Faces[f][r] {
				result.Faces[f][r][c] = colors[color]
			}
		}
	}
	return result
}
