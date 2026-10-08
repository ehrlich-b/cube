package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
	"github.com/spf13/cobra"
)

func TestReview5OneByOneStartsReturnError(t *testing.T) {
	for _, args := range [][]string{
		{"twist", "", "--start", "YB|Y/R/B/W/O/G", "--cfen"},
		{"generate-cfen", "", "--start", "YB|Y/R/B/W/O/G"},
		{"parse-cfen", "YB|Y/R/B/W/O/G"},
		{"identify", "YB|Y/R/B/W/O/G"},
		{"verify", "", "--start", "YB|Y/R/B/W/O/G", "--target", "YB|Y/R/B/W/O/G"},
	} {
		output, err := reviewCLI(t, args...)
		if err == nil || !strings.Contains(output, "dimension must be at least 2") {
			t.Errorf("%v: wanted a clear unsupported dimension error, got %v: %s", args, err, output)
		}
	}
}

func review5AlgorithmCube(t *testing.T, id, extra string) *cube.Cube {
	t.Helper()
	state, err := cfen.ParseCFEN(cube.LookupAlgorithm(id)[0].Pattern)
	if err != nil {
		t.Fatal(err)
	}
	c, err := state.ToCube()
	if err != nil {
		t.Fatal(err)
	}
	moves, err := cube.ParseMoves(extra)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.ApplyMoves(moves); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestReview5IdentifySuneWithDifferentPermutation(t *testing.T) {
	c := review5AlgorithmCube(t, "Sune", cube.LookupAlgorithm("PLL-H")[0].Moves)
	solver, err := cube.GetSolver("cfop")
	if err != nil {
		t.Fatal(err)
	}
	result, err := solver.Solve(c)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.Stages[5].CaseName(), "Sune") {
		t.Fatalf("CFOP did not recognize Sune: %v", result.Stages[5])
	}
	pattern, err := cfen.GenerateCFEN(c)
	if err != nil {
		t.Fatal(err)
	}
	output, err := reviewCLI(t, "identify", pattern, "--category", "OLL", "--suggest")
	if err != nil || !strings.Contains(output, "Sune (OLL-27)") {
		t.Fatalf("identify missed the Sune recognized by CFOP: %v: %s", err, output)
	}
}

func TestReview5IdentifyPLLUpToAUF(t *testing.T) {
	for _, auf := range []string{"", "U", "U2", "U'"} {
		c := review5AlgorithmCube(t, "T-Perm", auf)
		matches := findMatchingAlgorithms(c, "", "PLL")
		found := false
		for _, match := range matches {
			if match.Algorithm.CaseID == "PLL-T" {
				found = true
				assertReview5Suggestion(t, c, match, true)
			}
		}
		if !found {
			t.Errorf("T-Perm with %q AUF was not recognized", auf)
		}
	}
}

func assertReview5Suggestion(t *testing.T, input *cube.Cube, match AlgorithmMatch, pll bool) {
	t.Helper()
	state, err := cfen.FromCube(input, cfen.CFENOrientation{Up: cube.Yellow, Front: cube.Blue})
	if err != nil {
		t.Fatal(err)
	}
	replay, err := state.ToCube()
	if err != nil {
		t.Fatal(err)
	}
	moves, err := cube.ParseMoves(match.Algorithm.Moves)
	if err != nil {
		t.Fatal(err)
	}
	if err := replay.ApplyMoves(moves); err != nil {
		t.Fatal(err)
	}
	if !cube.OLLSolved(replay) || pll && !replay.IsSolved() || match.Algorithm.MoveCount != len(moves) {
		t.Fatalf("suggested %s sequence %q failed its stage goal", match.Algorithm.CaseID, match.Algorithm.Moves)
	}
	for slot := 0; slot < 4; slot++ {
		if !cube.F2LSlotSolved(replay, slot) {
			t.Fatalf("suggested %s damaged F2L slot %d", match.Algorithm.CaseID, slot)
		}
	}
}

func TestReview5IdentifyOLLUpToAUFAndPermutation(t *testing.T) {
	for _, auf := range []string{"", "U", "U2", "U'"} {
		c := review5AlgorithmCube(t, "Sune", cube.LookupAlgorithm("PLL-H")[0].Moves+" "+auf)
		matches := findMatchingAlgorithms(c, "", "OLL")
		if len(matches) == 0 {
			t.Fatalf("no OLL recognized after changing permutation and %q AUF", auf)
		}
		for _, match := range matches {
			assertReview5Suggestion(t, c, match, false)
		}
		if future := findMatchingAlgorithms(c, "", "PLL"); len(future) != 0 {
			t.Errorf("identified a future PLL case before OLL: %v", future)
		}
	}
}

