//go:build !js || !wasm

package cube

import "os"

var useLargePhase1 = os.Getenv("CUBE_LARGE_TABLES") == "1"
