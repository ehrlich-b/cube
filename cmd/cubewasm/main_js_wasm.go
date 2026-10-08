//go:build js && wasm

// Command cubewasm exposes the same cube engine used by the CLI as a JSON API.
package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"syscall/js"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
	"github.com/ehrlich-b/cube/internal/webapi"
)

type request struct {
	Op       string `json:"op"`
	CFEN     string `json:"cfen"`
	Moves    string `json:"moves"`
	Method   string `json:"method"`
	Target   string `json:"target"`
	MaxDepth int    `json:"maxDepth"`
	Size     int    `json:"size"`
	Profile  bool   `json:"profile"`
}

type snapshot struct {
	Size   int                 `json:"size"`
	CFEN   string              `json:"cfen"`
	Solved bool                `json:"solved"`
	Faces  map[string][]string `json:"faces"`
}

func parseMoves(text string) ([]cube.Move, error) {
	if len(text) > 65536 {
		return nil, fmt.Errorf("move input exceeds 65536 characters")
	}
	for _, token := range strings.Fields(text) {
		if !webapi.ValidMoveToken(token) {
			return nil, fmt.Errorf("invalid move %q; separate WCA moves with spaces", token)
		}
	}
	return cube.ParseMoves(text)
}

// Bound numeric runs before using the general NxN parser, just like cube learn.
func parseCFEN(text string, pattern bool) (*cfen.CFENState, error) {
	if len(text) > 600 || !strings.HasPrefix(text, "YB|") {
		return nil, fmt.Errorf("use 2x2 through 7x7 CFEN in YB storage order: U/R/F/D/L/B")
	}
	faces := strings.Split(strings.TrimPrefix(text, "YB|"), "/")
	if len(faces) != 6 {
		return nil, fmt.Errorf("CFEN needs six faces in U/R/F/D/L/B order")
	}
	for _, face := range faces {
		count, valid := webapi.FaceStickerCount(face)
		if !valid || (!pattern && strings.Contains(face, "?")) {
			return nil, fmt.Errorf("use colors W Y R O G B and runs 1–49; ? is only allowed in search targets")
		}
		if count > 49 {
			return nil, fmt.Errorf("CFEN runs must total at most 49 stickers per face")
		}
	}
	state, err := cfen.ParseCFEN(text)
	if err != nil {
		return nil, err
	}
	if state.Dimension < 2 || state.Dimension > 7 {
		return nil, fmt.Errorf("the website supports 2x2 through 7x7 cubes")
	}
	return state, nil
}

func view(c *cube.Cube) snapshot {
	text, _ := cfen.GenerateCFEN(c)
	faces := make(map[string][]string, 6)
	for f := cube.Front; f <= cube.Down; f++ {
		stickers := make([]string, 0, c.Size*c.Size)
		for _, row := range c.Faces[f] {
			for _, color := range row {
				stickers = append(stickers, color.String())
			}
		}
		faces[f.String()] = stickers
	}
	return snapshot{c.Size, text, c.IsSolved(), faces}
}

func tokens(moves []cube.Move) []string {
	result := make([]string, len(moves))
	for i, move := range moves {
		result[i] = move.String()
	}
	return result
}

