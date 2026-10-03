package cube

import "fmt"

// BeginnerLesson extends the verified first-layer checkpoints to a fully solved
// 3x3. Every checkpoint restores the prerequisites stated in its Check text.
type BeginnerLesson struct {
	Steps []FirstLayerStep
	Final *Cube
}

func (l *BeginnerLesson) Moves() []Move {
	var moves []Move
	for _, step := range l.Steps {
		moves = append(moves, step.Moves()...)
	}
	return moves
}

// PlanBeginner completes a physical 3x3 using a beginner layer-by-layer method.
// It is read only, keeps white down after normalization, and bounds its work:
// cross searches are bounded; remaining stages use fixed insertions or tiny
// finite searches over named beginner algorithms, never arbitrary face-turn DFS.
func PlanBeginner(c *Cube) (*BeginnerLesson, error) {
	if err := Validate3x3(c); err != nil {
		return nil, err
	}
	// A solved rotated cube needs no moves for a full solve. This also lets
	// repeated next at completion leave both stickers and grip unchanged.
	if c.IsSolved() {
		return &BeginnerLesson{Final: c.clone()}, nil
	}
	first, err := PlanFirstLayer(c)
	if err != nil {
		return nil, err
	}
	lesson := &BeginnerLesson{Steps: first.Steps}
	work := first.Final.clone()
	middleEdges := [4][2]Color{{Blue, Red}, {Blue, Orange}, {Green, Red}, {Green, Orange}}
	for _, colors := range middleEdges {
		if EdgeSolved(work, colors[0], colors[1]) {
			continue
		}
		var protected [][2]Color
		for _, other := range middleEdges {
			if EdgeSolved(work, other[0], other[1]) {
				protected = append(protected, other)
			}
		}
		actions, err := insertMiddleEdge(work, colors)
		if err != nil {
			return nil, err
		}
		if !FirstLayerSolved(work) || !EdgeSolved(work, colors[0], colors[1]) || !edgesSolved(work, protected) {
			return nil, fmt.Errorf("middle-edge checkpoint failed verification")
		}
		lesson.Steps = append(lesson.Steps, FirstLayerStep{
			Title:   "Insert the " + colorName(colors[0]) + "-" + colorName(colors[1]) + " middle edge",
			Actions: actions, Check: "Both stickers match their side centers. White and earlier middle edges are restored.", After: work.clone(),
		})
	}
	if !TwoLayersSolved(work) {
		return nil, fmt.Errorf("first two layers failed verification")
	}
	uSetups := []LessonAction{
		{"Turn only the top face to prepare the next algorithm; keep white down and blue front.", lessonMoves("U")},
		{"Turn only the top face a half turn to prepare the next algorithm.", lessonMoves("U2")},
		{"Turn only the top face counterclockwise to prepare the next algorithm.", lessonMoves("U'")},
	}
	crossAlgorithms := append(append([]LessonAction(nil), uSetups...), LessonAction{
		"Make more yellow cross arms with F R U R' U' F'. Check only after the entire group.", lessonMoves("F R U R' U' F'"),
	})
	if err := addLastLayerStep(lesson, work, "Make the yellow cross", "All four top edge stickers are yellow. The first two layers are restored.",
		YellowCrossSolved, yellowCrossKey, crossAlgorithms, 4); err != nil {
		return nil, err
	}
	edgeAlgorithms := append(append([]LessonAction(nil), uSetups...), LessonAction{
		"Cycle yellow edges with R U R' U R U2 R'. Yellow corner stickers may change; the cross and lower layers return after the group.", lessonMoves("R U R' U R U2 R'"),
	})
	if err := addLastLayerStep(lesson, work, "Match the yellow cross to the side centers", "Each yellow edge has yellow up and its other sticker matching its side center. The first two layers are restored.",
		YellowEdgesSolved, yellowEdgeKey, edgeAlgorithms, 4); err != nil {
		return nil, err
	}
	var cornerAlgorithms []LessonAction
	cycle := lessonMoves("U R U' L' U R' U' L")
	for _, setupText := range []string{"", "y", "y2", "y'"} {
		setup := lessonMoves(setupText)
		instruction := "Cycle three yellow corner positions with U R U' L' U R' U' L; their twists do not matter yet."
		if len(setup) > 0 {
			instruction = "Rotate the whole cube with " + setupText + ", cycle three corner positions, then undo that rotation. Keep white down."
		}
		moves := append(append(append([]Move(nil), setup...), cycle...), inverseSequence(setup)...)
		cornerAlgorithms = append(cornerAlgorithms, LessonAction{instruction, moves})
	}
	if err := addLastLayerStep(lesson, work, "Put the yellow corners in their home positions", "Each top corner contains the three colors of its surrounding centers, even if yellow faces sideways. All edges and the first two layers are restored.",
		YellowCornersPlaced, yellowCornerKey, cornerAlgorithms, 2); err != nil {
		return nil, err
	}
	if !YellowEdgesSolved(work) {
		return nil, fmt.Errorf("placing yellow corners disturbed the solved yellow edges")
	}
	if !work.IsSolved() {
		actions, err := orientYellowCorners(work)
		if err != nil {
			return nil, err
		}
		lesson.Steps = append(lesson.Steps, FirstLayerStep{
			Title:   "Turn yellow corners upright — finish the entire four-corner sweep",
			Actions: actions, Check: "All six faces are uniform and match their centers. Complete the final top turn before checking or typing next.", After: work.clone(),
		})
	}
	if !work.IsSolved() {
		return nil, fmt.Errorf("beginner lesson did not solve all six faces")
	}
	lesson.Final = work.clone()
	if len(lesson.Moves()) > 600 {
		return nil, fmt.Errorf("beginner lesson exceeded the 600-move bound")
	}
	return lesson, nil
}

