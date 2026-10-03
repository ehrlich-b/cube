package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
)

func executeSolve(args ...string) (string, error) {
	cmd := newSolveCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestDefaultFullSolveAndCFENResume(t *testing.T) {
	c := cube.NewCube(3)
	scramble, _ := cube.ParseMoves("x R U F2 L' B")
	c.ApplyMoves(scramble)
	start, _ := cfen.GenerateCFEN(c)
	for _, args := range [][]string{{"x R U F2 L' B", "--headless"}, {"--start", start, "--headless"}} {
		out, err := executeSolve(args...)
		if err != nil {
			t.Fatal(err)
		}
		moves, err := parseLessonMoves(out)
		if err != nil || len(moves) == 0 {
			t.Fatalf("invalid full headless answer: %q (%v)", out, err)
		}
		check, _ := parseLessonState(start)
		check.ApplyMoves(moves)
		if !check.IsSolved() {
			t.Fatal("default full solve returned partial answer")
		}
	}
	out, err := executeSolve("--start", start, "--cfen", "--headless")
	if err != nil || out != "YB|Y9/R9/B9/W9/O9/G9" {
		t.Fatalf("full saved-state result: %s (%v)", out, err)
	}
	out, err = executeSolve("", "--headless")
	if err != nil || out != "" {
		t.Fatal("already solved cube must return an empty answer")
	}
}

func TestFullCLIRejectsUnsupportedOrInvalidInputs(t *testing.T) {
	for _, args := range [][]string{
		{"R", "--algorithm", "cfop"}, {"R", "--algorithm", "kociemba"}, {"R", "--dimension", "4"},
		{"R3"}, {"--start", "YB|Y999999999/R9/B9/W9/O9/G9"}, {"--start", "YB|?9/R9/B9/W9/O9/G9"},
		{"R", "--goal", "typo"},
	} {
		out, err := executeSolve(args...)
		if err == nil || out != "" {
			t.Fatalf("invalid args %v produced success/output: %q (%v)", args, out, err)
		}
	}
}

func TestFullLessonDefaultAndSavedGoal(t *testing.T) {
	cmd := newLearnCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"R U F2 L' B"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"middle edge", "Make the yellow cross", "Match the yellow cross", "yellow corners in their home", "finish the entire four-corner sweep", "Cube complete:"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("default full lesson missing %q", want)
		}
	}
	if strings.Contains(out.String(), "this lesson stops at the first layer") {
		t.Fatal("full lesson stopped prematurely")
	}
	for _, goal := range []string{"full", "first-layer"} {
		for _, input := range []string{"quit\n", ""} {
			out, err := executeLearn(input, "R", "--goal", goal, "--interactive")
			if err != nil || !strings.Contains(out, "--goal "+goal+" --interactive") {
				t.Fatal("quit/EOF lost selected goal")
			}
		}
	}
}
