package web

import (
	"embed"
	"io/fs"
)

// DistFS is the embedded UI build (synced from what2cook-ui/dist via `make build`).
//
//go:embed all:dist
var DistFS embed.FS

// UI returns the dist subdirectory as an fs.FS for Gin static / SPA serving.
func UI() (fs.FS, error) {
	return fs.Sub(DistFS, "dist")
}
