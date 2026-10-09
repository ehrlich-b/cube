package cli

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
	"github.com/spf13/cobra"
)

var nxnMoveToken = regexp.MustCompile(`^(?:[1-7]?[URFDLB]w?|[MESxyz])(?:'|2)?$`)
var nxnStateRun = regexp.MustCompile(`([WYROGB])([0-9]*)`)

func fullSolveInput(cmd *cobra.Command, args []string) (*cube.Cube, error) {
	n, _ := cmd.Flags().GetInt("dimension")
	if n < 2 {
		return nil, fmt.Errorf("cube dimension must be at least 2 (got %d)", n)
	}
	if n == 3 {
		return firstLayerInput(cmd, args)
	}
	if n > 7 {
		return nil, fmt.Errorf("full solving supports dimensions 2-7 (got %d)", n)
	}
	c := cube.NewCube(n)
	start, _ := cmd.Flags().GetString("start")
	if start != "" {
		if len(start) > 1024 || !strings.HasPrefix(start, "YB|") {
			return nil, fmt.Errorf("use concrete YB CFEN in U/R/F/D/L/B storage order")
		}
		faces := strings.Split(strings.TrimPrefix(start, "YB|"), "/")
		if len(faces) != 6 {
			return nil, fmt.Errorf("CFEN needs six faces")
		}
		// Bound every run before the general parser allocates its expansion.
		for _, face := range faces {
			runs := nxnStateRun.FindAllStringSubmatch(face, -1)
			var joined strings.Builder
			count := 0
			for _, run := range runs {
				joined.WriteString(run[0])
				length := 1
				if run[2] != "" {
					var err error
					length, err = strconv.Atoi(run[2])
					if err != nil || length < 1 || length > n*n {
						return nil, fmt.Errorf("CFEN run exceeds dimension %d", n)
					}
				}
				count += length
			}
			if joined.String() != face || count != n*n {
				return nil, fmt.Errorf("concrete CFEN needs %d real stickers per face for --dimension %d", n*n, n)
			}
		}
		state, err := cfen.ParseCFEN(start)
		if err != nil {
			return nil, err
		}
		c, err = state.ToCube()
		if err != nil {
			return nil, err
		}
	}
	if len(args) > 0 {
		moves, err := parseSolveMoves(args[0], n)
		if err != nil {
			return nil, err
		}
		if err := c.ApplyMoves(moves); err != nil {
			return nil, err
		}
	}
	return c, nil
}
