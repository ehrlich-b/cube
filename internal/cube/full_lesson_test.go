package cube

import (
	"math/rand"
	"strings"
	"testing"
)

func allFacesByStickers(c *Cube) bool {
	for _, face := range c.Faces {
		for _, row := range face {
			for _, color := range row {
				if color != face[1][1] {
					return false
				}
			}
		}
	}
	return true
}

func TestBeginnerFullReplay(t *testing.T) {
	r := rand.New(rand.NewSource(609140))
	maxMoves, totalMoves := 0, 0
	for trial := 0; trial < 500; trial++ {
		c := NewCube(3)
		c.ApplyMoves(randomScramble(r, 25))
		before := c.clone()
		lesson, err := PlanBeginner(c)
		if err != nil {
			t.Fatalf("trial %d: %v", trial, err)
		}
		if !facesEqual(c, before) {
			t.Fatal("full planner mutated input")
		}
		physical := c.clone()
		for _, step := range lesson.Steps {
			physical.ApplyMoves(step.Moves())
			if !facesEqual(physical, step.After) {
				t.Fatalf("trial %d: %s checkpoint mismatch", trial, step.Title)
			}
			if err := Validate3x3(physical); err != nil {
				t.Fatalf("trial %d: %s invalid: %v", trial, step.Title, err)
			}
			if strings.Contains(step.Title, "middle edge") && !whiteLayerByStickers(physical) {
				t.Fatal("middle insertion broke white")
			}
			if strings.Contains(step.Title, "yellow") && !TwoLayersSolved(physical) {
				t.Fatal("last-layer checkpoint broke first two layers")
			}
		}
		if !allFacesByStickers(physical) || !facesEqual(physical, lesson.Final) {
			t.Fatalf("trial %d not fully solved", trial)
		}
		n := len(lesson.Moves())
		totalMoves += n
		if n > maxMoves {
			maxMoves = n
		}
		if n > 600 {
			t.Fatal("full lesson exceeded move bound")
		}
	}
	t.Logf("500 full lessons: mean %.1f moves, maximum %d", float64(totalMoves)/500, maxMoves)
}

func TestBeginnerRecoveryAtEveryStage(t *testing.T) {
	c := NewCube(3)
	c.ApplyMoves(lessonMoves("x R2 F' U B2 L D' R U2 F B'"))
	lesson, err := PlanBeginner(c)
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range lesson.Steps {
		// Includes an interrupted final twist sequence, extra turn and grip change.
		states := []*Cube{step.After.clone()}
		before := c.clone()
		for _, action := range step.Actions {
			for _, move := range action.Moves {
				before.ApplyMove(move)
				if move.Face == Down && move.Rotation == NoRotation {
					states = append(states, before.clone())
				}
			}
		}
		for _, state := range states {
			state.ApplyMoves(lessonMoves("R U' x y'"))
			resume, err := PlanBeginner(state)
			if err != nil {
				t.Fatalf("%s recovery: %v", step.Title, err)
			}
			state.ApplyMoves(resume.Moves())
			if !allFacesByStickers(state) {
				t.Fatalf("%s recovery did not solve", step.Title)
			}
		}
		c = step.After.clone()
	}
}

func TestBeginnerSessionFinishesAndResets(t *testing.T) {
	c := NewCube(3)
	c.ApplyMoves(lessonMoves("z R U F2 L' B"))
	session, err := NewBeginnerSession(c)
	if err != nil {
		t.Fatal(err)
	}
	physical := c.clone()
	for i := 0; i < 20; i++ {
		step, err := session.Next()
		if err != nil {
			t.Fatal(err)
		}
		if step == nil {
			break
		}
		physical.ApplyMoves(step.Moves())
		if !facesEqual(physical, session.State()) {
			t.Fatal("full session checkpoint differs from actual moves")
		}
		inverse, ok := session.Undo()
		if !ok {
			t.Fatal("checkpoint was not undoable")
		}
		physical.ApplyMoves(inverse)
		step, err = session.Next()
		if err != nil || step == nil {
			t.Fatal("undo/retry failed")
		}
		physical.ApplyMoves(step.Moves())
	}
	if !allFacesByStickers(physical) {
		t.Fatal("full session did not finish all six faces")
	}
	for i := 0; i < 3; i++ {
		step, err := session.Next()
		if err != nil || step != nil || !facesEqual(physical, session.State()) {
			t.Fatal("completed next changed state")
		}
	}
	physical.ApplyMoves(session.Reset())
	if !facesEqual(physical, c) || !facesEqual(session.State(), c) {
		t.Fatal("full reset did not physically restore initial scramble")
	}
}

func TestBeginnerPreservesSolvedRotatedCube(t *testing.T) {
	for _, grip := range []string{"", "x", "y'", "z2 x"} {
		c := NewCube(3)
		c.ApplyMoves(lessonMoves(grip))
		lesson, err := PlanBeginner(c)
		if err != nil || len(lesson.Moves()) != 0 || !facesEqual(lesson.Final, c) {
			t.Fatalf("solved grip %q changed: %v", grip, err)
		}
	}
}

func BenchmarkBeginnerFull(b *testing.B) {
	c := NewCube(3)
	c.ApplyMoves(lessonMoves("x R2 F' U B2 L D' R U2 F B'"))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := PlanBeginner(c); err != nil {
			b.Fatal(err)
		}
	}
}
