package cube

import (
	"math/rand"
	"strings"
	"testing"
)

// This oracle checks face rows directly, independently of piece lookup or the
// solver's FirstLayerSolved predicate, which might share an addressing bug.
func whiteLayerByStickers(c *Cube) bool {
	for row := 0; row < 3; row++ {
		for col := 0; col < 3; col++ {
			if c.Faces[Down][row][col] != White {
				return false
			}
		}
	}
	for _, face := range []Face{Front, Back, Left, Right} {
		for col := 0; col < 3; col++ {
			if c.Faces[face][2][col] != c.Faces[face][1][1] {
				return false
			}
		}
	}
	return true
}

func TestFirstLayerOracle(t *testing.T) {
	r := rand.New(rand.NewSource(20261003))
	const trials = 500
	maxMoves, totalMoves := 0, 0
	for trial := 0; trial < trials; trial++ {
		c := NewCube(3)
		c.ApplyMoves(crossScrambleMoves(r, 25))
		// Exercise center changes as well as face turns, including slice/wide moves.
		for _, token := range []string{"x", "y'", "z2", "M", "E'", "S2", "Rw", "Fw'"} {
			if r.Intn(3) == 0 {
				moves, _ := ParseMoves(token)
				c.ApplyMoves(moves)
			}
		}
		before := c.clone()
		lesson, err := PlanFirstLayer(c)
		if err != nil {
			t.Fatalf("trial %d: %v\n%s", trial, err, c)
		}
		if !facesEqual(before, c) {
			t.Fatal("planning mutated its input")
		}
		check := before.clone()
		for i, step := range lesson.Steps {
			check.ApplyMoves(step.Moves())
			if !facesEqual(check, step.After) {
				t.Fatalf("trial %d checkpoint %d differs from replay", trial, i)
			}
			if strings.Contains(step.Title, "corner") && !WhiteCrossSolved(check) {
				t.Fatalf("trial %d corner checkpoint %d broke the cross", trial, i)
			}
			if err := Validate3x3(check); err != nil {
				t.Fatalf("trial %d checkpoint %d is invalid: %v", trial, i, err)
			}
		}
		if !whiteLayerByStickers(check) || !facesEqual(check, lesson.Final) {
			t.Fatalf("trial %d did not solve the white layer", trial)
		}
		moves := len(lesson.Moves())
		totalMoves += moves
		if moves > maxMoves {
			maxMoves = moves
		}
		if moves > 260 {
			t.Fatalf("trial %d exceeded the bounded lesson length: %d", trial, moves)
		}
		// Repeated planning is a no-op after success, never more insertion loops.
		repeat, err := PlanFirstLayer(check)
		if err != nil || len(repeat.Moves()) != 0 {
			t.Fatalf("trial %d repeated goal: %v", trial, err)
		}
	}
	t.Logf("%d scrambles: mean %.1f moves, max %d (includes whole-cube rotations)", trials, float64(totalMoves)/trials, maxMoves)
}

func TestFirstLayerRecovery(t *testing.T) {
	c := NewCube(3)
	moves, _ := ParseMoves("R2 F' U B2 L D' R U2 F B'")
	c.ApplyMoves(moves)
	lesson, err := PlanFirstLayer(c)
	if err != nil {
		t.Fatal(err)
	}
	for i, step := range lesson.Steps {
		for _, token := range []string{"", "R", "x y'", "R U R'", "M'"} {
			// An interruption, extra turn or changed orientation at any checkpoint.
			work := step.After.clone()
			deviation, _ := ParseMoves(token)
			work.ApplyMoves(deviation)
			resume, err := PlanFirstLayer(work)
			if err != nil {
				t.Fatalf("checkpoint %d, deviation %q: %v", i, token, err)
			}
			work.ApplyMoves(resume.Moves())
			if !whiteLayerByStickers(work) {
				t.Fatalf("checkpoint %d, deviation %q: recovery failed", i, token)
			}
		}
	}
}

