package cfen

import (
	"strings"
	"testing"

	"github.com/ehrlich-b/cube/internal/cube"
)

var malformedCFENs = []struct {
	name        string
	cfenStr     string
	errContains string
}{
	{"empty", "", "expected 'orientation|faces'"},
	{"orientation-only", "YB", "expected 'orientation|faces'"},
	{"five-faces", "YB|W9/R9/B9/Y9/O9", "expected 6 faces"},
	{"seven-faces", "YB|W9/R9/B9/Y9/O9/G9/X9", "expected 6 faces"},
	{"unknown-orientation-color", "XY|W9/R9/B9/Y9/O9/G9", "unknown color character"},
	{"one-char-orientation", "Y|W9/R9/B9/Y9/O9/G9", "orientation must be exactly 2 characters"},
	{"face-no-tokens", "YB|Z9/R9/B9/Y9/O9/G9", "no valid color tokens found"},
	{"face-not-square", "YB|W8/R9/B9/Y9/O9/G9", "not a perfect square"},
	{"face-wrong-size", "YB|W9/R4/B9/Y9/O9/G9", "expected 9"},
}

const validWildcardCFEN = "YB|?9/R9/B9/Y9/O9/G9"

const solvedCanonicalCFEN = "YB|Y9/R9/B9/W9/O9/G9"

func TestErrorPaths_ParseCFEN(t *testing.T) {
	for _, tc := range malformedCFENs {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseCFEN(tc.cfenStr)
			if err == nil {
				t.Fatalf("ParseCFEN(%q): expected error, got nil", tc.cfenStr)
			}
			if !strings.Contains(err.Error(), tc.errContains) {
				t.Errorf("ParseCFEN(%q) error %q: missing key phrase %q", tc.cfenStr, err.Error(), tc.errContains)
			}
		})
	}

	state, err := ParseCFEN(validWildcardCFEN)
	if err != nil {
		t.Fatalf("ParseCFEN(%q): unexpected error: %v", validWildcardCFEN, err)
	}
	if state == nil {
		t.Fatalf("ParseCFEN(%q): expected non-nil state", validWildcardCFEN)
	}
}

func TestErrorPaths_ValidateCFENAgreesWithParseCFEN(t *testing.T) {
	for _, tc := range malformedCFENs {
		t.Run(tc.name, func(t *testing.T) {
			_, parseErr := ParseCFEN(tc.cfenStr)
			validateErr := ValidateCFEN(tc.cfenStr)

			if (parseErr == nil) != (validateErr == nil) {
				t.Fatalf("ValidateCFEN(%q) err=%v, ParseCFEN err=%v: outcome mismatch", tc.cfenStr, validateErr, parseErr)
			}
			if validateErr == nil {
				t.Fatalf("ValidateCFEN(%q): expected error, got nil", tc.cfenStr)
			}
			if !strings.Contains(validateErr.Error(), tc.errContains) {
				t.Errorf("ValidateCFEN(%q) error %q: missing key phrase %q", tc.cfenStr, validateErr.Error(), tc.errContains)
			}
		})
	}

	if err := ValidateCFEN(validWildcardCFEN); err != nil {
		t.Errorf("ValidateCFEN(%q): unexpected error: %v", validWildcardCFEN, err)
	}
}

func TestErrorPaths_ValidateCFENAcceptsSolved(t *testing.T) {
	if err := ValidateCFEN(solvedCanonicalCFEN); err != nil {
		t.Errorf("ValidateCFEN(%q): unexpected error: %v", solvedCanonicalCFEN, err)
	}
}

func TestErrorPaths_ParseColorWildcard(t *testing.T) {
	got, err := parseColor('?')
	if err != nil {
		t.Fatalf("parseColor('?'): unexpected error: %v", err)
	}
	if got != cube.Grey {
		t.Errorf("parseColor('?') = %v (%d), want cube.Grey (%d)", got, got, cube.Grey)
	}
}

func TestErrorPaths_ParseColorAllColors(t *testing.T) {
	mapping := map[rune]cube.Color{
		'W': cube.White,
		'Y': cube.Yellow,
		'R': cube.Red,
		'O': cube.Orange,
		'G': cube.Green,
		'B': cube.Blue,
	}
	for ch, want := range mapping {
		got, err := parseColor(ch)
		if err != nil {
			t.Errorf("parseColor(%q): unexpected error: %v", ch, err)
			continue
		}
		if got != want {
			t.Errorf("parseColor(%q) = %v (%d), want %v (%d)", ch, got, got, want, want)
		}
	}
}

func TestErrorPaths_ParseColorInvalid(t *testing.T) {
	for _, ch := range []rune{'Z', 'x', '1'} {
		got, err := parseColor(ch)
		if err == nil {
			t.Errorf("parseColor(%q): expected error, got nil", ch)
			continue
		}
		if !strings.Contains(err.Error(), "unknown color character") {
			t.Errorf("parseColor(%q) error %q: missing key phrase \"unknown color character\"", ch, err.Error())
		}
		if got != cube.White {
			t.Errorf("parseColor(%q) returned color %v, want zero value cube.White", ch, got)
		}
	}
}
