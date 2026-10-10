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

func TestLookupAlgorithm_SubstringMatchesSuneAndAntiSune(t *testing.T) {
	got := LookupAlgorithm("sune")
	expectAlgNames(t, got, []string{"Sune", "Anti-Sune"})
}

func TestLookupAlgorithm_SubstringByMoves(t *testing.T) {
	// "U R U2 R'" is a substring (in token terms) of Sune's moves
	// "R U R' U R U2 R'", and does not appear in any other live algorithm.
	got := LookupAlgorithm("U R U2 R'")
	expectAlgNames(t, got, []string{"Sune"})
}

func TestLookupAlgorithm_SubstringByMovesCaseInsensitive(t *testing.T) {
	// Lowercased query matches lowercased move strings. "R U R'" (uppercased)
	// appears in Sune, Sexy Move, Cross OLL, and T-Perm — but NOT in
	// Anti-Sune (whose moves start "R U2 ..."). The query itself is
	// lowercase, proving case-insensitive matching.
	got := LookupAlgorithm("r u r'")
	expectAlgNames(t, got, []string{"Sune", "Sexy Move", "Cross OLL", "T-Perm"})
}

func TestLookupAlgorithm_SubstringByDescription(t *testing.T) {
	// Only Sune's description contains "orient corners".
	got := LookupAlgorithm("orient corners")
	expectAlgNames(t, got, []string{"Sune"})
}

func TestLookupAlgorithm_SubstringByCaseID(t *testing.T) {
	// Query is lowercased by LookupAlgorithm; "oll-27" appears only in Sune's
	// CaseID ("OLL-27").
	got := LookupAlgorithm("oll-27")
	expectAlgNames(t, got, []string{"Sune"})
}

func TestLookupAlgorithm_NoMatch(t *testing.T) {
	got := LookupAlgorithm("zzz-nonexistent")
	if len(got) != 0 {
		t.Fatalf("expected 0 results, got %d: %v", len(got), algNames(got))
	}
}

func TestLookupByMoves_ExactMatchSune(t *testing.T) {
	got := LookupByMoves("R U R' U R U2 R'")
	expectAlgNames(t, got, []string{"Sune"})
}

func TestLookupByMoves_SubstringIsNotExact(t *testing.T) {
	// "U R U2 R'" is only a substring of Sune's moves, not the full string, so
	// an exact-match lookup must return nothing.
	got := LookupByMoves("U R U2 R'")
	if len(got) != 0 {
		t.Fatalf("expected exact-match lookup of a substring to return 0 results, got %d: %v", len(got), algNames(got))
	}
}

func TestGetByCategory_OLL(t *testing.T) {
	got := GetByCategory("OLL")
	expectAlgNames(t, got, []string{"Sune", "Anti-Sune", "Cross OLL"})
}

func TestGetByCategory_OLLLowercase(t *testing.T) {
	got := GetByCategory("oll")
	expectAlgNames(t, got, []string{"Sune", "Anti-Sune", "Cross OLL"})
}

func TestGetByCategory_PLL(t *testing.T) {
	got := GetByCategory("PLL")
	expectAlgNames(t, got, []string{"T-Perm"})
}

func TestGetByCategory_TriggerIsUnreachable(t *testing.T) {
	// SURPRISING (pinned, not fixed): Sexy Move's stored Category is the
	// mixed-case string "Trigger", but GetByCategory uppercases the query to
	// "TRIGGER" before doing an exact compare. So Sexy Move can never be found
	// via GetByCategory — not even with "TRIGGER" or "trigger" — and is only
	// reachable through LookupAlgorithm / LookupByMoves. This is a data bug in
	// the live server (the task's claim that all stored Categories are already
	// uppercase does not hold).
	if got := GetByCategory("Trigger"); len(got) != 0 {
		t.Fatalf("GetByCategory(%q) = %v, want 0 results (exact-match against uppercased query)", "Trigger", algNames(got))
	}
	if got := GetByCategory("TRIGGER"); len(got) != 0 {
		t.Fatalf("GetByCategory(%q) = %v, want 0 results (stored category is 'Trigger', not 'TRIGGER')", "TRIGGER", algNames(got))
	}
	// Confirms Sexy Move IS in the live database and findable by moves.
	if got := LookupByMoves("R U R' U'"); len(got) != 1 || got[0].Name != "Sexy Move" {
		t.Fatalf("Sexy Move should be present in the live database, LookupByMoves returned %v", algNames(got))
	}
}

func TestGetByCategory_F2L(t *testing.T) {
	// There are no F2L entries in the live database today.
	got := GetByCategory("F2L")
	if len(got) != 0 {
		t.Fatalf("expected 0 F2L results, got %d: %v", len(got), algNames(got))
	}
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
	// The whole point of this task: the live database is exactly the 5 entries
	// after the "Temporarily commenting out" marker, no matter how many dead
	// entries hide inside /* ... */. Pin the constituent sets for the
	// categories reachable via GetByCategory today (Trigger is unreachable —
	// see TestGetByCategory_TriggerIsUnreachable).
	oll := algNames(GetByCategory("OLL"))
	pll := algNames(GetByCategory("PLL"))
	if strings.Join(oll, ",") != "Anti-Sune,Cross OLL,Sune" {
		t.Fatalf("OLL set = %v, want [Sune Anti-Sune Cross OLL]", oll)
	}
	if strings.Join(pll, ",") != "T-Perm" {
		t.Fatalf("PLL set = %v, want [T-Perm]", pll)
	}
	if len(AlgorithmDatabase) != 5 {
		t.Fatalf("AlgorithmDatabase has %d live entries, want 5 (the TODO.md count of 63 includes the commented-out block)", len(AlgorithmDatabase))
	}
}
