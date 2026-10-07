# Database Verification Tools

This directory contains specialized tools for working with the cube algorithm database. These tools use the cube package as a library and are separate from the main CLI to keep it clean and focused.

Patterns are **recognition states**: the inverse of the algorithm applied to a
solved cube. Verification applies the original algorithm to that pattern and
checks the canonical solved target of the recorded dimension.

## Available Tools

### `import-algorithms`

```sh
CUBE_CACHE_DIR=$PWD/.scratch/cube-cache taskpolicy -b nice -n 15 go run -p 2 ./tools/import-algorithms
```

This imports all nine `alg_dumps/*.csv` files into the embedded
`internal/cube/algorithms_data.json`. It normalizes lowercase wide notation,
parentheses/repetition and prime/half-turn spellings, strips citation artifacts,
merges equal normalized sequences within a dimension, and retains source rows,
aliases and every category. Five built-in entries remain part of the reproducible
import. Unresolved textual references and stage-invalid 3×3 CFOP entries go to
`alg_dumps/quarantine.json` with the complete original row and an explicit reason.
No quarantine entry enters the live database. The report records 159 rows,
144 accepted, 18 merged, 15 quarantined, and 131 verified unique algorithms
(117 for 3×3 and 14 for other sizes), plus counts by category.

All generated patterns pass the same CFEN parse/replay/match semantics as
`cube verify`. OLL/PLL also preserve F2L, PLL preserves top orientation, and F2L
preserves the cross after restoring the center frame. Other dimension case and
parity descriptions are retained without a claim of independent validation.
Inverse IDs and mirror IDs are detected by exact sticker permutations, with
self-pairs included; mirror uses a left/right reflection. A missing counterpart
leaves its ID empty. Every entry still includes its inverse move sequence.

### `verify-algorithm`
Verify a specific algorithm from the database using its predefined CFEN patterns.

```bash
# Build the tool
make build-tools

# List all algorithms with CFEN patterns
./dist/tools/verify-algorithm --list

# Verify a specific algorithm
./dist/tools/verify-algorithm "T-Perm"

# Verify with detailed output
./dist/tools/verify-algorithm "Sune" --verbose
```

### `verify-database` 
Verify all algorithms in the database that have CFEN patterns defined.

```bash
# Verify all algorithms
./dist/tools/verify-database

# Verify with detailed output
./dist/tools/verify-database --verbose

# Verify only specific categories
./dist/tools/verify-database --category OLL
./dist/tools/verify-database --category PLL
```

## Building

```bash
# Build just the tools
make build-tools

# Build main CLI + tools
make build-all-local

# Build everything including cross-platform
make build-all
```

## Purpose

These tools allow database curators and algorithm researchers to:

1. **Validate algorithm correctness** - Ensure algorithms actually solve their intended cases
2. **Batch verify collections** - Test entire categories or the full database
3. **Debug CFEN patterns** - Identify issues with start/target state definitions
4. **Quality assurance** - Maintain database integrity as it grows

The tools are designed to be used by developers and researchers working on expanding the algorithm database, while keeping the main `cube` CLI focused on end-user functionality.
