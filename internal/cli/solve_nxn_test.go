package cli

import (
	"strconv"
	"testing"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
)

func TestNxNCLIContractAndResume(t *testing.T) {
	for _, n := range []int{2, 4, 5, 6, 7} {
		scramble := "R U R' U' x"
		if n > 3 {
			scramble = "Rw Uw Fw 2R' 2U2 x"
		}
		state := cube.NewCube(n)
		moves, _ := cube.ParseMoves(scramble)
		if err := state.ApplyMoves(moves); err != nil {
			t.Fatal(err)
		}
		start, _ := cfen.GenerateCFEN(state)
		for _, input := range [][]string{{scramble}, {"--start", start}, {"U2", "--start", start}} {
			args := append(append([]string{}, input...), "--dimension", strconv.Itoa(n), "--headless")
			out, err := executeSolve(args...)
			if err != nil {
				t.Fatalf("%v: %v", args, err)
			}
			checkState, _ := cfen.ParseCFEN(start)
			check, _ := checkState.ToCube()
			if input[0] == "U2" {
				m, _ := cube.ParseMoves("U2")
				check.ApplyMoves(m)
			}
			solution, err := cube.ParseMoves(out)
			if err != nil || len(solution) == 0 {
				t.Fatalf("invalid %dx%d solution: %s (%v)", n, n, out, err)
			}
			if err := check.ApplyMoves(solution); err != nil || !check.IsSolved() {
				t.Fatalf("CLI solution replay failed: %v", err)
			}
			final, _ := cfen.GenerateCFEN(check)
			home, _ := cfen.GenerateCFEN(cube.NewCube(n))
			if final != home {
				t.Fatalf("solution did not match the solved center frame: %s", final)
			}
		}
		out, err := executeSolve("--dimension", strconv.Itoa(n), "--start", start, "--cfen", "--headless")
		home, _ := cfen.GenerateCFEN(cube.NewCube(n))
		if err != nil || out != home {
			t.Fatalf("CFEN output contract: %s (%v)", out, err)
		}
	}
}

func TestNxNCLIRejectsInvalidInputs(t *testing.T) {
	for _, args := range [][]string{
		{"R", "--dimension", "4", "--optimal"},
		{"R", "--dimension", "4", "--method", "beginner"},
		{"R", "--dimension", "5", "--method", "cfop"},
		{"R", "--dimension", "4", "--goal", "first-layer"},
		{"5R", "--dimension", "4"}, {"M", "--dimension", "4"},
		{"R3", "--dimension", "4"}, {"R''", "--dimension", "5"},
		{"--dimension", "4", "--start", "YB|Y99999999999999999999999/R16/B16/W16/O16/G16"},
		{"--dimension", "4", "--start", "YB|?16/R16/B16/W16/O16/G16"},
		{"--dimension", "4", "--start", "YB|Y9/R9/B9/W9/O9/G9"},
		{"--dimension", "4", "--start", "YB|Y0Y16/R16/B16/W16/O16/G16"},
		{"R", "--dimension", "4", "--time-limit", "0s"},
		{"R", "--dimension", "5", "--target-length", "31"},
	} {
		out, err := executeSolve(args...)
		if err == nil || out != "" {
			t.Fatalf("invalid args %v emitted a solution: %q (%v)", args, out, err)
		}
	}
}
