package cgreader

import (
	"embed"
	"io/fs"
	"net/http"
	"sync"
)

//go:embed static/*
var staticRaw embed.FS

var StaticFiles, _ = fs.Sub(staticRaw, "static")

func StaticHandler() http.Handler {
	return http.FileServer(http.FS(StaticFiles))
}

var (
	assetsPathMu sync.RWMutex
	assetsPath   = "/static/"
)

// AssetsPath returns the current URL prefix for static assets.
func AssetsPath() string {
	assetsPathMu.RLock()
	defer assetsPathMu.RUnlock()
	return assetsPath
}

// SetAssetsPath overrides the default URL prefix for static assets.
// The path should have a trailing slash, e.g. "/assets/".
func SetAssetsPath(path string) {
	assetsPathMu.Lock()
	defer assetsPathMu.Unlock()
	assetsPath = path
}
