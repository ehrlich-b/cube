//go:build js && wasm

// Command cubewasm exposes the same cube engine used by the CLI as a JSON API.
package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"syscall/js"

	"github.com/ehrlich-b/cube/internal/cfen"
	"github.com/ehrlich-b/cube/internal/cube"
)

type request struct {
	Op       string `json:"op"`
	CFEN     string `json:"cfen"`
	Moves    string `json:"moves"`
	Method   string `json:"method"`
	Target   string `json:"target"`
	MaxDepth int    `json:"maxDepth"`
}

type snapshot struct {
	CFEN   string              `json:"cfen"`
	Solved bool                `json:"solved"`
	Faces  map[string][]string `json:"faces"`
}

var moveToken = regexp.MustCompile(`^(?:[URFDLB]w?|[MESxyz])(?:'|2)?$`)
var faceToken = regexp.MustCompile(`^(?:[WYROGB?][1-9]?)+$`)

func parseMoves(text string) ([]cube.Move, error) {
	if len(text) > 8192 {
		return nil, fmt.Errorf("move input exceeds 8192 characters")
	}
	for _, token := range strings.Fields(text) {
		if !moveToken.MatchString(token) {
			return nil, fmt.Errorf("invalid 3x3 move %q; separate WCA moves with spaces", token)
		}
	}
	return cube.ParseMoves(text)
}

// Bound numeric runs before using the general NxN parser, just like cube learn.
func parseCFEN(text string, pattern bool) (*cfen.CFENState, error) {
	if len(text) > 128 || !strings.HasPrefix(text, "YB|") {
		return nil, fmt.Errorf("use 3x3 CFEN in YB storage order: U/R/F/D/L/B")
	}
	faces := strings.Split(strings.TrimPrefix(text, "YB|"), "/")
	if len(faces) != 6 {
		return nil, fmt.Errorf("CFEN needs six faces in U/R/F/D/L/B order")
	}
	for _, face := range faces {
		if !faceToken.MatchString(face) || (!pattern && strings.Contains(face, "?")) {
			return nil, fmt.Errorf("use colors W Y R O G B and runs 1–9; ? is only allowed in search targets")
		}
	}
	state, err := cfen.ParseCFEN(text)
	if err != nil {
		return nil, err
	}
	if state.Dimension != 3 {
		return nil, fmt.Errorf("the website supports 3x3 cubes")
	}
	return state, nil
}

func view(c *cube.Cube) snapshot {
	text, _ := cfen.GenerateCFEN(c)
	faces := make(map[string][]string, 6)
	for f := cube.Front; f <= cube.Down; f++ {
		stickers := make([]string, 0, 9)
		for _, row := range c.Faces[f] {
			for _, color := range row {
				stickers = append(stickers, color.String())
			}
		}
		faces[f.String()] = stickers
	}
	return snapshot{text, c.IsSolved(), faces}
}

func tokens(moves []cube.Move) []string {
	result := make([]string, len(moves))
	for i, move := range moves {
		result[i] = move.String()
	}
	return result
}

func dispatch(req request) (any, error) {
	c := cube.NewCube(3)
	if req.CFEN != "" {
		state, err := parseCFEN(req.CFEN, false)
		if err != nil {
			return nil, err
		}
		c, err = state.ToCube()
		if err != nil {
			return nil, err
		}
		if err := cube.Validate3x3(c); err != nil {
			return nil, err
		}
	}
	moves, err := parseMoves(req.Moves)
	if err != nil {
		return nil, err
	}
	c.ApplyMoves(moves)
	switch req.Op {
	case "state", "twist":
		return map[string]any{"state": view(c), "moves": tokens(moves)}, nil
	case "solve":
		method := req.Method
		if method == "" || method == "auto" {
			method = "kociemba"
		}
		solver, err := cube.GetSolver(method)
		if err != nil {
			return nil, err
		}
		result, err := solver.Solve(c)
		if err != nil {
			return nil, err
		}
		c.ApplyMoves(result.Solution)
		if !c.IsSolved() {
			return nil, fmt.Errorf("%s returned a solution that does not solve all six faces", method)
		}
		return map[string]any{"state": view(c), "moves": tokens(result.Solution), "method": method}, nil
	case "learn":
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
		if req.MaxDepth < 0 || req.MaxDepth > 10 {
			return nil, fmt.Errorf("search depth must be between 0 and 10")
		}
		target, err := parseCFEN(req.Target, true)
		if err != nil {
			return nil, err
		}
		goal, err := target.ToCube()
		if err != nil {
			return nil, err
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
	if len(args) != 1 || args[0].Type() != js.TypeString || len(args[0].String()) > 16384 {
		return encode(map[string]any{"ok": false, "error": "pass one JSON request of at most 16384 characters"})
	}
	var req request
	if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
		return encode(map[string]any{"ok": false, "error": err.Error()})
	}
	data, err := dispatch(req)
	if err != nil {
		return encode(map[string]any{"ok": false, "error": err.Error()})
	}
	return encode(map[string]any{"ok": true, "data": data})
}

func main() {
	callback := js.FuncOf(invoke)
	js.Global().Set("cubeAPI", callback)
	select {}
}