func lessonMoves(text string) []Move {
	moves, err := ParseMoves(text)
	if err != nil {
		panic("invalid built-in beginner algorithm: " + err.Error())
	}
	return moves
}

// TwoLayersSolved is center-relative and works in any rigid orientation.
func TwoLayersSolved(c *Cube) bool {
	return FirstLayerSolved(c) && EdgeSolved(c, Blue, Red) && EdgeSolved(c, Blue, Orange) &&
		EdgeSolved(c, Green, Red) && EdgeSolved(c, Green, Orange)
}

func YellowCrossSolved(c *Cube) bool {
	return c.Faces[Up][0][1] == Yellow && c.Faces[Up][1][0] == Yellow &&
		c.Faces[Up][1][2] == Yellow && c.Faces[Up][2][1] == Yellow
}

func YellowEdgesSolved(c *Cube) bool {
	return EdgeSolved(c, Yellow, Blue) && EdgeSolved(c, Yellow, Red) &&
		EdgeSolved(c, Yellow, Green) && EdgeSolved(c, Yellow, Orange)
}

func YellowCornersPlaced(c *Cube) bool {
	for _, coords := range cornerFacelets[:4] {
		if !cornerColorsMatch(sticker(c, coords[0]), sticker(c, coords[1]), sticker(c, coords[2]),
			c.faceColor(coords[0].Face), c.faceColor(coords[1].Face), c.faceColor(coords[2].Face)) {
			return false
		}
	}
	return true
}

func edgeOnFace(edge EdgePiece, face Face) bool {
	a, _, _ := CubieToFacePos(edge.A, 3)
	b, _, _ := CubieToFacePos(edge.B, 3)
	return a == face || b == face
}

func edgeAt(c *Cube, colors [2]Color, a, b Face) bool {
	edge, err := FindEdge(c, colors[0], colors[1])
	return err == nil && edgeOnFace(edge, a) && edgeOnFace(edge, b)
}

