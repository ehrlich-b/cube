# Algorithm lookup

Run from the repository root after `make build`. Use the database's case IDs
for specific entries. Common names and substring searches can return several
matches; names such as `h-perm` are not a substitute for an actual database ID.

```sh
./dist/cube lookup OLL-27
```

Selected output lines:

```text cube-output contains
OLL-27 - Sune
Moves: R U R' U R U2 R'
Dimension: 3x3
```

OLL-27 is Sune; OLL-26 is Anti-Sune. For PLL, use PLL-T, PLL-H, PLL-Ua,
PLL-Ub, PLL-Y and PLL-Z. These are stored IDs, independent of common-name
aliases.

```sh
./dist/cube lookup OLL-26
./dist/cube lookup PLL-T
./dist/cube lookup PLL-H
./dist/cube lookup PLL-Ua
./dist/cube lookup PLL-Ub
./dist/cube lookup PLL-Y
./dist/cube lookup PLL-Z
```

## Browse categories and moves

```sh
./dist/cube lookup --category OLL
./dist/cube lookup --category PLL
./dist/cube lookup --category F2L
./dist/cube lookup --category Trigger
./dist/cube lookup --pattern "R U R' U'"
./dist/cube lookup --all
```

`--pattern` matches the complete stored move-sequence text; it does not
recognize a cube state.
Category counts and imported-row provenance are in the
[database report](../alg_dumps/import-report.json). Entries for other dimensions
are algorithms to inspect, not full-cube solvers.

## Preview an algorithm

```sh
./dist/cube lookup OLL-27 --preview
```

Selected output lines:

```text cube-output contains
OLL-27 - Sune
Preview (applied to solved cube):
Top face after algorithm:
```

The lookup preview applies the algorithm **to a solved cube** and shows its
top face. That is different from starting with the algorithm's recognition
case. To study a scramble, use `twist` or
[the beginner lesson](./beginner.md).
