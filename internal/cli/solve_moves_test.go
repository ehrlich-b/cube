package cli

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/ehrlich-b/cube/internal/cube"
)

func TestSolveInputNumberedMovesMatchTwist(t *testing.T) {
	for _, size := range []int{3, 5} {
		for _, face := range []string{"R", "U", "F", "L", "D", "B"} {
			for layer := 1; layer <= size; layer++ {
				for _, wide := range []string{"", "w"} {
					for _, suffix := range []string{"", "'", "2"} {
						text := fmt.Sprintf("%d%s%s%s", layer, face, wide, suffix)
						t.Run(fmt.Sprintf("%dx%d/%s", size, size, text), func(t *testing.T) {
							cmd := newSolveCommand()
							if err := cmd.Flags().Set("dimension", strconv.Itoa(size)); err != nil {
								t.Fatal(err)
							}
							got, err := fullSolveInput(cmd, []string{text})
							if err != nil {
								t.Fatal(err)
							}
							want := cube.NewCube(size)
							moves, err := cube.ParseMoves(text)
							if err != nil {
								t.Fatal(err)
							}
							if err := want.ApplyMoves(moves); err != nil {
								t.Fatal(err)
							}
							if got.String() != want.String() {
								t.Fatal("solve input differs from twist's move application")
							}
						})
					}
				}
			}
		}
	}
}

func TestSolveAndLearnNumberedTurnReplay(t *testing.T) {
	for _, size := range []int{3, 5} {
		for _, text := range []string{"2R", "2R'", "2R2"} {
			t.Run(fmt.Sprintf("%dx%d/%s", size, size, text), func(t *testing.T) {
				out, err := executeSolve(text, "--dimension", strconv.Itoa(size), "--headless")
				if err != nil {
					t.Fatal(err)
				}
				replay := func(solution string) {
					t.Helper()
					moves, err := cube.ParseMoves(text + " " + solution)
					if err != nil {
						t.Fatal(err)
					}
					c := cube.NewCube(size)
					if err := c.ApplyMoves(moves); err != nil || !c.IsSolved() {
						t.Fatalf("numbered turn solution failed replay: %v", err)
					}
				}
				replay(out)
				if size == 3 {
					out, err = executeLearn("", text, "--goal", "full")
					if err != nil {
						t.Fatal(err)
					}
					parts := strings.SplitN(out, "Solution: ", 2)
					if len(parts) != 2 || !strings.Contains(out, "Cube complete:") {
						t.Fatal("missing complete numbered turn lesson")
					}
					replay(strings.SplitN(parts[1], "\n", 2)[0])
				}
			})
		}
	}
}

func TestLessonMoveErrorListsNumberedNotation(t *testing.T) {
	_, err := parseLessonMoves("Q")
	if err == nil {
		t.Fatal("invalid move accepted")
	}
	for _, want := range []string{"3x3", "numbered 2R", "wide Rw or 2Rw", "slice M E S", "rotations x y z", "' or 2"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("move error omits %q: %v", want, err)
		}
	}
}