func insertMiddleEdge(work *Cube, colors [2]Color) ([]LessonAction, error) {
	var actions []LessonAction
	apply := func(instruction string, moves []Move) {
		if len(moves) > 0 {
			work.ApplyMoves(moves)
			actions = append(actions, LessonAction{instruction, moves})
		}
	}
	right := lessonMoves("U R U' R' U' F' U F")
	left := lessonMoves("U' L' U L U F U' F'")
	edge, _ := FindEdge(work, colors[0], colors[1])
	if !edgeOnFace(edge, Up) {
		var setup []Move
		for turns := 0; !edgeAt(work, colors, Front, Right) && turns < 4; turns++ {
			work.ApplyMoves(lessonMoves("y"))
			setup = append(setup, lessonMoves("y")...)
		}
		if !edgeAt(work, colors, Front, Right) {
			return nil, fmt.Errorf("could not bring misplaced middle edge to front-right")
		}
		if len(setup) > 0 {
			actions = append(actions, LessonAction{"Rotate the whole cube to bring the misplaced edge to the middle front-right slot.", setup})
		}
		apply("Use a right insertion to lift that misplaced edge into the top layer. The white layer returns at the end.", right)
		apply("Return to white down, blue front.", inverseSequence(setup))
	}
	var up []Move
	for turns := 0; !edgeAt(work, colors, Up, Front) && turns < 4; turns++ {
		work.ApplyMoves(lessonMoves("U"))
		up = append(up, lessonMoves("U")...)
	}
	if !edgeAt(work, colors, Up, Front) {
		return nil, fmt.Errorf("could not bring middle edge to top front")
	}
	if len(up) > 0 {
		actions = append(actions, LessonAction{"Turn the top face to bring this non-yellow edge to the top front.", up})
	}
	// Match the side sticker, rather than assuming which of the two colors faces
	// outward. Rotate the cube and top oppositely so the edge remains at UF.
	var setup []Move
	for turns := 0; work.Faces[Front][0][1] != work.faceColor(Front) && turns < 4; turns++ {
		pair := lessonMoves("y U'")
		work.ApplyMoves(pair)
		setup = append(setup, pair...)
	}
	if work.Faces[Front][0][1] != work.faceColor(Front) {
		return nil, fmt.Errorf("could not line up the middle-edge side sticker")
	}
	if len(setup) > 0 {
		actions = append(actions, LessonAction{"Turn the whole cube with y and the top back with U' until the edge's side sticker matches the front center.", setup})
	}
	if work.Faces[Up][2][1] == work.faceColor(Right) {
		apply("The top sticker matches the right center: insert to the right with U R U' R' U' F' U F.", right)
	} else if work.Faces[Up][2][1] == work.faceColor(Left) {
		apply("The top sticker matches the left center: insert to the left with U' L' U L U F U' F'.", left)
	} else {
		return nil, fmt.Errorf("middle edge's top sticker matches neither side destination")
	}
	// Undo only the whole-cube part of the alignment; the U turns were actual
	// setup turns, not a frame change, and must not be undone after insertion.
	var rotations []Move
	for _, move := range setup {
		if move.Rotation != NoRotation {
			rotations = append(rotations, move)
		}
	}
	apply("Return to white down, blue front.", inverseSequence(rotations))
	return actions, nil
}

func yellowCrossKey(c *Cube) string {
	key := byte(0)
	for i, coord := range []Coord{{Up, 0, 1}, {Up, 1, 2}, {Up, 2, 1}, {Up, 1, 0}} {
		if sticker(c, coord) == Yellow {
			key |= 1 << i
		}
	}
	return string([]byte{key})
}

func yellowEdgeKey(c *Cube) string {
	return string([]byte{byte(c.Faces[Front][0][1]), byte(c.Faces[Right][0][1]), byte(c.Faces[Back][0][1]), byte(c.Faces[Left][0][1])})
}

func yellowCornerKey(c *Cube) string {
	key := make([]byte, 0, 12)
	for _, coords := range cornerFacelets[:4] {
		a, b, d := colorTripleSorted(sticker(c, coords[0]), sticker(c, coords[1]), sticker(c, coords[2]))
		key = append(key, byte(a), byte(b), byte(d))
	}
	return string(key)
}

