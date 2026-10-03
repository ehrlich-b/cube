# Cube user guide

Start with [the first-layer lesson](./first-layer.md). It teaches the white
cross and corners on a 3×3 cube and supports checkpoint playback and recovery.

```sh
make build
./dist/cube learn "R U F2 L' B" --interactive --color
./dist/cube solve "R U F2 L' B" --goal first-layer --headless
```

For move exploration, use `cube twist`; for algorithm lookup, use `cube lookup`.
The [advanced examples](./advanced.md) and [algorithm guide](./algorithms.md)
cover other toolkit features. Some older examples assume full-cube solvers or
a removed web interface: check the current [README](../README.md) for status
and command syntax. Full-cube beginner, CFOP and Kociemba solvers remain stubs.
