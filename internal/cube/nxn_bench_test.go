package cube

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

// Export independent fixtures with test/nxn_oracle.py --export-cases, then set
// CUBE_NXN_CASES to that file and run this test with -cpuprofile/-memprofile.
// Rebuild reduction tables for each case to include fresh-process setup costs.
func TestNxNProfileCases(t *testing.T) {
	path := os.Getenv("CUBE_NXN_CASES")
	if path == "" {
		t.Skip("set CUBE_NXN_CASES to exported independent oracle fixtures")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Size  int
		State string
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for i, fixture := range cases {
		if fixture.Size < 2 || fixture.Size > 7 || len(fixture.State) != 6*fixture.Size*fixture.Size {
			t.Fatalf("invalid fixture %d", i)
		}
		c := NewCube(fixture.Size)
		for pos, color := range fixture.State {
			f, r, col := indexToCoord(pos, c.Size)
			id := strings.IndexRune("WYROBG", color)
			if id < 0 {
				t.Fatalf("invalid color in fixture %d", i)
			}
			c.Faces[f][r][col] = Color(id)
		}
		reductionCache.tables[c.Size] = nil
		started := time.Now()
		result, err := (&ReductionSolver{}).Solve(c)
		if err != nil {
			t.Fatalf("fixture %d: %v", i, err)
		}
		t.Logf("fixture %d: %dx%d %d moves %v", i, c.Size, c.Size, result.Steps, time.Since(started))
		if err := c.ApplyMoves(result.Solution); err != nil || !c.IsSolved() || !nxnCenterMatched(c) {
			t.Fatalf("fixture %d failed replay: %v", i, err)
		}
	}
}
