//go:build js && wasm

package main

import (
	"syscall/js"

	cgreader "github.com/pnzj-dev/comics-galore-reader"
)

func main() {
	cgreader.RegisterJS()

	// Keep the WASM alive.
	select {}
}

// Ensure js is referenced (for build).
var _ js.Value
