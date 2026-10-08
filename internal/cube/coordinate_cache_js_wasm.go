//go:build js && wasm

package cube

import "time"

// Browsers obtain verified assets from the static site, never an emulated
// filesystem cache. Avoid linking the native gob encoder and filesystem code.
func coordinateCachePath() string                           { return "" }
func loadCoordinateTablesLimit(time.Time) *coordinateTables { return nil }
func saveCoordinateBytes([]byte) error                      { return nil }
func saveCoordinateTables(*coordinateTables) error          { return nil }
func fallbackCoordinateTables(time.Time) *coordinateTables  { return nil }
func decodeCoordinateTables(data []byte, deadline time.Time) *coordinateTables {
	return decodeBrowserCoordinates(data, deadline)
}
