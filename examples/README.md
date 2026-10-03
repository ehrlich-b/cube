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
cover other toolkit features. Some older examples assume full-cube solvers or
a removed web interface: check the current [README](../README.md) for status
and command syntax. The beginner solver works on 3×3. CFOP and Kociemba remain unimplemented.
