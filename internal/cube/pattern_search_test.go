package cube

import "testing"

func TestWildcardSearchHugeMaxDepth(t *testing.T) {
	start, target := NewCube(3), NewCube(3)
	for f := range target.Faces {
		for r := range target.Faces[f] {
			for col := range target.Faces[f][r] {
				target.Faces[f][r][col] = Grey
			}
		}
	}
	target.Faces[Up][0][0] = Blue
	want, found := stickerPatternBFS(start, target, nil, 3)
	got, ok := FindPattern(start, target, nil, int(^uint(0)>>1))
	if !found || !ok || len(got) != len(want) {
		t.Fatalf("huge depth: got %v/%v, shortest sticker BFS %v/%v", got, ok, want, found)
	}
	start.ApplyMoves(got)
	if !matchesStickerPattern(start, target) {
		t.Fatal("wildcard answer failed sticker replay")
	}
}
