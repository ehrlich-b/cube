package cli

import (
	"io"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/ehrlich-b/cube/internal/cube"
)

func captureFindResult(t *testing.T, run func() error) ([]cube.Move, bool) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	stdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = stdout }()
	err = run()
	w.Close()
	output, readErr := io.ReadAll(r)
	if err != nil || readErr != nil {
		t.Fatalf("find: %v; read output: %v", err, readErr)
	}
	for _, line := range strings.Split(string(output), "\n") {
		if text, ok := strings.CutPrefix(line, "1. "); ok {
			if strings.HasPrefix(text, "(already at target)") {
				return []cube.Move{}, true
			}
			notation, _, _ := strings.Cut(text, " (")
			moves, err := cube.ParseMoves(notation)
			if err != nil {
				t.Fatal(err)
			}
			return moves, true
		}
	}
	return nil, false
}

func TestFindSolvedCenterMovingMoves(t *testing.T) {
	for _, tc := range []struct {
		scramble, alphabet, want string
		depth                    int
	}{
		{"M", "M' R' L", "M'", 2},
		{"M", "M'", "M'", 1},
		{"Rw", "Rw'", "Rw'", 1},
		{"Rw'", "Rw", "Rw", 1},
		{"2R", "2R'", "2R'", 1},
		{"x R", "R' x'", "R'", 1},
	} {
		t.Run(tc.scramble+"/"+tc.alphabet, func(t *testing.T) {
			moves, err := cube.ParseMoves(tc.alphabet)
			if err != nil {
				t.Fatal(err)
			}
			options := findSearchOptions{dimension: 3, moves: moves}
			for _, kind := range []string{"sequence", "pattern"} {
				t.Run(kind, func(t *testing.T) {
					got, ok := captureFindResult(t, func() error {
						if kind == "sequence" {
							return runSequenceSearchWithOptions(tc.scramble, tc.depth, false, "", options)
						}
						return runPatternSearchWithOptions("solved", tc.depth, tc.scramble, false, options)
					})
					if !ok || cube.FormatMoves(got) != tc.want {
						t.Fatalf("got %q/%v, want shortest %q", cube.FormatMoves(got), ok, tc.want)
					}
				})
			}
		})
	}
}

// This test-only BFS uses complete sticker states and no production search
// helpers, cubie coordinates or pruning rules.
func independentStickerDistance(start *cube.Cube, moves []cube.Move, maxDepth int, target func(*cube.Cube) bool) (int, bool) {
	type node struct {
		c     *cube.Cube
		depth int
	}
	queue := []node{{copyCube(start), 0}}
	seen := map[string]bool{start.String(): true}
	for head := 0; head < len(queue); head++ {
		current := queue[head]
		if target(current.c) {
			return current.depth, true
		}
		if current.depth == maxDepth {
			continue
		}
		for _, move := range moves {
			next := copyCube(current.c)
			next.ApplyMove(move)
			key := next.String()
			if !seen[key] {
				seen[key] = true
				queue = append(queue, node{next, current.depth + 1})
			}
		}
	}
	return 0, false
}

func TestFindCenterMovingAlphabetsMatchIndependentStickerBFS(t *testing.T) {
	rng := rand.New(rand.NewSource(2026100707))
	for _, alphabet := range []string{
		"M M' E E' R R' L L'",
		"Rw Rw' Uw Uw' R R' U U'",
		"x x' y y' R R' U U'",
	} {
		t.Run(alphabet, func(t *testing.T) {
			moves, err := cube.ParseMoves(alphabet)
			if err != nil {
				t.Fatal(err)
			}
			for n := 0; n < 30; n++ {
				var scramble []cube.Move
				start := cube.NewCube(3)
				for i := 0; i < 1+n%3; i++ {
					m := moves[rng.Intn(len(moves))]
					scramble = append(scramble, m)
					start.ApplyMove(m)
				}
				before := start.String()
				want, found := independentStickerDistance(start, moves, 3, func(c *cube.Cube) bool { return c.IsSolved() })
				for _, kind := range []string{"sequence", "pattern"} {
					got, ok := captureFindResult(t, func() error {
						options := findSearchOptions{dimension: 3, moves: moves}
						if kind == "sequence" {
							return runSequenceSearchWithOptions(cube.FormatMoves(scramble), 3, false, "", options)
						}
						return runPatternSearchWithOptions("solved", 3, cube.FormatMoves(scramble), false, options)
					})
					if !found || !ok || len(got) != want {
						t.Fatalf("case %d %s scramble %s: find=%s/%v, sticker BFS=%d/%v", n, kind, cube.FormatMoves(scramble), cube.FormatMoves(got), ok, want, found)
					}
					check := copyCube(start)
					check.ApplyMoves(got)
					if !check.IsSolved() || start.String() != before {
						t.Fatal("solution failed replay or search mutated start")
					}
				}
				// Concrete targets keep their exact sticker orientation, even
				// when the alphabet can move centers.
				target := copyCube(start)
				for i := 0; i < 1+n%3; i++ {
					target.ApplyMove(moves[rng.Intn(len(moves))])
				}
				want, found = independentStickerDistance(start, moves, 3, func(c *cube.Cube) bool { return c.String() == target.String() })
				got, ok := cube.FindPattern(start, target, moves, 3)
				if !found || !ok || len(got) != want {
					t.Fatalf("case %d concrete target: find=%s/%v, sticker BFS=%d/%v", n, cube.FormatMoves(got), ok, want, found)
				}
				check := copyCube(start)
				check.ApplyMoves(got)
				if check.String() != target.String() {
					t.Fatal("concrete answer failed sticker replay")
				}
			}
		})
	}
}