func TestFirstLayerCornerCases(t *testing.T) {
	// Inverse trigger cases cover white facing up/side/down, in the correct or
	// wrong bottom slot, and already solved corners that must be preserved.
	for _, rotation := range []string{"", "y", "y2", "y'"} {
		for repeats := 0; repeats < 6; repeats++ {
			for turns := 0; turns < 4; turns++ {
				c := NewCube(3)
				notation := rotation + " " + strings.Repeat("U R U' R' ", repeats) + strings.Repeat("U ", turns)
				moves, _ := ParseMoves(notation)
				c.ApplyMoves(moves)
				lesson, err := PlanFirstLayer(c)
				if err != nil {
					t.Fatalf("%s: %v", notation, err)
				}
				c.ApplyMoves(lesson.Moves())
				if !whiteLayerByStickers(c) {
					t.Fatalf("%s: white layer not solved", notation)
				}
			}
		}
	}
}

func TestValidate3x3RejectsUnreachable(t *testing.T) {
	set := func(c *Cube, p Coord, color Color) { c.Faces[p.Face][p.Row][p.Col] = color }
	swap := func(c *Cube, a, b Coord) {
		x, y := sticker(c, a), sticker(c, b)
		set(c, a, y)
		set(c, b, x)
	}
	cases := []struct {
		name string
		edit func(*Cube)
		want string
	}{
		{"single edge flip", func(c *Cube) { swap(c, edgeFacelets[0][0], edgeFacelets[0][1]) }, "flipped edges"},
		{"single corner twist", func(c *Cube) {
			p := cornerFacelets[0]
			a, b, d := sticker(c, p[0]), sticker(c, p[1]), sticker(c, p[2])
			set(c, p[0], b)
			set(c, p[1], d)
			set(c, p[2], a)
		}, "corner twists"},
		{"two edges exchanged", func(c *Cube) {
			for i := 0; i < 2; i++ {
				swap(c, edgeFacelets[0][i], edgeFacelets[1][i])
			}
		}, "parity"},
		{"mirrored corner", func(c *Cube) { swap(c, cornerFacelets[0][1], cornerFacelets[0][2]) }, "mirrored"},
		{"duplicate pieces", func(c *Cube) { swap(c, edgeFacelets[0][1], edgeFacelets[9][1]) }, "duplicate edge"},
		{"wrong counts", func(c *Cube) { set(c, Coord{Up, 0, 0}, White) }, "stickers"},
		{"wildcard", func(c *Cube) { set(c, Coord{Up, 0, 0}, Grey) }, "wildcards"},
		{"out of range color", func(c *Cube) { set(c, Coord{Up, 0, 0}, Color(-1)) }, "real colors"},
		{"impossible centers", func(c *Cube) { swap(c, Coord{Front, 1, 1}, Coord{Right, 1, 1}) }, "centers"},
		{"missing row", func(c *Cube) { c.Faces[Front] = nil }, "three rows"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewCube(3)
			tc.edit(c)
			err := Validate3x3(c)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want error containing %q", err, tc.want)
			}
			if _, err := PlanFirstLayer(c); err == nil {
				t.Fatal("lesson accepted invalid state")
			}
		})
	}
	for _, c := range []*Cube{nil, NewCube(2), NewCube(4)} {
		if _, err := PlanFirstLayer(c); err == nil {
			t.Fatal("accepted unsupported cube")
		}
	}
}

func TestValidate3x3Reachable(t *testing.T) {
	r := rand.New(rand.NewSource(41))
	for i := 0; i < 200; i++ {
		c := NewCube(3)
		c.ApplyMoves(randomScramble(r, 30))
		if err := Validate3x3(c); err != nil {
			t.Fatalf("reachable state %d: %v", i, err)
		}
	}
}

func BenchmarkFirstLayer(b *testing.B) {
	c := NewCube(3)
	moves, _ := ParseMoves("R2 F' U B2 L D' R U2 F B' x")
	c.ApplyMoves(moves)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := PlanFirstLayer(c); err != nil {
			b.Fatal(err)
		}
	}
}
