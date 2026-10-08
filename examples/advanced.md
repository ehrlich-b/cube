# Moves, verification and search

Run these examples from the repository root after `make build`. Solving is
3×3 only; use `twist` to inspect moves on other dimensions.

## Explore notation

```sh
./dist/cube twist "M E' S2 x y' z2" --dimension 3 --color
./dist/cube twist "Rw Uw2 Fw'" --dimension 4 --color
./dist/cube twist "2R 3L'" --dimension 5 --color
```

`M/E/S` turn the middle layer of odd cubes. `Rw` turns the outer two layers;
`3Rw` turns three. `2R` turns only the second layer from the right. Layer
depth must fit the cube. `x/y/z` rotate the whole cube like `R/U/F`.

## Verify a transformation

`verify` takes **one** positional argument: the moves. Supply concrete CFEN
with `--start` and a concrete or wildcard target with `--target`. Without
flags, both states default to solved 3×3. A mismatch exits with status 1.

```sh
./dist/cube verify "R U U' R'" \
  --start "YB|Y9/R9/B9/W9/O9/G9" --target "YB|Y9/R9/B9/W9/O9/G9"
```

Expected output:

```text cube-output
✅ PASS: Algorithm correctly transforms start to target state
Algorithm: R U U' R'
Move count: 4
```

For a known scramble, generate its state, obtain moves only, then verify those
moves from that state. This shell block is also run by `make test-docs`:

```sh cube-check
scramble="R U F2 L' B"
start=$(./dist/cube twist "$scramble" --cfen)
solution=$(./dist/cube solve "$scramble" --headless)
./dist/cube verify "$solution" --start "$start" \
  --target "YB|Y9/R9/B9/W9/O9/G9" --headless
```

## Combine adjacent moves

Optimization merges adjacent turns of the same layer; it does not search for
the shortest equivalent algorithm.

```sh
./dist/cube optimize "R R U U' F F F"
```

```text cube-output
Original:  R R U U' F F F (7 moves)
Optimized: R2 F' (2 moves)
Saved 5 move(s)
```

```sh
./dist/cube optimize "x x 3Rw 3Rw"
```

```text cube-output
Original:  x x 3Rw 3Rw (4 moves)
Optimized: x2 3Rw2 (2 moves)
Saved 2 move(s)
```

## Find a shortest sequence

Search returns one shortest sequence within the depth limit. It prints a
banner before the numbered result; the banner is not a move sequence.

```sh
./dist/cube find sequence "R U" --max-moves 2
```

```text cube-output
Searching for solutions to 'R U' (max 2 moves)...

Found 1 solution(s):
1. U' R' (2 moves)
```

Extract only the first numbered result, require it to be nonempty, then verify
it. A depth limit with no result cannot be used as a solution.

```sh cube-check
scramble="R U"
search=$(./dist/cube find sequence "$scramble" --max-moves 2)
solution=$(printf '%s\n' "$search" | sed -n 's/^1\. \(.*\) ([0-9][0-9]* moves.*$/\1/p')
test -n "$solution"
start=$(./dist/cube twist "$scramble" --cfen)
./dist/cube verify "$solution" --start "$start" --headless
./dist/cube optimize "$solution"
```

```text cube-output
Original:  U' R' (2 moves)
Optimized: U' R' (2 moves)
```

Use a named pattern or a wildcard CFEN target for a partial goal:

```sh
./dist/cube find pattern cross --from "R U" --max-moves 2
./dist/cube find --target "YB|Y9/?9/?9/?9/?9/?9" \
  --start "YB|Y9/R9/B9/W9/O9/G9" --max-moves 0
./dist/cube show "R U R' U'" --highlight-oll --color
```

For full solves, see [solving](./solving.md). For stable lookup IDs, see
[the algorithm guide](./algorithms.md).
