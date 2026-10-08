package cfen

import (
	"math/rand"
	"reflect"
	"regexp"
	"testing"
)

func TestFaceScannerMatchesGrammar(t *testing.T) {
	re := regexp.MustCompile(`([WYROGB?])(\d*)`)
	rng := rand.New(rand.NewSource(4087))
	alphabet := []byte("WYROGB?0123456789x/ _\xff\xc3\xa9")
	for i := 0; i < 20000; i++ {
		text := make([]byte, rng.Intn(100))
		for j := range text {
			text[j] = alphabet[rng.Intn(len(alphabet))]
		}
		if got, want := scanFaceRuns(string(text)), re.FindAllStringSubmatch(string(text), -1); !reflect.DeepEqual(got, want) {
			t.Fatalf("scanner differs on %q: %v != %v", text, got, want)
		}
	}
}
