# Complete beginner milestone verification — 2026-10-03

Environment: Bryan's Mac, Darwin arm64, Apple M4. Runs were local, with
`GOMAXPROCS=2` and build/module caches under `/tmp`. No paid service, WSL,
push, deploy, credential change, private source upload or canonical board edit
was used.

This extension starts at the preserved first-layer commit `c82b2ca`, on branch
`codex/cube-full-lesson-20261003` in
`/Users/ehrlich/repos/cube-first-layer-20261003`. The original checkout stays
on its clean `main` at `f34d3e5`. The
[earlier receipt](./verification-2026-10-03.md) describes the partial milestone
and its then-current limitations; this receipt records the full extension.

## Delivered behavior

- Default `cube solve` returns a complete 3×3 beginner solution, with normal,
  headless moves and final CFEN output. It verifies the result before output.
- Default `cube learn` covers white cross/corners, middle edges, yellow cross,
  yellow edge alignment, corner placement and corner orientation.
- `--goal first-layer` preserves the partial lesson and partial solve.
- Existing actual-move, undo/reset, save/resume and repeated-step mechanics
  operate across the full solve, including changed grips and interrupted steps.
- The final corner sweep is one checkpoint: its lower layers temporarily mix
  and return only after the complete sweep and final top turn.
- Solved rotated cubes need no moves and keep their grip. Unavailable
  CFOP/Kociemba algorithms and unsupported dimensions produce CLI errors.

The unchanged solver-contract assertions actively check all five beginner
cases. Only the ten CFOP/Kociemba empty API-stub cases still skip. The E2E
harness's former beginner placeholder skips have become real solve-and-verify
checks; its final run has no skips counted as passes.

## Checks and outcomes

| Check | Outcome |
| --- | --- |
| Focused full-planner/session/solver-contract Go tests | Green, input unchanged and all six faces checked |
| Randomized full replay | 500 deterministic 25-turn scrambles, all checkpoints valid |
| Recovery | Checkpoint states and true intermediate D-turn prefixes recover after extra moves/rotations |
| Aggregate `make fmt build-all-local test-all test-first-layer test-beginner vet` | Exit 0 |
| Binary E2E harness | 122/122, zero failures |
| Independent full-state oracle | 211 physical states, 2,157 printed checkpoints |
| Independent invalid full-input cases | 11 rejected |
| Emitted resume commands executed | Six, covering quit, unterminated quit and EOF for both goals |
| Solved grips | Four preserved exactly |
| Full live session | 16 confirmed checkpoints plus 29 resumed checkpoint replays |
| Live recovery and inverses | Wrong turn, x/y grip change, interrupted final sweep, undo/retry and reset match physical stickers |
| Prior partial-goal oracle | 120 legal cubie states across 24 orientations, four impossible states, nine invalid cases, seven scripts, three live sessions |
| Fresh independent `go test ./... -count=1` and `go vet ./...` | Green |
| Whitespace check | `git diff --check` green |

The independent reviewer uses a separate Python 3D sticker model and constructs
physical cubies with balanced orientation and matching parity. It does not use
Go-generated scrambles for the 120 uniform legal cubie states. Additional
fixtures cover every yellow cross pattern (8), edge permutation (24), even
corner permutation (12), balanced corner twist pattern (27), and middle-edge
setup cases (8 from top, 12 misplaced). Yellow-stage fixtures explicitly retain
the first two layers; each printed yellow checkpoint independently checks those
layers and the applicable cross, edge or corner-slot goal.

The reviewer identified missing goal preservation in resume output and a test
cursor that did not represent a true cumulative intermediate prefix. Both were
fixed and verified. Final independent review found no remaining defect.

## Performance and UX evidence

The 500-scramble Go corpus averaged **209.2 moves**, maximum **291**. The separate
120-state uniform cubie corpus averaged **209.6 moves**, maximum **281**. These
counts include whole-cube rotations; the method prioritizes familiar algorithms
and restored checkpoints over short solutions.

`BenchmarkBeginnerFull`, one fixed rotated scramble, three one-second runs:
**1.783–1.894 ms/op**, approximately **5.27 MB/op**, **21,077 allocations/op**.
This is one fixture, not a worst-case bound or a memory optimization claim.
The final aggregate oracle's 120 separate CLI processes measured median
**8.12 ms**, maximum **51.32 ms**, including process startup. The independent
reviewer's preceding run measured median **5.34 ms**, maximum **18.62 ms**.
These are observations on this Mac, not general latency guarantees.

UX evidence is independent simulated physical replay of the instructions,
prompts and recovery commands. **No human beginner trial with a physical cube
occurred.** Recovery requires known actual moves or a complete observed CFEN.
The input supports standard 3×3 colors in concrete YB storage order; actual
centers determine the grip. Human readability, CFOP/Kociemba, larger-cube solving
and shorter solutions remain future work.

## Reproduce

```sh
GOCACHE=/tmp/cube-go-cache GOMODCACHE=/tmp/cube-go-mod GOMAXPROCS=2 \
  make fmt build-all-local test-all test-first-layer test-beginner vet
GOCACHE=/tmp/cube-go-cache GOMODCACHE=/tmp/cube-go-mod GOMAXPROCS=2 \
  go test ./internal/cube -run 'TestBeginner|TestInvariant_Solver' -v -count=1
GOCACHE=/tmp/cube-go-cache GOMODCACHE=/tmp/cube-go-mod GOMAXPROCS=2 \
  go test ./internal/cube -run '^$' -bench '^BenchmarkBeginnerFull$' -benchmem -benchtime=1s -count=3
./dist/cube learn "R U F2 L' B" --interactive --color
./dist/cube solve "x R U F2 L' B" --headless
```

Final aggregate, focused and benchmark output is retained locally in
`/tmp/cube-full-final-checks.log`, `/tmp/cube-full-focused-checks.log` and
`/tmp/cube-full-benchmark.log`. These temporary files are not needed to reproduce.
