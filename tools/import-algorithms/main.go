// Command import-algorithms reproducibly curates the CSV dumps into the embedded database.
package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
)

type rejected struct {
	Source string
	Row    []string
	Reason string
}

type summary struct {
	Files, Rows, Imported, Merged, Quarantined, Verified3x3, VerifiedOther int
	Categories                                                             map[string]int
	InversePairs, MirrorPairs                                              int
}

var citation = regexp.MustCompile(`:contentReference\[oaicite:\d+\]\{index=\d+\}`)
var token = regexp.MustCompile(`^(?:[1-9][0-9]*)?(?:[URFDLB]w?|[urfdlb]|[MESxyz])(?:2'?|'2|')?$`)
var repeatedGroup = regexp.MustCompile(`\(([^()]*)\)\s*\^([2-9])`)

func clean(s string) string { return strings.TrimSpace(citation.ReplaceAllString(s, "")) }

func normalize(s string) (string, error) {
	s = strings.NewReplacer("’", "'", "′", "'").Replace(clean(s))
	annotatedDoubleTurn := strings.Join(strings.Fields(s), " ") == "U U (U2)"
	for repeatedGroup.MatchString(s) {
		s = repeatedGroup.ReplaceAllStringFunc(s, func(group string) string {
			parts := repeatedGroup.FindStringSubmatch(group)
			n, _ := strconv.Atoi(parts[2])
			return "(" + strings.Repeat(parts[1]+" ", n) + ")"
		})
	}
	depth := 0
	for _, ch := range s {
		if ch == '(' {
			depth++
		}
		if ch == ')' {
			depth--
		}
		if depth < 0 {
			return "", fmt.Errorf("unbalanced parentheses")
		}
	}
	if depth != 0 {
		return "", fmt.Errorf("unbalanced parentheses")
	}
	s = strings.NewReplacer("(", " ", ")", " ").Replace(s)
	// The dump explicitly annotates this double turn; the parenthesis is not a third move.
	if annotatedDoubleTurn {
		s = "U2"
	}
	var result []cube.Move
	for _, word := range strings.Fields(s) {
		if !token.MatchString(word) {
			return "", fmt.Errorf("unresolved reference or invalid move %q", word)
		}
		if strings.ContainsAny(word, "urfdlb") {
			for _, f := range "urfdlb" {
				word = strings.ReplaceAll(word, string(f), strings.ToUpper(string(f))+"w")
			}
		}
		m, err := cube.ParseMove(word)
		if err != nil {
			return "", err
		}
		result = append(result, m)
	}
	if len(result) == 0 {
		return "", fmt.Errorf("empty algorithm")
	}
	return cube.FormatMoves(result), nil
}

func inverse(moves []cube.Move) []cube.Move {
	result := make([]cube.Move, len(moves))
	for i, m := range moves {
		m.Clockwise = !m.Clockwise
		result[len(moves)-1-i] = m
	}
	return result
}

// Reflection in the left/right plane reverses handedness. M and x lie along
// the reflected axis, so their axial rotation direction stays the same.
func mirror(moves []cube.Move) []cube.Move {
	result := append([]cube.Move(nil), moves...)
	for i := range result {
		m := &result[i]
		if m.Slice != cube.M_Slice && m.Rotation != cube.X_Rotation {
			m.Clockwise = !m.Clockwise
		}
		if m.Slice == cube.NoSlice && m.Rotation == cube.NoRotation {
			if m.Face == cube.Right {
				m.Face = cube.Left
			} else if m.Face == cube.Left {
				m.Face = cube.Right
			}
		}
	}
	return result
}

