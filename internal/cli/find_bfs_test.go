package cli

import (
	"testing"

	"github.com/ehrlich-b/cube/internal/cube"
)

func mustMove(t *testing.T, notation string) cube.Move {
	t.Helper()
	m, err := cube.ParseMove(notation)
	if err != nil {
		t.Fatalf("parse move %q: %v", notation, err)
	}
	return m
}

func upFaceAllYellow(c *cube.Cube) bool {
	for row := 0; row < c.Size; row++ {
		for col := 0; col < c.Size; col++ {
			if c.Faces[cube.Up][row][col] != cube.Yellow {
				return false
			}
		}
	}
	return true
}

func TestBFSAlreadyAtTargetShortCircuit(t *testing.T) {
	alwaysTrue := func(*cube.Cube) bool { return true }
	results := breadthFirstSearch(cube.NewCube(3), alwaysTrue, 10)

	if len(results) != 1 {
		t.Fatalf("expected exactly 1 sentinel result, got %d", len(results))
	}
	if results[0].notation != "(already at target)" {
		t.Errorf("expected notation %q, got %q", "(already at target)", results[0].notation)
	}
	if results[0].moves == nil {
		t.Errorf("expected a non-nil empty moves slice, got nil")
	}
	if len(results[0].moves) != 0 {
		t.Errorf("expected an empty moves slice, got %d moves", len(results[0].moves))
	}
}

func TestBFSFindableTargetAtKnownDepth(t *testing.T) {
	start := cube.NewCube(3)

	target := cube.NewCube(3)
	target.ApplyMove(mustMove(t, "R"))
	want := target.String()

	isTarget := func(c *cube.Cube) bool { return c.String() == want }
	results := breadthFirstSearch(start, isTarget, 1)

	if len(results) != 1 {
		t.Fatalf("expected exactly 1 result, got %d", len(results))
	}
	if results[0].notation != "R" {
		t.Errorf("expected notation %q, got %q", "R", results[0].notation)
	}
	if len(results[0].moves) != 1 {
		t.Errorf("expected exactly 1 move, got %d", len(results[0].moves))
	}
}

func TestBFSUnreachableTargetReturnsEmpty(t *testing.T) {
	alwaysFalse := func(*cube.Cube) bool { return false }
	results := breadthFirstSearch(cube.NewCube(3), alwaysFalse, 1)

	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestBFSZeroMaxDepthReturnsEmpty(t *testing.T) {
	target := cube.NewCube(3)
	target.ApplyMove(mustMove(t, "R"))
	want := target.String()

	results := breadthFirstSearch(cube.NewCube(3), func(c *cube.Cube) bool { return c.String() == want }, 0)

	if len(results) != 0 {
		t.Fatalf("expected 0 results at maxDepth=0, got %d", len(results))
	}
}

func TestBFSResultCapLimitsResultsToTen(t *testing.T) {
	start := cube.NewCube(3)
	start.ApplyMove(mustMove(t, "F"))

	results := breadthFirstSearch(start, upFaceAllYellow, 6)

	if len(results) != 10 {
		t.Fatalf("expected exactly 10 results (the cap), got %d", len(results))
	}
}

func TestBFSResultCapSearchSpaceExhaustedBeforeCap(t *testing.T) {
	start := cube.NewCube(3)
	start.ApplyMove(mustMove(t, "F"))

	results := breadthFirstSearch(start, upFaceAllYellow, 4)

	if len(results) != 6 {
		t.Fatalf("expected exactly 6 results (search space exhausted before the cap), got %d", len(results))
	}
}

func TestCopyCubeIndependence(t *testing.T) {
	original := cube.NewCube(3)
	copied := copyCube(original)

	copied.ApplyMove(mustMove(t, "R"))

	if !original.IsSolved() {
		t.Errorf("original cube was mutated by an operation on its copy; expected it to stay solved")
	}
	if copied.IsSolved() {
		t.Errorf("expected the copy to be unsolved after applying a move")
	}
}

func TestCopyCubePreservesSize(t *testing.T) {
	original := cube.NewCube(4)
	copied := copyCube(original)

	if copied.Size != 4 {
		t.Fatalf("expected copied Size to be 4, got %d", copied.Size)
	}

	for face := 0; face < 6; face++ {
		for row := 0; row < original.Size; row++ {
			for col := 0; col < original.Size; col++ {
				if copied.Faces[face][row][col] != original.Faces[face][row][col] {
					t.Errorf("sticker mismatch at face=%d row=%d col=%d: got %v want %v",
						face, row, col, copied.Faces[face][row][col], original.Faces[face][row][col])
				}
			}
		}
	}
}
