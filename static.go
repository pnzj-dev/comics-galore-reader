package cgreaderwasm

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/*
var staticRaw embed.FS

var StaticFiles, _ = fs.Sub(staticRaw, "static")

func StaticHandler() http.Handler {
	return http.FileServer(http.FS(StaticFiles))
}
