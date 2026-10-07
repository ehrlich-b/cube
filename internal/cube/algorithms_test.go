package cube

import (
	"sort"
	"strings"
	"testing"
)

// algNames returns the sorted set of names from a slice of algorithms, so tests
// can assert exact resulting sets without depending on database ordering.
func algNames(algs []Algorithm) []string {
	names := make([]string, 0, len(algs))
	for _, a := range algs {
		names = append(names, a.Name)
	}
	sort.Strings(names)
	return names
}

// expectAlgNames asserts that the queried algorithms are exactly the wanted
// names (as a set). algs must be non-nil (or at least non-empty) when len(want)
// > 0; a nil/nil-ish result for a non-empty want is reported as a failure.
func expectAlgNames(t *testing.T, got []Algorithm, want []string) {
	t.Helper()
	gotNames := algNames(got)
	wantNames := append([]string(nil), want...)
	sort.Strings(wantNames)
	if len(gotNames) != len(wantNames) {
		t.Fatalf("got %d results %v, want %d results %v", len(gotNames), gotNames, len(wantNames), wantNames)
	}
	for i := range gotNames {
		if gotNames[i] != wantNames[i] {
			t.Fatalf("results mismatch: got %v, want %v", gotNames, wantNames)
		}
	}
}

// Lookup semantics use a controlled corpus so new imports cannot make a
// substring assertion accidentally depend on unrelated algorithm text.
func lookupFixture(t *testing.T) {
	t.Helper()
	saved := AlgorithmDatabase
	AlgorithmDatabase = []Algorithm{
		{Name: "Sune", CaseID: "OLL-27", Category: "OLL", Moves: "R U R' U R U2 R'", MoveCount: 7, Description: "Orient corners"},
		{Name: "Anti-Sune", CaseID: "OLL-26", Category: "OLL", Moves: "R U2 R' U' R U' R'"},
		{Name: "Cross OLL", CaseID: "OLL-CROSS", Category: "OLL", Moves: "F R U R' U' F'"},
		{Name: "T-Perm", CaseID: "PLL-T", Category: "PLL", Moves: "R U R' F' R U R' U' R' F R2 U' R'"},
		{Name: "Sexy Move", CaseID: "TRIG-1", Category: "Trigger", Moves: "R U R' U'", Aliases: []string{"Original trigger name"}},
		{Name: "Pair", CaseID: "F2L-1", Category: "F2L", Moves: "U R U' R'"},
	}
	t.Cleanup(func() { AlgorithmDatabase = saved })
}

func TestLookupAlgorithmSemantics(t *testing.T) {
	lookupFixture(t)
	for _, test := range []struct {
		query string
		want  []string
	}{
		{"sune", []string{"Sune", "Anti-Sune"}},
		{"U R U2 R'", []string{"Sune"}},
		{"r u r'", []string{"Sune", "Sexy Move", "Cross OLL", "T-Perm"}},
		{"orient corners", []string{"Sune"}},
		{"oll-27", []string{"Sune"}},
		{"Original trigger name", []string{"Sexy Move"}},
		{"zzz-nonexistent", nil},
	} {
		expectAlgNames(t, LookupAlgorithm(test.query), test.want)
	}
}

func TestLookupByMovesSemantics(t *testing.T) {
	lookupFixture(t)
	expectAlgNames(t, LookupByMoves("R U R' U R U2 R'"), []string{"Sune"})
	expectAlgNames(t, LookupByMoves("U R U2 R'"), nil)
}

func TestGetByCategorySemantics(t *testing.T) {
	lookupFixture(t)
	for _, category := range []string{"OLL", "oll", " OLL "} {
		expectAlgNames(t, GetByCategory(category), []string{"Sune", "Anti-Sune", "Cross OLL"})
	}
	expectAlgNames(t, GetByCategory("PLL"), []string{"T-Perm"})
	for _, category := range []string{"Trigger", "TRIGGER", "trigger"} {
		expectAlgNames(t, GetByCategory(category), []string{"Sexy Move"})
	}
	expectAlgNames(t, GetByCategory("F2L"), []string{"Pair"})
}

func TestCalculateMoveCount_KnownAlgorithms(t *testing.T) {
	sune := Algorithm{Moves: "R U R' U R U2 R'"}
	if got := sune.CalculateMoveCount(); got != 7 {
		t.Fatalf("Sune.CalculateMoveCount() = %d, want 7", got)
	}
	tperm := Algorithm{Moves: "R U R' F' R U R' U' R' F R2 U' R'"}
	if got := tperm.CalculateMoveCount(); got != 13 {
		t.Fatalf("T-Perm.CalculateMoveCount() = %d, want 13", got)
	}
}

func TestCalculateMoveCount_EmptyMoves(t *testing.T) {
	alg := Algorithm{Moves: ""}
	if got := alg.CalculateMoveCount(); got != 0 {
		t.Fatalf("CalculateMoveCount() for empty moves = %d, want 0", got)
	}
}

func TestCalculateMoveCount_InvalidMovesSwallowsError(t *testing.T) {
	// CalculateMoveCount swallows the ParseScramble error and returns 0 — it
	// must not panic or surface the error.
	alg := Algorithm{Moves: "not a valid move sequence !!"}
	if got := alg.CalculateMoveCount(); got != 0 {
		t.Fatalf("CalculateMoveCount() for unparseable moves = %d, want 0 (error swallowed)", got)
	}
}