func canonical(c *cube.Cube) *cube.Cube {
	// Only rigid rotations, with no changes to the algorithm's physical effect.
	type node struct{ state *cube.Cube }
	queue := []node{{c}}
	seen := map[string]bool{}
	rotations, _ := cube.ParseMoves("x y z")
	home := cube.NewCube(3)
	for head := 0; head < len(queue); head++ {
		current := queue[head].state
		key, want := "", ""
		for f := cube.Front; f <= cube.Down; f++ {
			key += current.Faces[f][1][1].String()
			want += home.Faces[f][1][1].String()
		}
		if key == want {
			return current
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		for _, m := range rotations {
			text, _ := cfen.GenerateCFEN(current)
			state, _ := cfen.ParseCFEN(text)
			next, _ := state.ToCube()
			next.ApplyMove(m)
			queue = append(queue, node{next})
		}
	}
	return nil
}

func f2lSolved(c *cube.Cube) bool {
	for f := cube.Front; f <= cube.Down; f++ {
		if f == cube.Up {
			continue
		}
		first := 1
		if f == cube.Down {
			first = 0
		}
		for r := first; r < 3; r++ {
			for _, color := range c.Faces[f][r] {
				if color != c.Faces[f][1][1] {
					return false
				}
			}
		}
	}
	return true
}

func prepare(a *cube.Algorithm) error {
	text, err := normalize(a.Moves)
	if err != nil {
		return err
	}
	a.Moves = text
	moves, _ := cube.ParseMoves(text)
	for _, m := range moves {
		if m.Layer >= a.Dimension || m.WideDepth > a.Dimension {
			return fmt.Errorf("move %s exceeds dimension %d", m, a.Dimension)
		}
		if m.Slice != cube.NoSlice && a.Dimension%2 == 0 {
			return fmt.Errorf("slice %s needs an odd dimension", m)
		}
	}
	a.MoveCount = len(moves)
	a.Inverse = cube.FormatMoves(inverse(moves))
	start := cube.NewCube(a.Dimension)
	start.ApplyMoves(inverse(moves))
	a.Pattern, err = cfen.GenerateCFEN(start)
	if err != nil {
		return err
	}
	target, _ := cfen.GenerateCFEN(cube.NewCube(a.Dimension))
	// Exercise the same public CFEN verify path as the CLI, including parse/round trip.
	state, err := cfen.ParseCFEN(a.Pattern)
	if err != nil {
		return err
	}
	replay, err := state.ToCube()
	if err != nil {
		return err
	}
	replay.ApplyMoves(moves)
	goal, err := cfen.ParseCFEN(target)
	if err != nil {
		return err
	}
	matched, err := goal.MatchesCube(replay)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("generated recognition pattern does not verify")
	}
	if a.Dimension == 3 {
		if err := cube.Validate3x3(start); err != nil {
			return err
		}
		after := cube.NewCube(3)
		after.ApplyMoves(moves)
		after = canonical(after)
		if after == nil {
			return fmt.Errorf("unreachable center frame")
		}
		switch a.Category {
		case "OLL", "PLL":
			if !f2lSolved(after) {
				return fmt.Errorf("claimed %s algorithm disturbs the first two layers", a.Category)
			}
			if a.Category == "PLL" {
				for _, row := range after.Faces[cube.Up] {
					for _, color := range row {
						if color != cube.Yellow {
							return fmt.Errorf("claimed PLL algorithm changes last-layer orientation")
						}
					}
				}
			}
		case "F2L":
			if !cube.WhiteCrossSolved(after) {
				return fmt.Errorf("claimed F2L algorithm disturbs the white cross")
			}
		}
	}
	return nil
}

func appendUnique(list []string, values ...string) []string {
	for _, v := range values {
		if v == "" {
			continue
		}
		found := false
		for _, old := range list {
			if v == old {
				found = true
				break
			}
		}
		if !found {
			list = append(list, v)
		}
	}
	return list
}

func signature(dimension int, moves []cube.Move) string {
	c := cube.NewCube(dimension)
	// Unique labels distinguish same-color centers and wings on larger cubes;
	// relationships describe the full sticker permutation, not just a colored
	// solved-cube effect that could hide a nontrivial center permutation.
	label := 0
	for f := range c.Faces {
		for r := range c.Faces[f] {
			for col := range c.Faces[f][r] {
				c.Faces[f][r][col] = cube.Color(label)
				label++
			}
		}
	}
	c.ApplyMoves(moves)
	var key strings.Builder
	fmt.Fprintf(&key, "%d:", dimension)
	for _, face := range c.Faces {
		for _, row := range face {
			for _, color := range row {
				key.WriteRune(rune(color))
			}
		}
	}
	return key.String()
}

func relationships(db []cube.Algorithm) (int, int) {
	effects := map[string]int{}
	for i, a := range db {
		moves, _ := cube.ParseMoves(a.Moves)
		key := signature(a.Dimension, moves)
		if _, ok := effects[key]; !ok {
			effects[key] = i
		}
	}
	inverses, mirrors := map[string]bool{}, map[string]bool{}
	for i := range db {
		a := &db[i]
		moves, _ := cube.ParseMoves(a.Moves)
		pair := func(other int) string {
			ids := []string{a.CaseID, db[other].CaseID}
			sort.Strings(ids)
			return strings.Join(ids, ":")
		}
		if j, ok := effects[signature(a.Dimension, inverse(moves))]; ok {
			a.InverseID = db[j].CaseID
			inverses[pair(j)] = true
		}
		if j, ok := effects[signature(a.Dimension, mirror(moves))]; ok {
			a.Mirror = db[j].CaseID
			mirrors[pair(j)] = true
		}
	}
	return len(inverses), len(mirrors)
}

