package cli

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
)

func TestShowAlgorithmDimension(t *testing.T) {
	for _, id := range []string{"2x2-OLL-1", "4x4-PLL-PARITY"} {
		t.Run(id, func(t *testing.T) {
			alg := cube.LookupAlgorithm(id)[0]
			want := cube.NewCube(alg.Dimension)
			pattern, err := cfen.ParseCFEN(alg.Pattern)
			if err != nil {
				t.Fatal(err)
			}
			start, err := pattern.ToCube()
			if err != nil {
				t.Fatal(err)
			}
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer r.Close()
			saved := os.Stdout
			os.Stdout = w
			defer func() { os.Stdout = saved }()
			err = showAlgCmd.RunE(showAlgCmd, []string{id})
			w.Close()
			output, readErr := io.ReadAll(r)
			if err != nil || readErr != nil {
				t.Fatal(err, readErr)
			}
			if !strings.Contains(string(output), "🎯 FINAL STATE:\n"+want.String()) {
				t.Fatalf("%s rendered a different cube size; wanted %d stickers", id, 6*alg.Dimension*alg.Dimension)
			}
			if !strings.Contains(string(output), "START STATE:\n"+start.String()) {
				t.Fatal("demonstration did not start at its recognition state")
			}
		})
	}
}

func TestAnimatedAlgorithmDimension(t *testing.T) {
	// One move exercises the animated renderer without waiting for stdin.
	alg := cube.Algorithm{Moves: "R", Dimension: 2}
	want := cube.NewCube(2)
	start := cube.NewCube(2)
	move, _ := cube.ParseMove("R'")
	start.ApplyMove(move)
	alg.Pattern, _ = cfen.GenerateCFEN(start)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	saved := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = saved }()
	err = showAlgorithmAnimated(alg, false)
	w.Close()
	output, readErr := io.ReadAll(r)
	if err != nil || readErr != nil {
		t.Fatal(err, readErr)
	}
	if !strings.Contains(string(output), "🎯 FINAL STATE:\n"+want.String()) ||
		!strings.Contains(string(output), "START STATE:\n"+start.String()) {
		t.Fatal("animated 2x2 renders the wrong number of stickers")
	}
}
