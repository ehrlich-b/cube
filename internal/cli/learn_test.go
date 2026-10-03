package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
)

func executeLearn(input string, args ...string) (string, error) {
	cmd := newLearnCommand()
	cmd.Flags().Set("goal", "first-layer")
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetIn(strings.NewReader(input))
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestLearnLessonAndResume(t *testing.T) {
	out, err := executeLearn("", "x R U F2 L' B")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Rotate only when instructed", "Hold white down", "Place the white-", "Repeat R U R' U'", "First layer complete", "Middle and last layers still need solving"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in output", want)
		}
	}
	// Resume from a printed checkpoint; replay all displayed flat moves to verify
	// CFEN and CLI output agree, using the engine independently from plan.Final.
	checkpoints := strings.Split(out, "After this checkpoint: ")
	for _, checkpoint := range checkpoints[1:] {
		state := strings.SplitN(checkpoint, "\n", 2)[0]
		resume, err := executeLearn("", "--start", state)
		if err != nil {
			t.Fatal(err)
		}
		parts := strings.Split(resume, "First-layer moves: ")
		if len(parts) != 2 {
			t.Fatal("missing moves in resumed lesson")
		}
		moves, err := cube.ParseMoves(strings.SplitN(parts[1], "\n", 2)[0])
		if err != nil {
			t.Fatal(err)
		}
		c, err := parseLessonState(state)
		if err != nil {
			t.Fatal(err)
		}
		c.ApplyMoves(moves)
		if !cube.FirstLayerSolved(c) {
			t.Fatal("CLI checkpoint failed recovery")
		}
	}
}

func TestLearnRejectsInvalidInputs(t *testing.T) {
	for _, args := range [][]string{
		{"R3"}, {"R''"}, {"0R"}, {"99Rw"}, {"xw"}, {"R", "U"}, {"--dimension", "2"},
		{"--start", "YB|?9/R9/B9/W9/O9/G9"}, {"--start", "WB|W9/R9/B9/Y9/O9/G9"},
		{"--start", "YB|Y999999999/R9/B9/W9/O9/G9"}, {"--start", "YB|Y4/R4/B4/W4/O4/G4"},
		{"--start", "YB|Y9/R9/B9/W9/O9"}, {"--start", "YB|Y8W/R9/B9/W9/O9/G9"},
		{strings.Repeat("R ", 5000)},
	} {
		if out, err := executeLearn("", args...); err == nil {
			t.Fatalf("accepted %v: %s", args, out)
		}
	}
}

func TestLearnInteractionRecoveryAndInvalidCommands(t *testing.T) {
	input := "moves R3\nnonsense\n" + strings.Repeat("R", 20000) + "\nmoves x y' R U'\nstate\nundo\nundo\nreset\nreset\nnext\nhint\nshow\nhelp\nquit\n"
	out, err := executeLearn(input, "R U F2 L' B", "--interactive")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"No state changed: invalid", "Unknown command", "input line too long", "Recorded actual moves", "Undo on your physical cube", "Nothing to undo", "Reset on your physical cube", "Confirmed:", "Resume: cube learn --start"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q", want)
		}
	}
	c := cube.NewCube(3)
	moves, _ := cube.ParseMoves("R U F2 L' B")
	c.ApplyMoves(moves)
	state, _ := cfen.GenerateCFEN(c)
	if !strings.Contains(out, "learn> "+state+"\n") && !strings.Contains(out, state) {
		t.Fatal("reset input state was not displayed")
	}
}

func TestLearnRepeatedNextAndEndOfInput(t *testing.T) {
	out, err := executeLearn("next\nnext\nnext\n", "U", "--interactive")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(out, "First layer already complete; no moves applied.") != 3 {
		t.Fatal("repeated next did not report a no-op")
	}
	if !strings.Contains(out, "Input ended. Resume:") {
		t.Fatal("EOF did not preserve a resume command")
	}
	// A final command without a newline still executes and saves state.
	out, err = executeLearn("quit", "--interactive")
	if err != nil || !strings.Contains(out, "Resume: cube learn --start") {
		t.Fatal("unterminated final input failed")
	}
}

func TestLearnRotatedSolvedStateWording(t *testing.T) {
	out, err := executeLearn("next\nquit\n", "x", "--interactive")
	if err != nil {
		t.Fatal(err)
	}
	initial := strings.Split(out, "Confirmed:")[0]
	if strings.Contains(initial, "First layer complete: white face and all four side bottom rows") {
		t.Fatal("claimed canonical bottom rows before orienting rotated cube")
	}
	if !strings.Contains(initial, "white layer pieces are already solved") || !strings.Contains(out, "First layer complete: white face") {
		t.Fatal("rotated completion/orientation guidance missing")
	}
}

func TestFirstLayerGoalSolveModes(t *testing.T) {
	// The actual solve command is exercised separately by binary E2E tests.
	// These direct runner tests pin output contracts and error propagation.
	cmd := newLearnCommand()
	cmd.Flags().String("algorithm", "beginner", "")
	cmd.Flags().Bool("headless", true, "")
	cmd.Flags().Bool("cfen", false, "")
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := runFirstLayerSolve(cmd, []string{"x R U F2 L' B"}); err != nil {
		t.Fatal(err)
	}
	moves, err := parseLessonMoves(out.String())
	if err != nil || len(moves) == 0 {
		t.Fatalf("invalid headless output: %q (%v)", out.String(), err)
	}
	c, _ := firstLayerInput(cmd, []string{"x R U F2 L' B"})
	c.ApplyMoves(moves)
	if !cube.FirstLayerSolved(c) {
		t.Fatal("headless moves failed goal")
	}
	out.Reset()
	cmd.Flags().Set("cfen", "true")
	if err := runFirstLayerSolve(cmd, []string{"R U F2 L' B"}); err != nil {
		t.Fatal(err)
	}
	final, err := parseLessonState(out.String())
	if err != nil || !cube.FirstLayerSolved(final) {
		t.Fatal("CFEN output failed goal")
	}
	cmd.Flags().Set("algorithm", "cfop")
	if err := runFirstLayerSolve(cmd, []string{"R"}); err == nil {
		t.Fatal("accepted incompatible solver")
	}
}
