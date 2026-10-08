// Command export-web-tables produces the browser's compact coordinate asset.
package main

import (
	"fmt"
	"os"

	"github.com/ehrlich-b/cube/internal/cube"
)

func main() {
	data, err := cube.ExportBrowserCoordinates()
	if err == nil {
		err = os.WriteFile("web/coordinates-web-v1.bin.gz", data, 0644)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
