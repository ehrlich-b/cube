package cube

import (
	"reflect"
	"testing"
)

// TestCubieFacePosRoundTrip verifies that CubieToFacePos and FacePosToCubie are
// exact inverses for every address across several cube sizes.
func TestCubieFacePosRoundTrip(t *testing.T) {
	for _, size := range []int{2, 3, 4, 5} {
		for addr := 1; addr <= 6*size*size; addr++ {
			face, row, col := CubieToFacePos(CubieAddress(addr), size)
			if got := FacePosToCubie(face, row, col, size); got != CubieAddress(addr) {
				t.Fatalf("size %d addr %d: FacePosToCubie(CubieToFacePos(...)) = %d, want %d", size, addr, got, addr)
			}
		}
	}
}

// TestGetFacePositionsUpDown3 pins the verified address ranges for the Up and
// Down faces of a 3x3 cube.
func TestGetFacePositionsUpDown3(t *testing.T) {
	wantUp := []CubieAddress{37, 38, 39, 40, 41, 42, 43, 44, 45}
	if got := GetFacePositions(Up, 3); !reflect.DeepEqual(got, wantUp) {
		t.Errorf("GetFacePositions(Up, 3) = %v, want %v", got, wantUp)
	}

	wantDown := []CubieAddress{46, 47, 48, 49, 50, 51, 52, 53, 54}
	if got := GetFacePositions(Down, 3); !reflect.DeepEqual(got, wantDown) {
		t.Errorf("GetFacePositions(Down, 3) = %v, want %v", got, wantDown)
	}
}

// TestGet3x3SpecificPositions verifies the alias map has the expected number of
// keys and pins the verified TC and BC entries.
func TestGet3x3SpecificPositions(t *testing.T) {
	positions := Get3x3SpecificPositions()

	if len(positions) != 18 {
		t.Errorf("Get3x3SpecificPositions() has %d keys, want 18", len(positions))
	}

	wantTC := []CubieAddress{1, 3, 7, 9}
	if got := positions["TC"]; !reflect.DeepEqual(got, wantTC) {
		t.Errorf("TC = %v, want %v", got, wantTC)
	}

	wantBC := []CubieAddress{46, 48, 52, 54}
	if got := positions["BC"]; !reflect.DeepEqual(got, wantBC) {
		t.Errorf("BC = %v, want %v", got, wantBC)
	}
}

// assertParseCubieSpec checks ParseCubieSpec against a wanted result or, when
// wantErr is true, asserts it returns a non-nil error and a nil slice.
func assertParseCubieSpec(t *testing.T, spec string, size int, want []CubieAddress, wantErr bool) {
	t.Helper()
	got, err := ParseCubieSpec(spec, size)
	if wantErr {
		if err == nil {
			t.Errorf("ParseCubieSpec(%q, %d) = %v, want an error", spec, size, got)
			return
		}
		if got != nil {
			t.Errorf("ParseCubieSpec(%q, %d) returned non-nil slice %v on error", spec, size, got)
		}
		return
	}
	if err != nil {
		t.Fatalf("ParseCubieSpec(%q, %d) unexpected error: %v", spec, size, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseCubieSpec(%q, %d) = %v, want %v", spec, size, got, want)
	}
}

// TestParseCubieSpecBasic covers numbers, ranges, aliases, mixed specs,
// duplicates and whitespace tolerance.
func TestParseCubieSpecBasic(t *testing.T) {
	assertParseCubieSpec(t, "1,2,3", 3, []CubieAddress{1, 2, 3}, false)
	assertParseCubieSpec(t, "1-9", 3, []CubieAddress{1, 2, 3, 4, 5, 6, 7, 8, 9}, false)
	assertParseCubieSpec(t, "TC,BC", 3, []CubieAddress{1, 3, 7, 9, 46, 48, 52, 54}, false)
	assertParseCubieSpec(t, "1,TC,5-7", 3, []CubieAddress{1, 1, 3, 7, 9, 5, 6, 7}, false)
	assertParseCubieSpec(t, "5,5", 3, []CubieAddress{5, 5}, false)
	assertParseCubieSpec(t, " 1 , 2 ", 3, []CubieAddress{1, 2}, false)
}

// TestParseCubieSpecEmpty verifies empty and whitespace-only specs return a
// zero-length slice with no error.
func TestParseCubieSpecEmpty(t *testing.T) {
	for _, spec := range []string{"", "   "} {
		got, err := ParseCubieSpec(spec, 3)
		if err != nil {
			t.Fatalf("ParseCubieSpec(%q) unexpected error: %v", spec, err)
		}
		if len(got) != 0 {
			t.Errorf("ParseCubieSpec(%q) = %v, want a zero-length slice", spec, got)
		}
	}
}

// TestParseCubieSpecErrors verifies each documented error case returns an error.
func TestParseCubieSpecErrors(t *testing.T) {
	tests := []struct {
		name string
		spec string
		size int
	}{
		{"out of range address", "999", 3},
		{"out of range range endpoint", "1-999", 3},
		{"inverted range", "9-1", 3},
		{"alias at size 4", "TC", 4},
		{"alias at size 2", "TC", 2},
		{"unrecognized token", "XYZ", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertParseCubieSpec(t, tt.spec, tt.size, nil, true)
		})
	}
}