func TestUpdateMoveCount_RestoresDocumentedCount(t *testing.T) {
	alg := AlgorithmDatabase[0] // Sune
	if alg.Name != "Sune" {
		t.Fatalf("test precondition: AlgorithmDatabase[0] = %q, want Sune; the live database changed", alg.Name)
	}
	alg.MoveCount = 0 // deliberately zeroed copy

	if err := alg.UpdateMoveCount(); err != nil {
		t.Fatalf("UpdateMoveCount() returned unexpected error: %v", err)
	}
	if alg.MoveCount != 7 {
		t.Fatalf("UpdateMoveCount() set MoveCount = %d, want 7", alg.MoveCount)
	}
}

func TestUpdateMoveCount_InvalidMovesPropagatesError(t *testing.T) {
	// Unlike CalculateMoveCount (which swallows parse errors), UpdateMoveCount
	// returns the parser error.
	alg := Algorithm{Moves: "!!!"}
	err := alg.UpdateMoveCount()
	if err == nil {
		t.Fatal("UpdateMoveCount() returned nil error for unparseable moves, want non-nil")
	}
	if alg.MoveCount != 0 {
		t.Fatalf("UpdateMoveCount() set MoveCount = %d on failure, want 0", alg.MoveCount)
	}
}

func TestLookupReturnsCopies_ValueSemantics(t *testing.T) {
	lookupFixture(t)
	// LookupAlgorithm / LookupByMoves / GetByCategory range over
	// AlgorithmDatabase by value and append copies. Mutating a returned
	// algorithm must NOT affect the slice returned on a later lookup, nor the
	// underlying database — this is a surprising trap for anyone who thinks
	// they can update a stored algorithm through a lookup result.
	got := LookupByMoves("R U R' U R U2 R'")
	if len(got) != 1 {
		t.Fatalf("test precondition: exact Sune lookup returned %d results, want 1", len(got))
	}
	got[0].Name = "Mutated Sune"
	got[0].MoveCount = 999

	again := LookupByMoves("R U R' U R U2 R'")
	if len(again) != 1 || again[0].Name != "Sune" || again[0].MoveCount != 7 {
		t.Fatalf("mutating a lookup result leaked into the database: got %+v", again)
	}

	stored := AlgorithmDatabase[0]
	if stored.Name != "Sune" || stored.MoveCount != 7 {
		t.Fatalf("mutating a lookup result mutated AlgorithmDatabase in place: got %+v", stored)
	}
}

func TestAlgorithmDatabaseMoveCountsMatchCalculation(t *testing.T) {
	// Sanity check on the live database contents: every entry's stored
	// MoveCount must agree with a fresh re-derivation. Also serves as a guard
	// that the top-of-file dead-code mismatch (see commented-out block
	// replica) never sneaks back in with a wrong count.
	for i := range AlgorithmDatabase {
		alg := &AlgorithmDatabase[i]
		want := alg.CalculateMoveCount()
		if alg.MoveCount != want {
			t.Errorf("AlgorithmDatabase[%d] %s: stored MoveCount = %d, re-derived = %d", i, alg.Name, alg.MoveCount, want)
		}
		if want == 0 {
			t.Errorf("AlgorithmDatabase[%d] %s: re-derived move count is 0, moves %q must parse", i, alg.Name, alg.Moves)
		}
	}
}

func TestLookupAlgorithm_MovesQueryIsCaseInsensitiveOnWholeSequence(t *testing.T) {
	// Pins the substring semantics against the token stream: the query must not
	// split moves mid-token ("RU2" inside "R U2 R'" should not match, because
	// "RU2" as a token sequence never appears).
	got := LookupAlgorithm("RU2R'")
	if len(got) != 0 {
		t.Fatalf("expected '%s' (no separators) to not match as a move substring, got %v", "RU2R'", algNames(got))
	}
}

func TestAlgorithmDatabaseSanity(t *testing.T) {
	if len(AlgorithmDatabase) < 100 {
		t.Fatal("comprehensive database was not loaded")
	}
	ids := map[string]bool{}
	for _, a := range AlgorithmDatabase {
		if ids[a.CaseID] || a.CaseID == "" {
			t.Fatal("duplicate or empty case ID", a.CaseID)
		}
		ids[a.CaseID] = true
		if a.Pattern == "" || a.Dimension < 2 || len(a.Sources) == 0 {
			t.Fatal("missing recognition/provenance", a.CaseID)
		}
		if strings.Contains(a.Recognition, "contentReference") {
			t.Fatal("unstripped citation", a.CaseID)
		}
	}
	for _, category := range []string{"OLL", "PLL", "F2L", "Trigger", "2x2-CLL", "4x4-PARITY"} {
		if len(GetByCategory(category)) == 0 {
			t.Fatal("missing category", category)
		}
	}
	if a := LookupAlgorithm("OLL-27"); len(a) != 1 || a[0].InverseID != "OLL-26" {
		t.Fatal("Sune inverse relationship missing", a)
	}
}