func importFiles(dir string) ([]cube.Algorithm, []rejected, summary, error) {
	db := []cube.Algorithm{
		{Name: "Sune", CaseID: "OLL-27", Category: "OLL", Moves: "R U R' U R U2 R'", Dimension: 3, Description: "Orient corners when one is correctly oriented", Recognition: "One corner oriented, headlights on left", Probability: 4.63},
		{Name: "Anti-Sune", CaseID: "OLL-26", Category: "OLL", Moves: "R U2 R' U' R U' R'", Dimension: 3, Description: "Orient corners with the inverse of Sune", Recognition: "One corner oriented, headlights on right", Probability: 4.63},
		{Name: "Cross OLL", CaseID: "OLL-CROSS", Category: "OLL", Moves: "F R U R' U' F'", Dimension: 3, Description: "Form yellow cross on top face", Recognition: "Need yellow cross (dot, line, or L-shape)"},
		{Name: "T-Perm", CaseID: "PLL-T", Category: "PLL", Moves: "R U R' F' R U R' U' R' F R2 U' R'", Dimension: 3, Description: "Swaps two adjacent corners and two edges", Recognition: "Headlights with opposite edge swap", Probability: 4.17},
		{Name: "Sexy Move", CaseID: "TRIG-1", Category: "Trigger", Moves: "R U R' U'", Dimension: 3, Description: "Most common trigger in cubing", Recognition: "F2L pair building/breaking trigger"},
	}
	keys := map[string]int{}
	key := func(a cube.Algorithm) string { return fmt.Sprintf("%d:%s", a.Dimension, a.Moves) }
	for i := range db {
		if err := prepare(&db[i]); err != nil {
			return nil, nil, summary{}, err
		}
		db[i].Sources = []string{"builtin"}
		keys[key(db[i])] = i
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.csv"))
	if err != nil {
		return nil, nil, summary{}, err
	}
	report := summary{Files: len(files), Categories: map[string]int{}}
	var quarantine []rejected
	for _, name := range files {
		file, err := os.Open(name)
		if err != nil {
			return nil, nil, report, err
		}
		reader := csv.NewReader(file)
		reader.FieldsPerRecord = -1
		for line := 1; ; line++ {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				file.Close()
				return nil, nil, report, fmt.Errorf("%s:%d: %w", name, line, err)
			}
			report.Rows++
			source := fmt.Sprintf("%s:%d", filepath.Base(name), line)
			if len(row) != 7 {
				quarantine = append(quarantine, rejected{source, row, "expected seven CSV fields"})
				continue
			}
			category := strings.TrimPrefix(clean(row[2]), "CFOP-")
			dimension := 3
			if len(category) >= 4 && category[1:3] == "x"+category[:1] {
				dimension, _ = strconv.Atoi(category[:1])
			}
			a := cube.Algorithm{CaseID: clean(row[0]), Name: clean(row[1]), Category: category, Moves: row[3], Dimension: dimension, Description: clean(row[4]), Recognition: clean(row[5]), Sources: []string{source}, References: []string{clean(row[6])}}
			if err := prepare(&a); err != nil {
				quarantine = append(quarantine, rejected{source, row, err.Error()})
				continue
			}
			report.Imported++
			if i, ok := keys[key(a)]; ok {
				old := &db[i]
				old.Sources = appendUnique(old.Sources, source)
				old.Aliases = appendUnique(old.Aliases, a.CaseID, a.Name)
				old.References = appendUnique(old.References, a.References...)
				if a.Category != old.Category {
					old.Categories = appendUnique(old.Categories, a.Category)
				}
				if old.Recognition == "" {
					old.Recognition = a.Recognition
				}
				if old.Description == "" {
					old.Description = a.Description
				}
				report.Merged++
			} else {
				// Case IDs may have alternative algorithms, so retain both under distinct IDs.
				for _, old := range db {
					if old.CaseID == a.CaseID {
						a.Aliases = appendUnique(a.Aliases, a.CaseID)
						a.CaseID += "-CSV"
						break
					}
				}
				keys[key(a)] = len(db)
				db = append(db, a)
			}
		}
		file.Close()
	}
	report.Quarantined = len(quarantine)
	for _, a := range db {
		report.Categories[a.Category]++
		for _, category := range a.Categories {
			report.Categories[category]++
		}
		if a.Dimension == 3 {
			report.Verified3x3++
		} else {
			report.VerifiedOther++
		}
	}
	report.InversePairs, report.MirrorPairs = relationships(db)
	return db, quarantine, report, nil
}

func writeJSON(name string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(name, append(data, '\n'), 0644)
}

func main() {
	dir := flag.String("input", "alg_dumps", "CSV directory")
	out := flag.String("output", "internal/cube/algorithms_data.json", "database output")
	flag.Parse()
	db, quarantine, report, err := importFiles(*dir)
	if err == nil {
		err = writeJSON(*out, db)
	}
	if err == nil {
		err = writeJSON(filepath.Join(*dir, "quarantine.json"), quarantine)
	}
	if err == nil {
		err = writeJSON(filepath.Join(*dir, "import-report.json"), report)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	data, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(data))
}
