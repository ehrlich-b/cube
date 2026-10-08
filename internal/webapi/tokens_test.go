package webapi

import (
	"math/rand"
	"regexp"
	"strconv"
	"testing"
)

func TestBrowserTokenGrammar(t *testing.T) {
	moves := regexp.MustCompile(`^(?:[1-7]?[URFDLB]w?|[MESxyz])(?:'|2)?$`)
	faces := regexp.MustCompile(`^(?:[WYROGB?](?:[1-9][0-9]?)?)+$`)
	runs := regexp.MustCompile(`[WYROGB?]([0-9]*)`)
	rng := rand.New(rand.NewSource(9173))
	alphabet := []byte("URFDLBMESxyzw'0123456789WYROGB? _\n\xff")
	for i := 0; i < 100000; i++ {
		b := make([]byte, rng.Intn(12))
		for j := range b {
			b[j] = alphabet[rng.Intn(len(alphabet))]
		}
		text := string(b)
		if ValidMoveToken(text) != moves.MatchString(text) {
			t.Fatalf("move grammar differs on %q", text)
		}
		count, valid := FaceStickerCount(text)
		if valid != faces.MatchString(text) {
			t.Fatalf("face grammar differs on %q", text)
		}
		if valid {
			want := 0
			for _, run := range runs.FindAllStringSubmatch(text, -1) {
				n := 1
				if run[1] != "" {
					n, _ = strconv.Atoi(run[1])
				}
				want += n
			}
			if count != want {
				t.Fatalf("face count differs on %q: %d != %d", text, count, want)
			}
		}
	}
}