func dispatch(req request) (any, error) {
	size := req.Size
	if size == 0 {
		size = 3
	}
	if size < 2 || size > 7 {
		return nil, fmt.Errorf("cube size must be between 2 and 7")
	}
	c := cube.NewCube(size)
	if req.CFEN != "" {
		state, err := parseCFEN(req.CFEN, false)
		if err != nil {
			return nil, err
		}
		c, err = state.ToCube()
		if err != nil {
			return nil, err
		}
		if req.Size != 0 && req.Size != c.Size {
			return nil, fmt.Errorf("CFEN size does not match requested size")
		}
		if c.Size == 3 {
			if err := cube.Validate3x3(c); err != nil {
				return nil, err
			}
		} else {
			counts := make(map[cube.Color]int)
			for _, face := range c.Faces {
				for _, row := range face {
					for _, color := range row {
						counts[color]++
					}
				}
			}
			for color := cube.White; color <= cube.Green; color++ {
				if counts[color] != c.Size*c.Size {
					return nil, fmt.Errorf("color %s has %d stickers; expected %d", color, counts[color], c.Size*c.Size)
				}
			}
		}
	}
	moves, err := parseMoves(req.Moves)
	if err != nil {
		return nil, err
	}
	if req.Op == "sequence" {
		// Worker playback preparation crosses the WASM/JSON boundary once per
		// batch, rather than once per move on the UI thread.
		if len(moves) > 32 {
			return nil, fmt.Errorf("sequence batches allow at most 32 moves")
		}
		frames := make([]snapshot, 0, len(moves))
		for _, move := range moves {
			if err := c.ApplyMoves([]cube.Move{move}); err != nil {
				return nil, err
			}
			frames = append(frames, view(c))
		}
		return map[string]any{"frames": frames}, nil
	}
	if err := c.ApplyMoves(moves); err != nil {
		return nil, err
	}
	switch req.Op {
	case "state", "twist":
		return map[string]any{"state": view(c), "moves": tokens(moves)}, nil
	case "solve":
		method := req.Method
		if c.Size != 3 {
			if method != "" && method != "auto" && method != "reduction" {
				return nil, fmt.Errorf("%s is 3x3-only; use reduction for this size", method)
			}
			method = "reduction"
		} else if method == "" || method == "auto" {
			method = "kociemba"
		}
		var solver cube.Solver
		if method == "reduction" && c.Size != 3 {
			solver = &cube.ReductionSolver{}
		} else {
			solver, err = cube.GetSolver(method)
			if err != nil {
				return nil, err
			}
		}
		if !c.IsSolved() {
			if (c.Size != 3 || method == "kociemba") && !cube.BrowserSolverAssetLoaded("coordinates") {
				return nil, fmt.Errorf("load the coordinate solver asset before solving")
			}
			if c.Size > 3 && !cube.BrowserSolverAssetLoaded(fmt.Sprintf("nxn-%d", c.Size)) {
				return nil, fmt.Errorf("load the %dx%d solver asset before solving", c.Size, c.Size)
			}
		}
		result, err := solver.Solve(c)
		if err != nil {
			return nil, err
		}
		if err := c.ApplyMoves(result.Solution); err != nil {
			return nil, err
		}
		if !c.IsSolved() {
			return nil, fmt.Errorf("%s returned a solution that does not solve all six faces", method)
		}
		stages := make([]any, 0, len(result.Stages))
		for _, stage := range result.Stages {
			stages = append(stages, map[string]any{"name": stage.Name, "cases": stage.Cases, "moves": tokens(stage.Moves), "turns": cube.TurnCount(stage.Moves), "after": view(stage.After)})
		}
		return map[string]any{"state": view(c), "moves": tokens(result.Solution), "method": method, "stages": stages, "solveMs": float64(result.Duration.Microseconds()) / 1000}, nil
	case "learn":
		if c.Size != 3 {
			return nil, fmt.Errorf("lessons are 3x3-only")
		}
		lesson, err := cube.PlanBeginner(c)
		if err != nil {
			return nil, err
		}
		steps := make([]any, 0, len(lesson.Steps))
		for _, step := range lesson.Steps {
			actions := make([]any, 0, len(step.Actions))
			for _, action := range step.Actions {
				actions = append(actions, map[string]any{"instruction": action.Instruction, "moves": tokens(action.Moves)})
			}
			steps = append(steps, map[string]any{"title": step.Title, "check": step.Check, "actions": actions, "moves": tokens(step.Moves()), "after": view(step.After)})
		}
		return map[string]any{"steps": steps, "moves": tokens(lesson.Moves()), "state": view(lesson.Final)}, nil
	case "find":
		if c.Size != 3 {
			return nil, fmt.Errorf("search is 3x3-only")
		}
		if req.MaxDepth < 0 || req.MaxDepth > 10 {
			return nil, fmt.Errorf("search depth must be between 0 and 10")
		}
		target, err := parseCFEN(req.Target, true)
		if err != nil {
			return nil, err
		}
		if target.Dimension != 3 {
			return nil, fmt.Errorf("search targets must be 3x3")
		}
		goal, err := target.ToCube()
		if err != nil {
			return nil, err
		}
		if req.MaxDepth > 0 && !cube.BrowserSolverAssetLoaded("coordinates") {
			return nil, fmt.Errorf("load the coordinate solver asset before searching")
		}
		moves, found := cube.FindPattern(c, goal, nil, req.MaxDepth)
		c.ApplyMoves(moves)
		return map[string]any{"found": found, "moves": tokens(moves), "state": view(c)}, nil
	default:
		return nil, fmt.Errorf("unknown operation %q", req.Op)
	}
}

func invoke(_ js.Value, args []js.Value) (result any) {
	encode := func(value any) string {
		data, _ := json.Marshal(value)
		return string(data)
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			result = encode(map[string]any{"ok": false, "error": fmt.Sprint(recovered)})
		}
	}()
	if len(args) != 1 || args[0].Type() != js.TypeString || len(args[0].String()) > 131072 {
		return encode(map[string]any{"ok": false, "error": "pass one JSON request of at most 131072 characters"})
	}
	var req request
	if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
		return encode(map[string]any{"ok": false, "error": err.Error()})
	}
	var before, after runtime.MemStats
	if req.Profile {
		runtime.ReadMemStats(&before)
	}
	data, err := dispatch(req)
	if err != nil {
		return encode(map[string]any{"ok": false, "error": err.Error()})
	}
	if req.Profile {
		runtime.ReadMemStats(&after)
		data.(map[string]any)["runtime"] = map[string]any{
			"heapBytes": after.HeapAlloc, "heapSystemBytes": after.HeapSys,
			"allocatedBytes": after.TotalAlloc - before.TotalAlloc, "collections": after.NumGC - before.NumGC,
		}
	}
	return encode(map[string]any{"ok": true, "data": data})
}

func main() {
	callback := js.FuncOf(invoke)
	js.Global().Set("cubeAPI", callback)
	install := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) != 3 || args[0].Type() != js.TypeString || args[1].Type() != js.TypeObject || args[2].Type() != js.TypeString || !args[1].InstanceOf(js.Global().Get("Uint8Array")) {
			return "pass a solver asset name, Uint8Array and SHA-256 digest"
		}
		length := args[1].Get("byteLength").Int()
		if length == 0 || length > 10<<20 {
			return "invalid solver asset length"
		}
		data := make([]byte, length)
		js.CopyBytesToGo(data, args[1])
		if err := cube.InstallBrowserSolverAsset(args[0].String(), data, args[2].String()); err != nil {
			return err.Error()
		}
		return ""
	})
	js.Global().Set("cubeLoadSolverAsset", install)
	select {}
}
