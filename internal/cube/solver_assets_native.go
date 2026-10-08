//go:build !js || !wasm

package cube

import _ "embed"

// Native executables remain self-contained. Browser builds fetch these exact
// assets only when a solver first needs them.
//
//go:embed tables/coordinates-v5.bin.gz
var embeddedCoordinates []byte

// Fixed-size setup trees and verified commutators, generated and exhaustively
// compared by TestNxNEmbeddedTables with CUBE_GENERATE_NXN=1.
//
//go:embed tables/nxn-4-v1.bin.gz
var nxnAsset4 []byte

//go:embed tables/nxn-5-v1.bin.gz
var nxnAsset5 []byte

//go:embed tables/nxn-6-v1.bin.gz
var nxnAsset6 []byte

//go:embed tables/nxn-7-v1.bin.gz
var nxnAsset7 []byte
