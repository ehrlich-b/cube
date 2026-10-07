package cube

import (
	"fmt"
	"strings"
	"testing"
)

func TestFaceString(t *testing.T) {
	cases := []struct {
		face Face
		want string
	}{
		{Front, "F"},
		{Back, "B"},
		{Left, "L"},
		{Right, "R"},
		{Up, "U"},
		{Down, "D"},
	}
	for _, tc := range cases {
		if got := tc.face.String(); got != tc.want {
			t.Errorf("Face.String() = %q, want %q", got, tc.want)
		}
	}
}

func TestColorString(t *testing.T) {
	cases := []struct {
		color Color
		want  string
	}{
		{White, "W"},
		{Yellow, "Y"},
		{Red, "R"},
		{Orange, "O"},
		{Blue, "B"},
		{Green, "G"},
		{Grey, "."},
	}
	for _, tc := range cases {
		if got := tc.color.String(); got != tc.want {
			t.Errorf("Color.String() = %q, want %q", got, tc.want)
		}
	}
}

func TestColorUnicodeString(t *testing.T) {
	// None of the 7 valid colors may panic or return an empty string.
	for c := White; c <= Grey; c++ {
		got := c.UnicodeString()
		if got == "" {
			t.Errorf("Color(%d).UnicodeString() returned empty string", c)
		}
	}
	// Spot-check the two most important values.
	if got, want := White.UnicodeString(), "⬜"; got != want {
		t.Errorf("White.UnicodeString() = %q, want %q", got, want)
	}
	if got, want := Grey.UnicodeString(), "⬛"; got != want {
		t.Errorf("Grey.UnicodeString() = %q, want %q", got, want)
	}
}

func TestColorColoredString(t *testing.T) {
	// The ANSI-wrapped form must still embed the color's own letter.
	cases := []struct {
		color  Color
		letter string
	}{
		{White, "W"},
		{Grey, "."},
	}
	for _, tc := range cases {
		got := tc.color.ColoredString()
		if !strings.Contains(got, tc.letter) {
			t.Errorf("Color(%d).ColoredString() = %q, want it to contain %q", tc.color, got, tc.letter)
		}
	}
}

// TestGreyMonoStringPanics_KnownDefect pins a real, reachable defect:
// MonoString's backing table only has 6 entries (White..Green) but Color has
// 7 values including Grey (index 6, the pattern-matching wildcard). Calling
// Grey.MonoString() therefore panics with "index out of range".
func TestGreyMonoStringPanics_KnownDefect(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Grey.MonoString() did NOT panic; expected index out of range")
		}
		msg := fmt.Sprintf("%v", r)
		if !strings.Contains(msg, "index out of range") {
			t.Fatalf("Grey.MonoString() panicked with %q, want it to contain \"index out of range\"", msg)
		}
	}()
	_ = Grey.MonoString()
}

func TestMonoStringNoPanicForValidColors(t *testing.T) {
	// White through Green (indices 0..5) are backed by the 6-entry table.
	for c := White; c <= Green; c++ {
		if got := c.MonoString(); got == "" {
			t.Errorf("Color(%d).MonoString() returned empty string", c)
		}
	}
}

func TestFormatStickerDispatch(t *testing.T) {
	c := NewCube(3)
	if got, want := c.FormatSticker(White, false, false), White.String(); got != want {
		t.Errorf("FormatSticker(White, false, false) = %q, want %q", got, want)
	}
	if got, want := c.FormatSticker(White, true, false), White.ColoredString(); got != want {
		t.Errorf("FormatSticker(White, true, false) = %q, want %q", got, want)
	}
	if got, want := c.FormatSticker(White, false, true), White.UnicodeString(); got != want {
		t.Errorf("FormatSticker(White, false, true) = %q, want %q", got, want)
	}
}

func TestCubeStringStructure(t *testing.T) {
	c := NewCube(3)
	lines := strings.Split(c.String(), "\n")

	// Verified runtime layout for a solved 3x3: 3 rows of Up, a blank line,
	// 3 rows of Left+Front+Right+Back, a blank line, 3 rows of Down, then the
	// trailing empty element from the final "\n". Total 12 elements.
	if want := 12; len(lines) != want {
		t.Fatalf("len(strings.Split(String(), \"\\n\")) = %d, want %d", len(lines), want)
	}

	nonEmpty := 0
	for _, l := range lines {
		if l != "" {
			nonEmpty++
		}
	}
	// 3 rows x 3 sections = 9 non-empty content lines.
	if want := 9; nonEmpty != want {
		t.Errorf("non-empty lines = %d, want %d", nonEmpty, want)
	}

	// 3 empty strings: two real blank separators plus the trailing one.
	if want := 3; len(lines)-nonEmpty != want {
		t.Errorf("empty lines = %d, want %d", len(lines)-nonEmpty, want)
	}
}

func TestStringEqualsStringWithColorFalse(t *testing.T) {
	c := NewCube(3)
	plain := c.String()
	withColor := c.StringWithColor(false)
	if plain != withColor {
		t.Error("Cube.String() and Cube.StringWithColor(false) must be byte-identical")
	}
}

func TestUnfoldedStringColorDiffers(t *testing.T) {
	c := NewCube(3)
	off := c.UnfoldedString(false, false)
	on := c.UnfoldedString(true, false)
	if off == on {
		t.Error("UnfoldedString(true, false) must differ from UnfoldedString(false, false) (ANSI codes)")
	}
	if !strings.Contains(on, "\033[") {
		t.Error("UnfoldedString(true, false) should contain ANSI escape sequences")
	}
}

func TestCubeStringSizeIndependence(t *testing.T) {
	for _, size := range []int{2, 4} {
		c := NewCube(size)
		lines := strings.Split(c.String(), "\n")

		// Same pattern as size 3: 3*size non-empty content lines + 3 empty strings.
		if wantNonEmpty := 3 * size; len(lines) != wantNonEmpty+3 {
			t.Errorf("size %d: len(lines) = %d, want %d", size, len(lines), wantNonEmpty+3)
		}

		nonEmpty := 0
		for _, l := range lines {
			if l != "" {
				nonEmpty++
			}
		}
		if wantNonEmpty := 3 * size; nonEmpty != wantNonEmpty {
			t.Errorf("size %d: non-empty lines = %d, want %d", size, nonEmpty, wantNonEmpty)
		}
		if wantEmpty := 3; len(lines)-nonEmpty != wantEmpty {
			t.Errorf("size %d: empty lines = %d, want %d", size, len(lines)-nonEmpty, wantEmpty)
		}
	}
}
