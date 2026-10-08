//go:build js && wasm

package cube

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

var embeddedCoordinates, nxnAsset4, nxnAsset5, nxnAsset6, nxnAsset7 []byte

func BrowserSolverAssetLoaded(name string) bool {
	switch name {
	case "coordinates":
		return len(embeddedCoordinates) != 0
	case "nxn-4":
		return len(nxnAsset4) != 0
	case "nxn-5":
		return len(nxnAsset5) != 0
	case "nxn-6":
		return len(nxnAsset6) != 0
	case "nxn-7":
		return len(nxnAsset7) != 0
	}
	return false
}

// InstallBrowserSolverAsset accepts only complete, integrity-checked assets.
// The expected digest comes from the fingerprinted site's generated manifest.
// Solver table decoders also check their dimensions and move fingerprints.
func InstallBrowserSolverAsset(name string, data []byte, digest string) error {
	if len(data) == 0 || len(data) > 10<<20 {
		return fmt.Errorf("invalid solver asset length")
	}
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != digest {
		return fmt.Errorf("solver asset integrity check failed")
	}
	switch name {
	case "coordinates":
		embeddedCoordinates = data
	case "nxn-4":
		nxnAsset4 = data
	case "nxn-5":
		nxnAsset5 = data
	case "nxn-6":
		nxnAsset6 = data
	case "nxn-7":
		nxnAsset7 = data
	default:
		return fmt.Errorf("unknown solver asset %q", name)
	}
	return nil
}
