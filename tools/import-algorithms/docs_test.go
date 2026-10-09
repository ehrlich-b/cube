package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"testing"
)

func TestDocumentedDatabaseCountsMatchImportReport(t *testing.T) {
	root := filepath.Join("..", "..")
	data, err := os.ReadFile(filepath.Join(root, "alg_dumps", "import-report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report summary
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	total := report.Verified3x3 + report.VerifiedOther
	for _, document := range []struct {
		name   string
		checks map[string][]int
	}{
		{"README.md", map[string][]int{
			`contain \*\*(\d+) rows\*\*`:                                                     {report.Rows},
			`accepts \*\*(\d+)\*\*, merges \*\*(\d+)\*\*`:                                    {report.Imported, report.Merged},
			`quarantines \*\*(\d+)\*\*`:                                                      {report.Quarantined},
			`\*\*(\d+) unique entries: (\d+) for 3×3, (\d+) for other sizes\*\*`:             {total, report.Verified3x3, report.VerifiedOther},
			`\| F2L / OLL / PLL \| (\d+) / (\d+) / (\d+) \|`:                                 {report.Categories["F2L"], report.Categories["OLL"], report.Categories["PLL"]},
			`\| Trigger / Advanced \| (\d+) / (\d+) \|`:                                      {report.Categories["Trigger"], report.Categories["Advanced"]},
			`\| Roux CMLL / LSE \| (\d+) / (\d+) \|`:                                         {report.Categories["ROUX-CMLL"], report.Categories["ROUX-LSE"]},
			`\| 2×2 CLL / EG1 / EG2 / OLL / PBL \| (\d+) / (\d+) / (\d+) / (\d+) / (\d+) \|`: {report.Categories["2x2-CLL"], report.Categories["2x2-EG1"], report.Categories["2x2-EG2"], report.Categories["2x2-OLL"], report.Categories["2x2-PBL"]},
			`\| 4×4 / 5×5 / 6×6 parity \| (\d+) / (\d+) / (\d+) \|`:                          {report.Categories["4x4-PARITY"], report.Categories["5x5-PARITY"], report.Categories["6x6-PARITY"]},
			`\*\*(\d+) inverse pairs and (\d+) mirror pairs\*\*`:                             {report.InversePairs, report.MirrorPairs},
		}},
		{"TODO.md", map[string][]int{
			`Algorithm DB: \*\*(\d+) unique entries`:                         {total},
			`(\d+)/(\d+) CSV rows accepted, (\d+) merged, (\d+) quarantined`: {report.Imported, report.Rows, report.Merged, report.Quarantined},
		}},
		{"docs/solvers.md", map[string][]int{
			`All (\d+) database\s+entries carry inverse-to-solved verification patterns; (\d+) raw rows are quarantined`: {total, report.Quarantined},
		}},
	} {
		data, err := os.ReadFile(filepath.Join(root, document.name))
		if err != nil {
			t.Fatal(err)
		}
		for pattern, want := range document.checks {
			matches := regexp.MustCompile(pattern).FindStringSubmatch(string(data))
			if matches == nil {
				t.Errorf("%s: database count statement missing (%s)", document.name, pattern)
				continue
			}
			var got []int
			for _, match := range matches[1:] {
				n, err := strconv.Atoi(match)
				if err != nil {
					t.Fatal(err)
				}
				got = append(got, n)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("%s: %q has counts %v, import report requires %v", document.name, matches[0], got, want)
			}
		}
	}
}
