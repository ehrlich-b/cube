# Database Verification Tools

This directory contains specialized tools for working with the cube algorithm database. These tools use the cube package as a library and are separate from the main CLI to keep it clean and focused.

## Available Tools

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