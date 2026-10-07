package cube

import (
	"fmt"
	"strings"
	"testing"
)

func TestInvalidMovesRejectWithoutMutation(t *testing.T) {
	for _, test := range []struct {
		size    int
		text    string
		message string
	}{
		{3, "4R", "layer"}, {4, "5L", "layer"}, {5, "6U", "layer"},
		{3, "4Rw", "depth"}, {4, "5Fw", "depth"},
		{2, "M", "slice"}, {4, "M", "slice"}, {4, "E", "slice"}, {4, "S", "slice"},
	} {
		t.Run(fmt.Sprintf("%s/%dx%d", test.text, test.size, test.size), func(t *testing.T) {
			move, err := ParseMove(test.text)
			if err != nil {
				t.Fatal(err)
			}
			c := NewCube(test.size)
			before := c.clone()
			if err := c.ApplyMove(move); err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("wanted %s error, got %v", test.message, err)
			}
			if !facesEqual(c, before) {
				t.Fatal("invalid move changed stickers")
			}
			valid, _ := ParseMove("R")
			if err := c.ApplyMoves([]Move{valid, move}); err == nil {
				t.Fatal("invalid sequence accepted")
			}
			if !facesEqual(c, before) {
				t.Fatal("invalid sequence applied its valid prefix")
			}
		})
	}
}

func TestMoveValidationAllowsFullWidthAndOddSlices(t *testing.T) {
	for _, size := range []int{3, 4, 5} {
		for _, face := range []string{"R", "L", "U", "D", "F", "B"} {
			move := Move{Face: Right, Layer: size - 1, Clockwise: true}
			parsed, _ := ParseMove(face)
			move.Face = parsed.Face
			c := NewCube(size)
			if err := c.ApplyMove(move); err != nil || c.IsSolved() {
				t.Fatalf("valid far layer rejected on size %d: %v", size, err)
			}
			move.Layer, move.Wide, move.WideDepth = 0, true, size
			if err := NewCube(size).ApplyMove(move); err != nil {
				t.Fatal(err)
			}
		}
		if size%2 == 1 {
			for _, text := range []string{"M", "E", "S"} {
				move, _ := ParseMove(text)
				c := NewCube(size)
				if err := c.ApplyMove(move); err != nil || c.IsSolved() {
					t.Fatalf("valid odd slice %s rejected on size %d: %v", text, size, err)
				}
			}
		}
	}
}

func TestParseRejectsInvalidLayerModifiers(t *testing.T) {
	for _, text := range []string{"0R", "0Rw", "1M", "2E", "Sw", "1x", "2y", "zw"} {
		if _, err := ParseMove(text); err == nil {
			t.Errorf("invalid modifier %q accepted", text)
		}
	}
}
