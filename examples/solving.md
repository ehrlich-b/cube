# Solving 2×2 through 7×7

Run from the repository root after `make build`. The default full solver supports
**2×2 through 7×7**; beginner, CFOP, optimal search and lessons support **3×3 only**.
The quoted moves describe a scramble applied to a solved cube;
for an already scrambled physical cube, provide its observed state with
`--start`. A new random scramble describes a different cube.

## Inspect moves, then solve

```sh
./dist/cube twist "R U R' U'" --color
./dist/cube solve "R U R' U'" --headless
```

Face turns are clockwise when looking directly at that face. An apostrophe
reverses the turn; `2` means a half turn. Keep spaces between moves.

The default method is Kociemba, a two-phase solver. It returns verified moves;
it does not promise a shortest solution. Beginner uses layer-by-layer move
groups. CFOP prints Cross, F2L, OLL and PLL checkpoints with database case names.

```sh
./dist/cube solve "R U F2 L' B" --method kociemba --headless
./dist/cube solve "R U F2 L' B" --method beginner --headless
./dist/cube solve "R U F2 L' B" --method cfop
```

Return the final solved state instead of moves:

```sh
./dist/cube solve "R U F2 L' B" --cfen --headless
```

```text cube-output
YB|Y9/R9/B9/W9/O9/G9
```

## Learn or stop after the white layer

For a complete physical-cube walkthrough, follow
[the beginner lesson](./beginner.md). Its interactive mode waits for your
checkpoint confirmations and can record actual moves and recover mistakes.

```sh
./dist/cube learn "R U F2 L' B" --interactive --color
./dist/cube solve "R U F2 L' B" --goal first-layer --headless
```

`--goal first-layer` solves only the white cross and white corners; the other
layers can remain scrambled. See [the white-layer guide](./first-layer.md).

## Solve other sizes

For 4×4 through 7×7, reduction solves centers and pairs every edge orbit, then
finishes through the public 3×3 Kociemba solver. Odd fixed centers determine
the cube's orientation. The 2×2 uses the same solver's corners path.

On 4×4/5×5, small searches build opposite center blocks using coordinates for
four centers of one color. After U/D, outer turns and horizontal slices preserve
those blocks while searches complete the side centers. On 6×6/7×7, bar
commutators place several center rows together. Color-based three-piece cycles
finish incomplete blocks without assigning unnecessary labels to centers.

Edges use [slice–extract–restore pairing](https://speedcubing.com/chris/4speedsolve21.html)
and slice–flip–slice for the last pairs. A two-turn outer setup search chooses
useful pairings; pure wing cycles finish remaining cases while preserving
centers and earlier edge orbits. Even cubes keep naturally paired edges rather
than sending every wing to a predetermined slot. Fifteen-turn OLL correction
and a half-width PLL algorithm make that reduced edge state legal; later
orbits match it. Odd cubes pair their wings to the central edges. Adjacent
turns on one physical axis are canceled and packed into wide blocks.

The full sequence is verified on the original cube. Solutions are not promised
to be optimal. `--time-limit` and `--target-length` control the final 3×3 search;
the reduction happens first.

```sh
./dist/cube solve "R U R' U'" --dimension 2 --headless
./dist/cube solve "Rw U2 Rw'" --dimension 4 --headless
./dist/cube solve "2R 3L' Uw2" --dimension 5 --headless
```

```sh
./dist/cube solve "2R 3U Fw" --dimension 6 --cfen
```

```text cube-output
YB|Y36/R36/B36/W36/O36/G36
```

```sh
./dist/cube solve "M E S x Rw" --dimension 7 --cfen
```

```text cube-output
YB|Y49/R49/B49/W49/O49/G49
```

The two even-cube parity cases also solve completely:

```sh
./dist/cube solve "2R2 B2 U2 2L U2 2R' U2 2R U2 F2 2R F2 2L' B2 2R2" --dimension 4 --headless
./dist/cube solve "2R2 U2 2R2 Uw2 2R2 Uw2" --dimension 4 --headless
```

Use `--dimension N --start "YB|..."` for an observed NxN state. The CFEN must
contain exactly N² concrete stickers per face; wildcards are search patterns.
An optional scramble is applied after that saved state, as for 3×3 solving.

`make test-nxn` independently samples and replays 100 uniform legal states and
100 uniform-move scrambles per size in fresh CLI processes. It also checks all
24 grips, isolated/combined OLL and PLL parity, printed step counts and final
CFEN. Background-QoS measurements on this Mac (2026-10-08), over all 200 cases
per size, were (the uniform column covers just the 100 uniform states):

| Size | Mean / max moves | Uniform mean / max | Fresh process mean / max |
|---|---:|---:|---:|
| 2×2 | 18.79 / 20 | 18.66 / 20 | 177.60 / 587.11 ms |
| 4×4 | 92.72 / 128 | 92.16 / 115 | 183.91 / 452.92 ms |
| 5×5 | 192.65 / 253 | 193.19 / 253 | 185.61 / 734.83 ms |
| 6×6 | 457.31 / 522 | 452.05 / 508 | 215.58 / 790.62 ms |
| 7×7 | 638.53 / 712 | 640.83 / 712 | 182.37 / 401.88 ms |

Outer, numbered-slice, wide and half turns, and grip rotations each count once. Times
include startup, decoding embedded reduction tables and loading an already
populated 3×3 disk cache alongside reduction. These are sample results, not
worst-case guarantees. The oracle
fails if the combined or uniform mean moves for any size exceeds its
documented value by more than 5%, and requires both documentation tables
to agree.

See [advanced examples](./advanced.md) for CFEN verification and bounded search,
and [algorithm lookup](./algorithms.md) for stored case IDs.
