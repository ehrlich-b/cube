# Learn the white first layer

This lesson solves the white cross and all four white corners on a 3×3 cube.
The result has a white bottom face and a matching bottom row on each side.
The middle and last layers may still be scrambled.

Build from the repository root:

```sh
make build
./dist/cube learn "R U F2 L' B" --interactive --color
```

For this example, hold a **solved** physical cube white down, yellow up,
blue front and red right, then apply the scramble `R U F2 L' B`.
The CLI now represents that scrambled cube. If you already have a cube in
another state, supply its move history from solved, or use `--start` below.

## Follow a checkpoint

The lesson prints the current cube, the next piece to place, its `Do:` move
groups, and the stickers to check. Find the named edge (two colors) or corner
(three colors) before turning. Follow the groups in order. Type `next` after
performing the complete checkpoint on your physical cube.

- `R/L/U/D/F/B` mean right/left/top/bottom/front/back. Clockwise is viewed
  directly at the face you turn, including the back and bottom faces.
- `'` means counterclockwise; `2` means a half turn.
- `x/y/z` rotate the **whole cube** like `R/U/F`. `x` brings front to top;
  `y` brings right to front. Keep the new orientation until told to return.
- Corner insertion repeats `R U R' U'` a stated number of times. Complete
  every four-move group; a partly executed group can temporarily disturb white.
- Cross checkpoints restore earlier cross pieces. Corner checkpoints restore
  the full cross and every previously solved white corner.

The net shows U above, L F R B across, and D below. Letter colors are
W=white, Y=yellow, R=red, O=orange, B=blue, G=green.

## Recover a mistake

If you performed different moves from the suggestion, use
`moves <actual sequence>` **instead of `next`**. Record everything you did
since the last displayed state; the tool applies those actual moves and plans
a fresh lesson. For example:

```text
learn> moves R U' x
```

This is useful for an interrupted group, extra turn, or rotated grip. If you
already confirmed a checkpoint with `next`, `moves` starts from the newly
displayed state. If you cannot reconstruct what happened, quit and restart
with the observed sticker state using `--start`.

`hint` or `show` repeats the current state and instructions without advancing.
`undo` restores the last checkpoint or recorded batch and prints its inverse:
perform those inverse moves on your physical cube too. `reset` does the same
for everything since the session began. Repeated `undo`, `reset`, or `next`
after completion safely leaves the modeled state unchanged.

Invalid commands or moves leave the state unchanged. `help` lists commands.

## Save and resume

`state` prints CFEN. `quit` and end of input print an exact resume command:

```sh
./dist/cube learn --start 'YB|Y9/R9/B9/W9/O9/G9' --interactive
```

Replace the example with your saved state. Keep the physical orientation of
that state; perform any normalization rotations only when the lesson asks.
Adding a scramble argument applies those moves **after** the saved state.

Lesson CFEN uses `YB|` as the storage frame, followed by six faces in
U/R/F/D/L/B order, each read row by row as viewed directly at that face.
Runs like `W9` mean nine white stickers. Actual center stickers determine the
grip; the prefix does not assert that yellow is currently up. Other CFEN frame
prefixes and wildcard patterns are deliberately rejected by this lesson.

To derive a state from a known move history:

```sh
./dist/cube twist "x R U F2 L' B" --cfen
```

The validator rejects missing/duplicate pieces, bad counts or centers, mirrored
corners, unbalanced flips/twists, and mismatched permutation parity. These errors
usually mean a sticker was recorded incorrectly or a physical piece was twisted.

## Print the lesson or just the moves

```sh
# All checkpoints and their resulting states, without waiting for input
./dist/cube learn "R U F2 L' B"

# Moves only, with an explicit partial goal
./dist/cube solve "R U F2 L' B" --goal first-layer --headless

# Final CFEN only
./dist/cube solve "R U F2 L' B" --goal first-layer --cfen --headless
```

The solver uses bounded search for each cross edge and simple repeated triggers
for corners. It favors teachable checkpoints over a short move count. It
supports 3×3 only, with this project's standard color scheme. Full-cube
`solve` algorithms remain unimplemented; their existing contract still requires
that any nonempty full solution actually solve every layer.

Tests include independent physical-state and printed-move replay oracles.
They establish state correctness; a human beginner trial with a physical cube
has not yet been performed.