func TestReview5CFOPRecognitionSuggestionsReplay(t *testing.T) {
	// Exercise aliases, composed OLL solutions and final PLL AUF as well as
	// the Sune regression. Recognition must leave the supplied cube intact.
	count := 0
	for _, alg := range cube.AlgorithmDatabase {
		if alg.Dimension != 3 || !alg.HasCategory("OLL") && !alg.HasCategory("PLL") {
			continue
		}
		category := "PLL"
		extra := "U"
		if alg.HasCategory("OLL") {
			category = "OLL"
			extra = cube.LookupAlgorithm("PLL-H")[0].Moves + " U"
		}
		c := review5AlgorithmCube(t, alg.CaseID, extra)
		// Imported inverses containing grip rotations may have a different
		// bottom layer. The CFOP fast path requires four solved F2L slots.
		lastLayer := cube.WhiteCrossSolved(c)
		for slot := 0; slot < 4; slot++ {
			lastLayer = lastLayer && cube.F2LSlotSolved(c, slot)
		}
		if !lastLayer {
			continue
		}
		count++
		t.Run(alg.CaseID, func(t *testing.T) {
			before := c.String()
			matches := cfopAlgorithmMatches(c, category)
			if len(matches) != 1 {
				t.Fatalf("wanted a verified %s stage, got %v", category, matches)
			}
			assertReview5Suggestion(t, c, matches[0], category == "PLL")
			if c.String() != before {
				t.Fatal("recognition mutated the input cube")
			}
		})
	}
	if count < 20 {
		t.Fatalf("only exercised %d last-layer cases", count)
	}
}

func TestReview5MatchCFENNoMatchStillExitsZero(t *testing.T) {
	output, err := reviewCLI(t, "match-cfen", "YB|Y9/R9/B9/W9/O9/G9", "YB|R9/Y9/B9/W9/O9/G9")
	if err != nil || !strings.Contains(output, "NO MATCH") {
		t.Fatalf("NO MATCH exit contract changed: %v: %s", err, output)
	}
}

// Execute examples from the rendered help, including nested commands. This
// catches stale CFEN, placeholder moves and handlers that report failure at 0.
func TestReview5EveryHelpExampleRuns(t *testing.T) {
	total := 0
	var visit func(*cobra.Command)
	visit = func(command *cobra.Command) {
		args := strings.Fields(command.CommandPath())[1:]
		help, err := reviewCLI(t, append(args, "--help")...)
		if err != nil {
			t.Fatalf("%s help: %v: %s", command.CommandPath(), err, help)
		}
		help = strings.ReplaceAll(help, "\\\n", " ")
		count := 0
		examples := false
		for _, line := range strings.Split(help, "\n") {
			line = strings.TrimSpace(line)
			if line == "Examples:" {
				examples = true
				continue
			}
			if line == "Usage:" || line == "Flags:" || line == "Global Flags:" || line == "Available Commands:" {
				examples = false
			}
			if !examples || !strings.HasPrefix(line, "cube ") {
				continue
			}
			count++
			total++
			t.Run(fmt.Sprintf("%s/%d", command.CommandPath(), count), func(t *testing.T) {
				// Help examples are shell command lines. The shell supplies their
				// quoting and comments; cube is a local child-process test helper.
				ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
				defer cancel()
				script := "cube() { taskpolicy -b nice -n 15 \"$CUBE_HELP_TEST_BIN\" -test.run=^TestCLIReviewHelper$ -- \"$@\"; }\n" + line
				child := exec.CommandContext(ctx, "taskpolicy", "-b", "nice", "-n", "15", "bash", "-eu", "-c", script)
				child.Env = append(os.Environ(), "CUBE_REVIEW_HELPER=1", "CUBE_HELP_TEST_BIN="+os.Args[0])
				child.Stdin = strings.NewReader(strings.Repeat("\n", 100) + "quit\n")
				output, err := child.CombinedOutput()
				if err != nil || strings.Contains(string(output), "❌ FAIL") ||
					strings.Contains(string(output), "❌ NO MATCH") || strings.Contains(string(output), "No algorithms found.") {
					t.Fatalf("help example failed: %s\n%v: %s", line, err, output)
				}
				if strings.HasPrefix(line, "cube identify ") && strings.Contains(line, "--suggest") &&
					!strings.Contains(string(output), "RECOMMENDED ACTIONS:") {
					t.Fatalf("help example did not produce suggestions: %s\n%s", line, output)
				}
			})
		}
		for _, child := range command.Commands() {
			visit(child)
		}
	}
	visit(rootCmd)
	if total == 0 {
		t.Fatal("no help examples executed")
	}
	t.Logf("executed %d help examples", total)
}
