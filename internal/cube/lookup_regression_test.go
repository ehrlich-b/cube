package cube

import "testing"

func TestLookupPrefersExactIdentity(t *testing.T) {
	for _, id := range []string{"OLL-1", "oll-2", "2x2-OLL-1"} {
		got := LookupAlgorithm(id)
		if len(got) == 0 || !(got[0].CaseID == id || (id == "oll-2" && got[0].CaseID == "OLL-2")) {
			t.Fatalf("%q: first result is not exact: %v", id, got)
		}
	}
}

func TestLookupCommonPermutationNames(t *testing.T) {
	for query, ids := range map[string][]string{
		"h-perm": {"PLL-H"}, "Y-Perm": {"PLL-Y"}, "z-PERM": {"PLL-Z"},
		"U-Perm": {"PLL-Ua", "PLL-Ub"}, "u perm": {"PLL-Ua", "PLL-Ub"},
	} {
		got := LookupAlgorithm(query)
		if len(got) != len(ids) {
			t.Errorf("%q got %d results, want %v", query, len(got), ids)
			continue
		}
		for i, id := range ids {
			if got[i].CaseID != id {
				t.Errorf("%q result %d = %s, want %s", query, i, got[i].CaseID, id)
			}
		}
	}
}
