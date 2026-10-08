//go:build js && wasm

package cube

// The site's solvers use compact coordinates. The opt-in 140.67 MiB database
// and its filesystem builder belong to native CLI builds only.
const useLargePhase1 = false
