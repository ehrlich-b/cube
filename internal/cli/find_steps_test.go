package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/ehrlich-b/cube/internal/cube"
)

func TestFindSequenceNumberedSteps(t *testing.T) {
	for _, test := range []struct {
		name, scramble, alphabet, steps string
	}{
		{"faces", "R U", "", "   Steps:\n   1. U'\n   2. R'\n"},
		{"restricted", "R2", "R", "   Steps:\n   1. R\n   2. R\n"},
	} {
		for _, showSteps := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/steps=%t", test.name, showSteps), func(t *testing.T) {
				r, w, err := os.Pipe()
				if err != nil {
					t.Fatal(err)
				}
				defer r.Close()
				stdout := os.Stdout
				os.Stdout = w
				defer func() { os.Stdout = stdout }()
				var moves []cube.Move
				if test.alphabet != "" {
					moves, err = cube.ParseMoves(test.alphabet)
					if err != nil {
						t.Fatal(err)
					}
				}
				err = runSequenceSearchWithOptions(test.scramble, 2, showSteps, "", findSearchOptions{dimension: 3, moves: moves})
				w.Close()
				output, readErr := io.ReadAll(r)
				if err != nil || readErr != nil {
					t.Fatalf("find: %v; read output: %v", err, readErr)
				}
				if showSteps && !strings.Contains(string(output), test.steps) {
					t.Fatalf("missing numbered moves %q:\n%s", test.steps, output)
				}
				if !showSteps && strings.Contains(string(output), "Steps:") {
					t.Fatalf("printed steps without --steps:\n%s", output)
				}
			})
		}
	}
}
