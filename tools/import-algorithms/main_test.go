package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
)

func TestNormalizeNotation(t *testing.T) {
	for _, test := range []struct{ input, want string }{
		{"(r U R’ U') (r' F R F')", "Rw U R' U' Rw' F R F'"},
		{"R2' U2’ F'2", "R2 U2 F2"},
		{"(R U R' U')^2", "R U R' U' R U R' U'"},
		{"U U (U2)", "U2"},
		{"U U U2", "U U U2"},
		{"2R2 B2 2L'", "2R2 B2 2L'"},
	} {
		got, err := normalize(test.input)
		if err != nil || got != test.want {
			t.Fatalf("normalize(%q)=%q,%v; want %q", test.input, got, err, test.want)
		}
	}
	for _, input := range []string{"(apply 4x4 OLL parity alg)", "(no algorithm needed)", "(R U", "R U)", "R3", "R ^12", ""} {
		if _, err := normalize(input); err == nil {
			t.Fatal("invalid input accepted", input)
		}
	}
}

func TestImportMergeAndQuarantine(t *testing.T) {
	dir := t.TempDir()
	rows := [][]string{
		{"F2L-TEST", "Pair", "CFOP-F2L", "U (R U' R')", "Insert pair", "Pair on top:contentReference[oaicite:2]{index=2}", "local fixture"},
		{"TRIG-TEST", "Alias pair", "Trigger", "U R U' R'", "Same moves", "Pair", "local fixture"},
		{"OLL-BAD", "Wrong category", "CFOP-OLL", "R", "Orient LL", "Invalid stage", "local fixture"},
		{"REF", "Reference", "Advanced", "(use other algorithm)", "Unresolved", "Reference", "local fixture"},
	}
	for i, group := range [][][]string{rows[:1], rows[1:]} {
		file, err := os.Create(filepath.Join(dir, string(rune('a'+i))+".csv"))
		if err != nil {
			t.Fatal(err)
		}
		writer := csv.NewWriter(file)
		writer.WriteAll(group)
		if err := writer.Error(); err != nil {
			t.Fatal(err)
		}
		file.Close()
	}
	db, quarantine, report, err := importFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if report.Rows != 4 || report.Imported != 2 || report.Merged != 1 || report.Quarantined != 2 || len(db) != 6 {
		t.Fatalf("wrong accounting: %+v", report)
	}
	if len(quarantine) != 2 || !strings.Contains(quarantine[0].Reason, "first two layers") || !strings.Contains(quarantine[1].Reason, "reference") {
		t.Fatal("missing quarantine reasons", quarantine)
	}
	pair := db[len(db)-1]
	if !pair.HasCategory("F2L") || !pair.HasCategory("Trigger") || len(pair.Sources) != 2 || !matches(pair.Aliases, "TRIG-TEST") {
		t.Fatal("merged provenance/category lost", pair)
	}
	if strings.Contains(pair.Recognition, "contentReference") {
		t.Fatal("citation artifact retained")
	}
	second, rejectedAgain, secondReport, err := importFiles(dir)
	if err != nil || !reflect.DeepEqual(db, second) || !reflect.DeepEqual(quarantine, rejectedAgain) || !reflect.DeepEqual(report, secondReport) {
		t.Fatal("import is not deterministic", err)
	}
}

func matches(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

func TestCommittedDatabaseIsReproducibleAndVerified(t *testing.T) {
	root := filepath.Join("..", "..")
	db, quarantine, report, err := importFiles(filepath.Join(root, "alg_dumps"))
	if err != nil {
		t.Fatal(err)
	}
	if report.Files != 9 || report.Rows != 159 || report.Imported+report.Quarantined != report.Rows || len(db) != 5+report.Imported-report.Merged {
		t.Fatalf("unaccounted input: %+v", report)
	}
	for name, value := range map[string]any{"internal/cube/algorithms_data.json": db, "alg_dumps/quarantine.json": quarantine, "alg_dumps/import-report.json": report} {
		want, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, append(want, '\n')) {
			t.Fatalf("%s is stale; run tools/import-algorithms", name)
		}
	}
	for _, a := range db {
		start, err := cfen.ParseCFEN(a.Pattern)
		if err != nil {
			t.Fatal(a.CaseID, err)
		}
		c, err := start.ToCube()
		if err != nil {
			t.Fatal(a.CaseID, err)
		}
		moves, err := cube.ParseMoves(a.Moves)
		if err != nil {
			t.Fatal(a.CaseID, err)
		}
		c.ApplyMoves(moves)
		target, _ := cfen.GenerateCFEN(cube.NewCube(a.Dimension))
		goal, _ := cfen.ParseCFEN(target)
		ok, err := goal.MatchesCube(c)
		if err != nil || !ok {
			t.Fatalf("%s does not verify against its pattern: %v", a.CaseID, err)
		}
	}
	t.Logf("imported %d rows, merged %d, quarantined %d; verified %d 3x3 + %d other; categories %+v", report.Imported, report.Merged, report.Quarantined, report.Verified3x3, report.VerifiedOther, report.Categories)
}

func TestInverseMirrorRelationships(t *testing.T) {
	moves, _ := cube.ParseMoves("R U2 M E S x y z Rw 2R")
	want := "L' U2 M E' S' x y' z' Lw' 2L'"
	if got := cube.FormatMoves(mirror(moves)); got != want {
		t.Fatalf("mirror=%q want %q", got, want)
	}
	if cube.FormatMoves(mirror(mirror(moves))) != cube.FormatMoves(moves) {
		t.Fatal("mirror is not an involution")
	}
	db, _, _, err := importFiles(filepath.Join("..", "..", "alg_dumps"))
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]cube.Algorithm{}
	for _, a := range db {
		byID[a.CaseID] = a
	}
	for _, a := range db {
		moves, _ := cube.ParseMoves(a.Moves)
		for id, transformed := range map[string][]cube.Move{a.InverseID: inverse(moves), a.Mirror: mirror(moves)} {
			if id == "" {
				continue
			}
			other, ok := byID[id]
			if !ok {
				t.Fatal("unknown relationship", id)
			}
			seq, _ := cube.ParseMoves(other.Moves)
			if signature(a.Dimension, transformed) != signature(other.Dimension, seq) {
				t.Fatal("wrong relationship", a.CaseID, id)
			}
		}
	}
}