// Search only the stage's small quotient: 8 legal orientation patterns, 24 edge
// permutations, or 12 corner permutations. The macro effects on that quotient
// do not depend on ignored twists/permutations. Bounds guard future changes too.
func beginnerMacroPath(start *Cube, target func(*Cube) bool, key func(*Cube) string, macros []LessonAction, maxDepth int) ([]LessonAction, error) {
	if target(start) {
		return nil, nil
	}
	type node struct {
		state                *Cube
		parent, macro, depth int
	}
	queue := []node{{state: start.clone(), parent: -1}}
	seen := map[string]bool{key(start): true}
	for head := 0; head < len(queue); head++ {
		current := queue[head]
		if current.depth >= maxDepth {
			continue
		}
		for index, macro := range macros {
			state := current.state.clone()
			state.ApplyMoves(macro.Moves)
			k := key(state)
			if seen[k] {
				continue
			}
			seen[k] = true
			queue = append(queue, node{state, head, index, current.depth + 1})
			if target(state) {
				var reverse []LessonAction
				for at := len(queue) - 1; queue[at].parent >= 0; at = queue[at].parent {
					reverse = append(reverse, macros[queue[at].macro])
				}
				actions := make([]LessonAction, len(reverse))
				for i := range reverse {
					actions[len(reverse)-1-i] = reverse[i]
				}
				return actions, nil
			}
			if len(queue) > 64 {
				return nil, fmt.Errorf("last-layer macro stage exceeded its finite-state bound")
			}
		}
	}
	return nil, fmt.Errorf("no beginner algorithm path within %d groups", maxDepth)
}

func addLastLayerStep(lesson *BeginnerLesson, work *Cube, title, check string, target func(*Cube) bool, key func(*Cube) string, macros []LessonAction, maxDepth int) error {
	actions, err := beginnerMacroPath(work, target, key, macros, maxDepth)
	if err != nil {
		return fmt.Errorf("%s: %w", title, err)
	}
	if len(actions) == 0 {
		return nil
	}
	for _, action := range actions {
		work.ApplyMoves(action.Moves)
	}
	if !target(work) || !TwoLayersSolved(work) {
		return fmt.Errorf("%s checkpoint failed verification", title)
	}
	lesson.Steps = append(lesson.Steps, FirstLayerStep{title, actions, check, work.clone()})
	return nil
}

func orientYellowCorners(work *Cube) ([]LessonAction, error) {
	var actions []LessonAction
	trigger := lessonMoves("R' D' R D")
	for corner := 1; corner <= 4; corner++ {
		var turns []Move
		repeats := 0
		for ; work.Faces[Up][2][2] != Yellow && repeats < 6; repeats += 2 {
			work.ApplyMoves(trigger)
			work.ApplyMoves(trigger)
			turns = append(turns, trigger...)
			turns = append(turns, trigger...)
		}
		if work.Faces[Up][2][2] != Yellow {
			return nil, fmt.Errorf("could not orient yellow corner %d within the trigger bound", corner)
		}
		if len(turns) > 0 {
			actions = append(actions, LessonAction{
				fmt.Sprintf("Corner %d of 4 at top front-right: repeat R' D' R D exactly %d times. Keep your grip; lower layers can look mixed until ALL four corners and the final top turn are finished.", corner, repeats), turns,
			})
		}
		u := lessonMoves("U")
		work.ApplyMoves(u)
		instruction := "Turn only U to bring the next corner to top front-right. Do not rotate the whole cube or repair the lower layers."
		if corner == 4 {
			instruction = "Make the final U turn to restore the top-edge alignment. Now check all six faces and type next."
		}
		actions = append(actions, LessonAction{instruction, u})
	}
	if !work.IsSolved() {
		return nil, fmt.Errorf("four-corner sweep did not restore a fully solved cube")
	}
	return actions, nil
}
