package cube

import "testing"

func TestLessonSessionRepeatedStepsAndHistory(t *testing.T) {
	c := NewCube(3)
	scramble, _ := ParseMoves("x R2 F' U B2 L D' R U2 F B'")
	c.ApplyMoves(scramble)
	session, err := NewFirstLayerSession(c)
	if err != nil {
		t.Fatal(err)
	}
	for step := 0; step < 12; step++ {
		before := session.State()
		next, err := session.Next()
		if err != nil {
			t.Fatal(err)
		}
		if next == nil {
			break
		}
		physical := before.clone()
		physical.ApplyMoves(next.Moves())
		if !facesEqual(physical, session.State()) {
			t.Fatal("next differs from physical replay")
		}
		inverse, ok := session.Undo()
		physical.ApplyMoves(inverse)
		if !ok || !facesEqual(physical, before) || !facesEqual(session.State(), before) {
			t.Fatal("undo differs from physical replay")
		}
		if _, err := session.Next(); err != nil {
			t.Fatal(err)
		}
	}
	if !whiteLayerByStickers(session.State()) {
		t.Fatal("session did not finish")
	}
	complete := session.State()
	for i := 0; i < 4; i++ {
		next, err := session.Next()
		if err != nil || next != nil || !facesEqual(complete, session.State()) {
			t.Fatal("repeated next changed completed cube")
		}
	}
	resetMoves := session.Reset()
	complete.ApplyMoves(resetMoves)
	if !facesEqual(complete, c) || !facesEqual(session.State(), c) {
		t.Fatal("reset did not restore input")
	}
	if len(session.Reset()) != 0 {
		t.Fatal("repeated reset produced extra moves")
	}
	if _, ok := session.Undo(); ok {
		t.Fatal("reset must clear history")
	}
	copy := session.State()
	copy.ApplyMoves(scramble)
	if !facesEqual(session.State(), c) {
		t.Fatal("State exposed mutable session storage")
	}
}

func TestLessonSessionActualMoves(t *testing.T) {
	c := NewCube(3)
	moves, _ := ParseMoves("R U F2 L' B")
	c.ApplyMoves(moves)
	session, err := NewFirstLayerSession(c)
	if err != nil {
		t.Fatal(err)
	}
	actual, _ := ParseMoves("R U' x y2 M E' S2 Rw")
	physical := c.clone()
	physical.ApplyMoves(actual)
	if err := session.Record(actual); err != nil {
		t.Fatal(err)
	}
	if !facesEqual(physical, session.State()) {
		t.Fatal("record implicitly applied suggested moves")
	}
	for i := 0; i < 12; i++ {
		step, err := session.Next()
		if err != nil {
			t.Fatal(err)
		}
		if step == nil {
			break
		}
		physical.ApplyMoves(step.Moves())
		if !facesEqual(physical, session.State()) {
			t.Fatal("replanned checkpoint differs from replay")
		}
	}
	if !whiteLayerByStickers(physical) {
		t.Fatal("actual-move recovery did not finish")
	}
	physical.ApplyMoves(session.Reset())
	if !facesEqual(physical, c) {
		t.Fatal("reset failed after actual moves and checkpoints")
	}
}

func TestLessonSessionRejectsBadMovesAtomically(t *testing.T) {
	c := NewCube(3)
	session, _ := NewFirstLayerSession(c)
	for _, invalid := range []Move{{Face: Face(99)}, {Layer: 100}, {Wide: true, WideDepth: 100}, {Rotation: X_Rotation, Slice: M_Slice}} {
		if err := session.Record([]Move{{Face: Right, Clockwise: true}, invalid}); err == nil {
			t.Fatal("accepted invalid move batch")
		}
		if !facesEqual(session.State(), c) {
			t.Fatal("partially recorded invalid batch")
		}
		if _, ok := session.Undo(); ok {
			t.Fatal("failed record added undo history")
		}
	}
}
