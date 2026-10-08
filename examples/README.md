# Cube user guide

Start with [the complete beginner lesson](./beginner.md). It teaches a full
3×3 solve and supports checkpoint playback and recovery. The
[white-layer sublesson](./first-layer.md) remains available with `--goal first-layer`.

```sh
make build
./dist/cube learn "R U F2 L' B" --interactive --color
./dist/cube solve "R U F2 L' B" --goal first-layer --headless
```

For move exploration, use `cube twist`; for algorithm lookup, use `cube lookup`.
The [advanced examples](./advanced.md) and [algorithm guide](./algorithms.md)
cover verification, bounded search and stored case IDs. The
[solving guide](./solving.md) covers Kociemba (the default), Beginner and CFOP.
The default full solver supports 2×2 through 7×7. Beginner, CFOP, optimal search
and lessons support 3×3 only; the move engine also supports larger sizes.

`make test-docs` checks CLI examples here and in the root README against
`dist/cube`. Shell fences (`sh` / `bash`) mark runnable CLI commands. The check
sends `quit` to interactive lessons after startup. A `sh cube-check` fence
runs a complete shell workflow, including its variable assignments and pipes.
An adjacent `text cube-output` fence checks exact stdout;
`text cube-output contains` lists complete lines that must appear in order.
Build, install and publishing commands are not run by this check.
