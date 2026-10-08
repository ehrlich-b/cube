package cfen

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/ehrlich-b/cube/internal/cube"
)

func TestReview5RejectOneByOneCFEN(t *testing.T) {
	_, err := ParseCFEN("YB|Y/R/B/W/O/G")
	if err == nil || !strings.Contains(err.Error(), "dimension must be at least 2") {
		t.Fatalf("wanted an unsupported dimension error, got %v", err)
	}
}

func TestReview5ToCubeRejectsUnsupportedDimension(t *testing.T) {
	for _, size := range []int{0, 1} {
		state := &CFENState{Dimension: size, Orientation: CFENOrientation{cube.Yellow, cube.Blue}}
		if c, err := state.ToCube(); err == nil || !strings.Contains(err.Error(), "dimension must be at least 2") || c != nil {
			t.Errorf("dimension %d: wanted an error without a resized cube, got %v, %v", size, c, err)
		}
	}
}

func TestReview5WildcardExportRoundTrip(t *testing.T) {
	state, err := ParseCFEN("YB|Y9/?9/?9/?9/?9/?9")
	if err != nil {
		t.Fatal(err)
	}
	c, err := state.ToCube()
	if err != nil {
		t.Fatal(err)
	}
	text, err := GenerateCFEN(c)
	if err != nil {
		t.Fatal(err)
	}
	if text != state.String() || text != "YB|Y9/?9/?9/?9/?9/?9" {
		t.Fatalf("wildcard export = %q", text)
	}
	if restored, err := ParseCFEN(text); err != nil || !reflect.DeepEqual(state, restored) {
		t.Fatalf("wildcard export did not round-trip: %v", err)
	}
}

func TestReview5RandomWildcardPatternsRoundTrip(t *testing.T) {
	rng := rand.New(rand.NewSource(20261007))
	for size := 2; size <= 7; size++ {
		for sample := 0; sample < 50; sample++ {
			var faces [6]string
			for f := range faces {
				var stickers strings.Builder
				for i := 0; i < size*size; i++ {
					stickers.WriteByte("WYROBG?"[rng.Intn(7)])
				}
				faces[f] = stickers.String()
			}
			original, err := ParseCFEN("YB|" + strings.Join(faces[:], "/"))
			if err != nil {
				t.Fatal(err)
			}
			t.Run(fmt.Sprintf("%d/%d", size, sample), func(t *testing.T) {
				restored, err := ParseCFEN(original.String())
				if err != nil || !reflect.DeepEqual(original, restored) {
					t.Fatalf("pattern serialization did not round-trip: %v; CFEN %s", err, original.String())
				}
				c, err := original.ToCube()
				if err != nil {
					t.Fatal(err)
				}
				text, err := GenerateCFEN(c)
				if err != nil {
					t.Fatal(err)
				}
				restored, err = ParseCFEN(text)
				if err != nil || !reflect.DeepEqual(original, restored) {
					t.Fatalf("cube export did not round-trip: %v; CFEN %s", err, text)
				}
			})
		}
	}
}
