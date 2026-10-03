package cube

import "fmt"

type lessonHistory struct {
	before *Cube
	moves  []Move
}

// FirstLayerSession tracks the modeled physical cube. Each checkpoint or batch
// of recorded actual moves is atomic and undoable; replanning always uses the
// current state, never a stale step number.
type FirstLayerSession struct {
	initial *Cube
	current *Cube
	history []lessonHistory
}

func NewFirstLayerSession(c *Cube) (*FirstLayerSession, error) {
	if err := Validate3x3(c); err != nil {
		return nil, err
	}
	return &FirstLayerSession{initial: c.clone(), current: c.clone()}, nil
}

// State returns a snapshot that callers may safely inspect or change.
func (s *FirstLayerSession) State() *Cube { return s.current.clone() }

func (s *FirstLayerSession) Next() (*FirstLayerStep, error) {
	lesson, err := PlanFirstLayer(s.current)
	if err != nil || len(lesson.Steps) == 0 {
		return nil, err
	}
	step := lesson.Steps[0]
	if err := s.Record(step.Moves()); err != nil {
		return nil, err
	}
	return &step, nil
}

// Record applies actual moves made since the displayed state. It never adds the
// suggested moves implicitly, making a wrong turn recoverable without guessing.
func (s *FirstLayerSession) Record(moves []Move) error {
	if len(moves) == 0 {
		return nil
	}
	for _, move := range moves {
		if move.Face < Front || move.Face > Down || move.Layer < 0 || move.Layer > 2 ||
			move.WideDepth < 0 || move.WideDepth > 3 || move.Slice < NoSlice || move.Slice > S_Slice ||
			move.Rotation < NoRotation || move.Rotation > Z_Rotation ||
			(move.Slice != NoSlice && move.Rotation != NoRotation) ||
			((move.Slice != NoSlice || move.Rotation != NoRotation) && (move.Wide || move.Layer != 0 || move.WideDepth != 0)) ||
			(move.Wide && move.Layer != 0) || (!move.Wide && move.WideDepth != 0) {
			return fmt.Errorf("recorded moves must be valid 3x3 turns")
		}
	}
	work := s.current.clone()
	work.ApplyMoves(moves)
	if err := Validate3x3(work); err != nil {
		return err
	}
	s.history = append(s.history, lessonHistory{s.current, append([]Move(nil), moves...)})
	s.current = work
	return nil
}

// Undo restores the previous state and returns the physical moves that undo the
// last batch. No history means no change, including after repeated undo calls.
func (s *FirstLayerSession) Undo() ([]Move, bool) {
	if len(s.history) == 0 {
		return nil, false
	}
	last := s.history[len(s.history)-1]
	s.history = s.history[:len(s.history)-1]
	s.current = last.before
	return inverseSequence(last.moves), true
}

// Reset returns the physical inverse of all recorded batches in reverse order.
func (s *FirstLayerSession) Reset() []Move {
	var undo []Move
	for i := len(s.history) - 1; i >= 0; i-- {
		undo = append(undo, inverseSequence(s.history[i].moves)...)
	}
	s.current = s.initial.clone()
	s.history = nil
	return undo
}
