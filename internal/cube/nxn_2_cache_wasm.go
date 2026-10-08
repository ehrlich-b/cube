//go:build wasm

package cube

// Wasm already obtains its coordinates through solverTables. Exclude the
// native cache's imports too: a runtime guard still links gob initialization.
func nxnTwoByTwoTables() {}
