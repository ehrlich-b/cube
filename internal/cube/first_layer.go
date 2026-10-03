package cube

import (
	"fmt"
	"strings"
)

// LessonAction is one setup or insertion within a piece checkpoint.
type LessonAction struct {
	Instruction string
	Moves       []Move
}

// FirstLayerStep is a checkpoint after placing a piece, or changing orientation.
// After is a private snapshot; callers can replay Moves to independently check it.
type FirstLayerStep struct {
	Title   string
	Actions []LessonAction
	Check   string
	After   *Cube
}

func (s FirstLayerStep) Moves() []Move {
	var moves []Move
	for _, action := range s.Actions {
		moves = append(moves, action.Moves...)
	}
	return moves
}

// FirstLayerLesson deliberately has a different result type from a full solver.
// Its moves solve the white layer, not necessarily the other two layers.
type FirstLayerLesson struct {
	Steps []FirstLayerStep
	Final *Cube
}

func (l *FirstLayerLesson) Moves() []Move {
	var moves []Move
	for _, step := range l.Steps {
		moves = append(moves, step.Moves()...)
	}
	return moves
}

// PlanFirstLayer creates a deterministic, bounded beginner lesson from a valid
// 3x3 state, without mutating it. Cross stages use the existing short search;
// corners use U setups and repetitions of R U R' U', with white kept down.
func PlanFirstLayer(c *Cube) (*FirstLayerLesson, error) {
	if err := Validate3x3(c); err != nil {
		return nil, err
	}
	work, rotations, _ := canonical3x3(c)
	lesson := &FirstLayerLesson{}
	if len(rotations) > 0 {
		lesson.Steps = append(lesson.Steps, FirstLayerStep{
			Title:   "Hold white down and blue front",
			Actions: []LessonAction{{"Rotate the whole cube; do not turn an individual face.", rotations}},
			Check:   "Centers: white down, yellow up, blue front, red right.", After: work.clone(),
		})
	}
	for k, edge := range whiteCrossEdges {
		seq, ok := solveEdgeStage(work, whiteCrossEdges[:k+1])
		if !ok {
			return nil, fmt.Errorf("could not place cross edge %d within the search bound", k+1)
		}
		if len(seq) == 0 {
			continue
		}
		work.ApplyMoves(seq)
		if !edgesSolved(work, whiteCrossEdges[:k+1]) {
			return nil, fmt.Errorf("cross checkpoint %d failed verification", k+1)
		}
		name := colorName(edge[1])
		lesson.Steps = append(lesson.Steps, FirstLayerStep{
			Title:   "Place the white-" + name + " edge",
			Actions: []LessonAction{{"Move the two-color edge home. Earlier cross checkpoints are restored at the end of this sequence.", seq}},
			Check:   "White touches the down center; " + name + " touches its matching side center.", After: work.clone(),
		})
	}
	if !WhiteCrossSolved(work) {
		return nil, fmt.Errorf("white cross failed verification")
	}
	whiteCorners := [4][3]Color{{White, Blue, Red}, {White, Blue, Orange}, {White, Green, Red}, {White, Green, Orange}}
	for _, colors := range whiteCorners {
		if CornerSolved(work, colors[0], colors[1], colors[2]) {
			continue
		}
		// Remember all already solved corners, including those outside this order.
		var protected [][3]Color
		for _, other := range whiteCorners {
			if CornerSolved(work, other[0], other[1], other[2]) {
				protected = append(protected, other)
			}
		}
		actions, err := insertWhiteCorner(work, colors)
		if err != nil {
			return nil, err
		}
		if !WhiteCrossSolved(work) || !CornerSolved(work, colors[0], colors[1], colors[2]) {
			return nil, fmt.Errorf("corner checkpoint failed verification")
		}
		for _, other := range protected {
			if !CornerSolved(work, other[0], other[1], other[2]) {
				return nil, fmt.Errorf("corner insertion disturbed an earlier corner")
			}
		}
		name := "white-" + colorName(colors[1]) + "-" + colorName(colors[2])
		lesson.Steps = append(lesson.Steps, FirstLayerStep{
			Title: "Place the " + name + " corner", Actions: actions,
			Check: "All three corner stickers match their centers. The white cross and earlier corners are restored.", After: work.clone(),
		})
	}
	if !FirstLayerSolved(work) {
		return nil, fmt.Errorf("first-layer goal failed verification")
	}
	lesson.Final = work.clone()
	if len(lesson.Moves()) > 260 {
		return nil, fmt.Errorf("lesson exceeded the 260-move bound")
	}
	return lesson, nil
}

