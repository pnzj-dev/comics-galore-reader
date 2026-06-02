// Package cgfiber provides Fiber v3 handlers for serving the
// comic-reader's embedded static assets (CSS, JS, WASM glue).
package cgfiber

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	cgreader "github.com/pnzj-dev/comics-galore-reader"
)

// StaticHandler returns a fiber.Handler that serves the embedded
// static files (tailwind.css, wasm_exec.js, cg-reader-wasm.js).
// Mount it at the path configured via cgreader.SetAssetsPath().
//
//	app.Get("/cgrstatic/*", cgfiber.StaticHandler())
func StaticHandler() fiber.Handler {
	return adaptor.HTTPHandler(cgreader.StaticHandler())
}
