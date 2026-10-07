package cli

import (
	"testing"

	"github.com/ehrlich-b/cube/internal/cube"
)

func TestKociembaCLIOptions(t *testing.T) {
	out, err := executeSolve("R U F2 L' B", "--target-length", "12", "--time-limit", "100ms", "--headless")
	if err != nil {
		t.Fatal(err)
	}
	c := cube.NewCube(3)
	scramble, _ := cube.ParseMoves("R U F2 L' B")
	solution, err := cube.ParseMoves(out)
	if err != nil || len(solution) == 0 {
		t.Fatal("missing solution", out, err)
	}
	c.ApplyMoves(scramble)
	c.ApplyMoves(solution)
	if !c.IsSolved() || cube.TurnCount(solution) > 12 {
		t.Fatal("custom Kociemba target failed", out)
	}
	for _, flags := range [][]string{
		{"--target-length", "0"}, {"--target-length", "31"},
		{"--time-limit", "0"}, {"--time-limit", "-1s"}, {"--time-limit", "1ns"},
		{"--optimal", "--target-length", "20"},
		{"--method", "beginner", "--time-limit", "1s"},
		{"--method", "cfop", "--target-length", "20"},
		{"--goal", "first-layer", "--target-length", "20"},
		{"--goal", "first-layer", "--time-limit", "1s"},
	} {
		out, err := executeSolve(append([]string{"R U F2 L' B", "--headless"}, flags...)...)
		if err == nil || out != "" {
			t.Fatal("invalid/expired options emitted a solution", flags, out, err)
		}
	}
}
