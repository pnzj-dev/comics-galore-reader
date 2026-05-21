package main

import (
	"syscall/js"

	cgreaderwasm "github.com/ponzejo/cg-reader-wasm"
)

func main() {
	cgreaderwasm.RegisterJS()

	// Keep the WASM alive.
	select {}
}

// Ensure js is referenced (for build).
var _ js.Value
