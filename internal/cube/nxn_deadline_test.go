package cube_test

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
)

func TestNxNReducedSixSurvivesDeadline(t *testing.T) {
	data, err := os.ReadFile("../../test/fixtures/reduced-six.cfen")
	if err != nil {
		t.Fatal(err)
	}
	state, err := cfen.ParseCFEN(strings.TrimSpace(string(data)))
	if err != nil {
		t.Fatal(err)
	}
	c, err := state.ToCube()
	if err != nil {
		t.Fatal(err)
	}
	if c.Size != 6 || c.IsSolved() {
		t.Fatal("fixture must be an unsolved 6x6")
	}
	before, err := cfen.GenerateCFEN(c)
	if err != nil {
		t.Fatal(err)
	}
	// The already reduced outer-turn state has the same unsolved 3x3 on
	// every wing orbit. Its explicit hard budget expires before any search.
	reduced := cube.NewCube(3)
	for face := range reduced.Faces {
		for row, sourceRow := range []int{0, 1, 5} {
			for col, sourceCol := range []int{0, 1, 5} {
				reduced.Faces[face][row][col] = c.Faces[face][sourceRow][sourceCol]
			}
		}
	}
	if result, err := cube.SolveKociemba(reduced, cube.KociembaOptions{TargetLength: 20, TimeLimit: time.Nanosecond}); result != nil || err == nil || !strings.Contains(err.Error(), "time limit exceeded") {
		t.Fatal("fixture must exhaust the hard 3x3 budget", result, err)
	}
	baseline, err := (&cube.ReductionSolver{}).Solve(c)
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []int{20, 1} {
		result, err := cube.SolveNxN(c, cube.KociembaOptions{TargetLength: target, TimeLimit: time.Nanosecond})
		if err != nil {
			t.Fatalf("target %d: tiny reduction budget lost a solvable state: %v", target, err)
		}
		if result == nil || result.Steps != len(result.Solution) || len(result.Solution) == 0 {
			t.Fatal("invalid reduction result")
		}
		if !reflect.DeepEqual(result.Solution, baseline.Solution) {
			t.Fatal("tiny budget changed the first deterministic solution")
		}
		unchanged, err := cfen.GenerateCFEN(c)
		if err != nil || unchanged != before {
			t.Fatal("reduction mutated its input", err)
		}
		check, err := state.ToCube()
		if err != nil {
			t.Fatal(err)
		}
		if err := check.ApplyMoves(result.Solution); err != nil || !reflect.DeepEqual(check.Faces, cube.NewCube(6).Faces) {
			t.Fatal("solution failed full canonical sticker replay", err)
		}
		t.Logf("target %d: tiny budget returned %d verified moves", target, result.Steps)
	}
}