func insertWhiteCorner(work *Cube, colors [3]Color) ([]LessonAction, error) {
	var actions []LessonAction
	apply := func(instruction string, moves []Move) {
		if len(moves) > 0 {
			work.ApplyMoves(moves)
			actions = append(actions, LessonAction{instruction, moves})
		}
	}
	trigger, _ := ParseMoves("R U R' U'")
	y, _ := ParseMoves("y")
	corner, _ := FindCorner(work, colors[0], colors[1], colors[2])
	if cornerOnFace(corner, Down) {
		// Lift an unsolved bottom corner through its own slot; solved slots survive.
		var setup []Move
		for turns := 0; !cornerAt(work, colors, Down, Front, Right) && turns < 4; turns++ {
			work.ApplyMoves(y)
			setup = append(setup, y...)
		}
		if len(setup) > 0 {
			actions = append(actions, LessonAction{"Turn the whole cube so this misplaced corner is at the bottom front-right.", setup})
		}
		apply("Lift this corner into the top layer with one right-hand trigger.", trigger)
		apply("Return to white down, blue front.", inverseSequence(setup))
	}
	var setup []Move
	for turns := 0; !cornerColorsMatch(White, work.faceColor(Front), work.faceColor(Right), colors[0], colors[1], colors[2]) && turns < 4; turns++ {
		work.ApplyMoves(y)
		setup = append(setup, y...)
	}
	if len(setup) > 0 {
		actions = append(actions, LessonAction{"Turn the whole cube so the corner's destination is bottom front-right.", setup})
	}
	var up []Move
	u, _ := ParseMoves("U")
	for turns := 0; !cornerAt(work, colors, Up, Front, Right) && turns < 4; turns++ {
		work.ApplyMoves(u)
		up = append(up, u...)
	}
	if !cornerAt(work, colors, Up, Front, Right) {
		return nil, fmt.Errorf("could not bring white corner above its destination")
	}
	if len(up) > 0 {
		actions = append(actions, LessonAction{"Turn only the top face to put this corner above its destination at top front-right.", up})
	}
	var insertion []Move
	repeats := 0
	for ; !CornerSolved(work, colors[0], colors[1], colors[2]) && repeats < 6; repeats++ {
		work.ApplyMoves(trigger)
		insertion = append(insertion, trigger...)
	}
	if !CornerSolved(work, colors[0], colors[1], colors[2]) {
		return nil, fmt.Errorf("corner insertion exceeded six right-hand triggers")
	}
	actions = append(actions, LessonAction{fmt.Sprintf("Repeat R U R' U' exactly %d time(s). Finish every four-move group before checking the corner.", repeats), insertion})
	apply("Return to white down, blue front.", inverseSequence(setup))
	return actions, nil
}

func cornerOnFace(corner CornerPiece, face Face) bool {
	for _, addr := range []CubieAddress{corner.A, corner.B, corner.C} {
		f, _, _ := CubieToFacePos(addr, 3)
		if f == face {
			return true
		}
	}
	return false
}

func cornerAt(c *Cube, colors [3]Color, a, b, d Face) bool {
	corner, err := FindCorner(c, colors[0], colors[1], colors[2])
	return err == nil && cornerOnFace(corner, a) && cornerOnFace(corner, b) && cornerOnFace(corner, d)
}

func inverseSequence(moves []Move) []Move {
	inv := make([]Move, len(moves))
	for i, move := range moves {
		inv[len(moves)-1-i] = invertMove(move)
	}
	return inv
}

func colorName(c Color) string {
	return []string{"white", "yellow", "red", "orange", "blue", "green"}[c]
}

// FormatMoves returns notation suitable for display or ParseMoves.
func FormatMoves(moves []Move) string {
	tokens := make([]string, len(moves))
	for i, move := range moves {
		tokens[i] = move.String()
	}
	return strings.Join(tokens, " ")
}
