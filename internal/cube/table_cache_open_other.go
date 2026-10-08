//go:build !unix

package cube

import "os"

// Windows and WebAssembly do not expose Unix filesystem FIFOs.
func openTableCacheFile(path string) (*os.File, error) {
	return os.Open(path)
}

func setTableCacheBlocking(f *os.File) error { return nil }
