package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
)

// Execute the real Cobra command surface in a child process, including legacy
// handlers that call os.Exit. No separate binary build or remote access needed.
func TestCLIReviewHelper(t *testing.T) {
	if os.Getenv("CUBE_REVIEW_HELPER") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			rootCmd.SetArgs(os.Args[i+1:])
			if err := Execute(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}
	os.Exit(2)
}

func reviewCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	command := exec.Command("taskpolicy", append([]string{"-b", "nice", "-n", "15", os.Args[0], "-test.run=^TestCLIReviewHelper$", "--"}, args...)...)
	command.Env = append(os.Environ(), "CUBE_REVIEW_HELPER=1")
	output, err := command.CombinedOutput()
	return string(output), err
}

func TestReviewInvalidMovesReturnErrorsAcrossCommands(t *testing.T) {
	for _, args := range [][]string{
		{"twist", "4R"}, {"show", "4R"}, {"generate-cfen", "4R"},
		{"verify", "4R"}, {"verify-cfen", "4R", "", "--target", "YB|Y9/R9/B9/W9/O9/G9"},
		{"find", "sequence", "4R", "--max-moves", "0"},
		{"find", "pattern", "solved", "--from", "4R", "--max-moves", "0"},
		{"find", "pattern", "solved", "--moves", "4R", "--max-moves", "0"},
		{"twist", "4Rw"}, {"twist", "M", "-d", "4"}, {"twist", "E", "-d", "2"},
		{"show", "S", "-d", "4"}, {"generate-cfen", "M", "--dimension", "4"},
		{"verify", "M", "--start", "YB|Y16/R16/B16/W16/O16/G16", "--target", "YB|Y16/R16/B16/W16/O16/G16"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			output, err := reviewCLI(t, args...)
			if err == nil || strings.Contains(output, "panic:") ||
				(!strings.Contains(output, "layer") && !strings.Contains(output, "slice")) {
				t.Fatalf("wanted a clear move error; got %v: %s", err, output)
			}
		})
	}
}

func TestReviewInvalidDimensions(t *testing.T) {
	for _, name := range []string{"twist", "show", "generate-cfen"} {
		for _, dimension := range []string{"0", "1", "-1"} {
			output, err := reviewCLI(t, name, "R", "--dimension", dimension)
			if err == nil || !strings.Contains(output, "dimension must be at least 2") {
				t.Errorf("%s -d %s: wanted dimension error, got %v: %s", name, dimension, err, output)
			}
		}
	}
}

func TestReviewCommonPermutationPreviews(t *testing.T) {
	for _, name := range []string{"h-perm", "U-Perm", "Y-Perm", "Z-Perm"} {
		output, err := reviewCLI(t, "lookup", name, "--preview")
		if err != nil || !strings.Contains(output, "Top face after algorithm:") {
			t.Errorf("%s preview failed: %v: %s", name, err, output)
		}
	}
}

func TestReviewInvalidNotationFailsAcrossCommands(t *testing.T) {
	for _, args := range [][]string{
		{"show", "INVALID"}, {"twist", "INVALID"}, {"generate-cfen", "INVALID"},
		{"optimize", "INVALID"}, {"verify", "INVALID"}, {"solve", "INVALID"}, {"learn", "INVALID"},
		{"find", "sequence", "INVALID"}, {"find", "pattern", "solved", "--from", "INVALID"},
		{"verify-cfen", "INVALID", "", "--target", "YB|Y9/R9/B9/W9/O9/G9"},
	} {
		if output, err := reviewCLI(t, args...); err == nil {
			t.Errorf("%v exited successfully: %s", args, output)
		}
	}
}

func TestReviewIdentifyRecognition(t *testing.T) {
	if got := findMatchingAlgorithms(cube.NewCube(3), "YB|Y9/R9/B9/W9/O9/G9", "OLL"); len(got) != 0 {
		t.Fatalf("solved state matched %d OLL cases", len(got))
	}
	alg := cube.LookupAlgorithm("Sune")[0]
	state, err := cfen.ParseCFEN(alg.Pattern)
	if err != nil {
		t.Fatal(err)
	}
	c, err := state.ToCube()
	if err != nil {
		t.Fatal(err)
	}
	got := findMatchingAlgorithms(c, alg.Pattern, "oll")
	if len(got) != 1 || got[0].Algorithm.CaseID != alg.CaseID {
		t.Fatalf("Sune recognition returned %v", got)
	}
}

func TestReviewShowAlgorithmSolvesRecognition(t *testing.T) {
	for _, id := range []string{"Sune", "OLL-1", "OLL-2", "2x2-OLL-1", "4x4-PLL-PARITY"} {
		output, err := reviewCLI(t, "show-alg", id)
		alg := cube.LookupAlgorithm(id)[0]
		if err != nil || !strings.Contains(output, "START STATE:") ||
			!strings.Contains(output, "🎯 FINAL STATE:\n"+cube.NewCube(alg.Dimension).String()) {
			t.Errorf("%s did not demonstrate recognition to solved: %v: %s", id, err, output)
		}
	}
}

func TestReviewLookupQueryAndCategory(t *testing.T) {
	output, err := reviewCLI(t, "lookup", "Sune", "--category", "PLL")
	if err != nil || !strings.Contains(output, "No algorithms found") {
		t.Fatalf("query ignored with category: %v: %s", err, output)
	}
	output, err = reviewCLI(t, "lookup", "Sune", "--category", "OLL")
	if err != nil || !strings.Contains(output, "OLL-27 - Sune") || strings.Contains(output, "T-Perm") {
		t.Fatalf("combined filters lost Sune: %v: %s", err, output)
	}
}

func TestReviewVerifySuneHelpExample(t *testing.T) {
	// Read the actual documented start state so this fails for the old example.
	start := strings.Split(strings.Split(verifyCmd.Long, "--start \"")[1], "\"")[0]
	output, err := reviewCLI(t, "verify", "R U R' U R U2 R'", "--start", start, "--target", "YB|Y9/?9/?9/?9/?9/?9")
	if err != nil || !strings.Contains(output, "PASS") {
		t.Fatalf("documented Sune example failed: %v: %s", err, output)
	}
}
